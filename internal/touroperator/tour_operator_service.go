package touroperator

import (
	"errors"
	"fmt"
	"time"

	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

// AdministratorRoleCode identifies platform administrators, who may manage
// the staff of any tour operator.
const AdministratorRoleCode = "ADMINISTRATOR"

// OperatorRoleCode is the role given to tour company staff.
const OperatorRoleCode = "TOUR_OPERATOR"

var (
	// ErrForbidden is returned when the requester is neither staff of the
	// operator nor a platform administrator.
	ErrForbidden = errors.New("you can only manage users of your own tour operator")
	// ErrRoleNotAllowed is returned when an operator tries to grant a role it
	// does not hold itself.
	ErrRoleNotAllowed = errors.New("you can only assign roles you hold yourself")
)

type Input struct {
	Name               string     `json:"name" binding:"required"`
	Code               string     `json:"code" binding:"required"`
	LegalName          *string    `json:"legal_name"`
	RegistrationNumber *string    `json:"registration_number"`
	LicenseNumber      *string    `json:"license_number"`
	Email              string     `json:"email" binding:"required,email"`
	Phone              string     `json:"phone" binding:"required"`
	Website            *string    `json:"website" binding:"omitempty,url"`
	Address            *string    `json:"address"`
	GeographyUUID      *uuid.UUID `json:"geography_uuid"`
	Description        *string    `json:"description"`
	IsVerified         bool       `json:"is_verified"`
	IsActive           *bool      `json:"is_active"`
}

// GeographyRef is the compact form of the operator's base location.
type GeographyRef struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type Response struct {
	UUID               uuid.UUID     `json:"uuid"`
	Name               string        `json:"name"`
	Code               string        `json:"code"`
	LegalName          *string       `json:"legal_name"`
	RegistrationNumber *string       `json:"registration_number"`
	LicenseNumber      *string       `json:"license_number"`
	Email              string        `json:"email"`
	Phone              string        `json:"phone"`
	Website            *string       `json:"website"`
	Address            *string       `json:"address"`
	Geography          *GeographyRef `json:"geography"`
	Description        *string       `json:"description"`
	IsVerified         bool          `json:"is_verified"`
	IsActive           bool          `json:"is_active"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// ToResponse converts a TourOperator into its API representation.
func ToResponse(operator TourOperator) Response {
	var geo *GeographyRef
	if operator.Geography != nil {
		geo = &GeographyRef{UUID: operator.Geography.UUID, Name: operator.Geography.Name, Code: operator.Geography.Code}
	}
	return Response{
		UUID:               operator.UUID,
		Name:               operator.Name,
		Code:               operator.Code,
		LegalName:          operator.LegalName,
		RegistrationNumber: operator.RegistrationNumber,
		LicenseNumber:      operator.LicenseNumber,
		Email:              operator.Email,
		Phone:              operator.Phone,
		Website:            operator.Website,
		Address:            operator.Address,
		Geography:          geo,
		Description:        operator.Description,
		IsVerified:         operator.IsVerified,
		IsActive:           operator.IsActive,
		CreatedAt:          operator.CreatedAt,
		UpdatedAt:          operator.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	operators, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(operators))
	for i, o := range operators {
		items[i] = ToResponse(o)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	operator, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(operator), nil
}

// apply copies the input onto the operator and resolves its geography.
func apply(operator *TourOperator, input Input) error {
	operator.Name = input.Name
	operator.Code = input.Code
	operator.LegalName = input.LegalName
	operator.RegistrationNumber = input.RegistrationNumber
	operator.LicenseNumber = input.LicenseNumber
	operator.Email = input.Email
	operator.Phone = input.Phone
	operator.Website = input.Website
	operator.Address = input.Address
	operator.Description = input.Description
	operator.IsVerified = input.IsVerified
	if input.IsActive != nil {
		operator.IsActive = *input.IsActive
	}

	operator.GeographyID, operator.Geography = nil, nil
	if input.GeographyUUID != nil {
		geo, err := geography.FindByUUID(*input.GeographyUUID)
		if err != nil {
			return fmt.Errorf("geography_uuid: %w", err)
		}
		operator.GeographyID, operator.Geography = &geo.ID, &geo
	}
	return nil
}

func Create(input Input) (Response, error) {
	operator := TourOperator{IsActive: true}
	if err := apply(&operator, input); err != nil {
		return Response{}, err
	}
	if err := Insert(&operator); err != nil {
		return Response{}, err
	}
	resp := ToResponse(operator)
	ws.GlobalHub.Publish(ws.Event{Entity: "tourOperator", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input Input) (Response, error) {
	operator, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	if err := apply(&operator, input); err != nil {
		return Response{}, err
	}
	if err := Save(&operator); err != nil {
		return Response{}, err
	}
	resp := ToResponse(operator)
	ws.GlobalHub.Publish(ws.Event{Entity: "tourOperator", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "tourOperator", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}

// ── Staff (tour_operator_users) ─────────────────────────────────────────────

// authorize loads the operator and the requester, and checks the requester is
// either staff of this operator or a platform administrator.
func authorize(operatorUUID, requesterUUID uuid.UUID) (TourOperator, user.User, bool, error) {
	operator, err := FindByUUID(operatorUUID)
	if err != nil {
		return TourOperator{}, user.User{}, false, err
	}
	requester, err := user.FindByUUID(requesterUUID)
	if err != nil {
		return TourOperator{}, user.User{}, false, err
	}
	for _, r := range requester.Roles {
		if r.Code == AdministratorRoleCode {
			return operator, requester, true, nil
		}
	}
	member, err := IsStaff(operator.ID, requester.ID)
	if err != nil {
		return TourOperator{}, user.User{}, false, err
	}
	if !member {
		return TourOperator{}, user.User{}, false, ErrForbidden
	}
	return operator, requester, false, nil
}

func GetStaff(operatorUUID, requesterUUID uuid.UUID, params pagination.Params) (pagination.Result[user.Response], error) {
	operator, _, _, err := authorize(operatorUUID, requesterUUID)
	if err != nil {
		return pagination.Result[user.Response]{}, err
	}
	staff, total, err := FindStaff(operator.ID, params)
	if err != nil {
		return pagination.Result[user.Response]{}, err
	}
	items := make([]user.Response, len(staff))
	for i, u := range staff {
		items[i] = user.ToResponse(u)
	}
	return pagination.Result[user.Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

// CreateStaff creates a user account and links it to the operator. Operator
// staff may only grant roles they hold themselves; administrators may grant any.
func CreateStaff(operatorUUID, requesterUUID uuid.UUID, input user.Input) (user.Response, error) {
	operator, requester, isAdmin, err := authorize(operatorUUID, requesterUUID)
	if err != nil {
		return user.Response{}, err
	}
	if !isAdmin {
		held := make(map[uuid.UUID]bool, len(requester.Roles))
		for _, r := range requester.Roles {
			held[r.UUID] = true
		}
		for _, id := range input.RoleIDs {
			if !held[id] {
				return user.Response{}, ErrRoleNotAllowed
			}
		}
	}

	created, err := user.Create(input)
	if err != nil {
		return user.Response{}, err
	}
	account, err := user.FindByUUID(created.UUID)
	if err == nil {
		err = AttachStaff(operator.ID, account.ID)
	}
	if err != nil {
		// Do not leave an orphan account the operator cannot see.
		_ = user.Remove(created.UUID)
		return user.Response{}, err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "tourOperatorUser", Action: ws.ActionCreated, Payload: created})
	return created, nil
}

// RemoveStaff unlinks a user from the operator. The user account is kept.
func RemoveStaff(operatorUUID, requesterUUID, userUUID uuid.UUID) error {
	operator, _, _, err := authorize(operatorUUID, requesterUUID)
	if err != nil {
		return err
	}
	account, err := user.FindByUUID(userUUID)
	if err != nil {
		return err
	}
	if err := DetachStaff(operator.ID, account.ID); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "tourOperatorUser", Action: ws.ActionDeleted, Payload: map[string]string{
		"tour_operator_uuid": operator.UUID.String(),
		"user_uuid":          account.UUID.String(),
	}})
	return nil
}
