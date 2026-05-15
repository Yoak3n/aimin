package database

import (
	"github.com/Yoak3n/aimin/blood/schema"

	"gorm.io/gorm"
)

func InitDatabase(db *gorm.DB, dim int) error {
	if err := AddVectorExtension(db); err != nil {
		return err
	}
	if err := ConversationTable(db, dim); err != nil {
		return err
	}
	if err := MetacognitionTable(db); err != nil {
		return err
	}
	return db.AutoMigrate(schema.SummaryMemoryTable{}, schema.EntityTable{}, schema.MetacognitionLinkTable{})
}

func AddVectorExtension(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error
}
