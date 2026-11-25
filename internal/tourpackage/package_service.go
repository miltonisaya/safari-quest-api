package tourpackage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"safari-quest-api/internal/attraction"
	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/dbtypes"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrOutsideCompany   = errors.New("You can only manage packages in your company")
	ErrInvalidReference = errors.New("invalid reference")
	ErrInvalidPackage   = errors.New("Check the dates, price, and group size")
	ErrAttachment       = errors.New("Attachments must belong to the same company")
	ErrInvalidItinerary = errors.New("Check the itinerary: every day needs a title and description, meals must be breakfast, lunch or dinner, and there can't be more days than the trip lasts")
	ErrInvalidDetails   = errors.New("Difficulty must be easy, moderate or challenging, and currency a 3-letter code such as USD")
)

// DayInput is one itinerary day; days are numbered in the order sent.
type DayInput struct {
	Title          string     `json:"title" binding:"required"`
	Description    string     `json:"description" binding:"required"`
	Accommodation  *string    `json:"accommodation"`
	Meals          []string   `json:"meals"`
	AttractionUUID *uuid.UUID `json:"attraction_uuid"`
}

type Input struct {
	Name              string      `json:"name" binding:"required"`
	Description       string      `json:"description" binding:"required"`
	OperatorUUID      uuid.UUID   `json:"operator_uuid" binding:"required"`
	StartDate         string      `json:"start_date" binding:"required"`
	EndDate           string      `json:"end_date" binding:"required"`
	Pax               int         `json:"pax" binding:"required,min=1"`
	Price             float64     `json:"price" binding:"min=0"`
	PackageType       string      `json:"package_type" binding:"required"`
	FileResourceUUIDs []uuid.UUID `json:"file_resource_uuids"`
	Highlights        []string    `json:"highlights"`
	Inclusions        []string    `json:"inclusions"`
	Exclusions        []string    `json:"exclusions"`
	Difficulty        *string     `json:"difficulty"`
	MeetingPoint      *string     `json:"meeting_point"`
	Location          *string     `json:"location"`
	Currency          string      `json:"currency"` // ISO 4217; USD when empty
	EcoCertified      bool        `json:"eco_certified"`
	AttractionUUIDs   []uuid.UUID `json:"attraction_uuids"`
	Itinerary         []DayInput  `json:"itinerary" binding:"dive"`
}

