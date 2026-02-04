package implements

import (
	"time"

	"github.com/Yoak3n/aimin/blood/schema"
)

func (d *Database) CreateConversationRecord(r schema.ConversationRecord) error {
	db := d.GetPostgresSQL()
	return db.Create(&r).Error
}

func (d *Database) GetConversationRecord(id string) (schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	var r schema.ConversationRecord
	err := db.Where("id = ?", id).First(&r).Error
	return r, err
}

func (d *Database) GetAllConversations() ([]schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	var records []schema.ConversationRecord
	err := db.Order("updated_at desc").Find(&records).Error
	return records, err
}

func (d *Database) UpdateConversationOnlyTime(id string, update time.Time) {
	db := d.GetPostgresSQL()
	db.Model(&schema.ConversationRecord{}).Where("id = ?", id).Update("updated_at", update)
}
