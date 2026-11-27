package setting

import "testing"

func TestDefinitionsAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Definitions {
		if seen[d.Key] {
			t.Errorf("duplicate key %s", d.Key)
		}
		seen[d.Key] = true
		if d.Default < d.Min || d.Default > d.Max {
			t.Errorf("%s: default %d outside %d..%d", d.Key, d.Default, d.Min, d.Max)
		}
		if d.Label == "" || d.Description == "" {
			t.Errorf("%s needs a label and description", d.Key)
		}
	}
	if _, ok := definition(BookingConfirmHours); !ok {
		t.Fatal("booking confirmation setting missing")
	}
}
