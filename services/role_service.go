package services

import (
	"time"

	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"
	"safari-quest-api/repositories"

	"github.com/google/uuid"
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
	return toRoleResponse(role), nil
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
	return toRoleResponse(role), nil
}

func RoleDelete(uuid uuid.UUID) error {
	return repositories.RoleDelete(uuid)
}
