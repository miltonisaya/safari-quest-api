package tourpackage

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sortColumns = map[string]string{
	"name":         "packages.name",
	"start_date":   "packages.start_date",
	"end_date":     "packages.end_date",
	"pax":          "packages.pax",
	"price":        "packages.price",
	"package_type": "packages.package_type",
	"created_at":   "packages.created_at",
	"updated_at":   "packages.updated_at",
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Operator").Preload("FileResources.FileResourceType").Preload("FileResources.Operator")
}

func FindAll(params pagination.Params, operatorIDs []uint, unrestricted bool) ([]Package, int64, error) {
	var items []Package
	var total int64

	db := database.GORM_DB.Model(&Package{})
	if !unrestricted {
		if len(operatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			db = db.Where("packages.operator_id IN ?", operatorIDs)
		}
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where(
			"packages.name ILIKE ? OR packages.description ILIKE ? OR packages.package_type ILIKE ?",
			term, term, term,
		)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "packages.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (Package, error) {
	var item Package
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

func Insert(item *Package, files []fileresource.FileResource) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Operator", "FileResources").Create(item).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(item).Association("FileResources").Replace(files)
	})
}

func Save(item *Package, files []fileresource.FileResource) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Operator", "FileResources").Save(item).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(item).Association("FileResources").Replace(files)
	})
}

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&Package{}).Error
}
