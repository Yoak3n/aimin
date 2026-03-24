package config

import (
	"encoding/json"
	"fmt"
)

const (
	LLMTypeChat      LLMType = "chat"
	LLMTypeEmbedding LLMType = "embedding"
)

type LLMType string

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

func (l *LLMConfig) Tag() string {
	return "llm"
}

type ActiveLLMConfig struct {
	ChatModel      string `json:"chat_model"`
	EmbeddingModel string `json:"embedding_model"`
}
