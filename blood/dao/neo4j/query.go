package database

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type NodeDegree struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Degree int64  `json:"degree"`
}

func (n *Neo4jDB) eagerQuery(query string, params map[string]any) (*neo4j.EagerResult, error) {
	if n.conn == nil {
		n.reconnectNeuroDatabase()
	}
	return neo4j.ExecuteQuery(n.ctx, n.conn, query, params, neo4j.EagerResultTransformer)
}

func (n *Neo4jDB) GetNode(nodeType string, name string) (*schema.Node, error) {
	nodeType = strings.TrimSpace(nodeType)
	name = normalizeNodeName(name)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}

	query := fmt.Sprintf("MATCH (n:`%s` {name: $name}) RETURN n LIMIT 1", nodeType)
	res, err := n.eagerQuery(query, map[string]any{"name": name})
	if err != nil {
		return nil, err
	}
	if len(res.Records) == 0 {
		return nil, nil
	}

	v, ok := res.Records[0].Get("n")
	if !ok {
		return nil, fmt.Errorf("neo4j response missing field n")
	}
	nn, ok := v.(neo4j.Node)
	if !ok {
		return nil, fmt.Errorf("neo4j response n has unexpected type %T", v)
	}

	out := &schema.Node{
		Label: name,
		Type:  nodeType,
		Attr:  nn.Props,
	}
	if out.Attr == nil {
		out.Attr = map[string]any{}
	}
	return out, nil
}

func (n *Neo4jDB) FindNodesByType(nodeType string, limit int) ([]schema.Node, error) {
	nodeType = strings.TrimSpace(nodeType)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf("MATCH (n:`%s`) RETURN n ORDER BY n.name ASC LIMIT $limit", nodeType)
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
	if err != nil {
		return nil, err
	}

	out := make([]schema.Node, 0, len(res.Records))
	for _, r := range res.Records {
		v, ok := r.Get("n")
		if !ok {
			continue
		}
		nn, ok := v.(neo4j.Node)
		if !ok {
			continue
		}
		name, _ := nn.Props["name"].(string)
		out = append(out, schema.Node{
			Label: name,
			Type:  nodeType,
			Attr:  nn.Props,
		})
	}
	return out, nil
}

func (n *Neo4jDB) FindNodesByNameContains(nodeType string, keyword string, limit int) ([]schema.Node, error) {
	nodeType = strings.TrimSpace(nodeType)
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit <= 0 {
		limit = 20
	}

	match := "MATCH (n)"
	if nodeType != "" {
		if !validIdent(nodeType) {
			return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
		}
		match = fmt.Sprintf("MATCH (n:`%s`)", nodeType)
	}

	query := match + " WHERE toLower(coalesce(n.name, '')) CONTAINS toLower($keyword) RETURN n ORDER BY n.name ASC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"keyword": keyword, "limit": limit})
	if err != nil {
		return nil, err
	}

	out := make([]schema.Node, 0, len(res.Records))
	for _, r := range res.Records {
		v, ok := r.Get("n")
		if !ok {
			continue
		}
		nn, ok := v.(neo4j.Node)
		if !ok {
			continue
		}
		name, _ := nn.Props["name"].(string)
		typ := nodeType
		if typ == "" && len(nn.Labels) > 0 {
			typ = nn.Labels[0]
		}
		out = append(out, schema.Node{
			Label: name,
			Type:  typ,
			Attr:  nn.Props,
		})
	}
	return out, nil
}

