package user

import (
	"fmt"
	"os"
	"safari-quest-api/internal/role"
	"time"

	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Input struct {
	FirstName  string      `json:"first_name" binding:"required"`
	MiddleName string      `json:"middle_name"`
	LastName   string      `json:"last_name" binding:"required"`
	Email      string      `json:"email" binding:"required,email"`
	Sex        string      `json:"sex" binding:"required"`
	Mobile     string      `json:"mobile" binding:"required"`
	Address    string      `json:"address" binding:"required"`
	RoleIDs    []uuid.UUID `json:"role_ids" binding:"required,min=1"`
}

type UpdateInput struct {
	FirstName  string      `json:"first_name" binding:"required"`
	MiddleName string      `json:"middle_name"`
	LastName   string      `json:"last_name" binding:"required"`
	Email      string      `json:"email" binding:"required,email"`
	Sex        string      `json:"sex" binding:"required"`
	Mobile     string      `json:"mobile" binding:"required"`
	Address    string      `json:"address" binding:"required"`
	RoleIDs    []uuid.UUID `json:"role_ids" binding:"required,min=1"`
}

type Response struct {
	UUID            uuid.UUID       `json:"uuid"`
	FirstName       string          `json:"first_name"`
	MiddleName      string          `json:"middle_name"`
	LastName        string          `json:"last_name"`
	Email           string          `json:"email"`
	Sex             string          `json:"sex"`
	Mobile          string          `json:"mobile"`
	Address         string          `json:"address"`
	IsActive        bool            `json:"is_active"`
	EmailVerifiedAt *time.Time      `json:"email_verified_at"`
	Roles           []role.Response `json:"roles"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ToResponse converts a User into its API representation.
func ToResponse(user User) Response {
	roles := make([]role.Response, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = role.ToResponse(r)
	}
	return Response{
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

func resolveRoles(roleIDs []uuid.UUID) ([]role.Role, error) {
	roles, err := role.FindByUUIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(roleIDs) {
		return nil, fmt.Errorf("one or more role_ids are invalid")
	}
	return roles, nil
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	users, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(users))
	for i, user := range users {
		items[i] = ToResponse(user)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	user, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(user), nil
}

func Create(input Input) (Response, error) {
	roles, err := resolveRoles(input.RoleIDs)
	if err != nil {
		return Response{}, err
	}

	var defaultPassword = os.Getenv("DEFAULT_PASSWORD")
	hashed, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return Response{}, err
	}

	user := User{
		FirstName:  input.FirstName,
		MiddleName: input.MiddleName,
		LastName:   input.LastName,
		Email:      input.Email,
		Password:   string(hashed),
		Sex:        input.Sex,
		Mobile:     input.Mobile,
		Address:    input.Address,
	}
	if err := Insert(&user, roles); err != nil {
		return Response{}, err
	}
	resp := ToResponse(user)
	ws.GlobalHub.Publish(ws.Event{Entity: "user", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input UpdateInput) (Response, error) {
	user, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}

	roles, err := resolveRoles(input.RoleIDs)
	if err != nil {
		return Response{}, err
	}

	user.FirstName = input.FirstName
	user.MiddleName = input.MiddleName
	user.LastName = input.LastName
	user.Email = input.Email
	user.Sex = input.Sex
	user.Mobile = input.Mobile
	user.Address = input.Address

	if err := Save(&user, roles); err != nil {
		return Response{}, err
	}
	user.Roles = roles
	resp := ToResponse(user)
	ws.GlobalHub.Publish(ws.Event{Entity: "user", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "user", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
