package booking

import (
	"errors"
	"strings"
	"time"

	"safari-quest-api/internal/user"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var ErrEmptyMessage = errors.New("Write a message of 1 to 2000 characters")

// MessageResponse is one chat message as both sides see it.
type MessageResponse struct {
	UUID        uuid.UUID  `json:"uuid"`
	BookingUUID uuid.UUID  `json:"booking_uuid"`
	FromCompany bool       `json:"from_company"`
	SenderName  string     `json:"sender_name"`
	Body        string     `json:"body"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ThreadResponse is a booking's conversation, as listed for the bell and the
// Messages page. Unread counts messages from the other side not yet opened.
type ThreadResponse struct {
	BookingUUID  uuid.UUID        `json:"booking_uuid"`
	Code         string           `json:"code"`
	TripName     string           `json:"trip_name"`
	StartDate    string           `json:"start_date"`
	Status       Status           `json:"status"`
	Company      string           `json:"company"`
	Traveler     string           `json:"traveler"`
	IsCompany    bool             `json:"is_company"` // the viewer answers for the company
	Unread       int              `json:"unread"`
	LastMessage  *MessageResponse `json:"last_message"`
	MessageCount int              `json:"message_count"`
}

type ConversationResponse struct {
	ThreadResponse
	Messages []MessageResponse `json:"messages"`
}

type MessageInput struct {
	Body string `json:"body" binding:"required"`
}

// companySide reports whether the viewer speaks for the company on b.
func companySide(b Booking, v Viewer) bool { return b.User.UUID != v.UserUUID }

func threadOf(b Booking, v Viewer) ThreadResponse {
	return ThreadResponse{
		BookingUUID: b.UUID, Code: b.Code, TripName: b.Package.Name, StartDate: b.Package.StartDate.Format("2006-01-02"),
		Status: b.Status, Company: b.TourOperator.Name, Traveler: strings.TrimSpace(b.User.FirstName + " " + b.User.LastName),
		IsCompany: companySide(b, v),
	}
}

func messageResponse(m Message, bookingUUID uuid.UUID, names map[uint]string) MessageResponse {
	name := "SafariQuest"
	if m.SenderID != nil {
		if n, ok := names[*m.SenderID]; ok {
			name = n
		}
	}
	return MessageResponse{
		UUID: m.UUID, BookingUUID: bookingUUID, FromCompany: m.FromCompany, SenderName: name,
		Body: m.Body, ReadAt: m.ReadAt, CreatedAt: m.CreatedAt,
	}
}

// Conversation returns a booking's messages and marks the other side's as read.
func Conversation(bookingUUID uuid.UUID, v Viewer) (ConversationResponse, error) {
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return ConversationResponse{}, err
	}
	if !canSee(b, v) {
		return ConversationResponse{}, ErrNotYours
	}
	fromCompany := companySide(b, v)
	// Opening the thread reads what the other side sent.
	read, err := markRead(b.ID, !fromCompany, time.Now())
	if err != nil {
		return ConversationResponse{}, err
	}
	msgs, names, err := messagesOf(b.ID)
	if err != nil {
		return ConversationResponse{}, err
	}
	out := ConversationResponse{ThreadResponse: threadOf(b, v), Messages: make([]MessageResponse, len(msgs))}
	for i, m := range msgs {
		out.Messages[i] = messageResponse(m, b.UUID, names)
	}
	out.MessageCount = len(msgs)
	if len(msgs) > 0 {
		last := out.Messages[len(msgs)-1]
		out.LastMessage = &last
	}
	if read > 0 {
		announce(b, ws.Event{Entity: "message", Action: ws.ActionUpdated, Payload: map[string]any{"booking_uuid": b.UUID, "read": true}})
	}
	return out, nil
}

// Send posts a message on a booking's conversation and pushes it to both sides.
func Send(bookingUUID uuid.UUID, v Viewer, in MessageInput) (MessageResponse, error) {
	body := strings.TrimSpace(in.Body)
	if body == "" || len([]rune(body)) > 2000 {
		return MessageResponse{}, ErrEmptyMessage
	}
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return MessageResponse{}, err
	}
	if !canSee(b, v) {
		return MessageResponse{}, ErrNotYours
	}
	sender, err := user.FindByUUID(v.UserUUID)
	if err != nil {
		return MessageResponse{}, err
	}
	m := Message{UUID: uuid.New(), BookingID: b.ID, SenderID: &sender.ID, FromCompany: companySide(b, v), Body: body}
	if err := insertMessage(&m); err != nil {
		return MessageResponse{}, err
	}
	resp := messageResponse(m, b.UUID, map[uint]string{sender.ID: strings.TrimSpace(sender.FirstName + " " + sender.LastName)})
	announce(b, ws.Event{Entity: "message", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

// Threads lists the viewer's conversations that have messages, newest first.
func Threads(v Viewer) ([]ThreadResponse, error) {
	rows, err := threadsFor(v)
	if err != nil {
		return nil, err
	}
	out := make([]ThreadResponse, 0, len(rows))
	for _, r := range rows {
		t := threadOf(r.Booking, v)
		t.Unread, t.MessageCount = r.Unread, r.Count
		last := messageResponse(r.Last, r.Booking.UUID, map[uint]string{})
		last.SenderName = r.LastSender
		t.LastMessage = &last
		out = append(out, t)
	}
	return out, nil
}

// announce sends an event to the booking's traveler, its company's staff and
// administrators only.
func announce(b Booking, event ws.Event) {
	users, err := participants(b.UUID)
	if err != nil {
		return
	}
	ws.GlobalHub.PublishTo(users, event)
}
