package booking

import (
	"fmt"
	"strings"

	"safari-quest-api/config"
	"safari-quest-api/internal/email"

	"gorm.io/gorm"
)

// Booking emails are written to the email outbox in the same transaction as
// the booking change, then sent asynchronously through RabbitMQ (see
// internal/email). If the change rolls back, no email exists to send.

// mailFn returns the emails a save causes; it runs inside the save's
// transaction after the booking row is written.
type mailFn func(tx *gorm.DB, b *Booking) ([]email.Draft, error)

func appLink(path string) string {
	return strings.TrimRight(config.App.AppURL, "/") + path
}

// details is the booking summary shown in every booking email.
func details(b *Booking) []email.Detail {
	kind := fmt.Sprintf("Shared seats · group %d", b.UnitNumber)
	if b.BookingType == TypePrivate {
		kind = fmt.Sprintf("Private group %d", b.UnitNumber)
	}
	travellers := fmt.Sprintf("%d", b.Pax)
	if b.ChargedTravellers > b.Pax && b.Pricing != "per_group" {
		travellers += fmt.Sprintf(" (charged for %d)", b.ChargedTravellers)
	}
	return []email.Detail{
		{Label: "Reference", Value: b.Code},
		{Label: "Trip", Value: b.Package.Name},
		{Label: "Dates", Value: b.Package.StartDate.Format("Mon 2 Jan 2006") + " – " + b.Package.EndDate.Format("Mon 2 Jan 2006")},
		{Label: "Booking", Value: kind},
		{Label: "Travellers", Value: travellers},
		{Label: "Total", Value: fmt.Sprintf("%s %s", b.Package.Currency, money(b.TotalPrice))},
		{Label: "Company", Value: b.TourOperator.Name},
	}
}

func deadline(b *Booking) string {
	if b.ExpiresAt == nil {
		return ""
	}
	return " by " + b.ExpiresAt.Format("Mon 2 Jan 2006, 15:04 MST")
}

func staffDrafts(tx *gorm.DB, b *Booking, kind string, c email.Content) ([]email.Draft, error) {
	var emails []string
	err := tx.Table("users").
		Joins("JOIN tour_operator_users tou ON tou.user_id = users.id").
		Where("tou.tour_operator_id = ? AND users.is_active = ? AND users.deleted_at IS NULL", b.Package.OperatorID, true).
		Pluck("users.email", &emails).Error
	out := make([]email.Draft, len(emails))
	for i, e := range emails {
		out[i] = email.Draft{Kind: kind, To: e, Content: c}
	}
	return out, err
}

func traveler(b *Booking, kind string, c email.Content) email.Draft {
	return email.Draft{Kind: kind, To: b.User.Email, Content: c}
}

// mailsCreated: the traveler hears the request arrived (or was recorded as
// confirmed) and the company's staff hear there's a booking to answer.
func mailsCreated(tx *gorm.DB, b *Booking) ([]email.Draft, error) {
	voucher := appLink("/my-bookings/" + b.UUID.String() + "/voucher")
	if b.Status == StatusConfirmed {
		return []email.Draft{traveler(b, "booking_confirmed", email.Content{
			Subject:    "Booking confirmed: " + b.Package.Name,
			Heading:    "You're booked",
			Paragraphs: []string{fmt.Sprintf("Hello %s, %s has booked you on this trip and confirmed it.", b.User.FirstName, b.TourOperator.Name)},
			Details:    details(b),
			Button:     &email.Button{Label: "View your voucher", URL: voucher},
		})}, nil
	}
	drafts := []email.Draft{traveler(b, "booking_received", email.Content{
		Subject:   "Booking received: " + b.Package.Name,
		Preheader: fmt.Sprintf("%s will confirm your place%s.", b.TourOperator.Name, deadline(b)),
		Heading:   "Thanks, we've got your booking",
		Paragraphs: []string{
			fmt.Sprintf("Hello %s, %s will confirm your place%s. We'll email you as soon as they do.", b.User.FirstName, b.TourOperator.Name, deadline(b)),
		},
		Details: details(b),
		Button:  &email.Button{Label: "Add traveller details", URL: appLink("/my-bookings")},
		After:   []string{"Adding who is travelling helps the company confirm your booking sooner."},
	})}
	staff, err := staffDrafts(tx, b, "booking_new_staff", email.Content{
		Subject:    "New booking to confirm: " + b.Code,
		Preheader:  fmt.Sprintf("%s %s booked %s.", b.User.FirstName, b.User.LastName, b.Package.Name),
		Heading:    "A new booking is waiting for you",
		Paragraphs: []string{fmt.Sprintf("Please confirm or decline it%s, or it expires and the seats are released.", deadline(b))},
		Details:    append(details(b), email.Detail{Label: "Traveler", Value: fmt.Sprintf("%s %s · %s", b.User.FirstName, b.User.LastName, b.User.Email)}),
		Button:     &email.Button{Label: "Confirm or decline", URL: appLink("/bookings")},
	})
	return append(drafts, staff...), err
}

