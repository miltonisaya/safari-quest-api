package touroperator

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var tourOperatorSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

var staffSortColumns = map[string]string{
	"first_name": "users.first_name",
	"last_name":  "users.last_name",
	"email":      "users.email",
	"created_at": "users.created_at",
	"updated_at": "users.updated_at",
}

func FindAll(params pagination.Params) ([]TourOperator, int64, error) {
	var operators []TourOperator
	var total int64

	db := database.GORM_DB.Model(&TourOperator{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ? OR legal_name ILIKE ? OR email ILIKE ?", term, term, term, term)
	}

	db.Count(&total)

	result := db.
		Preload("Geography").
		Order(params.OrderClause(tourOperatorSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&operators)

	return operators, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (TourOperator, error) {
	var operator TourOperator
	result := database.GORM_DB.Preload("Geography").First(&operator, "uuid = ?", uuid)
	return operator, result.Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]TourOperator, error) {
	var operators []TourOperator
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&operators)
	return operators, result.Error
}

// FindForCompanies returns active operators the caller may sell for.
// all ignores ids. shortcut: capped at 100 names, raise the limit when a caller belongs to more companies.
func FindForCompanies(all bool, ids []uint) ([]TourOperator, error) {
	query := database.GORM_DB.Model(&TourOperator{}).Where("is_active = ?", true)
	if !all {
		if len(ids) == 0 {
			return []TourOperator{}, nil
		}
		query = query.Where("id IN ?", ids)
	}
	var operators []TourOperator
	err := query.Order("name").Limit(100).Find(&operators).Error
	return operators, err
}

func Insert(operator *TourOperator) error {
	return database.GORM_DB.Omit("Geography", "Users").Create(operator).Error
}

func Save(operator *TourOperator) error {
	return database.GORM_DB.Omit("Geography", "Users").Save(operator).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&TourOperator{}).Error
}

// FindStaff returns the users linked to the operator through tour_operator_users.
func FindStaff(operatorID uint, params pagination.Params) ([]user.User, int64, error) {
	var staff []user.User
	var total int64

	db := database.GORM_DB.Model(&user.User{}).
		Joins("JOIN tour_operator_users tou ON tou.user_id = users.id").
		Where("tou.tour_operator_id = ?", operatorID)

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", term, term, term)
	}

	db.Count(&total)

	result := db.
		Preload("Roles").
		Order(params.OrderClause(staffSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&staff)

	return staff, total, result.Error
}

// IsStaff reports whether userID is linked to operatorID.
func IsStaff(operatorID, userID uint) (bool, error) {
	var count int64
	err := database.GORM_DB.Table("tour_operator_users").
		Where("tour_operator_id = ? AND user_id = ?", operatorID, userID).
		Count(&count).Error
	return count > 0, err
}

func AttachStaff(operatorID, userID uint) error {
	return database.GORM_DB.Exec(
		"INSERT INTO tour_operator_users (tour_operator_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		operatorID, userID,
	).Error
}

// DetachStaff unlinks the user from the operator; the user account is kept.
func DetachStaff(operatorID, userID uint) error {
	result := database.GORM_DB.Exec(
		"DELETE FROM tour_operator_users WHERE tour_operator_id = ? AND user_id = ?",
		operatorID, userID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
