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

func RoleGetByID(id uuid.UUID) (models.Role, error) {
	return repositories.RoleFindByID(id)
}

func RoleCreate(input RoleInput) (models.Role, error) {
	role := models.Role{
		Name: input.Name,
		Code: input.Code,
	}
	err := repositories.RoleCreate(&role)
	return role, err
}

func RoleUpdate(id uuid.UUID, input RoleInput) (models.Role, error) {
	role, err := repositories.RoleFindByID(id)
	if err != nil {
		return role, err
	}
	role.Name = input.Name
	role.Code = input.Code
	err = repositories.RoleUpdate(&role)
	return role, err
}

func RoleDelete(id uuid.UUID) error {
	return repositories.RoleDelete(id)
}
