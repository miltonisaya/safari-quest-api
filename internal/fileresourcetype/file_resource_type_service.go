package fileresourcetype

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

func ToResponse(item FileResourceType) Response {
	return Response{
		UUID:      item.UUID,
		Name:      item.Name,
		Code:      item.Code,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	rows, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(rows))
	for i, row := range rows {
		items[i] = ToResponse(row)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

func GetByUUID(id uuid.UUID) (Response, error) {
	row, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(row), nil
}

func Create(input Input) (Response, error) {
	row := FileResourceType{Name: input.Name, Code: input.Code}
	if err := Insert(&row); err != nil {
		return Response{}, err
	}
	resp := ToResponse(row)
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResourceType", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input) (Response, error) {
	row, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	row.Name = input.Name
	row.Code = input.Code
	if err := Save(&row); err != nil {
		return Response{}, err
	}
	resp := ToResponse(row)
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResourceType", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID) error {
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResourceType", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
