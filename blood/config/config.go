package config

import (
	"encoding/json"
	"os"
)

type OptionInterface interface {
	Info() (string, error)
	Tag() string
}

func WithDatabaseConfig(config *DatabaseConfig) OptionInterface {
	return config
}

func WithLLMConfig(config *LLMConfig) OptionInterface {
	return config
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		Workspace: DefaultWorkspace(),
	}
}

func NewConfiguration() *Configuration {
	cfg := DefaultConfiguration()
	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func GlobalConfiguration() *Configuration {
	once.Do(func() {
		conf = NewConfiguration()
	})
	return conf
}
