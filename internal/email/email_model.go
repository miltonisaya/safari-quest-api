package email

import (
	"time"

	"github.com/google/uuid"
)

// Outbox statuses.
const (
	StatusPending = "pending" // written, not yet on RabbitMQ
	StatusQueued  = "queued"  // published; the worker sends (or retries) it
	StatusSent    = "sent"
	StatusFailed  = "failed" // gave up after MaxAttempts; also in the dead-letter queue
)

// Outbox is one email waiting to be (or already) sent.
type Outbox struct {
	ID          uint      `gorm:"primaryKey"`
	UUID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Kind        string    `gorm:"not null"`
	ToAddress   string    `gorm:"not null"`
	Content     string    `gorm:"type:jsonb;not null"` // JSON of Content
	Status      string    `gorm:"not null;default:pending"`
	Attempts    int       `gorm:"not null;default:0"`
	LastError   *string
	PublishedAt *time.Time
	SentAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Outbox) TableName() string { return "email_outbox" }
