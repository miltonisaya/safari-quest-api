package user

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/role"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var userSortColumns = map[string]string{
	"first_name": "first_name",
	"last_name":  "last_name",
	"email":      "email",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]User, int64, error) {
	var users []User
	var total int64

	db := database.GORM_DB.Model(&User{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR mobile ILIKE ?", term, term, term, term)
	}

	db.Count(&total)

	result := db.
		Preload("Roles").
		Order(params.OrderClause(userSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&users)

	return users, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (User, error) {
	var user User
	result := database.GORM_DB.Preload("Roles").First(&user, "uuid = ?", uuid)
	return user, result.Error
}

func FindByEmail(email string) (User, error) {
	var user User
	result := database.GORM_DB.Preload("Roles.Authorities").Where("email = ?", email).First(&user)
	return user, result.Error
}

func Insert(user *User, roles []role.Role) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Roles.*").Create(user).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(user).Association("Roles").Replace(roles)
	})
}

func Save(user *User, roles []role.Role) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(user).Association("Roles").Replace(roles)
	})
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&User{}).Error
}

// HasAuthority checks whether the user identified by userUUID holds any
// role that grants the given authority code. It does this in a single JOIN
// query across users → user_roles → roles → role_authorities → authorities
// rather than preloading the full object graph, keeping the hot-path check fast.
func HasAuthority(userUUID uuid.UUID, authorityCode string) (bool, error) {
	var count int64
	result := database.GORM_DB.
		Table("users u").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("JOIN role_authorities ra ON ra.role_id = r.id").
		Joins("JOIN authorities a ON a.id = ra.authority_id").
		Where("u.uuid = ? AND a.code = ? AND u.deleted_at IS NULL AND r.deleted_at IS NULL AND a.deleted_at IS NULL", userUUID, authorityCode).
		Count(&count)
	return count > 0, result.Error
}
