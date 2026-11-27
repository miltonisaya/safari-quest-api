package email

import (
	"time"

	"safari-quest-api/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func findByUUID(id string) (Outbox, error) {
	var row Outbox
	err := database.GORM_DB.Where("uuid = ?", id).Take(&row).Error
	return row, err
}

// claimPending locks up to limit unpublished emails in tx. SKIP LOCKED lets
// several API instances relay at once without taking the same rows.
func claimPending(tx *gorm.DB, limit int) ([]Outbox, error) {
	var rows []Outbox
	err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ?", StatusPending).Order("id").Limit(limit).Find(&rows).Error
	return rows, err
}

func markQueued(tx *gorm.DB, id uint, at time.Time) error {
	return tx.Model(&Outbox{}).Where("id = ?", id).Updates(map[string]any{"status": StatusQueued, "published_at": at}).Error
}

func markSent(id uint, at time.Time) error {
	return database.GORM_DB.Model(&Outbox{}).Where("id = ?", id).Updates(map[string]any{
		"status": StatusSent, "sent_at": at, "attempts": gorm.Expr("attempts + 1"), "last_error": nil,
	}).Error
}

// markAttemptFailed records a failed send; status becomes failed on the last attempt.
func markAttemptFailed(id uint, attempts int, reason string, final bool) error {
	status := StatusQueued
	if final {
		status = StatusFailed
	}
	return database.GORM_DB.Model(&Outbox{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "attempts": attempts, "last_error": reason,
	}).Error
}
