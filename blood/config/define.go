package config

import (
	"encoding/json"
	"fmt"
)

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"dbname"`
	SSLMode   string `json:"sslmode"`
	TimeZone  string `json:"timezone"`
	Dimension int    `json:"dimension"`
}

const (
	LLMTypeChat      LLMType = "chat"
	LLMTypeEmbedding LLMType = "embedding"
)

type LLMType string

type Configuration struct {
	LLMs []LLMConfig `json:"llms"`
}

type LLMConfig struct {
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	APIKey   string  `json:"api_key"`
	APIUrl   string  `json:"api_url"`
	Type     LLMType `json:"type"`
}

func (l *LLMConfig) Info() (string, error) {
	ll := *l
	if ll.APIKey != "" {
		ll.APIKey = "****"
	}
	b, err := json.MarshalIndent(ll, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}
	return string(b), nil
}

func (d *DatabaseConfig) Info() (string, error) {
	dd := *d
	if dd.Password != "" {
		dd.Password = "****"
	}
	b, err := json.MarshalIndent(dd, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}
	return string(b), nil
}
func (l *LLMConfig) Tag() string {
	return "llm"
}

func (d *DatabaseConfig) Tag() string {
	return "database"
}
