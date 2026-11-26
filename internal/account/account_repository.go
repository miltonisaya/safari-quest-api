package account

import (
	"time"

	"safari-quest-api/database"
	"safari-quest-api/internal/user"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// issuedSince reports whether the user was sent a reset link after t.
func issuedSince(userID uint, t time.Time) (bool, error) {
	var count int64
	err := database.GORM_DB.Model(&PasswordResetToken{}).
		Where("user_id = ? AND created_at > ?", userID, t).Count(&count).Error
	return count > 0, err
}

// insertToken stores a new link and drops the user's older unused ones, so
// only the latest email works.
func insertToken(t *PasswordResetToken) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND used_at IS NULL", t.UserID).Delete(&PasswordResetToken{}).Error; err != nil {
			return err
		}
		return tx.Create(t).Error
	})
}

// redeem sets the password for a valid, unused, unexpired token and uses it
// up, in one transaction that locks the token row (no double use).
// It returns gorm.ErrRecordNotFound for unknown, used or expired tokens.
func redeem(hash, hashedPassword string, now time.Time) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var t PasswordResetToken
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, now).
			Take(&t).Error
		if err != nil {
			return err
		}
		res := tx.Model(&user.User{}).Where("id = ? AND is_active", t.UserID).Update("password", hashedPassword)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound // account removed or deactivated since the email
		}
		if err := tx.Model(&t).Update("used_at", now).Error; err != nil {
			return err
		}
		// Any other outstanding links stop working too.
		return tx.Where("user_id = ? AND used_at IS NULL", t.UserID).Delete(&PasswordResetToken{}).Error
	})
}

// saveProfile updates only the profile columns, never email, roles or status.
func saveProfile(id uint, fields map[string]any) error {
	return database.GORM_DB.Model(&user.User{}).Where("id = ?", id).Updates(fields).Error
}

func savePassword(id uint, hashed string) error {
	return database.GORM_DB.Model(&user.User{}).Where("id = ?", id).Update("password", hashed).Error
}
