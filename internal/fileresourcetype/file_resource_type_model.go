package fileresourcetype

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileResourceType is a kind of upload an administrator allows,
// such as an image, a document, or a video.
type FileResourceType struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *FileResourceType) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}
