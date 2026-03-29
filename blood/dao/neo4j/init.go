package database

import (
	"context"
	"fmt"
	"log"

	"github.com/Yoak3n/aimin/blood/config"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Neo4jDB struct {
	conn neo4j.Driver
	ctx  context.Context
}

func NewNeuroDB() *Neo4jDB {
	db := &Neo4jDB{
		conn: ConnectNeuroDatabase(),
		ctx:  context.Background(),
	}
	if err := db.EnsureConstraints(); err != nil {
		log.Println(err)
	}
	return db
}

func (n *Neo4jDB) reconnectNeuroDatabase() *neo4j.Driver {
	driver := ConnectNeuroDatabase()
	n.conn = driver
	if err := n.EnsureConstraints(); err != nil {
		log.Println(err)
	}
	return &driver
}

func ConnectNeuroDatabase() neo4j.Driver {
	globalCfg := config.GlobalConfiguration()
	cfg := globalCfg.Database.Neo4j

	uri := cfg.URI
	if uri == "" {
		uri = fmt.Sprintf("neo4j://%s:%d", cfg.Host, cfg.Port)
	}

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		panic(err)
	}
	log.Println("connect neo4j successfully")
	return driver
}

func (n *Neo4jDB) EnsureConstraints() error {
	if n == nil {
		return fmt.Errorf("neo4j db is nil")
	}
	if n.conn == nil {
		n.reconnectNeuroDatabase()
	}
	_, err := n.eagerQuery(
		"CREATE CONSTRAINT entity_name_unique IF NOT EXISTS FOR (n:Entity) REQUIRE n.name IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("ensure neo4j constraints failed: %w", err)
	}
	return nil
}
