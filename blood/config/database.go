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

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:      "localhost",
		Port:      5432,
		User:      "postgres",
		Password:  "123456",
		DBName:    "aimin",
		SSLMode:   "disable",
		TimeZone:  "Asia/Shanghai",
		Dimension: 2560,
	}
}
