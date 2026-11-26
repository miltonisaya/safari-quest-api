package account

import "time"

// PasswordResetToken is a one-time password reset link. Only the SHA-256 hash
// of the emailed token is stored.
type PasswordResetToken struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
