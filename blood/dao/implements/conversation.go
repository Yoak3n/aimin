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

func (d *Database) UpdateConversationOnlyTime(id string, update time.Time) {
	db := d.GetPostgresSQL()
	db.Model(&schema.ConversationRecord{}).Where("id = ?", id).Update("update_time", update)
}

func (d *Database) UpdateConversationRecord(r schema.ConversationRecord) error {
	db := d.GetPostgresSQL()
	return db.Save(&r).Error
}

func (d *Database) DeleteConversationRecord(id string) {
	db := d.GetPostgresSQL()
	db.Where("id = ?", id).Delete(&schema.ConversationRecord{})
}
