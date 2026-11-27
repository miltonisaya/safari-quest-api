package setting

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Keys of the known settings.
const (
	BookingConfirmHours     = "booking_confirm_hours"
	IDDocumentRetentionDays = "id_document_retention_days"
)

// Definition describes one setting an administrator can change.
type Definition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Group       string `json:"group"`
	Unit        string `json:"unit"`
	Default     int    `json:"default"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
}

// Definitions lists every setting, in the order the settings page shows them.
var Definitions = []Definition{
	{
		Key:         BookingConfirmHours,
		Label:       "Time to confirm a booking",
		Description: "Pending bookings that the tour company has not confirmed within this time expire and free their seats. Set 0 to let bookings wait indefinitely.",
		Group:       "Bookings",
		Unit:        "hours",
		Default:     48,
		Min:         0,
		Max:         720,
	},
	{
		Key:         IDDocumentRetentionDays,
		Label:       "Keep ID copies after the trip",
		Description: "Uploaded copies of group leaders' IDs are deleted automatically this many days after the trip ends. The rest of the booking is kept.",
		Group:       "Privacy",
		Unit:        "days",
		Default:     90,
		Min:         7,
		Max:         3650,
	},
}

var ErrUnknownSetting = errors.New("unknown setting")

func definition(key string) (Definition, bool) {
	for _, d := range Definitions {
		if d.Key == key {
			return d, true
		}
	}
	return Definition{}, false
}

// Int returns a setting's value, or its default when it was never set or the
// stored value is unusable.
func Int(key string) int {
	d, ok := definition(key)
	if !ok {
		return 0
	}
	if raw, found := findValue(key); found {
		if v, err := strconv.Atoi(raw); err == nil && v >= d.Min && v <= d.Max {
			return v
		}
	}
	return d.Default
}

// Response is a setting with its current value.
type Response struct {
	Definition
	Value     int        `json:"value"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// UpdateInput maps keys to new values; keys left out are unchanged.
type UpdateInput struct {
	Values map[string]int `json:"values" binding:"required"`
}

func GetAll() ([]Response, error) {
	rows, err := findAll()
	if err != nil {
		return nil, err
	}
	stored := make(map[string]Setting, len(rows))
	for _, r := range rows {
		stored[r.Key] = r
	}
	out := make([]Response, 0, len(Definitions))
	for _, d := range Definitions {
		resp := Response{Definition: d, Value: d.Default}
		if s, ok := stored[d.Key]; ok {
			if v, err := strconv.Atoi(s.Value); err == nil && v >= d.Min && v <= d.Max {
				resp.Value = v
			}
			at := s.UpdatedAt
			resp.UpdatedAt = &at
		}
		out = append(out, resp)
	}
	return out, nil
}

// Update validates every value against its definition before saving any.
func Update(input UpdateInput, userID uint) ([]Response, error) {
	rows := make([]Setting, 0, len(input.Values))
	now := time.Now()
	for key, v := range input.Values {
		d, ok := definition(key)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownSetting, key)
		}
		if v < d.Min || v > d.Max {
			return nil, fmt.Errorf("%w: %s must be between %d and %d %s", ErrOutOfRange, d.Label, d.Min, d.Max, d.Unit)
		}
		id := userID
		rows = append(rows, Setting{Key: key, Value: strconv.Itoa(v), UpdatedBy: &id, UpdatedAt: now})
	}
	if len(rows) > 0 {
		if err := upsert(rows); err != nil {
			return nil, err
		}
	}
	return GetAll()
}

var ErrOutOfRange = errors.New("value out of range")
