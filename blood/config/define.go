package config

import (
	"sync"
)

var (
	conf           *Configuration
	once           sync.Once
	configFilePath string
)

type Configuration struct {
	LLMs        []LLMConfig      `json:"llms"`
	Workspace   *Workspace       `json:"workspace"`
	Database    *Database        `json:"database"`
	ActiveLLM   ActiveLLMConfig  `json:"active_llm"`
	DisabledLLM map[string]int64 `json:"disabled_llm"`
}
