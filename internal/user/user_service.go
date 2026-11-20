package user

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"safari-quest-api/config"
	"safari-quest-api/internal/role"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCannotDeactivateSelf = errors.New("You cannot deactivate your own account")
	ErrCannotDeleteSelf     = errors.New("You cannot delete your own account")
	ErrDefaultPasswordUnset = errors.New("default password is not configured")
	ErrInvalidRoleFilter    = errors.New("Invalid role filter")
	ErrInvalidStatusFilter  = errors.New("Invalid status filter")
	ErrOutsideCompany       = errors.New("You can only manage users in your company")
	ErrRoleNotAllowed       = errors.New("You can only grant roles you hold")
	errUnauthenticated      = errors.New("Unauthorized")
)

const administratorRoleCode = "ADMINISTRATOR"

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
	FirstName     string      `json:"first_name" binding:"required"`
	MiddleName    string      `json:"middle_name"`
	LastName      string      `json:"last_name" binding:"required"`
	Email         string      `json:"email" binding:"required,email"`
	Sex           string      `json:"sex" binding:"required"`
	Mobile        string      `json:"mobile" binding:"required"`
	Address       string      `json:"address" binding:"required"`
	RoleIDs       []uuid.UUID `json:"role_ids" binding:"required,min=1"`
	IsActive      *bool       `json:"is_active"`
	ResetPassword bool        `json:"reset_password"`
}

type listFilter struct {
	RoleUUID      *uuid.UUID
	Active        *bool
	CompanyOnly   bool
	OperatorIDs   []uint
	HeldRoleUUIDs []uuid.UUID
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

func parseListFilter(roleUUID, active string) (listFilter, error) {
	var filter listFilter
	roleUUID = strings.TrimSpace(roleUUID)
	active = strings.TrimSpace(active)
	if roleUUID != "" {
		id, err := uuid.Parse(roleUUID)
		if err != nil {
			return filter, ErrInvalidRoleFilter
		}
		filter.RoleUUID = &id
	}
	if active != "" {
		switch active {
		case "true":
			v := true
			filter.Active = &v
		case "false":
			v := false
			filter.Active = &v
		default:
			return filter, ErrInvalidStatusFilter
		}
	}
	return filter, nil
}

// mayManage reports whether a non-administrator may act on a user.
// Administrators are unrestricted. Everyone else must share a company with
// the target, and the target must not hold a role the caller does not.
func mayManage(admin, sameCompany bool, callerRoles, targetRoles []uuid.UUID) error {
	if admin {
		return nil
	}
	if !sameCompany {
		return ErrOutsideCompany
	}
	held := make(map[uuid.UUID]struct{}, len(callerRoles))
	for _, id := range callerRoles {
		held[id] = struct{}{}
	}
	for _, id := range targetRoles {
		if _, ok := held[id]; !ok {
			return ErrOutsideCompany
		}
	}
	return nil
}

// mayGrant reports whether the caller may assign these roles.
func mayGrant(admin bool, callerRoles, requested []uuid.UUID) error {
	if admin {
		return nil
	}
	held := make(map[uuid.UUID]struct{}, len(callerRoles))
	for _, id := range callerRoles {
		held[id] = struct{}{}
	}
	for _, id := range requested {
		if _, ok := held[id]; !ok {
			return ErrRoleNotAllowed
		}
	}
	return nil
}

func roleUUIDs(roles []role.Role) []uuid.UUID {
	ids := make([]uuid.UUID, len(roles))
	for i, item := range roles {
		ids[i] = item.UUID
	}
	return ids
}

func isAdministrator(roles []role.Role) bool {
	for _, item := range roles {
		if item.Code == administratorRoleCode {
			return true
		}
	}
	return false
}

func guardOwnAccount(actor, target uuid.UUID, active *bool, deleting bool) error {
	if actor != target {
		return nil
	}
	if deleting {
		return ErrCannotDeleteSelf
	}
	if active != nil && !*active {
		return ErrCannotDeactivateSelf
	}
	return nil
}

func defaultPasswordHash() (string, error) {
	raw := strings.TrimSpace(config.App.AdminDefaultPassword)
	if raw == "" {
		return "", ErrDefaultPasswordUnset
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func GetAll(params pagination.Params, filter listFilter) (pagination.Result[Response], error) {
	users, total, err := FindAll(params, filter)
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

	hashed, err := defaultPasswordHash()
	if err != nil {
		return Response{}, err
	}

	user := User{
		FirstName:  input.FirstName,
		MiddleName: input.MiddleName,
		LastName:   input.LastName,
		Email:      input.Email,
		Password:   hashed,
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

func Update(actor, id uuid.UUID, input UpdateInput) (Response, error) {
	if err := guardOwnAccount(actor, id, input.IsActive, false); err != nil {
		return Response{}, err
	}

	user, err := FindByUUID(id)
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
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	if input.ResetPassword {
		hashed, err := defaultPasswordHash()
		if err != nil {
			return Response{}, err
		}
		user.Password = hashed
	}

	if err := Save(&user, roles); err != nil {
		return Response{}, err
	}
	user.Roles = roles
	resp := ToResponse(user)
	ws.GlobalHub.Publish(ws.Event{Entity: "user", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(actor, id uuid.UUID) error {
	if err := guardOwnAccount(actor, id, nil, true); err != nil {
		return err
	}
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "user", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
