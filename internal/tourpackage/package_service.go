package tourpackage

import (
	"errors"
	"fmt"
	"time"

	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrOutsideCompany   = errors.New("You can only manage packages in your company")
	ErrInvalidReference = errors.New("invalid reference")
	ErrInvalidPackage   = errors.New("Check the dates, price, and group size")
	ErrAttachment       = errors.New("Attachments must belong to the same company")
)

type Input struct {
	Name              string      `json:"name" binding:"required"`
	Description       string      `json:"description" binding:"required"`
	OperatorUUID      uuid.UUID   `json:"operator_uuid" binding:"required"`
	StartDate         string      `json:"start_date" binding:"required"`
	EndDate           string      `json:"end_date" binding:"required"`
	Pax               int         `json:"pax" binding:"required,min=1"`
	Price             float64     `json:"price" binding:"min=0"`
	PackageType       string      `json:"package_type" binding:"required"`
	FileResourceUUIDs []uuid.UUID `json:"file_resource_uuids"`
}

type Ref struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type FileRef struct {
	UUID             uuid.UUID `json:"uuid"`
	Name             string    `json:"name"`
	Code             string    `json:"code"`
	Path             string    `json:"path"`
	FileResourceType Ref       `json:"file_resource_type"`
}

type Response struct {
	UUID          uuid.UUID `json:"uuid"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Operator      Ref       `json:"operator"`
	StartDate     string    `json:"start_date"`
	EndDate       string    `json:"end_date"`
	Pax           int       `json:"pax"`
	Price         float64   `json:"price"`
	PackageType   string    `json:"package_type"`
	FileResources []FileRef `json:"file_resources"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type List struct {
	Items     []Response      `json:"items"`
	Meta      pagination.Meta `json:"meta"`
	Operators []Ref           `json:"operators"`
}

func ToResponse(item Package) Response {
	files := make([]FileRef, len(item.FileResources))
	for i, file := range item.FileResources {
		files[i] = FileRef{
			UUID: file.UUID,
			Name: file.Name,
			Code: file.Code,
			Path: file.Path,
			FileResourceType: Ref{
				UUID: file.FileResourceType.UUID,
				Name: file.FileResourceType.Name,
				Code: file.FileResourceType.Code,
			},
		}
	}
	return Response{
		UUID:          item.UUID,
		Name:          item.Name,
		Description:   item.Description,
		Operator:      Ref{UUID: item.Operator.UUID, Name: item.Operator.Name, Code: item.Operator.Code},
		StartDate:     item.StartDate.Format("2006-01-02"),
		EndDate:       item.EndDate.Format("2006-01-02"),
		Pax:           item.Pax,
		Price:         item.Price,
		PackageType:   item.PackageType,
		FileResources: files,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func parseTrip(input Input) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	if end.Before(start) || input.Pax < 1 || input.Price < 0 {
		return time.Time{}, time.Time{}, ErrInvalidPackage
	}
	return start, end, nil
}

func filesFor(operatorID uint, ids []uuid.UUID) ([]fileresource.FileResource, error) {
	if len(ids) == 0 {
		return []fileresource.FileResource{}, nil
	}
	files, err := fileresource.FindByUUIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(files) != len(ids) {
		return nil, fmt.Errorf("%w: one or more file_resource_uuids", ErrInvalidReference)
	}
	for _, file := range files {
		if file.OperatorID != operatorID {
			return nil, ErrAttachment
		}
	}
	return files, nil
}

func apply(item *Package, input Input, scope companyscope.Scope) ([]fileresource.FileResource, error) {
	start, end, err := parseTrip(input)
	if err != nil {
		return nil, err
	}
	operator, err := touroperator.FindByUUID(input.OperatorUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: operator_uuid", ErrInvalidReference)
	}
	if !scope.Allows(operator.ID) {
		return nil, ErrOutsideCompany
	}
	files, err := filesFor(operator.ID, input.FileResourceUUIDs)
	if err != nil {
		return nil, err
	}
	item.Name = input.Name
	item.Description = input.Description
	item.OperatorID = operator.ID
	item.Operator = operator
	item.StartDate = start
	item.EndDate = end
	item.Pax = input.Pax
	item.Price = input.Price
	item.PackageType = input.PackageType
	item.FileResources = files
	return files, nil
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
	item, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(item.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	return ToResponse(item), nil
}

func Create(input Input, scope companyscope.Scope) (Response, error) {
	item := Package{}
	files, err := apply(&item, input, scope)
	if err != nil {
		return Response{}, err
	}
	if err := Insert(&item, files); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(item.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(id uuid.UUID, input Input, scope companyscope.Scope) (Response, error) {
	item, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(item.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	files, err := apply(&item, input, scope)
	if err != nil {
		return Response{}, err
	}
	if err := Save(&item, files); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(item.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

func Delete(id uuid.UUID, scope companyscope.Scope) error {
	item, err := FindByUUID(id)
	if err != nil {
		return err
	}
	if !scope.Allows(item.OperatorID) {
		return ErrOutsideCompany
	}
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "package", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}
