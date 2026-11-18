package authority

import (
	"time"

	"github.com/google/uuid"
)

type ListResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetAll() ([]ListResponse, error) {
	authorities, err := FindAll()
	if err != nil {
		return nil, err
	}
	items := make([]ListResponse, len(authorities))
	for i, a := range authorities {
		items[i] = ListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}
