package geography

import (
	"safari-quest-api/internal/geographytype"
	"time"

	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

type Input struct {
	Name              string     `json:"name" binding:"required"`
	Code              string     `json:"code" binding:"required"`
	GeographyTypeUuid uuid.UUID  `json:"geographyTypeUuid" binding:"required"`
	ParentUUID        *uuid.UUID `json:"parentUuid"`
}

type Response struct {
	UUID          uuid.UUID              `json:"uuid"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"`
	GeographyType geographytype.Response `json:"geographyType"`
	ParentUUID    *uuid.UUID             `json:"parentUuid,omitempty"`
	Level         int                    `json:"level"`
	Children      []Response             `json:"children,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func toGeographyResponse(geography Geography, level int) Response {
	resp := Response{
		UUID:          geography.UUID,
		Name:          geography.Name,
		Code:          geography.Code,
		GeographyType: geographytype.ToResponse(geography.GeographyType),
		Level:         level,
		CreatedAt:     geography.CreatedAt,
		UpdatedAt:     geography.UpdatedAt,
	}
	if geography.Parent != nil {
		resp.ParentUUID = &geography.Parent.UUID
	}
	if len(geography.Children) > 0 {
		resp.Children = make([]Response, len(geography.Children))
		for i, child := range geography.Children {
			resp.Children[i] = toGeographyResponse(child, level+1)
		}
	}
	return resp
}

func GetAll(params pagination.Params, parentUUID *uuid.UUID) (pagination.Result[Response], error) {
	geographies, total, err := FindAll(params, parentUUID)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(geographies))
	for i, g := range geographies {
		items[i] = toGeographyResponse(g, 0)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(id uuid.UUID) (Response, error) {
	geography, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	return toGeographyResponse(geography, 0), nil
}

func Create(input Input) (Response, error) {
	geographyType, err := geographytype.FindByUUID(input.GeographyTypeUuid)
	if err != nil {
		return Response{}, err
	}
	var parentID *uint
	if input.ParentUUID != nil {
		parent, err := FindParentByUUID(*input.ParentUUID)
		if err != nil {
			return Response{}, err
		}
		parentID = &parent.ID
	}
	geography := Geography{
		Name:            input.Name,
		Code:            input.Code,
		GeographyTypeID: geographyType.ID,
		ParentID:        parentID,
	}
	if err := Insert(&geography); err != nil {
		return Response{}, err
	}
	geography.GeographyType = geographyType
	resp := toGeographyResponse(geography, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input) (Response, error) {
	geography, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	geographyType, err := geographytype.FindByUUID(input.GeographyTypeUuid)
	if err != nil {
		return Response{}, err
	}
	var parentID *uint
	if input.ParentUUID != nil {
		parent, err := FindParentByUUID(*input.ParentUUID)
		if err != nil {
			return Response{}, err
		}
		parentID = &parent.ID
	}
	geography.Name = input.Name
	geography.Code = input.Code
	geography.GeographyTypeID = geographyType.ID
	geography.ParentID = parentID
	if err := Save(&geography); err != nil {
		return Response{}, err
	}
	geography.GeographyType = geographyType
	resp := toGeographyResponse(geography, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID) error {
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "geography", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
