package agegroup

import (
	"safari-quest-api/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func FindAll() ([]AgeGroup, error) {
	var rows []AgeGroup
	err := database.GORM_DB.Order("min_age").Find(&rows).Error
	return rows, err
}

func FindByUUID(id uuid.UUID) (AgeGroup, error) {
	var row AgeGroup
	err := database.GORM_DB.Where("uuid = ?", id).Take(&row).Error
	return row, err
}

// inUse reports which groups active (pending or confirmed) bookings use.
func inUse(ids []uint) (map[uint]bool, error) {
	out := map[uint]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	var used []uint
	err := database.GORM_DB.Table("booking_travellers bt").
		Joins("JOIN bookings b ON b.id = bt.booking_id").
		Where("bt.age_group_id IN ? AND b.status IN ('pending', 'confirmed') AND b.deleted_at IS NULL", ids).
		Distinct().Pluck("bt.age_group_id", &used).Error
	for _, id := range used {
		out[id] = true
	}
	return out, err
}

// replaceAll saves the new list in one transaction: rows with a known UUID are
// updated, new ones inserted, and the rest deleted.
func replaceAll(keep []AgeGroup, remove []uint) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if len(remove) > 0 {
			if err := tx.Delete(&AgeGroup{}, remove).Error; err != nil {
				return err
			}
		}
		for i := range keep {
			if err := tx.Save(&keep[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
