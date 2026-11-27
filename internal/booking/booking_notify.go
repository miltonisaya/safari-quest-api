package booking

import (
	"fmt"
	"log"
	"strings"

	"safari-quest-api/config"
	"safari-quest-api/pkg/mailer"
)

// Emails about bookings go out in the background, so a slow or missing mail
// server never delays or fails the booking itself.

func appLink(path string) string {
	return strings.TrimRight(config.App.AppURL, "/") + path
}

func send(to []string, subject, body string) {
	for _, addr := range to {
		if addr == "" {
			continue
		}
		go func(addr string) {
			if err := mailer.Send(addr, subject, body); err != nil {
				log.Printf("[booking] could not email %s: %v\n", addr, err)
			}
		}(addr)
	}
}

func toStaff(b Booking, subject, body string) {
	emails, err := staffEmails(b.Package.OperatorID)
	if err != nil {
		log.Printf("[booking] could not find staff of company %d: %v\n", b.Package.OperatorID, err)
		return
	}
	send(emails, subject, body)
}

func toTraveler(b Booking, subject, body string) {
	send([]string{b.User.Email}, subject, body)
}

// summary is the booking block shared by every email.
func summary(b Booking) string {
	kind := fmt.Sprintf("Shared seats, group %d", b.UnitNumber)
	if b.BookingType == TypePrivate {
		kind = fmt.Sprintf("Private group %d", b.UnitNumber)
	}
	travellers := fmt.Sprintf("%d", b.Pax)
	if b.ChargedTravellers > b.Pax && b.Pricing != "per_group" {
		travellers += fmt.Sprintf(" (charged for %d)", b.ChargedTravellers)
	}
	return fmt.Sprintf(`Reference:   %s
Trip:        %s
Dates:       %s to %s
Booking:     %s
Travellers:  %s
Total:       %s %.2f
Company:     %s`,
		b.Code, b.Package.Name,
		b.Package.StartDate.Format("Mon 2 Jan 2006"), b.Package.EndDate.Format("Mon 2 Jan 2006"),
		kind, travellers, b.Package.Currency, b.TotalPrice, b.TourOperator.Name)
}

func deadline(b Booking) string {
	if b.ExpiresAt == nil {
		return ""
	}
	return " by " + b.ExpiresAt.Format("Mon 2 Jan 2006, 15:04 MST")
}

// notifyCreated tells the traveler their request arrived (or was recorded as
// confirmed) and tells the company's staff there is a booking to answer.
func notifyCreated(b Booking) {
	if b.Status == StatusConfirmed {
		toTraveler(b, "Booking confirmed: "+b.Package.Name, fmt.Sprintf(`Hello %s,

%s has booked you on this trip and confirmed it.

%s

Your confirmation voucher: %s

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, summary(b), appLink("/my-bookings/"+b.UUID.String()+"/voucher")))
		return
	}
	toTraveler(b, "Booking received: "+b.Package.Name, fmt.Sprintf(`Hello %s,

Thank you for booking. %s will confirm your place%s. We'll email you as soon as they do.

%s

See your bookings: %s

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, deadline(b), summary(b), appLink("/my-bookings")))

	toStaff(b, "New booking to confirm: "+b.Code, fmt.Sprintf(`A new booking is waiting for your answer%s.

%s
Traveler:    %s %s <%s>

Confirm or decline it: %s
`, deadline(b), summary(b), b.User.FirstName, b.User.LastName, b.User.Email, appLink("/bookings")))
}

// notifyStatus emails whoever needs to know about a status change.
// byTraveler is true when the traveler made the change (their own cancellation).
func notifyStatus(b Booking, byTraveler bool) {
	switch b.Status {
	case StatusConfirmed:
		toTraveler(b, "Booking confirmed: "+b.Package.Name, fmt.Sprintf(`Hello %s,

Good news: %s has confirmed your booking.

%s

Your confirmation voucher: %s

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, summary(b), appLink("/my-bookings/"+b.UUID.String()+"/voucher")))

	case StatusDeclined:
		reason := ""
		if b.DeclineReason != nil {
			reason = "\nTheir reason: " + *b.DeclineReason + "\n"
		}
		toTraveler(b, "Booking not accepted: "+b.Package.Name, fmt.Sprintf(`Hello %s,

Sorry, %s could not accept your booking %s.
%s
Your seats have been released and you have not been charged. Other trips with open seats: %s

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, b.Code, reason, appLink("/explore")))

	case StatusExpired:
		toTraveler(b, "Booking expired: "+b.Package.Name, fmt.Sprintf(`Hello %s,

%s did not confirm your booking %s in time, so it has expired and the seats were released. You can book again if seats are still open: %s

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, b.Code, appLink("/explore")))
		toStaff(b, "Booking expired: "+b.Code, fmt.Sprintf(`Booking %s expired because nobody confirmed it in time. Its seats are free again.

%s
`, b.Code, summary(b)))

	case StatusCancelled:
		if byTraveler {
			toStaff(b, "Booking cancelled by the traveler: "+b.Code, fmt.Sprintf(`%s %s cancelled this booking. Its seats are free again.

%s
`, b.User.FirstName, b.User.LastName, summary(b)))
			return
		}
		toTraveler(b, "Booking cancelled: "+b.Package.Name, fmt.Sprintf(`Hello %s,

%s has cancelled your booking %s. Contact them at %s if you have questions.

The SafariQuest team
`, b.User.FirstName, b.TourOperator.Name, b.Code, b.TourOperator.Email))
	}
}
