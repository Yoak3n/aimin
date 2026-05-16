package implements

import (
	"fmt"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

func (d *Database) CreateConversationRecord(r *schema.ConversationRecord) error {
	res := d.GetPostgresSQL().Exec(`INSERT INTO conversation
		(id, question, thoughts, answer, system, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		r.Id, r.Question, r.Thoughts, r.Answer, r.System, r.CreateAt, r.UpdatedAt)
	return res.Error
}

func (d *Database) CreateConversationTextChunks(chunks []schema.ConversationTextChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	db := d.GetPostgresSQL()

	const batchSize = 100
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		sql := `INSERT INTO conversation_text_chunk 
            (id, conversation_id, chunk_index, content, embedding, created_at, updated_at) VALUES `
		args := []any{}
		for j, chunk := range batch {
			if j > 0 {
				sql += ","
			}
			embeddingStr := util.Float32SliceToString(chunk.Embedding)
			base := j * 7
			sql += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7)
			args = append(args, chunk.Id, chunk.ConversationId, chunk.ChunkIndex,
				chunk.Content, embeddingStr, chunk.CreatedAt, chunk.UpdatedAt)
		}
		sql += ";"
		if err := db.Exec(sql, args...).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) GetReleventConversationRecords(embedding []float32, limit ...int) ([]schema.ConversationRecord, error) {
	db := d.GetPostgresSQL()
	if len(limit) == 0 {
		limit = append(limit, 5)
	}

	embeddingStr := util.Float32SliceToString(embedding)

	var conversationIds []string
	res := db.Raw(`SELECT DISTINCT conversation_id FROM (
		SELECT conversation_id FROM conversation_text_chunk
		ORDER BY embedding <-> $1::vector LIMIT $2
	) AS nearest`, embeddingStr, limit[0]).Scan(&conversationIds)
	if res.Error != nil {
		return nil, res.Error
	}

	records := make([]schema.ConversationRecord, 0)
	if len(conversationIds) == 0 {
		return records, nil
	}

	res = db.Where("id IN ?", conversationIds).Find(&records)
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
