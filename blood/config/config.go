package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type OptionInterface interface {
	Info() (string, error)
	Tag() string
}

func WithDatabaseConfig(config *Database) OptionInterface {
	return config
}

func WithLLMConfig(config *LLMConfig) OptionInterface {
	return config
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		Workspace: DefaultWorkspace(),
		Database:  DefaultDatabase(),
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
		conf.Save()
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
	if cfg.DisabledLLM == nil {
		cfg.DisabledLLM = map[string]int64{}
	} else {
		now := time.Now().Unix()
		for k, until := range cfg.DisabledLLM {
			if until <= now {
				delete(cfg.DisabledLLM, k)
			}
		}
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
		cfg.Database = DefaultDatabase()
	}
	if cfg.Database.Postgres == nil {
		cfg.Database.Postgres = DefaultDatabaseConfig()
	} else {
		if cfg.Database.Postgres.Host == "" {
			cfg.Database.Postgres.Host = DefaultDatabaseConfig().Host
		}
		if cfg.Database.Postgres.Port == 0 {
			cfg.Database.Postgres.Port = DefaultDatabaseConfig().Port
		}
		if cfg.Database.Postgres.User == "" {
			cfg.Database.Postgres.User = DefaultDatabaseConfig().User
		}
		if cfg.Database.Postgres.Password == "" {
			cfg.Database.Postgres.Password = DefaultDatabaseConfig().Password
		}
		if cfg.Database.Postgres.DBName == "" {
			cfg.Database.Postgres.DBName = DefaultDatabaseConfig().DBName
		}
		if cfg.Database.Postgres.SSLMode == "" {
			cfg.Database.Postgres.SSLMode = DefaultDatabaseConfig().SSLMode
		}
		if cfg.Database.Postgres.TimeZone == "" {
			cfg.Database.Postgres.TimeZone = DefaultDatabaseConfig().TimeZone
		}
		if cfg.Database.Postgres.Dimension == 0 {
			cfg.Database.Postgres.Dimension = DefaultDatabaseConfig().Dimension
		}
	}

	if cfg.Database.Neo4j == nil {
		cfg.Database.Neo4j = DefaultNeo4jConfig()
	} else {
		if cfg.Database.Neo4j.URI == "" {
			cfg.Database.Neo4j.URI = DefaultNeo4jConfig().URI
		}
		if cfg.Database.Neo4j.Host == "" {
			cfg.Database.Neo4j.Host = DefaultNeo4jConfig().Host
		}
		if cfg.Database.Neo4j.Port == 0 {
			cfg.Database.Neo4j.Port = DefaultNeo4jConfig().Port
		}
		if cfg.Database.Neo4j.User == "" {
			cfg.Database.Neo4j.User = DefaultNeo4jConfig().User
		}
		if cfg.Database.Neo4j.Password == "" {
			cfg.Database.Neo4j.Password = DefaultNeo4jConfig().Password
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