func (n *Neo4jDB) FindNeighbors(nodeType string, name string, relType string, limit int) ([]schema.Node, error) {
	nodeType = strings.TrimSpace(nodeType)
	name = normalizeNodeName(name)
	relType = strings.TrimSpace(relType)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}
	if relType != "" && !validIdent(relType) {
		return nil, fmt.Errorf("invalid relType=%q", relType)
	}
	if limit <= 0 {
		limit = 20
	}

	query := ""
	if relType != "" {
		query = fmt.Sprintf(
			"MATCH (n:`%s` {name: $name})-[:`%s`]-(m) "+
				"RETURN DISTINCT m AS m ORDER BY m.name ASC LIMIT $limit",
			nodeType, relType,
		)
	} else {
		query = fmt.Sprintf(
			"MATCH (n:`%s` {name: $name})--(m) "+
				"RETURN DISTINCT m AS m ORDER BY m.name ASC LIMIT $limit",
			nodeType,
		)
	}

	res, err := n.eagerQuery(query, map[string]any{"name": name, "limit": limit})
	if err != nil {
		return nil, err
	}

	out := make([]schema.Node, 0, len(res.Records))
	for _, r := range res.Records {
		v, ok := r.Get("m")
		if !ok {
			continue
		}
		nn, ok := v.(neo4j.Node)
		if !ok {
			continue
		}
		nm, _ := nn.Props["name"].(string)
		typ := ""
		if len(nn.Labels) > 0 {
			typ = nn.Labels[0]
		}
		out = append(out, schema.Node{
			Label: nm,
			Type:  typ,
			Attr:  nn.Props,
		})
	}
	return out, nil
}

func (n *Neo4jDB) FindRelationshipsByLink(link string, limit int) ([]schema.Edge, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil, fmt.Errorf("link is required")
	}
	if limit <= 0 {
		limit = 50
	}

	query := "MATCH (a)-[r {link: $link}]->(b) RETURN a AS a, r AS r, b AS b LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"link": link, "limit": limit})
	if err != nil {
		return nil, err
	}

	out := make([]schema.Edge, 0, len(res.Records))
	for _, rec := range res.Records {
		av, okA := rec.Get("a")
		rv, okR := rec.Get("r")
		bv, okB := rec.Get("b")
		if !okA || !okR || !okB {
			continue
		}
		an, okA := av.(neo4j.Node)
		rr, okR := rv.(neo4j.Relationship)
		bn, okB := bv.(neo4j.Node)
		if !okA || !okR || !okB {
			continue
		}

		aName, _ := an.Props["name"].(string)
		bName, _ := bn.Props["name"].(string)
		aType := ""
		bType := ""
		if len(an.Labels) > 0 {
			aType = an.Labels[0]
		}
		if len(bn.Labels) > 0 {
			bType = bn.Labels[0]
		}

		edge := schema.Edge{
			Type: rr.Type,
			Attr: map[string]any{},
			Subject: &schema.Node{
				Label: aName,
				Type:  aType,
				Attr:  an.Props,
			},
			Object: &schema.Node{
				Label: bName,
				Type:  bType,
				Attr:  bn.Props,
			},
		}
		for k, v := range rr.Props {
			edge.Attr[k] = v
		}
		out = append(out, edge)
	}
	return out, nil
}

func (n *Neo4jDB) FindLeastConnectedNodes(nodeType string, limit int) ([]NodeDegree, error) {
	nodeType = strings.TrimSpace(nodeType)
	if nodeType != "" && !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if limit <= 0 {
		limit = 10
	}

	match := "MATCH (n)"
	if nodeType != "" {
		match = fmt.Sprintf("MATCH (n:`%s`)", nodeType)
	}

	query := match + " WITH n, COUNT { (n)--() } AS degree RETURN head(labels(n)) AS type, n.name AS name, degree ORDER BY degree ASC, name ASC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
	if err != nil {
		return nil, err
	}

	out := make([]NodeDegree, 0, len(res.Records))
	for _, r := range res.Records {
		t, _ := r.Get("type")
		name, _ := r.Get("name")
		deg, _ := r.Get("degree")

		item := NodeDegree{}
		if ts, ok := t.(string); ok {
			item.Type = ts
		}
		if ns, ok := name.(string); ok {
			item.Name = ns
		}
		switch v := deg.(type) {
		case int64:
			item.Degree = v
		case int:
			item.Degree = int64(v)
		}
		out = append(out, item)
	}
	return out, nil
}
