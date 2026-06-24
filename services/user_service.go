package services

import (
	"fmt"
	"time"

	"safari-quest-api/models"
	"safari-quest-api/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserInput struct {
	FirstName  string      `json:"first_name" binding:"required"`
	MiddleName string      `json:"middle_name"`
	LastName   string      `json:"last_name" binding:"required"`
	Email      string      `json:"email" binding:"required,email"`
	Password   string      `json:"password" binding:"required,min=8"`
	Sex        string      `json:"sex" binding:"required"`
	Mobile     string      `json:"mobile" binding:"required"`
	Address    string      `json:"address" binding:"required"`
	RoleIDs    []uuid.UUID `json:"role_ids" binding:"required,min=1"`
}

type UserResponse struct {
	UUID            uuid.UUID      `json:"uuid"`
	FirstName       string         `json:"first_name"`
	MiddleName      string         `json:"middle_name"`
	LastName        string         `json:"last_name"`
	Email           string         `json:"email"`
	Sex             string         `json:"sex"`
	Mobile          string         `json:"mobile"`
	Address         string         `json:"address"`
	IsActive        bool           `json:"is_active"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at"`
	Roles           []RoleResponse `json:"roles"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func toUserResponse(user models.User) UserResponse {
	roles := make([]RoleResponse, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = toRoleResponse(role)
	}
	return UserResponse{
		UUID:            user.UUID,
		FirstName:       user.FirstName,
		MiddleName:      user.MiddleName,
		LastName:        user.LastName,
		Email:           user.Email,
		Sex:             user.Sex,
		Mobile:          user.Mobile,
		Address:         user.Address,
		IsActive:        user.IsActive,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Roles:           roles,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

func resolveRoles(roleIDs []uuid.UUID) ([]models.Role, error) {
	roles, err := repositories.RoleFindByUUIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(roleIDs) {
		return nil, fmt.Errorf("one or more role_ids are invalid")
	}
	return roles, nil
}

func UserGetAll() ([]UserResponse, error) {
	users, err := repositories.UserFindAll()
	if err != nil {
		return nil, err
	}
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = toUserResponse(user)
	}
	return response, nil
}

func UserGetByUUID(uuid uuid.UUID) (UserResponse, error) {
	user, err := repositories.UserFindByUUID(uuid)
	if err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(user), nil
}

func UserCreate(input UserInput) (UserResponse, error) {
	roles, err := resolveRoles(input.RoleIDs)
	if err != nil {
		return UserResponse{}, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserResponse{}, err
	}

	user := models.User{
		FirstName:  input.FirstName,
		MiddleName: input.MiddleName,
		LastName:   input.LastName,
		Email:      input.Email,
		Password:   string(hashed),
		Sex:        input.Sex,
		Mobile:     input.Mobile,
		Address:    input.Address,
	}
	if err := repositories.UserCreate(&user, roles); err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(user), nil
}

func UserUpdate(uuid uuid.UUID, input UserInput) (UserResponse, error) {
	user, err := repositories.UserFindByUUID(uuid)
	if err != nil {
		return UserResponse{}, err
	}

	roles, err := resolveRoles(input.RoleIDs)
	if err != nil {
		return UserResponse{}, err
	}

	user.FirstName = input.FirstName
	user.MiddleName = input.MiddleName
	user.LastName = input.LastName
	user.Email = input.Email
	user.Sex = input.Sex
	user.Mobile = input.Mobile
	user.Address = input.Address

	if err := repositories.UserUpdate(&user, roles); err != nil {
		return UserResponse{}, err
	}
	user.Roles = roles
	return toUserResponse(user), nil
}

func UserDelete(uuid uuid.UUID) error {
	return repositories.UserDelete(uuid)
}
