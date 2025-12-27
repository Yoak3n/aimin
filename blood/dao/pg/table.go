package database

import (
	"fmt"

	"gorm.io/gorm"
)

func TemporaryMemoryTable(db *gorm.DB) error {
	temporaryMemoryTableSQL := `
		CREATE TABLE IF NOT EXISTS temporary_memory (
			id TEXT PRIMARY KEY,
			topic TEXT NOT NULL,
			content TEXT,
			count INTEGER DEFAULT 0,
			last_simulated_time TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		);`
	return db.Exec(temporaryMemoryTableSQL).Error
}

// EnduringMemoryTable 创建长期记忆表,指定向量化的维度
func EnduringMemoryTable(db *gorm.DB, dim int) error {
	enduringMemoryTableTemplate := `
CREATE TABLE IF NOT EXISTS enduring_memory (
	id TEXT PRIMARY KEY, 
	topic TEXT NOT NULL,
	content TEXT, 
	count INTEGER DEFAULT 0,
	last_simulated_time TIMESTAMP,
	embedding VECTOR(%d),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);`
	enduringMemoryTableSQL := fmt.Sprintf(enduringMemoryTableTemplate, dim)

	return db.Exec(enduringMemoryTableSQL).Error
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
