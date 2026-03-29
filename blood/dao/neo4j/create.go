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
		subjectName := normalizeNodeName(chain.Subject)
		objectName := normalizeNodeName(chain.Object)
		if subjectName == "" || objectName == "" {
			return fmt.Errorf("invalid node name: subject=%q object=%q", chain.Subject, chain.Object)
		}

		subjectType := chain.SubjectType
		objectType := chain.ObjectType
		edgeType := chain.Predicate
		if !validIdent(subjectType) || !validIdent(objectType) || !validIdent(edgeType) {
			return fmt.Errorf("invalid label or relationship type: subject=%q object=%q rel=%q", subjectType, objectType, edgeType)
		}

		link := strings.TrimSpace(chain.Link)

		// 按 name 查找已有节点（不限 label，找到就复用，不改其标签）
		// 找不到才 CREATE 并指定 LLM 生成的类型标签
		var query string
		if link != "" {
			query = fmt.Sprintf(
				"WITH $subject AS sName, $object AS oName "+
					"OPTIONAL MATCH (sExist {name: sName}) "+
					"WITH sName, oName, sExist ORDER BY id(sExist) ASC LIMIT 1 "+
					"CALL { "+
					"  WITH sExist, sName "+
					"  WITH sExist, sName WHERE sExist IS NOT NULL "+
					"  RETURN sExist AS s "+
					"  UNION "+
					"  WITH sExist, sName "+
					"  WITH sExist, sName WHERE sExist IS NULL "+
					"  CREATE (s:`%s` {name: sName}) "+
					"  RETURN s "+
					"} "+
					"WITH s, oName "+
					"OPTIONAL MATCH (oExist {name: oName}) "+
					"WITH s, oName, oExist ORDER BY id(oExist) ASC LIMIT 1 "+
					"CALL { "+
					"  WITH oExist, oName "+
					"  WITH oExist, oName WHERE oExist IS NOT NULL "+
					"  RETURN oExist AS o "+
					"  UNION "+
					"  WITH oExist, oName "+
					"  WITH oExist, oName WHERE oExist IS NULL "+
					"  CREATE (o:`%s` {name: oName}) "+
					"  RETURN o "+
					"} "+
					"MERGE (s)-[r:`%s` {link: $link}]->(o)",
				subjectType, objectType, edgeType,
			)
		} else {
			query = fmt.Sprintf(
				"WITH $subject AS sName, $object AS oName "+
					"OPTIONAL MATCH (sExist {name: sName}) "+
					"WITH sName, oName, sExist ORDER BY id(sExist) ASC LIMIT 1 "+
					"CALL { "+
					"  WITH sExist, sName "+
					"  WITH sExist, sName WHERE sExist IS NOT NULL "+
					"  RETURN sExist AS s "+
					"  UNION "+
					"  WITH sExist, sName "+
					"  WITH sExist, sName WHERE sExist IS NULL "+
					"  CREATE (s:`%s` {name: sName}) "+
					"  RETURN s "+
					"} "+
					"WITH s, oName "+
					"OPTIONAL MATCH (oExist {name: oName}) "+
					"WITH s, oName, oExist ORDER BY id(oExist) ASC LIMIT 1 "+
					"CALL { "+
					"  WITH oExist, oName "+
					"  WITH oExist, oName WHERE oExist IS NOT NULL "+
					"  RETURN oExist AS o "+
					"  UNION "+
					"  WITH oExist, oName "+
					"  WITH oExist, oName WHERE oExist IS NULL "+
					"  CREATE (o:`%s` {name: oName}) "+
					"  RETURN o "+
					"} "+
					"MERGE (s)-[r:`%s`]->(o)",
				subjectType, objectType, edgeType,
			)
		}

		params := map[string]any{
			"subject": subjectName,
			"object":  objectName,
			"link":    link,
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
