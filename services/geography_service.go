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
	Name              string     `json:"name" binding:"required"`
	Code              string     `json:"code" binding:"required"`
	GeographyTypeUuid uuid.UUID  `json:"geographyTypeUuid" binding:"required"`
	ParentUUID        *uuid.UUID `json:"parentUuid"`
}

type GeographyResponse struct {
	UUID          uuid.UUID             `json:"uuid"`
	Name          string                `json:"name"`
	Code          string                `json:"code"`
	GeographyType GeographyTypeResponse `json:"geographyType"`
	ParentUUID    *uuid.UUID            `json:"parentUuid,omitempty"`
	Level         int                   `json:"level"`
	Children      []GeographyResponse   `json:"children,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

func toGeographyResponse(geography models.Geography, level int) GeographyResponse {
	resp := GeographyResponse{
		UUID:          geography.UUID,
		Name:          geography.Name,
		Code:          geography.Code,
		GeographyType: toGeographyTypeResponse(geography.GeographyType),
		Level:         level,
		CreatedAt:     geography.CreatedAt,
		UpdatedAt:     geography.UpdatedAt,
	}
	if geography.Parent != nil {
		resp.ParentUUID = &geography.Parent.UUID
	}
	if len(geography.Children) > 0 {
		resp.Children = make([]GeographyResponse, len(geography.Children))
		for i, child := range geography.Children {
			resp.Children[i] = toGeographyResponse(child, level+1)
		}
	}
	return resp
}

func GeographyGetAll(params pagination.Params, parentUUID *uuid.UUID) (pagination.Result[GeographyResponse], error) {
	geographies, total, err := repositories.GeographyFindAll(params, parentUUID)
	if err != nil {
		return pagination.Result[GeographyResponse]{}, err
	}
	items := make([]GeographyResponse, len(geographies))
	for i, g := range geographies {
		items[i] = toGeographyResponse(g, 0)
	}
	return pagination.Result[GeographyResponse]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GeographyGetByUUID(id uuid.UUID) (GeographyResponse, error) {
	geography, err := repositories.GeographyFindByUUID(id)
	if err != nil {
		return GeographyResponse{}, err
	}
	return toGeographyResponse(geography, 0), nil
}

func GeographyCreate(input GeographyInput) (GeographyResponse, error) {
	geographyType, err := repositories.GeographyTypeFindByUUID(input.GeographyTypeUuid)
	if err != nil {
		return GeographyResponse{}, err
	}
	var parentID *uint
	if input.ParentUUID != nil {
		parent, err := repositories.GeographyFindParentByUUID(*input.ParentUUID)
		if err != nil {
			return GeographyResponse{}, err
		}
		parentID = &parent.ID
	}
	geography := models.Geography{
		Name:            input.Name,
		Code:            input.Code,
		GeographyTypeID: geographyType.ID,
		ParentID:        parentID,
	}
	if err := repositories.GeographyCreate(&geography); err != nil {
		return GeographyResponse{}, err
	}
	geography.GeographyType = geographyType
	resp := toGeographyResponse(geography, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func GeographyUpdate(id uuid.UUID, input GeographyInput) (GeographyResponse, error) {
	geography, err := repositories.GeographyFindByUUID(id)
	if err != nil {
		return GeographyResponse{}, err
	}
	geographyType, err := repositories.GeographyTypeFindByUUID(input.GeographyTypeUuid)
	if err != nil {
		return GeographyResponse{}, err
	}
	var parentID *uint
	if input.ParentUUID != nil {
		parent, err := repositories.GeographyFindParentByUUID(*input.ParentUUID)
		if err != nil {
			return GeographyResponse{}, err
		}
		parentID = &parent.ID
	}
	geography.Name = input.Name
	geography.Code = input.Code
	geography.GeographyTypeID = geographyType.ID
	geography.ParentID = parentID
	if err := repositories.GeographyUpdate(&geography); err != nil {
		return GeographyResponse{}, err
	}
	geography.GeographyType = geographyType
	resp := toGeographyResponse(geography, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func GeographyDelete(id uuid.UUID) error {
	if err := repositories.GeographyDelete(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
