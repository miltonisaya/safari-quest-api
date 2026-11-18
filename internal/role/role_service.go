package role

import (
	"safari-quest-api/internal/authority"
	"time"

	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Input struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type Response struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a Role into its API representation.
func ToResponse(role Role) Response {
	return Response{
		UUID:      role.UUID,
		Name:      role.Name,
		Code:      role.Code,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func GetAll(params pagination.Params) (pagination.Result[Response], error) {
	roles, total, err := FindAll(params)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(roles))
	for i, role := range roles {
		items[i] = ToResponse(role)
	}
	return pagination.Result[Response]{
		Items: items,
		Meta:  pagination.NewMeta(total, params),
	}, nil
}

func GetByUUID(uuid uuid.UUID) (Response, error) {
	role, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	return ToResponse(role), nil
}

func Create(input Input) (Response, error) {
	role := Role{
		Name: input.Name,
		Code: input.Code,
	}
	if err := Insert(&role); err != nil {
		return Response{}, err
	}
	resp := ToResponse(role)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionCreated, Payload: resp})
	return resp, nil
}

func Update(uuid uuid.UUID, input Input) (Response, error) {
	role, err := FindByUUID(uuid)
	if err != nil {
		return Response{}, err
	}
	role.Name = input.Name
	role.Code = input.Code
	if err := Save(&role); err != nil {
		return Response{}, err
	}
	resp := ToResponse(role)
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionUpdated, Payload: resp})
	return resp, nil
}

type AuthoritiesInput struct {
	AuthorityUUIDs []uuid.UUID `json:"authority_uuids" binding:"required,min=1"`
}

func UpdateAuthorities(roleUUID uuid.UUID, input AuthoritiesInput) ([]authority.ListResponse, error) {
	role, err := FindByUUID(roleUUID)
	if err != nil {
		return nil, err
	}
	authorities, err := authority.FindByUUIDs(input.AuthorityUUIDs)
	if err != nil {
		return nil, err
	}
	if err := database.GORM_DB.Session(&gorm.Session{FullSaveAssociations: false}).
		Model(&role).Association("Authorities").Replace(authorities); err != nil {
		return nil, err
	}
	items := make([]authority.ListResponse, len(authorities))
	for i, a := range authorities {
		items[i] = authority.ListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}

func GetAuthorities(roleUUID uuid.UUID) ([]authority.ListResponse, error) {
	role, err := FindByUUIDWithAuthorities(roleUUID)
	if err != nil {
		return nil, err
	}
	items := make([]authority.ListResponse, len(role.Authorities))
	for i, a := range role.Authorities {
		items[i] = authority.ListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}

func Delete(uuid uuid.UUID) error {
	if err := Remove(uuid); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "role", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": uuid.String()}})
	return nil
}
