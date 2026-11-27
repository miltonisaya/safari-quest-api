package agegroup

import (
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidGroups = errors.New("Age groups must start at 0, run in order without gaps or overlaps, and only the oldest may have no upper age")
	ErrGroupInUse    = errors.New("An age group used by upcoming bookings can't be removed; rename or re-range it instead")
)

type Response struct {
	UUID   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	MinAge int       `json:"min_age"`
	MaxAge *int      `json:"max_age"`
}

func ToResponse(g AgeGroup) Response {
	return Response{UUID: g.UUID, Name: g.Name, MinAge: g.MinAge, MaxAge: g.MaxAge}
}

// Input is one row of the age groups editor; UUID is empty for a new group.
type Input struct {
	UUID   *uuid.UUID `json:"uuid"`
	Name   string     `json:"name" binding:"required,max=60"`
	MinAge int        `json:"min_age" binding:"min=0,max=120"`
	MaxAge *int       `json:"max_age"`
}

type ReplaceInput struct {
	Groups []Input `json:"groups" binding:"required,min=1,max=20,dive"`
}

func GetAll() ([]Response, error) {
	rows, err := FindAll()
	if err != nil {
		return nil, err
	}
	out := make([]Response, len(rows))
	for i, g := range rows {
		out[i] = ToResponse(g)
	}
	return out, nil
}

// validate checks the bands cover every age from 0 exactly once: each starts
// right after the previous one ends, and only the oldest is open-ended.
func validate(groups []Input) error {
	sort.Slice(groups, func(i, j int) bool { return groups[i].MinAge < groups[j].MinAge })
	next := 0
	for i, g := range groups {
		if strings.TrimSpace(g.Name) == "" || g.MinAge != next {
			return ErrInvalidGroups
		}
		if i == len(groups)-1 {
			if g.MaxAge != nil {
				return ErrInvalidGroups
			}
			return nil
		}
		if g.MaxAge == nil || *g.MaxAge < g.MinAge {
			return ErrInvalidGroups
		}
		next = *g.MaxAge + 1
	}
	return nil
}

// ReplaceAll saves the administrator's whole list at once.
func ReplaceAll(input ReplaceInput) ([]Response, error) {
	if err := validate(input.Groups); err != nil {
		return nil, err
	}
	current, err := FindAll()
	if err != nil {
		return nil, err
	}
	byUUID := map[uuid.UUID]AgeGroup{}
	for _, g := range current {
		byUUID[g.UUID] = g
	}
	keep := make([]AgeGroup, 0, len(input.Groups))
	kept := map[uuid.UUID]bool{}
	for _, in := range input.Groups {
		g := AgeGroup{UUID: uuid.New()}
		if in.UUID != nil {
			existing, ok := byUUID[*in.UUID]
			if !ok {
				return nil, ErrInvalidGroups
			}
			g = existing
			kept[g.UUID] = true
		}
		g.Name, g.MinAge, g.MaxAge = strings.TrimSpace(in.Name), in.MinAge, in.MaxAge
		keep = append(keep, g)
	}
	var remove []uint
	for _, g := range current {
		if !kept[g.UUID] {
			remove = append(remove, g.ID)
		}
	}
	used, err := inUse(remove)
	if err != nil {
		return nil, err
	}
	if len(used) > 0 {
		return nil, ErrGroupInUse
	}
	if err := replaceAll(keep, remove); err != nil {
		return nil, err
	}
	return GetAll()
}
