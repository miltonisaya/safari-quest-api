package services

import (
	"time"

	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/repositories"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

type GeographyInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type GeographyResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toGeographyResponse(role models.Geography) GeographyResponse {
	return GeographyResponse{
		UUID:      role.UUID,
		Name:      role.Name,
		Code:      role.Code,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func GeographyGetAll(params pagination.Params) (pagination.Result[GeographyResponse], error) {
	roles, total, err := repositories.GeographyFindAll(params)
	if err != nil {
		return pagination.Result[GeographyResponse]{}, err
	}
	items := make([]GeographyResponse, len(roles))
	for i, role := range roles {
		items[i] = toGeographyResponse(role)
	}
	return pagination.Result[GeographyResponse]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GeographyGetByUUID(uuid uuid.UUID) (GeographyResponse, error) {
	geography, err := repositories.GeographyFindByUUID(uuid)
	if err != nil {
		return GeographyResponse{}, err
	}
	return toGeographyResponse(geography), nil
}

func GeographyCreate(input GeographyInput) (GeographyResponse, error) {
	geography := models.Geography{
		Name: input.Name,
		Code: input.Code,
	}
	if err := repositories.GeographyCreate(&geography); err != nil {
		return GeographyResponse{}, err
	}
	resp := toGeographyResponse(geography)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func GeographyUpdate(uuid uuid.UUID, input GeographyInput) (GeographyResponse, error) {
	geography, err := repositories.GeographyFindByUUID(uuid)
	if err != nil {
		return GeographyResponse{}, err
	}
	geography.Name = input.Name
	geography.Code = input.Code
	if err := repositories.GeographyUpdate(&geography); err != nil {
		return GeographyResponse{}, err
	}
	resp := toGeographyResponse(geography)
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func GeographyDelete(uuid uuid.UUID) error {
	if err := repositories.GeographyDelete(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
