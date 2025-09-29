package db

import (
	"fmt"

	"gorm.io/gorm"
)

var gdb *gorm.DB

func InitDatabase(dsn string) error {
	db, err := OpenPostgreSQL(dsn)
	if err != nil {
		return err
	}
	if err := migrateDB(db); err != nil {
		return fmt.Errorf("migrate database failed: %w", err)
	}
	gdb = db
	return nil
}

func Close() error {
	idb, err := gdb.DB()
	if err != nil {
		return err
	}
	return idb.Close()
}

func migrateDB(db *gorm.DB) error {
	err := db.AutoMigrate()
	if err != nil {
		return err
	}

	return nil
}
