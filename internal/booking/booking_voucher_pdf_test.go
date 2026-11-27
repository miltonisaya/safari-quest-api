package booking

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"safari-quest-api/internal/tourpackage"
)

func TestVoucherPDF(t *testing.T) {
	now := time.Now()
	where, meet := "Serengeti, Tanzania", "Kilimanjaro airport – arrivals"
	d := Detail{
		Response: Response{Code: "SQ-TEST01", Status: StatusConfirmed, Pax: 2, BookingType: TypeShared, Group: 1,
			Pricing: "per_person", ChargedTravellers: 2, TotalPrice: 4900, ConfirmedAt: &now, Traveler: TravelerRef{Name: "Asha Mollel"}},
		Trip: tourpackage.Response{Name: "Serengeti Big Cats Photo Safari", PackageType: "Safari", Location: &where, MeetingPoint: &meet,
			StartDate: "2027-01-18", EndDate: "2027-01-22", Currency: "USD", Inclusions: []string{"Park fees"},
			Itinerary: []tourpackage.DayResponse{{Day: 1, Title: "Arrival", Meals: []string{"dinner"}}}},
		Company:            CompanyContact{Name: "Savanna Explorers", Email: "hello@example.com", Phone: "+255 700 000 000"},
		PreferredLanguages: []string{"en", "sw"},
		Travellers:         []TravellerResponse{{IsLead: true, FullName: "Asha Mollel", AgeGroup: "18–64", Nationality: "TZ"}, {FullName: "Juma Ali", AgeGroup: "5–15", Nationality: "KE"}},
	}
	data, err := VoucherPDF(d)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) || len(data) < 1500 {
		t.Fatalf("not a PDF (%d bytes)", len(data))
	}
	d.Status = StatusPending
	if _, err := VoucherPDF(d); !errors.Is(err, ErrNoVoucher) {
		t.Fatalf("pending booking should have no voucher: %v", err)
	}
	if got := money(1234567.5); got != "1,234,567.50" {
		t.Fatalf("money = %s", got)
	}
}
