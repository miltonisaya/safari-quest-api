package user

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/role"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var userSortColumns = map[string]string{
	"first_name": "users.first_name",
	"last_name":  "users.last_name",
	"email":      "users.email",
	"created_at": "users.created_at",
	"updated_at": "users.updated_at",
}

func FindAll(params pagination.Params, filter listFilter) ([]User, int64, error) {
	var users []User
	var total int64

	db := database.GORM_DB.Model(&User{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where(
			"users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ? OR users.mobile ILIKE ?",
			term, term, term, term,
		)
	}
	if filter.RoleUUID != nil {
		db = db.
			Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN roles ON roles.id = user_roles.role_id AND roles.deleted_at IS NULL").
			Where("roles.uuid = ?", *filter.RoleUUID)
	}
	if filter.Active != nil {
		db = db.Where("users.is_active = ?", *filter.Active)
	}
	if filter.CompanyOnly {
		if len(filter.OperatorIDs) == 0 || len(filter.HeldRoleUUIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			db = db.Where(`users.id IN (
				SELECT tou.user_id FROM tour_operator_users tou
				JOIN tour_operators t ON t.id = tou.tour_operator_id AND t.deleted_at IS NULL
				WHERE tou.tour_operator_id IN ?
			)`, filter.OperatorIDs).
				Where(`NOT EXISTS (
					SELECT 1 FROM user_roles ur
					JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
					WHERE ur.user_id = users.id AND r.uuid NOT IN ?
				)`, filter.HeldRoleUUIDs)
		}
	}

	db.Count(&total)

	result := db.
		Preload("Roles").
		Order(params.OrderClause(userSortColumns, "users.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&users)

	return users, total, result.Error
}

func OperatorIDsForUser(userID uint) ([]uint, error) {
	var ids []uint
	err := database.GORM_DB.Table("tour_operator_users AS tou").
		Joins("JOIN tour_operators t ON t.id = tou.tour_operator_id AND t.deleted_at IS NULL").
		Where("tou.user_id = ?", userID).
		Pluck("tou.tour_operator_id", &ids).Error
	return ids, err
}

func SharesCompany(operatorIDs []uint, userID uint) (bool, error) {
	if len(operatorIDs) == 0 {
		return false, nil
	}
	var count int64
	err := database.GORM_DB.Table("tour_operator_users AS tou").
		Joins("JOIN tour_operators t ON t.id = tou.tour_operator_id AND t.deleted_at IS NULL").
		Where("tou.tour_operator_id IN ? AND tou.user_id = ?", operatorIDs, userID).
		Count(&count).Error
	return count > 0, err
}

func AttachToCompanies(userID uint, operatorIDs []uint) error {
	for _, operatorID := range operatorIDs {
		err := database.GORM_DB.Exec(
			"INSERT INTO tour_operator_users (tour_operator_id, user_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			operatorID, userID,
		).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func Active(id uuid.UUID) (bool, error) {
	var row struct {
		IsActive bool
	}
	err := database.GORM_DB.Model(&User{}).Select("is_active").Where("uuid = ?", id).Take(&row).Error
	return row.IsActive, err
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
