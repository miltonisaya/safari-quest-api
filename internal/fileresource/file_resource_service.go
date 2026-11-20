package fileresource

import (
	"errors"
	"fmt"
	"time"

	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/fileresourcetype"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrOutsideCompany   = errors.New("You can only manage files in your company")
	ErrInvalidReference = errors.New("invalid reference")
)

type Input struct {
	Name                 string    `json:"name" form:"name" binding:"required"`
	Code                 string    `json:"code" form:"code" binding:"required"`
	FileResourceTypeUUID uuid.UUID `json:"file_resource_type_uuid" form:"file_resource_type_uuid" binding:"required"`
	OperatorUUID         uuid.UUID `json:"operator_uuid" form:"operator_uuid" binding:"required"`
}

type Ref struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type Response struct {
	UUID             uuid.UUID `json:"uuid"`
	Name             string    `json:"name"`
	Code             string    `json:"code"`
	DateUploaded     time.Time `json:"date_uploaded"`
	FileResourceType Ref       `json:"file_resource_type"`
	Operator         Ref       `json:"operator"`
	Path             string    `json:"path"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type List struct {
	Items     []Response      `json:"items"`
	Meta      pagination.Meta `json:"meta"`
	Operators []Ref           `json:"operators"`
}

func ToResponse(item FileResource) Response {
	return Response{
		UUID:         item.UUID,
		Name:         item.Name,
		Code:         item.Code,
		DateUploaded: item.DateUploaded,
		FileResourceType: Ref{
			UUID: item.FileResourceType.UUID,
			Name: item.FileResourceType.Name,
			Code: item.FileResourceType.Code,
		},
		Operator:  Ref{UUID: item.Operator.UUID, Name: item.Operator.Name, Code: item.Operator.Code},
		Path:      item.Path,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func operatorRefs(scope companyscope.Scope) ([]Ref, error) {
	rows, err := touroperator.FindForCompanies(scope.Admin, scope.OperatorIDs)
	if err != nil {
		return nil, err
	}
	refs := make([]Ref, len(rows))
	for i, row := range rows {
		refs[i] = Ref{UUID: row.UUID, Name: row.Name, Code: row.Code}
	}
	return refs, nil
}

func GetAll(params pagination.Params, scope companyscope.Scope) (List, error) {
	rows, total, err := FindAll(params, scope.OperatorIDs, scope.Admin)
	if err != nil {
		return List{}, err
	}
	items := make([]Response, len(rows))
	for i, row := range rows {
		items[i] = ToResponse(row)
	}
	operators, err := operatorRefs(scope)
	if err != nil {
		return List{}, err
	}
	return List{Items: items, Meta: pagination.NewMeta(total, params), Operators: operators}, nil
}

func GetByUUID(id uuid.UUID, scope companyscope.Scope) (Response, error) {
	row, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(row.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	return ToResponse(row), nil
}

func resolve(input Input, scope companyscope.Scope) (fileresourcetype.FileResourceType, touroperator.TourOperator, error) {
	kind, err := fileresourcetype.FindByUUID(input.FileResourceTypeUUID)
	if err != nil {
		return fileresourcetype.FileResourceType{}, touroperator.TourOperator{}, fmt.Errorf("%w: file_resource_type_uuid", ErrInvalidReference)
	}
	operator, err := touroperator.FindByUUID(input.OperatorUUID)
	if err != nil {
		return fileresourcetype.FileResourceType{}, touroperator.TourOperator{}, fmt.Errorf("%w: operator_uuid", ErrInvalidReference)
	}
	if !scope.Allows(operator.ID) {
		return fileresourcetype.FileResourceType{}, touroperator.TourOperator{}, ErrOutsideCompany
	}
	return kind, operator, nil
}

func Create(input Input, path string, scope companyscope.Scope) (Response, error) {
	kind, operator, err := resolve(input, scope)
	if err != nil {
		return Response{}, err
	}
	row := FileResource{
		Name:               input.Name,
		Code:               input.Code,
		DateUploaded:       time.Now().UTC(),
		FileResourceTypeID: kind.ID,
		OperatorID:         operator.ID,
		Path:               path,
	}
	if err := Insert(&row); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(row.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResource", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input, scope companyscope.Scope) (Response, error) {
	row, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(row.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	kind, operator, err := resolve(input, scope)
	if err != nil {
		return Response{}, err
	}
	row.Name = input.Name
	row.Code = input.Code
	row.FileResourceTypeID = kind.ID
	row.OperatorID = operator.ID
	if err := Save(&row); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(row.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResource", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID, scope companyscope.Scope) error {
	row, err := FindByUUID(id)
	if err != nil {
		return err
	}
	if !scope.Allows(row.OperatorID) {
		return ErrOutsideCompany
	}
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "fileResource", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
