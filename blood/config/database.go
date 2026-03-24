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

func (d *DatabaseConfig) Tag() string {
	return "database"
}
