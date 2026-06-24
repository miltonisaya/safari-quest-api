package services

import (
	"safari-quest-api/models"
	"safari-quest-api/repositories"

	"github.com/google/uuid"
)

type AuthorityInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

func AuthorityGetAll() ([]models.Role, error) {
	return repositories.RoleFindAll()
}

func AuthorityGetByUUID(uuid uuid.UUID) (models.Role, error) {
	return repositories.RoleFindByUUID(uuid)
}

func AuthorityCreate(input RoleInput) (models.Role, error) {
	role := models.Role{
		Name: input.Name,
		Code: input.Code,
	}
	err := repositories.RoleCreate(&role)
	return role, err
}

func AuthorityUpdate(uuid uuid.UUID, input AuthorityInput) (models.Authority, error) {
	authority, err := repositories.AuthorityFindByUUID(uuid)
	if err != nil {
		return authority, err
	}
	authority.Name = input.Name
	err = repositories.AuthorityUpdate(&authority)
	return authority, err
}

func AuthorityDelete(uuid uuid.UUID) error {
	return repositories.AuthorityDelete(uuid)
}