// mailsStatus emails whoever needs to know about a status change.
// byTraveler is true when the traveler made the change (their own cancellation).
func mailsStatus(byTraveler bool) mailFn {
	return func(tx *gorm.DB, b *Booking) ([]email.Draft, error) {
		switch b.Status {
		case StatusConfirmed:
			return []email.Draft{traveler(b, "booking_confirmed", email.Content{
				Subject:    "Booking confirmed: " + b.Package.Name,
				Preheader:  "Your place is confirmed. Your voucher is ready.",
				Heading:    "Good news: you're confirmed",
				Paragraphs: []string{fmt.Sprintf("Hello %s, %s has confirmed your booking.", b.User.FirstName, b.TourOperator.Name)},
				Details:    details(b),
				Button:     &email.Button{Label: "View your voucher", URL: appLink("/my-bookings/" + b.UUID.String() + "/voucher")},
			})}, nil

		case StatusDeclined:
			reason := ""
			if b.DeclineReason != nil {
				reason = *b.DeclineReason
			}
			return []email.Draft{traveler(b, "booking_declined", email.Content{
				Subject:    "Booking not accepted: " + b.Package.Name,
				Heading:    "Sorry, this booking wasn't accepted",
				Paragraphs: []string{fmt.Sprintf("Hello %s, %s couldn't accept your booking %s. Your seats have been released and you have not been charged.", b.User.FirstName, b.TourOperator.Name, b.Code)},
				Quote:      reason,
				Button:     &email.Button{Label: "Find another trip", URL: appLink("/explore")},
			})}, nil

		case StatusExpired:
			drafts := []email.Draft{traveler(b, "booking_expired", email.Content{
				Subject:    "Booking expired: " + b.Package.Name,
				Heading:    "Your booking expired",
				Paragraphs: []string{fmt.Sprintf("Hello %s, %s didn't confirm booking %s in time, so it expired and the seats were released.", b.User.FirstName, b.TourOperator.Name, b.Code)},
				Button:     &email.Button{Label: "Book again", URL: appLink("/explore")},
			})}
			staff, err := staffDrafts(tx, b, "booking_expired_staff", email.Content{
				Subject:    "Booking expired: " + b.Code,
				Heading:    "A booking expired",
				Paragraphs: []string{"Nobody confirmed this booking in time, so it expired and its seats are free again."},
				Details:    details(b),
			})
			return append(drafts, staff...), err

		case StatusCancelled:
			if byTraveler {
				return staffDrafts(tx, b, "booking_cancelled_by_traveler", email.Content{
					Subject:    "Booking cancelled by the traveler: " + b.Code,
					Heading:    "A traveler cancelled",
					Paragraphs: []string{fmt.Sprintf("%s %s cancelled this booking. Its seats are free again.", b.User.FirstName, b.User.LastName)},
					Details:    details(b),
				})
			}
			return []email.Draft{traveler(b, "booking_cancelled", email.Content{
				Subject:    "Booking cancelled: " + b.Package.Name,
				Heading:    "Your booking was cancelled",
				Paragraphs: []string{fmt.Sprintf("Hello %s, %s has cancelled your booking %s.", b.User.FirstName, b.TourOperator.Name, b.Code)},
				After:      []string{"Questions? Contact the company at " + b.TourOperator.Email + "."},
			})}, nil
		}
		return nil, nil
	}
}
