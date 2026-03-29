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
	return &Neo4jDB{
		conn: ConnectNeuroDatabase(),
		ctx:  context.Background(),
	}
}

func (n *Neo4jDB) reconnectNeuroDatabase() *neo4j.Driver {
	driver := ConnectNeuroDatabase()
	n.conn = driver
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
