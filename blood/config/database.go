package config

import (
	"encoding/json"
	"fmt"
)

// PostgresConfig 数据库配置结构体
type PostgresConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"dbname"`
	SSLMode   string `json:"sslmode"`
	TimeZone  string `json:"timezone"`
	Dimension int    `json:"dimension"`
}

type Neo4jConfig struct {
	URI      string `json:"uri"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Database struct {
	Postgres *PostgresConfig `json:"postgres"`
	Neo4j    *Neo4jConfig    `json:"neo4j"`
}

func (d *Database) Info() (string, error) {
	if d == nil {
		return "null", nil
	}
	out := &Database{
		Postgres: nil,
		Neo4j:    nil,
	}
	if d.Postgres != nil {
		pg := *d.Postgres
		if pg.Password != "" {
			pg.Password = "****"
		}
		out.Postgres = &pg
	}
	if d.Neo4j != nil {
		n4 := *d.Neo4j
		if n4.Password != "" {
			n4.Password = "****"
		}
		out.Neo4j = &n4
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}
	return string(b), nil
}

func (d *Database) Tag() string {
	return "database"
}

func (d *Database) UnmarshalJSON(data []byte) error {
	type v2 struct {
		Postgres *PostgresConfig `json:"postgres"`
		Neo4j    *Neo4jConfig    `json:"neo4j"`
	}
	var candidateV2 v2
	if err := json.Unmarshal(data, &candidateV2); err == nil {
		if candidateV2.Postgres != nil || candidateV2.Neo4j != nil {
			d.Postgres = candidateV2.Postgres
			d.Neo4j = candidateV2.Neo4j
			return nil
		}
	}

	var candidateV1 PostgresConfig
	if err := json.Unmarshal(data, &candidateV1); err == nil {
		if candidateV1.Host != "" ||
			candidateV1.Port != 0 ||
			candidateV1.User != "" ||
			candidateV1.Password != "" ||
			candidateV1.DBName != "" ||
			candidateV1.SSLMode != "" ||
			candidateV1.TimeZone != "" ||
			candidateV1.Dimension != 0 {
			d.Postgres = &candidateV1
		}
	}
	return nil
}

func (d *PostgresConfig) Info() (string, error) {
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

func (d *PostgresConfig) Tag() string {
	return "database"
}

func (n *Neo4jConfig) Info() (string, error) {
	nn := *n
	if nn.Password != "" {
		nn.Password = "****"
	}
	b, err := json.MarshalIndent(nn, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal error: %v", err)
	}
	return string(b), nil
}

func (n *Neo4jConfig) Tag() string {
	return "neo4j"
}

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *PostgresConfig {
	return &PostgresConfig{
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

func DefaultNeo4jConfig() *Neo4jConfig {
	return &Neo4jConfig{
		URI:      "",
		Host:     "localhost",
		Port:     7687,
		User:     "neo4j",
		Password: "12345678",
	}
}

func DefaultDatabase() *Database {
	return &Database{
		Postgres: DefaultDatabaseConfig(),
		Neo4j:    DefaultNeo4jConfig(),
	}
}