type DayResponse struct {
	Day           int      `json:"day"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Accommodation *string  `json:"accommodation"`
	Meals         []string `json:"meals"`
	Attraction    *Ref     `json:"attraction"`
}

type RatingResponse struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type Ref struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type FileRef struct {
	UUID             uuid.UUID `json:"uuid"`
	Name             string    `json:"name"`
	Code             string    `json:"code"`
	Path             string    `json:"path"`
	FileResourceType Ref       `json:"file_resource_type"`
}

type Response struct {
	UUID          uuid.UUID      `json:"uuid"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Operator      Ref            `json:"operator"`
	StartDate     string         `json:"start_date"`
	EndDate       string         `json:"end_date"`
	Pax           int            `json:"pax"`
	Price         float64        `json:"price"`
	PackageType   string         `json:"package_type"`
	Highlights    []string       `json:"highlights"`
	Inclusions    []string       `json:"inclusions"`
	Exclusions    []string       `json:"exclusions"`
	Difficulty    *string        `json:"difficulty"`
	MeetingPoint  *string        `json:"meeting_point"`
	Location      *string        `json:"location"`
	Currency      string         `json:"currency"`
	EcoCertified  bool           `json:"eco_certified"`
	Attractions   []Ref          `json:"attractions"`
	Itinerary     []DayResponse  `json:"itinerary"`
	Rating        RatingResponse `json:"rating"`
	FileResources []FileRef      `json:"file_resources"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type List struct {
	Items     []Response      `json:"items"`
	Meta      pagination.Meta `json:"meta"`
	Operators []Ref           `json:"operators"`
}

// ToResponseRated is ToResponse with the package's review summary filled in.
func ToResponseRated(item Package, rating Rating) Response {
	resp := ToResponse(item)
	resp.Rating = RatingResponse{Average: rating.Average, Count: rating.Count}
	return resp
}

func ToResponse(item Package) Response {
	attractions := make([]Ref, len(item.Attractions))
	for i, a := range item.Attractions {
		attractions[i] = Ref{UUID: a.UUID, Name: a.Name, Code: a.Code}
	}
	days := make([]DayResponse, len(item.Itinerary))
	for i, d := range item.Itinerary {
		var at *Ref
		if d.Attraction != nil {
			at = &Ref{UUID: d.Attraction.UUID, Name: d.Attraction.Name, Code: d.Attraction.Code}
		}
		days[i] = DayResponse{
			Day: d.DayNumber, Title: d.Title, Description: d.Description,
			Accommodation: d.Accommodation, Meals: nonNil(d.Meals), Attraction: at,
		}
	}
	files := make([]FileRef, len(item.FileResources))
	for i, file := range item.FileResources {
		files[i] = FileRef{
			UUID: file.UUID,
			Name: file.Name,
			Code: file.Code,
			Path: file.Path,
			FileResourceType: Ref{
				UUID: file.FileResourceType.UUID,
				Name: file.FileResourceType.Name,
				Code: file.FileResourceType.Code,
			},
		}
	}
	return Response{
		UUID:          item.UUID,
		Name:          item.Name,
		Description:   item.Description,
		Operator:      Ref{UUID: item.Operator.UUID, Name: item.Operator.Name, Code: item.Operator.Code},
		StartDate:     item.StartDate.Format("2006-01-02"),
		EndDate:       item.EndDate.Format("2006-01-02"),
		Pax:           item.Pax,
		Price:         item.Price,
		PackageType:   item.PackageType,
		Highlights:    nonNil(item.Highlights),
		Inclusions:    nonNil(item.Inclusions),
		Exclusions:    nonNil(item.Exclusions),
		Difficulty:    item.Difficulty,
		MeetingPoint:  item.MeetingPoint,
		Location:      item.Location,
		Currency:      item.Currency,
		EcoCertified:  item.EcoCertified,
		Attractions:   attractions,
		Itinerary:     days,
		FileResources: files,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func parseTrip(input Input) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	if end.Before(start) || input.Pax < 1 || input.Price < 0 {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	return start, end, nil
}

func filesFor(operatorID uint, ids []uuid.UUID) ([]fileresource.FileResource, error) {
	if len(ids) == 0 {
		return []fileresource.FileResource{}, nil
	}
	files, err := fileresource.FindByUUIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(files) != len(ids) {
		return nil, fmt.Errorf("%w: one or more file_resource_uuids", ErrInvalidReference)
	}
	for _, file := range files {
		if file.OperatorID != operatorID {
			return nil, ErrAttachment
		}
	}
	return files, nil
}

func nonNil(l dbtypes.StringList) []string {
	if l == nil {
		return []string{}
	}
	return l
}

func optional(s *string) *string {
	if s == nil {
		return nil
	}
	if t := strings.TrimSpace(*s); t != "" {
		return &t
	}
	return nil
}

// itineraryFor validates the days and resolves their attractions. Days are
// numbered 1..n in the order given and may not outnumber the trip's days.
func itineraryFor(input Input, start, end time.Time) ([]ItineraryDay, error) {
	tripDays := int(end.Sub(start).Hours()/24) + 1
	if len(input.Itinerary) > tripDays {
		return nil, ErrInvalidItinerary
	}
	days := make([]ItineraryDay, 0, len(input.Itinerary))
	for i, in := range input.Itinerary {
		title, desc := strings.TrimSpace(in.Title), strings.TrimSpace(in.Description)
		if title == "" || desc == "" {
			return nil, ErrInvalidItinerary
		}
		meals := dbtypes.StringList{}
		seen := map[string]bool{}
		for _, m := range in.Meals {
			m = strings.ToLower(strings.TrimSpace(m))
			if !validMeals[m] {
				return nil, ErrInvalidItinerary
			}
			if !seen[m] {
				seen[m] = true
				meals = append(meals, m)
			}
		}
		day := ItineraryDay{DayNumber: i + 1, Title: title, Description: desc, Accommodation: optional(in.Accommodation), Meals: meals}
		if in.AttractionUUID != nil {
			a, err := attraction.FindByUUID(*in.AttractionUUID)
			if err != nil {
				return nil, fmt.Errorf("%w: itinerary day %d attraction_uuid", ErrInvalidReference, i+1)
			}
			day.AttractionID = &a.ID
		}
		days = append(days, day)
	}
	return days, nil
}

func attractionsFor(ids []uuid.UUID) ([]attraction.Attraction, error) {
	if len(ids) == 0 {
		return []attraction.Attraction{}, nil
	}
	found, err := attraction.FindByUUIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(found) != len(ids) {
		return nil, fmt.Errorf("%w: one or more attraction_uuids", ErrInvalidReference)
	}
	return found, nil
}

func apply(item *Package, input Input, scope companyscope.Scope) (Links, error) {
	start, end, err := parseTrip(input)
	if err != nil {
		return Links{}, err
	}
	difficulty := optional(input.Difficulty)
	if difficulty != nil {
		lower := strings.ToLower(*difficulty)
		if !validDifficulty[lower] {
			return Links{}, ErrInvalidDetails
		}
		difficulty = &lower
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "USD"
	}
	if len(currency) != 3 || strings.Trim(currency, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return Links{}, ErrInvalidDetails
	}
	days, err := itineraryFor(input, start, end)
	if err != nil {
		return Links{}, err
	}
	attractions, err := attractionsFor(input.AttractionUUIDs)
	if err != nil {
		return Links{}, err
	}
	operator, err := touroperator.FindByUUID(input.OperatorUUID)
	if err != nil {
		return Links{}, fmt.Errorf("%w: operator_uuid", ErrInvalidReference)
	}
	if !scope.Allows(operator.ID) {
		return Links{}, ErrOutsideCompany
	}
	files, err := filesFor(operator.ID, input.FileResourceUUIDs)
	if err != nil {
		return Links{}, err
	}
	item.Name = input.Name
	item.Description = input.Description
	item.OperatorID = operator.ID
	item.Operator = operator
	item.StartDate = start
	item.EndDate = end
	item.Pax = input.Pax
	item.Price = input.Price
	item.PackageType = input.PackageType
	item.Highlights = dbtypes.StringList(input.Highlights).Clean()
	item.Inclusions = dbtypes.StringList(input.Inclusions).Clean()
	item.Exclusions = dbtypes.StringList(input.Exclusions).Clean()
	item.Difficulty = difficulty
	item.MeetingPoint = optional(input.MeetingPoint)
	item.Location = optional(input.Location)
	item.Currency = currency
	item.EcoCertified = input.EcoCertified
	item.FileResources = files
	return Links{Files: files, Attractions: attractions, Itinerary: days}, nil
}

func operatorRefs(scope companyscope.Scope) ([]Ref, error) {
	rows, err := touroperator.FindForCompanies(scope.Admin, scope.OperatorIDs)
	if err != nil {
		return nil, err
	}
	refs := make([]Ref, len(rows))
	for i, row := range rows {
		refs[i] = Ref{UUID: row.UUID, Name: row.Name, Code: row.Code}
	}
	return refs, nil
}

func GetAll(params pagination.Params, scope companyscope.Scope) (List, error) {
	rows, total, err := FindAll(params, scope.OperatorIDs, scope.Admin)
	if err != nil {
		return List{}, err
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	ratings, err := RatingsFor(ids)
	if err != nil {
		return List{}, err
	}
	items := make([]Response, len(rows))
	for i, row := range rows {
		items[i] = ToResponseRated(row, ratings[row.ID])
	}
	operators, err := operatorRefs(scope)
	if err != nil {
		return List{}, err
	}
	return List{Items: items, Meta: pagination.NewMeta(total, params), Operators: operators}, nil
}

func GetByUUID(id uuid.UUID, scope companyscope.Scope) (Response, error) {
	item, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(item.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	ratings, err := RatingsFor([]uint{item.ID})
	if err != nil {
		return Response{}, err
	}
	return ToResponseRated(item, ratings[item.ID]), nil
}

func Create(input Input, scope companyscope.Scope) (Response, error) {
	item := Package{}
	links, err := apply(&item, input, scope)
	if err != nil {
		return Response{}, err
	}
	if err := Insert(&item, links); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(item.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input, scope companyscope.Scope) (Response, error) {
	item, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(item.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	links, err := apply(&item, input, scope)
	if err != nil {
		return Response{}, err
	}
	if err := Save(&item, links); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(item.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID, scope companyscope.Scope) error {
	item, err := FindByUUID(id)
	if err != nil {
		return err
	}
	if !scope.Allows(item.OperatorID) {
		return ErrOutsideCompany
	}
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
