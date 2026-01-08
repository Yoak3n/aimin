package implements

import "blood/schema"

func (d *Database) CreateDialogueRecord(r schema.DialogueRecord) error {
	db := d.GetPostgresSQL()
	return db.Create(&r).Error
}

func (d *Database) QueryDialogueRecord(id string) schema.DialogueRecord {
	var ret schema.DialogueRecord
	db := d.GetPostgresSQL()
	db.Where("id = ?", id).First(&ret)
	return ret
}

func (d *Database) QueryDialogueRecords(conversationId string) []schema.DialogueRecord {
	var ret []schema.DialogueRecord
	db := d.GetPostgresSQL()
	db.Where("conversation_id = ?", conversationId).Order("created_at asc").Find(&ret)
	return ret
}

func (d *Database) QueryDialogueRecordsByLinks(Link []string) []schema.DialogueRecord {
	ret := make([]schema.DialogueRecord, len(Link))
	db := d.GetPostgresSQL()
	db.Where("link IN (?)", Link).Find(&ret)
	return ret
}

func (d *Database) UpdateDialogueRecord(r schema.DialogueRecord) error {
	db := d.GetPostgresSQL()
	return db.Save(&r).Error
}
func (d *Database) UpdatedDialogueRecordLink(id string, link string) error {
	db := d.GetPostgresSQL()
	return db.Model(&schema.DialogueRecord{}).Where("id = ?", id).Update("link", link).Error
}

func (d *Database) UpdatedDialogueRecordsLinks(ids []string, link string) error {
	db := d.GetPostgresSQL()
	return db.Model(&schema.DialogueRecord{}).Where("id IN (?)", ids).Update("link", link).Error
}
