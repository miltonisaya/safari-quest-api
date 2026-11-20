package attractioncategory

import (
	"time"

	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

type Input struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type Response struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts an AttractionCategory into its API representation.
func ToResponse(attractionCategory AttractionCategory) Response {
	return Response{
		UUID:      attractionCategory.UUID,
		Name:      attractionCategory.Name,
		Code:      attractionCategory.Code,
		CreatedAt: attractionCategory.CreatedAt,
		UpdatedAt: attractionCategory.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	categories, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(categories))
	for i, c := range categories {
		items[i] = ToResponse(c)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	category, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(category), nil
}

func Create(input Input) (Response, error) {
	category := AttractionCategory{
		Name: input.Name,
		Code: input.Code,
	}
	if err := Insert(&category); err != nil {
		return Response{}, err
	}
	resp := ToResponse(category)
	ws.GlobalHub.Publish(ws.Event{Entity: "attractionCategory", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input Input) (Response, error) {
	category, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	category.Name = input.Name
	category.Code = input.Code
	if err := Save(&category); err != nil {
		return Response{}, err
	}
	resp := ToResponse(category)
	ws.GlobalHub.Publish(ws.Event{Entity: "attractionCategory", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "attractionCategory", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
