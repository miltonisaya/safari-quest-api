package services

import (
	"time"

	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/repositories"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type RoleResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toRoleResponse(role models.Role) RoleResponse {
	return RoleResponse{
		UUID:      role.UUID,
		Name:      role.Name,
		Code:      role.Code,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func RoleGetAll(params pagination.Params) (pagination.Result[RoleResponse], error) {
	roles, total, err := repositories.RoleFindAll(params)
	if err != nil {
		return pagination.Result[RoleResponse]{}, err
	}
	items := make([]RoleResponse, len(roles))
	for i, role := range roles {
		items[i] = toRoleResponse(role)
	}
	return pagination.Result[RoleResponse]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func RoleGetByUUID(uuid uuid.UUID) (RoleResponse, error) {
	role, err := repositories.RoleFindByUUID(uuid)
	if err != nil {
		return RoleResponse{}, err
	}
	return toRoleResponse(role), nil
}

func RoleCreate(input RoleInput) (RoleResponse, error) {
	role := models.Role{
		Name: input.Name,
		Code: input.Code,
	}
	if err := repositories.RoleCreate(&role); err != nil {
		return RoleResponse{}, err
	}
	resp := toRoleResponse(role)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func RoleUpdate(uuid uuid.UUID, input RoleInput) (RoleResponse, error) {
	role, err := repositories.RoleFindByUUID(uuid)
	if err != nil {
		return RoleResponse{}, err
	}
	role.Name = input.Name
	role.Code = input.Code
	if err := repositories.RoleUpdate(&role); err != nil {
		return RoleResponse{}, err
	}
	resp := toRoleResponse(role)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

type RoleAuthoritiesInput struct {
	AuthorityUUIDs []uuid.UUID `json:"authority_uuids" binding:"required,min=1"`
}

func RoleUpdateAuthorities(roleUUID uuid.UUID, input RoleAuthoritiesInput) ([]AuthorityListResponse, error) {
	role, err := repositories.RoleFindByUUID(roleUUID)
	if err != nil {
		return nil, err
	}
	authorities, err := repositories.AuthorityFindByUUIDs(input.AuthorityUUIDs)
	if err != nil {
		return nil, err
	}
	if err := database.GORM_DB.Session(&gorm.Session{FullSaveAssociations: false}).
		Model(&role).Association("Authorities").Replace(authorities); err != nil {
		return nil, err
	}
	items := make([]AuthorityListResponse, len(authorities))
	for i, a := range authorities {
		items[i] = AuthorityListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}

func RoleGetAuthorities(roleUUID uuid.UUID) ([]AuthorityListResponse, error) {
	role, err := repositories.RoleFindByUUIDWithAuthorities(roleUUID)
	if err != nil {
		return nil, err
	}
	items := make([]AuthorityListResponse, len(role.Authorities))
	for i, a := range role.Authorities {
		items[i] = AuthorityListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}

func RoleDelete(uuid uuid.UUID) error {
	if err := repositories.RoleDelete(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
