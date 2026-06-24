package database

import (
	"database/sql"

	"safari-quest-api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GORM_DB *gorm.DB
var SQL_DB *sql.DB
var DB_MIGRATOR gorm.Migrator

func ConnectToDatabase() error {
	db, err := gorm.Open(postgres.Open(config.App.DBString), &gorm.Config{})
	if err != nil {
		return err
	}
	GORM_DB = db
	SQL_DB, _ = db.DB()
	DB_MIGRATOR = db.Migrator()
	return nil
}
