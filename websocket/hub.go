package websocket

import "github.com/google/uuid"

// Hub maintains the set of active clients and fans events out to all of them
// (Publish) or only to chosen users (PublishTo). All mutations to the clients
// map happen inside the Run goroutine so no mutex is needed.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	direct     chan directMessage
	register   chan *Client
	unregister chan *Client
}

// directMessage is delivered only to the connections of the listed users.
type directMessage struct {
	users map[uuid.UUID]bool
	data  []byte
}

var GlobalHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		direct:     make(chan directMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run must be started in its own goroutine. It processes register, unregister,
// and broadcast events sequentially so client map access is always single-threaded.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				h.deliver(client, message)
			}

		case message := <-h.direct:
			for client := range h.clients {
				if message.users[client.userID] {
					h.deliver(client, message.data)
				}
			}
		}
	}
}

func (h *Hub) deliver(client *Client, message []byte) {
	select {
	case client.send <- message:
	default:
		// Client's send buffer is full — assume it is dead.
		delete(h.clients, client)
		close(client.send)
	}
}

// PublishTo sends an Event only to the connections of the given users, for
// private data such as bookings and chat messages.
func (h *Hub) PublishTo(users []uuid.UUID, event Event) {
	if len(users) == 0 {
		return
	}
	data, err := event.Marshal()
	if err != nil {
		return
	}
	set := make(map[uuid.UUID]bool, len(users))
	for _, u := range users {
		set[u] = true
	}
	h.direct <- directMessage{users: set, data: data}
}

// Publish serialises an Event and queues it for broadcast to all clients.
func (h *Hub) Publish(event Event) {
	data, err := event.Marshal()
	if err != nil {
		return
	}
	h.broadcast <- data
}
