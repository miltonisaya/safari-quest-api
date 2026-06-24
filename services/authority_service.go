package services

import (
	"safari-quest-api/models"
	"safari-quest-api/repositories"

	"github.com/google/uuid"
)

type RoleInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

func RoleGetAll() ([]models.Role, error) {
	return repositories.RoleFindAll()
}

func RoleGetByUUID(uuid uuid.UUID) (models.Role, error) {
	return repositories.RoleFindByUUID(uuid)
}

func RoleCreate(input RoleInput) (models.Role, error) {
	role := models.Role{
		Name: input.Name,
		Code: input.Code,
	}
	err := repositories.RoleCreate(&role)
	return role, err
}

func RoleUpdate(uuid uuid.UUID, input RoleInput) (models.Role, error) {
	role, err := repositories.RoleFindByUUID(uuid)
	if err != nil {
		return role, err
	}
	role.Name = input.Name
	role.Code = input.Code
	err = repositories.RoleUpdate(&role)
	return role, err
}

func RoleDelete(uuid uuid.UUID) error {
	return repositories.RoleDelete(uuid)
}
