package geographytype

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

// ToResponse converts a GeographyType into its API representation.
func ToResponse(geographyType GeographyType) Response {
	return Response{
		UUID:      geographyType.UUID,
		Name:      geographyType.Name,
		Code:      geographyType.Code,
		CreatedAt: geographyType.CreatedAt,
		UpdatedAt: geographyType.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	gt, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(gt))
	for i, gt := range gt {
		items[i] = ToResponse(gt)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	gt, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(gt), nil
}

func Create(input Input) (Response, error) {
	gt := GeographyType{
		Name: input.Name,
		Code: input.Code,
	}
	if err := Insert(&gt); err != nil {
		return Response{}, err
	}
	resp := ToResponse(gt)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input Input) (Response, error) {
	gt, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	gt.Name = input.Name
	gt.Code = input.Code
	if err := Save(&gt); err != nil {
		return Response{}, err
	}
	resp := ToResponse(gt)
	ws.GlobalHub.Publish(ws.Event{Entity: "geographyType", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "geographyType", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
