package booking

import "testing"

func TestStatusTransitions(t *testing.T) {
	allowed := map[Status][]Status{
		StatusPending:   {StatusPending, StatusConfirmed, StatusCancelled},
		StatusConfirmed: {StatusConfirmed, StatusCompleted, StatusCancelled},
		StatusCompleted: {StatusCompleted},
		StatusCancelled: {StatusCancelled},
	}
	all := []Status{StatusPending, StatusConfirmed, StatusCompleted, StatusCancelled}
	for from, ok := range allowed {
		okSet := map[Status]bool{}
		for _, s := range ok {
			okSet[s] = true
		}
		for _, to := range all {
			if got := from.CanMoveTo(to); got != okSet[to] {
				t.Errorf("%s -> %s: got %v, want %v", from, to, got, okSet[to])
			}
		}
	}
}

func TestHoldsSeats(t *testing.T) {
	for status, want := range map[Status]bool{
		StatusPending: true, StatusConfirmed: true, StatusCompleted: false, StatusCancelled: false,
	} {
		if got := status.HoldsSeats(); got != want {
			t.Errorf("%s.HoldsSeats() = %v, want %v", status, got, want)
		}
	}
}

func TestValidStatus(t *testing.T) {
	if ValidStatus("shipped") {
		t.Error("unknown status accepted")
	}
	if !ValidStatus(StatusConfirmed) {
		t.Error("confirmed rejected")
	}
}

func TestNewCodeFormat(t *testing.T) {
	for _, c := range codeAlphabet {
		if c == '0' || c == 'O' || c == '1' || c == 'I' {
			t.Fatalf("ambiguous character %q in booking code alphabet", c)
		}
	}
}
