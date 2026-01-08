package implements

import (
	"blood/schema"
	"time"
)

func (d *Database) CreateConversationRecord(r schema.ConversationRecord) error {
	db := d.GetPostgresSQL()
	return db.Create(&r).Error
}

func (d *Database) GetConversationRecord(id string) schema.ConversationRecord {
	db := d.GetPostgresSQL()
	var r schema.ConversationRecord
	db.Where("id = ?", id).First(&r)
	return r
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
