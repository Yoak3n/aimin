package config

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
