package services

import (
	"time"

	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/repositories"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

type GeographyTypeInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type GeographyTypeResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toGeographyTypeResponse(role models.GeographyType) GeographyTypeResponse {
	return GeographyTypeResponse{
		UUID:      role.UUID,
		Name:      role.Name,
		Code:      role.Code,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func GeographyTypeGetAll(params pagination.Params) (pagination.Result[GeographyTypeResponse], error) {
	gt, total, err := repositories.GeographyTypeFindAll(params)
	if err != nil {
		return pagination.Result[GeographyTypeResponse]{}, err
	}
	items := make([]GeographyTypeResponse, len(gt))
	for i, gt := range gt {
		items[i] = toGeographyTypeResponse(gt)
	}
	return pagination.Result[GeographyTypeResponse]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GeographyTypeGetByUUID(uuid uuid.UUID) (GeographyTypeResponse, error) {
	gt, err := repositories.GeographyTypeFindByUUID(uuid)
	if err != nil {
		return GeographyTypeResponse{}, err
	}
	return toGeographyTypeResponse(gt), nil
}

func GeographyTypeCreate(input GeographyTypeInput) (GeographyTypeResponse, error) {
	gt := models.GeographyType{
		Name: input.Name,
		Code: input.Code,
	}
	if err := repositories.GeographyTypeCreate(&gt); err != nil {
		return GeographyTypeResponse{}, err
	}
	resp := toGeographyTypeResponse(gt)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func GeographyTypeUpdate(uuid uuid.UUID, input GeographyTypeInput) (GeographyTypeResponse, error) {
	gt, err := repositories.GeographyTypeFindByUUID(uuid)
	if err != nil {
		return GeographyTypeResponse{}, err
	}
	gt.Name = input.Name
	gt.Code = input.Code
	if err := repositories.GeographyTypeUpdate(&gt); err != nil {
		return GeographyTypeResponse{}, err
	}
	resp := toGeographyTypeResponse(gt)
	ws.GlobalHub.Publish(ws.Event{Entity: "geographyType", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func GeographyTypeDelete(uuid uuid.UUID) error {
	if err := repositories.GeographyDelete(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "geographyType", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
