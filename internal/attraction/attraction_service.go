package attraction

import (
	"errors"
	"fmt"
	"time"

	"safari-quest-api/internal/attractioncategory"
	attractiontype "safari-quest-api/internal/attractiontype"
	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

// ErrInvalidReference is returned when an input UUID points at a record that
// does not exist (type, category, geography or tour operator).
var ErrInvalidReference = errors.New("invalid reference")

type Input struct {
	Name                   string      `json:"name" binding:"required"`
	Code                   string      `json:"code" binding:"required"`
	Description            *string     `json:"description"`
	AttractionTypeUUID     uuid.UUID   `json:"attraction_type_uuid" binding:"required"`
	AttractionCategoryUUID uuid.UUID   `json:"attraction_category_uuid" binding:"required"`
	GeographyUUID          uuid.UUID   `json:"geography_uuid" binding:"required"`
	Latitude               *float64    `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude              *float64    `json:"longitude" binding:"omitempty,min=-180,max=180"`
	IsActive               *bool       `json:"is_active"`
	TourOperatorUUIDs      []uuid.UUID `json:"tour_operator_uuids"`
}

// Ref is the compact form of a related record.
type Ref struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type Response struct {
	UUID               uuid.UUID `json:"uuid"`
	Name               string    `json:"name"`
	Code               string    `json:"code"`
	Description        *string   `json:"description"`
	AttractionType     Ref       `json:"attraction_type"`
	AttractionCategory Ref       `json:"attraction_category"`
	Geography          Ref       `json:"geography"`
	Latitude           *float64  `json:"latitude"`
	Longitude          *float64  `json:"longitude"`
	IsActive           bool      `json:"is_active"`
	TourOperators      []Ref     `json:"tour_operators"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToResponse converts an Attraction into its API representation.
func ToResponse(a Attraction) Response {
	operators := make([]Ref, len(a.TourOperators))
	for i, o := range a.TourOperators {
		operators[i] = Ref{UUID: o.UUID, Name: o.Name, Code: o.Code}
	}
	return Response{
		UUID:               a.UUID,
		Name:               a.Name,
		Code:               a.Code,
		Description:        a.Description,
		AttractionType:     Ref{UUID: a.AttractionType.UUID, Name: a.AttractionType.Name, Code: a.AttractionType.Code},
		AttractionCategory: Ref{UUID: a.AttractionCategory.UUID, Name: a.AttractionCategory.Name, Code: a.AttractionCategory.Code},
		Geography:          Ref{UUID: a.Geography.UUID, Name: a.Geography.Name, Code: a.Geography.Code},
		Latitude:           a.Latitude,
		Longitude:          a.Longitude,
		IsActive:           a.IsActive,
		TourOperators:      operators,
		CreatedAt:          a.CreatedAt,
		UpdatedAt:          a.UpdatedAt,
	}
}

func GetAll(params pagination.Params, f Filters) (pagination.Result[Response], error) {
	attractions, total, err := FindAll(params, f)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(attractions))
	for i, a := range attractions {
		items[i] = ToResponse(a)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	a, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(a), nil
}

// apply copies the input onto the attraction and resolves every reference,
// returning the tour operators to link.
func apply(a *Attraction, input Input) ([]touroperator.TourOperator, error) {
	t, err := attractiontype.FindByUUID(input.AttractionTypeUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: attraction_type_uuid", ErrInvalidReference)
	}
	c, err := attractioncategory.FindByUUID(input.AttractionCategoryUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: attraction_category_uuid", ErrInvalidReference)
	}
	g, err := geography.FindByUUID(input.GeographyUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: geography_uuid", ErrInvalidReference)
	}
	operators := []touroperator.TourOperator{}
	if len(input.TourOperatorUUIDs) > 0 {
		operators, err = touroperator.FindByUUIDs(input.TourOperatorUUIDs)
		if err != nil {
			return nil, err
		}
		if len(operators) != len(input.TourOperatorUUIDs) {
			return nil, fmt.Errorf("%w: one or more tour_operator_uuids", ErrInvalidReference)
		}
	}

	a.Name = input.Name
	a.Code = input.Code
	a.Description = input.Description
	a.AttractionTypeID, a.AttractionType = t.ID, t
	a.AttractionCategoryID, a.AttractionCategory = c.ID, c
	a.GeographyID, a.Geography = g.ID, g
	a.Latitude = input.Latitude
	a.Longitude = input.Longitude
	if input.IsActive != nil {
		a.IsActive = *input.IsActive
	}
	a.TourOperators = operators
	return operators, nil
}

func Create(input Input) (Response, error) {
	a := Attraction{IsActive: true}
	operators, err := apply(&a, input)
	if err != nil {
		return Response{}, err
	}
	if err := Insert(&a, operators); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(a.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "attraction", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input Input) (Response, error) {
	a, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	operators, err := apply(&a, input)
	if err != nil {
		return Response{}, err
	}
	if err := Save(&a, operators); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(a.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "attraction", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "attraction", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
