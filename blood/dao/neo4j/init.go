package database

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"blood/schema"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type NeuroDB struct {
	conn neo4j.Driver
	port int
	ctx  context.Context
}

func NewNeuroDB(port int) *NeuroDB {
	return &NeuroDB{
		conn: ConnectNeuroDatabase(port),
		port: port,
		ctx:  context.Background(),
	}
}

func (n *NeuroDB) reconnectNeuroDatabase() *neo4j.Driver {
	dbUri := "neo4j://localhost:" + string(rune(n.port))
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "12345678", ""))
	if err != nil {
		panic(err)
	}
	n.conn = driver
	return &driver
}

func ConnectNeuroDatabase(port int) neo4j.Driver {
	dbUri := fmt.Sprintf("neo4j://localhost:%d", port)
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "12345678", ""))
	if err != nil {
		panic(err)
	}
	log.Println("connect neo4j successfully")
	return driver
}

var identRe = regexp.MustCompile(`^[\p{L}_][\p{L}\p{N}_]*$`)

func validIdent(s string) bool {
	return s != "" && identRe.MatchString(s)
}

func (n *NeuroDB) CreateNode(e []schema.EntityTable) error {
	if n.conn == nil {
		n.reconnectNeuroDatabase()
	}

	for _, chain := range e {
		subjectNode := schema.Node{
			Label: chain.Subject,
			Type:  chain.SubjectType,
		}
		objectNode := schema.Node{
			Label: chain.Object,
			Type:  chain.ObjectType,
		}
		edge := schema.Edge{
			Type: chain.Predicate,
			Attr: map[string]any{
				"link": chain.Link,
			},
		}
		if !validIdent(subjectNode.Type) || !validIdent(objectNode.Type) || !validIdent(edge.Type) {
			return fmt.Errorf("invalid label or relationship type: subject=%q object=%q rel=%q", subjectNode.Type, objectNode.Type, edge.Type)
		}
		query := fmt.Sprintf(
			"MERGE (subject:`%s` {name: $subject}) "+
				"MERGE (object:`%s` {name: $object}) "+
				"MERGE (subject)-[r:`%s`]->(object) "+
				"SET r.link = $link",
			subjectNode.Type, objectNode.Type, edge.Type,
		)
		params := map[string]any{
			"subject": subjectNode.Label,
			"object":  objectNode.Label,
			"link":    edge.Attr["link"],
		}
		res, err := neo4j.ExecuteQuery(n.ctx, n.conn, query, params, neo4j.EagerResultTransformer)
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(res.Summary)

	}
	return nil
}
