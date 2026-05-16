package database

import (
	"fmt"

	"gorm.io/gorm"
)

// ConversationTable 创建对话表,指定向量化的维度
func ConversationTable(db *gorm.DB) error {
	conversationTableTemplate := `CREATE TABLE IF NOT EXISTS conversation (
	id TEXT PRIMARY KEY, 
	question TEXT NOT NULL, 
	thoughts TEXT NOT NULL,
	answer TEXT NOT NULL,
	system TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);`

	return db.Exec(conversationTableTemplate).Error
}

func MetacognitionTable(db *gorm.DB) error {
	metacognitionTableSQL := `
		CREATE TABLE IF NOT EXISTS metacognition_table (
			id SERIAL PRIMARY KEY,
			insight_text TEXT,
			insight_type TEXT,
			confidence_score REAL,
			success_count INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);`
	return db.Exec(metacognitionTableSQL).Error
}

func ConversationTextChunkTable(db *gorm.DB, dim int) error {
	conversationTextChunkTableSQL := `
		CREATE TABLE IF NOT EXISTS conversation_text_chunk (
			id TEXT PRIMARY KEY,
			conversation_id TEXT,
			chunk_index INTEGER,
			content TEXT,
			embedding VECTOR(%d),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);`
	// 向量化content
	conversationTextChunkTableSQL = fmt.Sprintf(conversationTextChunkTableSQL, dim)

	return db.Exec(conversationTextChunkTableSQL).Error
}
