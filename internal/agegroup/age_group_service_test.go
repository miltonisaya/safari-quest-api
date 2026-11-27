package agegroup

import (
	"errors"
	"testing"
)

func n(v int) *int { return &v }

func TestValidateAgeGroups(t *testing.T) {
	ok := []Input{{Name: "Under 5", MinAge: 0, MaxAge: n(4)}, {Name: "5–15", MinAge: 5, MaxAge: n(15)}, {Name: "16+", MinAge: 16}}
	if err := validate(ok); err != nil {
		t.Fatalf("valid groups refused: %v", err)
	}
	bad := map[string][]Input{
		"gap":                  {{Name: "A", MinAge: 0, MaxAge: n(4)}, {Name: "B", MinAge: 6}},
		"overlap":              {{Name: "A", MinAge: 0, MaxAge: n(10)}, {Name: "B", MinAge: 5}},
		"not from 0":           {{Name: "A", MinAge: 1}},
		"oldest has a maximum": {{Name: "A", MinAge: 0, MaxAge: n(99)}},
		"middle open-ended":    {{Name: "A", MinAge: 0}, {Name: "B", MinAge: 1}},
		"blank name":           {{Name: " ", MinAge: 0}},
	}
	for name, groups := range bad {
		if err := validate(groups); !errors.Is(err, ErrInvalidGroups) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}
