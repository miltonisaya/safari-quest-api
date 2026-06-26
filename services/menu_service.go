package services

import (
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/repositories"
	"time"

	"github.com/google/uuid"
)

type AuthorityResponse struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

func toAuthorityResponse(authority models.Authority) AuthorityResponse {
	return AuthorityResponse{
		UUID: authority.UUID,
		Name: authority.Name,
		Code: authority.Code,
	}
}

type MenuInput struct {
	Name          string     `json:"name" binding:"required"`
	Type          string     `json:"type" binding:"required"`
	Route         *string    `json:"route"`
	Icon          *string    `json:"icon"`
	AuthorityUUID uuid.UUID  `json:"authorityUuid" binding:"required"`
	ParentID      *uint      `json:"parentId"`
	SortOrder     int        `json:"sortOrder"`
	Active        bool       `json:"active"`
}

type MenuResponse struct {
	UUID      uuid.UUID         `json:"uuid"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Route     *string           `json:"route,omitempty"`
	Icon      *string           `json:"icon,omitempty"`
	SortOrder int               `json:"sortOrder"`
	Active    bool              `json:"active"`
	Authority AuthorityResponse `json:"authority"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func toMenuResponse(menu models.Menu) MenuResponse {
	return MenuResponse{
		UUID:      menu.UUID,
		Name:      menu.Name,
		Type:      string(menu.Type),
		Route:     menu.Route,
		Icon:      menu.Icon,
		SortOrder: menu.SortOrder,
		Active:    menu.Active,
		Authority: toAuthorityResponse(menu.Authority),
		CreatedAt: menu.CreatedAt,
		UpdatedAt: menu.UpdatedAt,
	}
}

func MenuGetAll(params pagination.Params) (pagination.Result[MenuResponse], error) {
	menus, total, err := repositories.MenuFindAll(params)
	if err != nil {
		return pagination.Result[MenuResponse]{}, err
	}
	items := make([]MenuResponse, len(menus))
	for i, menu := range menus {
		items[i] = toMenuResponse(menu)
	}
	return pagination.Result[MenuResponse]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func MenuGetByUUID(id uuid.UUID) (MenuResponse, error) {
	menu, err := repositories.MenuFindByUUID(id)
	if err != nil {
		return MenuResponse{}, err
	}
	return toMenuResponse(menu), nil
}

func MenuCreate(input MenuInput) (MenuResponse, error) {
	authority, err := repositories.AuthorityFindByUUID(input.AuthorityUUID)
	if err != nil {
		return MenuResponse{}, err
	}
	menu := models.Menu{
		Name:        input.Name,
		Type:        models.MenuType(input.Type),
		Route:       input.Route,
		Icon:        input.Icon,
		AuthorityID: authority.ID,
		ParentID:    input.ParentID,
		SortOrder:   input.SortOrder,
		Active:      input.Active,
	}
	if err := repositories.MenuCreate(&menu); err != nil {
		return MenuResponse{}, err
	}
	menu.Authority = authority
	return toMenuResponse(menu), nil
}

func MenuUpdate(id uuid.UUID, input MenuInput) (MenuResponse, error) {
	menu, err := repositories.MenuFindByUUID(id)
	if err != nil {
		return MenuResponse{}, err
	}
	authority, err := repositories.AuthorityFindByUUID(input.AuthorityUUID)
	if err != nil {
		return MenuResponse{}, err
	}
	menu.Name = input.Name
	menu.Type = models.MenuType(input.Type)
	menu.Route = input.Route
	menu.Icon = input.Icon
	menu.AuthorityID = authority.ID
	menu.ParentID = input.ParentID
	menu.SortOrder = input.SortOrder
	menu.Active = input.Active
	if err := repositories.MenuUpdate(&menu); err != nil {
		return MenuResponse{}, err
	}
	menu.Authority = authority
	return toMenuResponse(menu), nil
}

func MenuDelete(id uuid.UUID) error {
	return repositories.MenuDelete(id)
}
