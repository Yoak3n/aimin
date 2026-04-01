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

type LabelCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type RelTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type PatternCount struct {
	From  string `json:"from"`
	Rel   string `json:"rel"`
	To    string `json:"to"`
	Count int64  `json:"count"`
}

type PropCount struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type LabelProps struct {
	Label string      `json:"label"`
	Props []PropCount `json:"props"`
}

func (n *Neo4jDB) SampleNodeNamesByLabel(label string, limit int) ([]string, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	query := "MATCH (n) WHERE $label IN labels(n) AND n.name IS NOT NULL " +
		"RETURN n.name AS name ORDER BY id(n) ASC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"label": label, "limit": limit})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(res.Records))
	for _, r := range res.Records {
		v, _ := r.Get("name")
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

func (n *Neo4jDB) SampleTriples(limit int) ([]schema.Edge, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	query := "MATCH (a)-[r]->(b) RETURN a AS a, r AS r, b AS b ORDER BY id(r) ASC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
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

func (n *Neo4jDB) GetTopLabels(limit int) ([]LabelCount, error) {
	if limit <= 0 {
		limit = 30
	}
	query := "CALL db.labels() YIELD label " +
		"CALL { WITH label MATCH (n) WHERE label IN labels(n) RETURN count(n) AS c } " +
		"RETURN label AS label, c AS c ORDER BY c DESC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
	if err != nil {
		return nil, err
	}
	out := make([]LabelCount, 0, len(res.Records))
	for _, r := range res.Records {
		lv, _ := r.Get("label")
		cv, _ := r.Get("c")
		item := LabelCount{}
		if s, ok := lv.(string); ok {
			item.Label = s
		}
		switch v := cv.(type) {
		case int64:
			item.Count = v
		case int:
			item.Count = int64(v)
		}
		if item.Label != "" {
			out = append(out, item)
		}
	}
	return out, nil
}

func (n *Neo4jDB) GetTopRelationshipTypes(limit int) ([]RelTypeCount, error) {
	if limit <= 0 {
		limit = 30
	}
	query := "CALL db.relationshipTypes() YIELD relationshipType " +
		"CALL { WITH relationshipType MATCH ()-[r]->() WHERE type(r)=relationshipType RETURN count(r) AS c } " +
		"RETURN relationshipType AS t, c AS c ORDER BY c DESC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
	if err != nil {
		return nil, err
	}
	out := make([]RelTypeCount, 0, len(res.Records))
	for _, r := range res.Records {
		tv, _ := r.Get("t")
		cv, _ := r.Get("c")
		item := RelTypeCount{}
		if s, ok := tv.(string); ok {
			item.Type = s
		}
		switch v := cv.(type) {
		case int64:
			item.Count = v
		case int:
			item.Count = int64(v)
		}
		if item.Type != "" {
			out = append(out, item)
		}
	}
	return out, nil
}

func (n *Neo4jDB) GetTopPatterns(limit int) ([]PatternCount, error) {
	if limit <= 0 {
		limit = 30
	}
	query := "MATCH (a)-[r]->(b) " +
		"RETURN head(labels(a)) AS from, type(r) AS rel, head(labels(b)) AS to, count(*) AS c " +
		"ORDER BY c DESC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{"limit": limit})
	if err != nil {
		return nil, err
	}
	out := make([]PatternCount, 0, len(res.Records))
	for _, r := range res.Records {
		fv, _ := r.Get("from")
		rv, _ := r.Get("rel")
		tv, _ := r.Get("to")
		cv, _ := r.Get("c")
		item := PatternCount{}
		if s, ok := fv.(string); ok {
			item.From = s
		}
		if s, ok := rv.(string); ok {
			item.Rel = s
		}
		if s, ok := tv.(string); ok {
			item.To = s
		}
		switch v := cv.(type) {
		case int64:
			item.Count = v
		case int:
			item.Count = int64(v)
		}
		if item.From != "" && item.Rel != "" && item.To != "" {
			out = append(out, item)
		}
	}
	return out, nil
}

func (n *Neo4jDB) SampleTopPropsByLabel(label string, sample int, propLimit int) ([]PropCount, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("label is required")
	}
	if sample <= 0 {
		sample = 200
	}
	if propLimit <= 0 {
		propLimit = 5
	}
	if sample > 2000 {
		sample = 2000
	}
	if propLimit > 20 {
		propLimit = 20
	}
	query := "MATCH (n) WHERE $label IN labels(n) " +
		"WITH n LIMIT $sample " +
		"UNWIND keys(n) AS k " +
		"RETURN k AS k, count(*) AS c ORDER BY c DESC LIMIT $limit"
	res, err := n.eagerQuery(query, map[string]any{
		"label":  label,
		"sample": sample,
		"limit":  propLimit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]PropCount, 0, len(res.Records))
	for _, r := range res.Records {
		kv, _ := r.Get("k")
		cv, _ := r.Get("c")
		item := PropCount{}
		if s, ok := kv.(string); ok {
			item.Key = s
		}
		switch v := cv.(type) {
		case int64:
			item.Count = v
		case int:
			item.Count = int64(v)
		}
		if item.Key != "" {
			out = append(out, item)
		}
	}
	return out, nil
}

func (n *Neo4jDB) FindNeighborEdges(nodeType string, name string, relTypes []string, limit int) ([]schema.Edge, error) {
	nodeType = strings.TrimSpace(nodeType)
	name = normalizeNodeName(name)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 200 {
		limit = 200
	}

	filter := ""
	if len(relTypes) > 0 {
		clean := make([]string, 0, len(relTypes))
		for _, rt := range relTypes {
			rt = strings.TrimSpace(rt)
			if rt == "" {
				continue
			}
			if !validIdent(rt) {
				return nil, fmt.Errorf("invalid relType=%q", rt)
			}
			clean = append(clean, rt)
		}
		relTypes = clean
		if len(relTypes) > 0 {
			filter = " AND type(r) IN $relTypes "
		}
	}

	query := fmt.Sprintf(
		"MATCH (seed:`%s` {name:$name})-[r]-(n) "+
			"WHERE 1=1 %s "+
			"RETURN seed AS seed, r AS r, n AS n, (id(startNode(r))=id(seed)) AS out "+
			"ORDER BY coalesce(r.weight,0) DESC, id(r) ASC LIMIT $limit",
		nodeType, filter,
	)

	params := map[string]any{"name": name, "limit": limit}
	if len(relTypes) > 0 {
		params["relTypes"] = relTypes
	}
	res, err := n.eagerQuery(query, params)
	if err != nil {
		return nil, err
	}

	out := make([]schema.Edge, 0, len(res.Records))
	for _, rec := range res.Records {
		sv, okS := rec.Get("seed")
		rv, okR := rec.Get("r")
		nv, okN := rec.Get("n")
		ov, okO := rec.Get("out")
		if !okS || !okR || !okN || !okO {
			continue
		}
		sn, okS := sv.(neo4j.Node)
		rr, okR := rv.(neo4j.Relationship)
		on, okN := nv.(neo4j.Node)
		if !okS || !okR || !okN {
			continue
		}
		outDir, _ := ov.(bool)

		sName, _ := sn.Props["name"].(string)
		oName, _ := on.Props["name"].(string)
		sType := nodeType
		oType := ""
		if len(on.Labels) > 0 {
			oType = on.Labels[0]
		}

		sub := &schema.Node{Label: sName, Type: sType, Attr: sn.Props}
		obj := &schema.Node{Label: oName, Type: oType, Attr: on.Props}
		if !outDir {
			sub, obj = obj, sub
		}

		edge := schema.Edge{
			Type:    rr.Type,
			Attr:    map[string]any{},
			Subject: sub,
			Object:  obj,
		}
		for k, v := range rr.Props {
			edge.Attr[k] = v
		}
		out = append(out, edge)
	}
	return out, nil
}
