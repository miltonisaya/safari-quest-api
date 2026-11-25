package review

import (
	"testing"

	"safari-quest-api/internal/user"
)

func TestValidRating(t *testing.T) {
	for stars, want := range map[int]bool{0: false, 1: true, 3: true, 5: true, 6: false, -1: false} {
		if got := ValidRating(stars); got != want {
			t.Errorf("ValidRating(%d) = %v, want %v", stars, got, want)
		}
	}
}

func TestReviewerNameHidesSurname(t *testing.T) {
	cases := map[string]user.User{
		"Amani T.": {FirstName: "Amani", LastName: "Traveler"},
		"Amani":    {FirstName: "Amani"},
		"Traveler": {},
		"Neema M.": {FirstName: " Neema ", LastName: "mollel"},
	}
	for want, u := range cases {
		if got := reviewerName(u); got != want {
			t.Errorf("reviewerName(%q %q) = %q, want %q", u.FirstName, u.LastName, got, want)
		}
	}
}
