package database

import (
	"blood/schema"

	"gorm.io/gorm"
)

func InitDatabase(db *gorm.DB, dim int) error {
	if err := AddVectorExtension(db); err != nil {
		return err
	}
	if err := TemporaryMemoryTable(db); err != nil {
		return err
	}
	if err := EnduringMemoryTable(db, dim); err != nil {
		return err
	}
	if err := MetacognitionTable(db); err != nil {
		return err
	}
	return db.AutoMigrate(schema.DialogueRecord{}, schema.EntityTable{}, schema.MetacognitionLinkTable{})
}

func AddVectorExtension(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
}
