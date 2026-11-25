package auth

import (
	"strings"

	"safari-quest-api/database"
	"safari-quest-api/internal/role"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/user"

	"gorm.io/gorm"
)

func emailTaken(email string) (bool, error) {
	var count int64
	// Unscoped: a soft-deleted account still owns its email (unique index).
	err := database.GORM_DB.Unscoped().Model(&user.User{}).Where("LOWER(email) = ?", email).Count(&count).Error
	return count > 0, err
}

// companyNameTaken: company names are unique, soft-deleted ones included.
func companyNameTaken(name string) (bool, error) {
	var count int64
	err := database.GORM_DB.Unscoped().Model(&touroperator.TourOperator{}).Where("LOWER(name) = LOWER(?)", name).Count(&count).Error
	return count > 0, err
}

func roleByCode(code string) (role.Role, error) {
	var r role.Role
	err := database.GORM_DB.Where("code = ?", code).First(&r).Error
	return r, err
}

// createAccount writes the user, their role and, for operators, the company
// and the tour_operator_users link in one transaction.
func createAccount(account *user.User, roleRow role.Role, operator *touroperator.TourOperator) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Roles.*").Create(account).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(account).Association("Roles").Replace([]role.Role{roleRow}); err != nil {
			return err
		}
		if operator == nil {
			return nil
		}
		if err := tx.Omit("Geography", "Users").Create(operator).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO tour_operator_users (tour_operator_id, user_id) VALUES (?, ?)", operator.ID, account.ID).Error
	})
}

// isUniqueViolation reports a Postgres unique_violation (SQLSTATE 23505),
// e.g. two sign-ups racing for the same email.
func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") || strings.Contains(msg, "duplicate key")
}
