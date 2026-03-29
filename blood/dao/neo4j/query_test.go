package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

func formatNodes(nodes []schema.Node) string {
	if len(nodes) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(nodes))
	for _, n := range nodes {
		parts = append(parts, fmt.Sprintf("%s(%s)", n.Label, n.Type))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatEdges(edges []schema.Edge) string {
	if len(edges) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(edges))
	for _, e := range edges {
		s := "?"
		o := "?"
		if e.Subject != nil {
			s = fmt.Sprintf("%s(%s)", e.Subject.Label, e.Subject.Type)
		}
		if e.Object != nil {
			o = fmt.Sprintf("%s(%s)", e.Object.Label, e.Object.Type)
		}
		link := ""
		if e.Attr != nil {
			if v, ok := e.Attr["link"]; ok {
				if ls, ok := v.(string); ok && strings.TrimSpace(ls) != "" {
					link = " link=" + ls
				}
			}
		}
		parts = append(parts, fmt.Sprintf("%s -[%s%s]-> %s", s, e.Type, link, o))
	}
	return "[" + strings.Join(parts, "; ") + "]"
}

func loadConfigForTest(t *testing.T) (cfg *config.Configuration, err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("load config panic: %v", r)
			cfg = nil
		}
	}()
	cfg = config.NewConfiguration()
	return cfg, nil
}

func newTestNeo4jDB(t *testing.T) (*Neo4jDB, func()) {
	t.Helper()

	uri := strings.TrimSpace(os.Getenv("NEO4J_URI"))
	user := strings.TrimSpace(os.Getenv("NEO4J_USER"))
	pass := os.Getenv("NEO4J_PASSWORD")

	if uri == "" || user == "" {
		cfg, err := loadConfigForTest(t)
		if err != nil || cfg == nil || cfg.Database == nil || cfg.Database.Neo4j == nil {
			t.Skip("需要设置环境变量 NEO4J_URI / NEO4J_USER / NEO4J_PASSWORD，或在 config.json 中配置 database.neo4j 才能运行 Neo4j 集成测试")
		}
		n4 := cfg.Database.Neo4j
		uri = strings.TrimSpace(n4.URI)
		if uri == "" {
			uri = fmt.Sprintf("neo4j://%s:%d", n4.Host, n4.Port)
		}
		user = strings.TrimSpace(n4.User)
		pass = n4.Password
		if uri == "" || user == "" {
			t.Skip("Neo4j 集成测试缺少连接信息：请检查 config.json 的 database.neo4j.uri/host/port/user")
		}
	}

	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(user, pass, ""))
	if err != nil {
		t.Skipf("neo4j driver 初始化失败: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	db := &Neo4jDB{conn: driver, ctx: ctx}

	_, err = db.eagerQuery("RETURN 1 AS ok", nil)
	if err != nil {
		cancel()
		_ = driver.Close(ctx)
		t.Skipf("neo4j 连接不可用: %v", err)
	}

	cleanup := func() {
		cancel()
		_ = driver.Close(ctx)
	}
	return db, cleanup
}

func TestNeo4jQueries_Basic(t *testing.T) {
	db, closeDB := newTestNeo4jDB(t)
	defer closeDB()

	prefix := "aimin_test_" + strings.ReplaceAll(t.Name(), "/", "_") + "_" + time.Now().Format("20060102150405")

	subjectA := prefix + "_Alice"
	subjectB := prefix + "_Bob"
	subjectC := prefix + "_Charlie"
	company := prefix + "_Acme"
	city := prefix + "_Shanghai"

	defer func() {
		_, _ = db.eagerQuery("MATCH (n) WHERE n.name STARTS WITH $prefix DETACH DELETE n", map[string]any{"prefix": prefix})
	}()

	chains := []schema.EntityTable{
		{
			Subject:     subjectA,
			SubjectType: "Person",
			Predicate:   "KNOWS",
			Object:      subjectB,
			ObjectType:  "Person",
			Link:        prefix + "_link_1",
		},
		{
			Subject:     subjectA,
			SubjectType: "Person",
			Predicate:   "WORKS_AT",
			Object:      company,
			ObjectType:  "Company",
			Link:        prefix + "_link_1",
		},
		{
			Subject:     company,
			SubjectType: "Company",
			Predicate:   "LOCATED_IN",
			Object:      city,
			ObjectType:  "City",
			Link:        prefix + "_link_2",
		},
		{
			Subject:     subjectC,
			SubjectType: "Person",
			Predicate:   "KNOWS",
			Object:      subjectA,
			ObjectType:  "Person",
			Link:        prefix + "_link_2",
		},
	}

	if err := db.CreateNode(chains); err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}

	gotAlice, err := db.GetNode("Person", subjectA)
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	t.Logf("GetNode(Person, %q) => %#v", subjectA, gotAlice)
	if gotAlice == nil || gotAlice.Label != subjectA || gotAlice.Type != "Person" {
		t.Fatalf("GetNode unexpected result: %#v", gotAlice)
	}

	neighbors, err := db.FindNeighbors("Person", subjectA, "KNOWS", 10)
	if err != nil {
		t.Fatalf("FindNeighbors failed: %v", err)
	}
	t.Logf("FindNeighbors(Person, %q, KNOWS) => %s", subjectA, formatNodes(neighbors))
	neighborNames := make(map[string]bool, len(neighbors))
	for _, n := range neighbors {
		neighborNames[n.Label] = true
	}
	if !neighborNames[subjectB] || !neighborNames[subjectC] {
		t.Fatalf("FindNeighbors missing expected nodes, got=%v", neighborNames)
	}

	edges1, err := db.FindRelationshipsByLink(prefix+"_link_1", 10)
	if err != nil {
		t.Fatalf("FindRelationshipsByLink failed: %v", err)
	}
	t.Logf("FindRelationshipsByLink(%q) => %s", prefix+"_link_1", formatEdges(edges1))
	if len(edges1) != 2 {
		t.Fatalf("FindRelationshipsByLink unexpected count: got=%d want=2", len(edges1))
	}

	leastCity, err := db.FindLeastConnectedNodes("City", 5)
	if err != nil {
		t.Fatalf("FindLeastConnectedNodes failed: %v", err)
	}
	t.Logf("FindLeastConnectedNodes(City) => %#v", leastCity)
	foundCity := false
	for _, n := range leastCity {
		if n.Name == city {
			foundCity = true
			break
		}
	}
	if !foundCity {
		t.Fatalf("FindLeastConnectedNodes missing expected node %q, got=%#v", city, leastCity)
	}

	contains, err := db.FindNodesByNameContains("", prefix+"_Ac", 10)
	if err != nil {
		t.Fatalf("FindNodesByNameContains failed: %v", err)
	}
	t.Logf("FindNodesByNameContains(keyword=%q) => %s", prefix+"_Ac", formatNodes(contains))
	foundCompany := false
	for _, n := range contains {
		if n.Label == company {
			foundCompany = true
			break
		}
	}
	if !foundCompany {
		t.Fatalf("FindNodesByNameContains did not return expected node, got=%v", contains)
	}
}
