package implements

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

func (d *Database) CreateConversationRecord(r *schema.ConversationRecord, embedding []float32) error {
	embeddingStr := util.Float32SliceToString(embedding)
	res := d.GetPostgresSQL().Exec(`INSERT INTO conversation
		(id, question, thoughts, answer, system, embedding) 
		VALUES ($1, $2, $3, $4, $5, $6);`,
		r.Id, r.Question, r.Thoughts, r.Answer, r.System, embeddingStr)
	return res.Error
}

func (d *Database) GetReleventConversationRecords(embedding []float32, limit ...int) ([]schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	if len(limit) == 0 {
		limit = append(limit, 5)
	}
	records := make([]schema.ConversationRecord, 0)
	embeddingStr := util.Float32SliceToString(embedding)
	res := db.Raw(`SELECT * FROM conversation 
	ORDER BY embedding <-> $1::vector LIMIT $2`, embeddingStr, limit[0]).Scan(&records)
	return records, res.Error
}

func (d *Database) GetConversationByID(id string) (schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	var r schema.ConversationRecord
	err := db.Where("id = ?", id).First(&r).Error
	return r, err
}

func (d *Database) UpdateConversationOnlyTime(id string, update time.Time) error {
	db := d.GetPostgresSQL()
	res := db.Model(&schema.ConversationRecord{}).Where("id = ?", id).Update("updated_at", update)
	return res.Error
}

func (d *Database) GetAllConversations() ([]schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	var records []schema.ConversationRecord
	err := db.Order("updated_at desc").Find(&records).Error
	return records, err
}

func (d *Database) GetRecentConversations(limit int) ([]schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	var records []schema.ConversationRecord
	err := db.Order("updated_at desc").Limit(limit).Find(&records).Error
	return records, err
}
