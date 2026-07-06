package services

import (
	"safari-quest-api/repositories"
	"time"

	"github.com/google/uuid"
)

type AuthorityListResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func AuthorityGetAll() ([]AuthorityListResponse, error) {
	authorities, err := repositories.AuthorityFindAll()
	if err != nil {
		return nil, err
	}
	items := make([]AuthorityListResponse, len(authorities))
	for i, a := range authorities {
		items[i] = AuthorityListResponse{
			UUID:      a.UUID,
			Name:      a.Name,
			Code:      a.Code,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}
	return items, nil
}
