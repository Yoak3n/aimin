package database

import (
	"fmt"

	"gorm.io/gorm"
)

// ConversationTable 创建对话表,指定向量化的维度
func ConversationTable(db *gorm.DB, dim int) error {
	conversationTableTemplate := `CREATE TABLE IF NOT EXISTS conversation (
	id TEXT PRIMARY KEY, 
	question TEXT NOT NULL, 
	thoughts TEXT NOT NULL,
	answer TEXT NOT NULL,
	system TEXT,
	embedding VECTOR(%d),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);`
	// 向量化question 和 answer
	conversationTableSQL := fmt.Sprintf(conversationTableTemplate, dim)

	return db.Exec(conversationTableSQL).Error
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