package email

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Enqueue writes emails to the outbox inside the caller's transaction, so
// they exist exactly when the change that causes them is committed. The relay
// publishes them to RabbitMQ; the worker sends them.
func Enqueue(tx *gorm.DB, drafts ...Draft) error {
	for _, d := range drafts {
		to := strings.TrimSpace(d.To)
		if to == "" {
			continue
		}
		body, err := json.Marshal(d.Content)
		if err != nil {
			return fmt.Errorf("email %s: %w", d.Kind, err)
		}
		row := Outbox{UUID: uuid.New(), Kind: d.Kind, ToAddress: to, Content: string(body), Status: StatusPending}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
