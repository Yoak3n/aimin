package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

var identRe = regexp.MustCompile(`^[\p{L}_][\p{L}\p{N}_]*$`)

func validIdent(s string) bool {
	return s != "" && identRe.MatchString(s)
}

func normalizeNodeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}

func (n *Neo4jDB) CreateNode(e []schema.EntityTable) error {
	if n.conn == nil {
		n.reconnectNeuroDatabase()
	}

	for _, chain := range e {
		subjectLabel := normalizeNodeName(chain.Subject)
		objectLabel := normalizeNodeName(chain.Object)
		if subjectLabel == "" || objectLabel == "" {
			return fmt.Errorf("invalid node name: subject=%q object=%q", chain.Subject, chain.Object)
		}

		subjectNode := schema.Node{
			Label: subjectLabel,
			Type:  chain.SubjectType,
		}
		objectNode := schema.Node{
			Label: objectLabel,
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

		query := ""
		if link, _ := edge.Attr["link"].(string); strings.TrimSpace(link) != "" {
			query = fmt.Sprintf(
				"MERGE (subject:`%s` {name: $subject}) "+
					"MERGE (object:`%s` {name: $object}) "+
					"MERGE (subject)-[r:`%s` {link: $link}]->(object)",
				subjectNode.Type, objectNode.Type, edge.Type,
			)
		} else {
			query = fmt.Sprintf(
				"MERGE (subject:`%s` {name: $subject}) "+
					"MERGE (object:`%s` {name: $object}) "+
					"MERGE (subject)-[r:`%s`]->(object)",
				subjectNode.Type, objectNode.Type, edge.Type,
			)
		}
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
