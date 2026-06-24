package services

import (
	"time"

	"safari-quest-api/models"
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

func RoleGetAll() ([]RoleResponse, error) {
	roles, err := repositories.RoleFindAll()
	if err != nil {
		return nil, err
	}
	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = toRoleResponse(role)
	}
	return response, nil
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
