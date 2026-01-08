package schema

import (
	"time"

	"gorm.io/gorm"
)

type ConversationRecord struct {
	Id           string `json:"id" gorm:"primary key"`
	Topic        string `json:"topic"`
	From         string `json:"from"`
	SystemPrompt string `json:"system_prompt"`
	CreateAt     time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type DialogueRecord struct {
	Id             string `json:"id" gorm:"primary key"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	Link           string
	ConversationId string `json:"conversation_id"`
	CreateAt       time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type TemporaryMemoryTable struct {
	Id                string    `gorm:"primary key"`
	Topic             string    `gorm:"topic"`
	Count             int       `gorm:"count"`
	Content           string    `gorm:"content"`
	LastSimulatedTime time.Time `gorm:"last_simulate_time"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
}

type EnduringMemoryTable struct {
	Id                string    `gorm:"primary key"`
	Topic             string    `json:"topic"`
	Content           string    `json:"content"`
	Count             int       `json:"count"`
	LastSimulatedTime time.Time `json:"last_simulate_time"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index"`
	// 嵌入向量  仅用作数据库比较，不需要读取
	// Embedding []float32 `gorm:"serializer:json"`
}

type EntityTable struct {
	gorm.Model
	Subject     string `json:"subject"`
	SubjectType string `json:"subject_type"`
	Predicate   string `json:"predicate"`
	Object      string `json:"object"`
	ObjectType  string `json:"object_type"`
	// Link is the link to the enduring memory.
	Link string `json:"link"`
}

// MetacognitionTable 元认知表
// 使用场景：
// 1. 当实体A和实体B之间的关系发生变化时，实体A可以通过元认知表记录下这个变化。
// 2. 当实体A需要进行决策时，实体A可以通过元认知表查询到之前记录的变化，从而进行决策。
// 3. 当实体A的决策成功时，实体A可以通过元认知表更新成功次数。
type MetacognitionTable struct {
	gorm.Model
	InsightText     string  `json:"insight_text"`
	InsightType     string  `json:"insight_type"`
	ConfidenceScore float64 `json:"confidence_score"`
	SuccessCount    int     `json:"success_count"`
}

// MetacognitionLinkTable 元认知链接表 用于存储元认知与持久记忆之间的关系
// 使用场景：
// 1. 当实体A和实体B之间的关系发生变化时，实体A可以通过元认知表记录下这个变化。
// 2. 当实体A需要进行决策时，实体A可以通过元认知表查询到之前记录的变化，从而进行决策。
// 3. 当实体A的决策成功时，实体A可以通过元认知表更新成功次数。
type MetacognitionLinkTable struct {
	gorm.Model
	MetacognitionId  uint    `json:"metacognition_id"`
	EnduringMemoryID string  `json:"enduring_memory_id"`
	Weight           float64 `json:"weight"`
}
