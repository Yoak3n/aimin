package config

type TTSConfig struct {
	APIKey string `json:"api_key"`
	APIUrl string `json:"api_url"`
	Model  string `json:"model"`
	Voice  string `json:"voice"`
}

func DefaultTTSConfig() *TTSConfig {
	return &TTSConfig{
		APIKey: "",
		APIUrl: "https://api.xiaomimimo.com/v1",
		Model:  "mimo-v2.5-tts",
		Voice:  "mimo_default",
	}
}
