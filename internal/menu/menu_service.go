package menu

import (
	"safari-quest-api/internal/authority"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"
	"time"

	"github.com/google/uuid"
)

type AuthorityResponse struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

func toAuthorityResponse(authority authority.Authority) AuthorityResponse {
	return AuthorityResponse{
		UUID: authority.UUID,
		Name: authority.Name,
		Code: authority.Code,
	}
}

type Input struct {
	Name          string     `json:"name" binding:"required"`
	Type          string     `json:"type" binding:"required"`
	Route         *string    `json:"route"`
	Icon          *string    `json:"icon"`
	AuthorityUUID uuid.UUID  `json:"authorityUuid" binding:"required"`
	ParentUUID    *uuid.UUID `json:"parentUuid"`
	SortOrder     int        `json:"sortOrder"`
	Active        bool       `json:"active"`
}

type Response struct {
	UUID       uuid.UUID         `json:"uuid"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Route      *string           `json:"route,omitempty"`
	Icon       *string           `json:"icon,omitempty"`
	SortOrder  int               `json:"sortOrder"`
	Level      int               `json:"level"`
	Active     bool              `json:"active"`
	ParentUUID *uuid.UUID        `json:"parentUuid,omitempty"`
	Authority  AuthorityResponse `json:"authority"`
	Children   []Response        `json:"children,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func toMenuResponse(menu Menu, level int) Response {
	children := make([]Response, len(menu.Children))
	for i, child := range menu.Children {
		children[i] = toMenuResponse(child, level+1)
	}
	var parentUUID *uuid.UUID
	if menu.Parent != nil {
		parentUUID = &menu.Parent.UUID
	}
	return Response{
		UUID:       menu.UUID,
		Name:       menu.Name,
		Type:       string(menu.Type),
		Route:      menu.Route,
		Icon:       menu.Icon,
		SortOrder:  menu.SortOrder,
		Level:      level,
		Active:     menu.Active,
		ParentUUID: parentUUID,
		Authority:  toAuthorityResponse(menu.Authority),
		Children:   children,
		CreatedAt:  menu.CreatedAt,
		UpdatedAt:  menu.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	menus, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(menus))
	for i, menu := range menus {
		items[i] = toMenuResponse(menu, 0)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(id uuid.UUID) (Response, error) {
	menu, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	return toMenuResponse(menu, 0), nil
}

func Create(input Input) (Response, error) {
	authority, err := authority.FindByUUID(input.AuthorityUUID)
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
	menu := Menu{
		Name:        input.Name,
		Type:        Type(input.Type),
		Route:       input.Route,
		Icon:        input.Icon,
		AuthorityID: authority.ID,
		ParentID:    parentID,
		SortOrder:   input.SortOrder,
		Active:      input.Active,
	}
	if err := Insert(&menu); err != nil {
		return Response{}, err
	}
	menu.Authority = authority
	resp := toMenuResponse(menu, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "menu", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input) (Response, error) {
	menu, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	authority, err := authority.FindByUUID(input.AuthorityUUID)
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
	menu.Name = input.Name
	menu.Type = Type(input.Type)
	menu.Route = input.Route
	menu.Icon = input.Icon
	menu.AuthorityID = authority.ID
	menu.ParentID = parentID
	menu.SortOrder = input.SortOrder
	menu.Active = input.Active
	if err := Save(&menu); err != nil {
		return Response{}, err
	}
	menu.Authority = authority
	resp := toMenuResponse(menu, 0)
	ws.GlobalHub.Publish(ws.Event{Entity: "menu", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID) error {
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "menu", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
