package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		Database:  DefaultDatabaseConfig(),
	}
}

func NewConfiguration() *Configuration {
	cfg := DefaultConfiguration()
	path, err := findConfigFilePath()
	if err != nil {
		panic(err)
	}
	configFilePath = path
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		panic(err)
	}
	normalizeConfiguration(cfg)
	return cfg
}

func GlobalConfiguration() *Configuration {
	once.Do(func() {
		conf = NewConfiguration()
	})
	return conf
}

func (c *Configuration) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	path := configFilePath
	if path == "" {
		path = "config.json"
	}
	return os.WriteFile(path, data, 0644)
}

func normalizeConfiguration(cfg *Configuration) {
	if cfg == nil {
		return
	}
	if cfg.Workspace == nil {
		cfg.Workspace = DefaultWorkspace()
	} else {
		if cfg.Workspace.MemoryDays == 0 {
			cfg.Workspace.MemoryDays = DefaultWorkspace().MemoryDays
		}
		if cfg.Workspace.ContextSize == 0 {
			cfg.Workspace.ContextSize = DefaultWorkspace().ContextSize
		}
		if cfg.Workspace.FileContentSize == 0 {
			cfg.Workspace.FileContentSize = DefaultWorkspace().FileContentSize
		}
	}
	if cfg.Database == nil {
		cfg.Database = DefaultDatabaseConfig()
	} else {
		if cfg.Database.Host == "" {
			cfg.Database.Host = DefaultDatabaseConfig().Host
		}
		if cfg.Database.Port == 0 {
			cfg.Database.Port = DefaultDatabaseConfig().Port
		}
		if cfg.Database.User == "" {
			cfg.Database.User = DefaultDatabaseConfig().User
		}
		if cfg.Database.Password == "" {
			cfg.Database.Password = DefaultDatabaseConfig().Password
		}
		if cfg.Database.DBName == "" {
			cfg.Database.DBName = DefaultDatabaseConfig().DBName
		}
		if cfg.Database.SSLMode == "" {
			cfg.Database.SSLMode = DefaultDatabaseConfig().SSLMode
		}
		if cfg.Database.TimeZone == "" {
			cfg.Database.TimeZone = DefaultDatabaseConfig().TimeZone
		}
		if cfg.Database.Dimension == 0 {
			cfg.Database.Dimension = DefaultDatabaseConfig().Dimension
		}
	}
}

func findConfigFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "config.json")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}
