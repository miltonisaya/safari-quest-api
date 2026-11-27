package booking

import "testing"

func TestDeclineAndExpiryRules(t *testing.T) {
	for _, next := range []Status{StatusDeclined, StatusExpired} {
		if !StatusPending.CanMoveTo(next) {
			t.Errorf("pending should move to %s", next)
		}
		if StatusConfirmed.CanMoveTo(next) {
			t.Errorf("a confirmed booking cannot become %s", next)
		}
		if next.HoldsSeats() {
			t.Errorf("%s bookings must free their seats", next)
		}
		if next.CanMoveTo(StatusConfirmed) || next.CanMoveTo(StatusPending) {
			t.Errorf("%s is final", next)
		}
		if !ValidStatus(next) {
			t.Errorf("%s should be a valid status", next)
		}
	}
}
