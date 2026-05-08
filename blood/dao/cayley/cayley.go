package cayleydb

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/dao/graph"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/cayleygraph/cayley"
	cgraph "github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/cayleygraph/cayley/quad"
)

var identRe = regexp.MustCompile(`^[\p{L}_][\p{L}\p{N}_]*$`)

type CayleyDB struct {
	store *cayley.Handle
	mu    sync.RWMutex
}

func NewNeuroDB() *CayleyDB {
	cfg := config.GlobalConfiguration().Database.Cayley
	if cfg == nil {
		cfg = config.DefaultCayleyConfig()
	}
	backend := strings.ToLower(strings.TrimSpace(cfg.Backend))
	if backend == "" {
		backend = "bolt"
	}
	if backend == "memstore" || backend == "memory" {
		store, err := cayley.NewMemoryGraph()
		if err != nil {
			panic(err)
		}
		return &CayleyDB{store: store}
	}

	dbPath := strings.TrimSpace(cfg.Path)
	dbPath = resolveDefaultDBPath(backend, dbPath)

	store, err := openPersistentGraph(backend, dbPath)
	if err != nil {
		log.Println("open cayley persistent store failed:", err)
		mem, memErr := cayley.NewMemoryGraph()
		if memErr != nil {
			panic(memErr)
		}
		return &CayleyDB{store: mem}
	}
	return &CayleyDB{store: store}
}

func (c *CayleyDB) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c == nil || c.store == nil {
		return nil
	}
	err := c.store.Close()
	c.store = nil
	return err
}

func resolveDefaultDBPath(backend string, configured string) string {
	base := configured
	if strings.TrimSpace(base) == "" {
		exe, err := os.Executable()
		if err != nil || strings.TrimSpace(exe) == "" {
			exe = "."
		}
		root := filepath.Dir(exe)
		cacheDir := filepath.Join(root, "data", "cache")
		if backend == "bolt" {
			base = filepath.Join(cacheDir, "cayley.bolt")
		} else {
			base = filepath.Join(cacheDir, "cayley")
		}
	}

	if backend == "bolt" {
		if st, err := os.Stat(base); err == nil && st.IsDir() {
			base = filepath.Join(base, "cayley.bolt")
		}
		_ = os.MkdirAll(filepath.Dir(base), 0755)
		return base
	}

	_ = os.MkdirAll(base, 0755)
	return base
}

func openPersistentGraph(backend string, dbPath string) (*cayley.Handle, error) {
	initErr := cgraph.InitQuadStore(backend, dbPath, nil)
	if initErr != nil {
		msg := strings.ToLower(initErr.Error())
		if !strings.Contains(msg, "exists") && !strings.Contains(msg, "already") {
			return nil, initErr
		}
	}
	return cayley.NewGraph(backend, dbPath, nil)
}

func (c *CayleyDB) CreateNode(e []schema.EntityTable) error {
	if c == nil || c.store == nil {
		return fmt.Errorf("cayley db is nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	tx := cayley.NewTransaction()
	for _, chain := range e {
		subjectName := normalizeNodeName(chain.Subject)
		objectName := normalizeNodeName(chain.Object)
		if subjectName == "" || objectName == "" {
			return fmt.Errorf("invalid node name: subject=%q object=%q", chain.Subject, chain.Object)
		}

		subjectType := strings.TrimSpace(chain.SubjectType)
		objectType := strings.TrimSpace(chain.ObjectType)
		edgeType := strings.TrimSpace(chain.Predicate)
		if !validIdent(subjectType) || !validIdent(objectType) || !validIdent(edgeType) {
			return fmt.Errorf("invalid label or relationship type: subject=%q object=%q rel=%q", subjectType, objectType, edgeType)
		}

		if err := c.ensureNodeTypeLocked(subjectName, subjectType, tx); err != nil {
			return err
		}
		if err := c.ensureNodeTypeLocked(objectName, objectType, tx); err != nil {
			return err
		}

		link := strings.TrimSpace(chain.Link)
		var lbl quad.Value
		if link != "" {
			lbl = quad.String(link)
		}
		tx.AddQuad(quad.Make(quad.String(subjectName), quad.String(edgeType), quad.String(objectName), lbl))
	}
	return c.store.ApplyTransaction(tx)
}

func (c *CayleyDB) GetNode(nodeType string, name string) (*schema.Node, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	nodeType = strings.TrimSpace(nodeType)
	name = normalizeNodeName(name)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok, err := c.getNodeTypeLocked(name)
	if err != nil {
		return nil, err
	}
	if !ok || t != nodeType {
		return nil, nil
	}

	out := &schema.Node{
		Label: name,
		Type:  nodeType,
		Attr:  map[string]any{"name": name},
	}
	return out, nil
}

func (c *CayleyDB) FindNodesByType(nodeType string, limit int) ([]schema.Node, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	nodeType = strings.TrimSpace(nodeType)
	if !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if limit <= 0 {
		limit = 20
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	names, err := c.subjectsByTypeLocked(nodeType)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if len(names) > limit {
		names = names[:limit]
	}

	out := make([]schema.Node, 0, len(names))
	for _, n := range names {
		out = append(out, schema.Node{
			Label: n,
			Type:  nodeType,
			Attr:  map[string]any{"name": n},
		})
	}
	return out, nil
}

func (c *CayleyDB) FindNodesByNameContains(nodeType string, keyword string, limit int) ([]schema.Node, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	nodeType = strings.TrimSpace(nodeType)
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit <= 0 {
		limit = 20
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	lower := strings.ToLower(keyword)
	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	out := make([]schema.Node, 0, limit)
	names := make([]string, 0, len(typeMap))
	for name := range typeMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if len(out) >= limit {
			break
		}
		if !strings.Contains(strings.ToLower(name), lower) {
			continue
		}
		typ := typeMap[name]
		if nodeType != "" {
			if !validIdent(nodeType) {
				return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
			}
			if typ != nodeType {
				continue
			}
		}
		out = append(out, schema.Node{
			Label: name,
			Type:  typ,
			Attr:  map[string]any{"name": name},
		})
	}
	return out, nil
}

func (c *CayleyDB) FindNeighbors(nodeType string, name string, relType string, limit int) ([]schema.Node, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
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

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok, err := c.getNodeTypeLocked(name)
	if err != nil {
		return nil, err
	}
	if !ok || t != nodeType {
		return nil, nil
	}

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	neighbors, err := c.neighborNamesLocked(name, relType, limit)
	if err != nil {
		return nil, err
	}

	out := make([]schema.Node, 0, len(neighbors))
	for _, nb := range neighbors {
		out = append(out, schema.Node{
			Label: nb,
			Type:  typeMap[nb],
			Attr:  map[string]any{"name": nb},
		})
	}
	return out, nil
}

func (c *CayleyDB) FindNeighborEdges(nodeType string, name string, relTypes []string, limit int) ([]schema.Edge, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
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

	filter := map[string]bool{}
	if len(relTypes) > 0 {
		for _, rt := range relTypes {
			rt = strings.TrimSpace(rt)
			if rt == "" {
				continue
			}
			if !validIdent(rt) {
				return nil, fmt.Errorf("invalid relType=%q", rt)
			}
			filter[rt] = true
		}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	t, ok, err := c.getNodeTypeLocked(name)
	if err != nil {
		return nil, err
	}
	if !ok || t != nodeType {
		return nil, nil
	}

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	edges, err := c.neighborEdgesLocked(name, filter)
	if err != nil {
		return nil, err
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		if edges[i].Object != nil && edges[j].Object != nil && edges[i].Object.Label != edges[j].Object.Label {
			return edges[i].Object.Label < edges[j].Object.Label
		}
		return false
	})

	if len(edges) > limit {
		edges = edges[:limit]
	}
	for i := range edges {
		if edges[i].Subject != nil {
			edges[i].Subject.Type = typeMap[edges[i].Subject.Label]
		}
		if edges[i].Object != nil {
			edges[i].Object.Type = typeMap[edges[i].Object.Label]
		}
	}
	return edges, nil
}

func (c *CayleyDB) FindRelationshipsByLink(link string, limit int) ([]schema.Edge, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	link = strings.TrimSpace(link)
	if link == "" {
		return nil, fmt.Errorf("link is required")
	}
	if limit <= 0 {
		limit = 50
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	it := c.store.QuadIterator(quad.Label, c.store.ValueOf(quad.String(link)))
	defer it.Close()

	out := make([]schema.Edge, 0, limit)
	ctx := context.Background()
	for it.Next(ctx) {
		if len(out) >= limit {
			break
		}
		q := c.store.Quad(it.Result())
		sub := quad.NativeOf(q.Subject)
		pred := quad.NativeOf(q.Predicate)
		obj := quad.NativeOf(q.Object)
		ss, _ := sub.(string)
		pp, _ := pred.(string)
		oo, _ := obj.(string)
		if ss == "" || pp == "" || oo == "" {
			continue
		}
		if pp == "type" {
			continue
		}
		out = append(out, schema.Edge{
			Type: pp,
			Attr: map[string]any{"link": link},
			Subject: &schema.Node{
				Label: ss,
				Type:  typeMap[ss],
				Attr:  map[string]any{"name": ss},
			},
			Object: &schema.Node{
				Label: oo,
				Type:  typeMap[oo],
				Attr:  map[string]any{"name": oo},
			},
		})
	}
	return out, nil
}

func (c *CayleyDB) SampleTriples(limit int) ([]schema.Edge, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	it := c.store.QuadsAllIterator()
	defer it.Close()

	out := make([]schema.Edge, 0, limit)
	ctx := context.Background()
	for it.Next(ctx) {
		if len(out) >= limit {
			break
		}
		q := c.store.Quad(it.Result())
		pred := quad.NativeOf(q.Predicate)
		pp, _ := pred.(string)
		if pp == "" || pp == "type" {
			continue
		}
		ss, _ := quad.NativeOf(q.Subject).(string)
		oo, _ := quad.NativeOf(q.Object).(string)
		if ss == "" || oo == "" {
			continue
		}
		edge := schema.Edge{
			Type: pp,
			Attr: map[string]any{},
			Subject: &schema.Node{
				Label: ss,
				Type:  typeMap[ss],
				Attr:  map[string]any{"name": ss},
			},
			Object: &schema.Node{
				Label: oo,
				Type:  typeMap[oo],
				Attr:  map[string]any{"name": oo},
			},
		}
		if q.Label != nil {
			if l, ok := quad.NativeOf(q.Label).(string); ok && l != "" {
				edge.Attr["link"] = l
			}
		}
		out = append(out, edge)
	}
	return out, nil
}

func (c *CayleyDB) SampleNodeNamesByLabel(label string, limit int) ([]string, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
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

	c.mu.RLock()
	defer c.mu.RUnlock()

	names, err := c.subjectsByTypeLocked(label)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if len(names) > limit {
		names = names[:limit]
	}
	return names, nil
}

func (c *CayleyDB) SampleTopPropsByLabel(label string, sample int, propLimit int) ([]graph.PropCount, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
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

	c.mu.RLock()
	defer c.mu.RUnlock()

	names, err := c.subjectsByTypeLocked(label)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	if len(names) > sample {
		names = names[:sample]
	}

	counts := map[string]int64{}
	for _, n := range names {
		it := c.store.QuadIterator(quad.Subject, c.store.ValueOf(quad.String(n)))
		ctx := context.Background()
		for it.Next(ctx) {
			q := c.store.Quad(it.Result())
			pp, _ := quad.NativeOf(q.Predicate).(string)
			if pp == "" || pp == "type" {
				continue
			}
			counts[pp]++
		}
		it.Close()
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > propLimit {
		keys = keys[:propLimit]
	}
	out := make([]graph.PropCount, 0, len(keys))
	for _, k := range keys {
		out = append(out, graph.PropCount{Key: k, Count: counts[k]})
	}
	return out, nil
}

func (c *CayleyDB) FindLeastConnectedNodes(nodeType string, limit int) ([]graph.NodeDegree, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	nodeType = strings.TrimSpace(nodeType)
	if nodeType != "" && !validIdent(nodeType) {
		return nil, fmt.Errorf("invalid nodeType=%q", nodeType)
	}
	if limit <= 0 {
		limit = 10
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}

	deg := map[string]int64{}
	it := c.store.QuadsAllIterator()
	defer it.Close()

	ctx := context.Background()
	for it.Next(ctx) {
		q := c.store.Quad(it.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		s, _ := quad.NativeOf(q.Subject).(string)
		o, _ := quad.NativeOf(q.Object).(string)
		if s != "" {
			deg[s]++
		}
		if o != "" {
			deg[o]++
		}
	}

	out := make([]graph.NodeDegree, 0, len(typeMap))
	for name, typ := range typeMap {
		if nodeType != "" && typ != nodeType {
			continue
		}
		out = append(out, graph.NodeDegree{Type: typ, Name: name, Degree: deg[name]})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Degree != out[j].Degree {
			return out[i].Degree < out[j].Degree
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (c *CayleyDB) GetTopLabels(limit int) ([]graph.LabelCount, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	if limit <= 0 {
		limit = 30
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}
	count := map[string]int64{}
	for _, t := range typeMap {
		if t != "" {
			count[t]++
		}
	}
	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if count[keys[i]] != count[keys[j]] {
			return count[keys[i]] > count[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	out := make([]graph.LabelCount, 0, len(keys))
	for _, k := range keys {
		out = append(out, graph.LabelCount{Label: k, Count: count[k]})
	}
	return out, nil
}

func (c *CayleyDB) GetTopRelationshipTypes(limit int) ([]graph.RelTypeCount, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	if limit <= 0 {
		limit = 30
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	count := map[string]int64{}
	it := c.store.QuadsAllIterator()
	defer it.Close()
	ctx := context.Background()
	for it.Next(ctx) {
		q := c.store.Quad(it.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		count[pp]++
	}

	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if count[keys[i]] != count[keys[j]] {
			return count[keys[i]] > count[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	out := make([]graph.RelTypeCount, 0, len(keys))
	for _, k := range keys {
		out = append(out, graph.RelTypeCount{Type: k, Count: count[k]})
	}
	return out, nil
}

func (c *CayleyDB) GetTopPatterns(limit int) ([]graph.PatternCount, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("cayley db is nil")
	}
	if limit <= 0 {
		limit = 30
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	typeMap, err := c.typeMapLocked()
	if err != nil {
		return nil, err
	}
	count := map[string]int64{}
	it := c.store.QuadsAllIterator()
	defer it.Close()
	ctx := context.Background()
	for it.Next(ctx) {
		q := c.store.Quad(it.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		ss, _ := quad.NativeOf(q.Subject).(string)
		oo, _ := quad.NativeOf(q.Object).(string)
		ft := typeMap[ss]
		tt := typeMap[oo]
		if ft == "" || tt == "" {
			continue
		}
		key := ft + "\x00" + pp + "\x00" + tt
		count[key]++
	}

	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if count[keys[i]] != count[keys[j]] {
			return count[keys[i]] > count[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	out := make([]graph.PatternCount, 0, len(keys))
	for _, k := range keys {
		parts := strings.Split(k, "\x00")
		if len(parts) != 3 {
			continue
		}
		out = append(out, graph.PatternCount{From: parts[0], Rel: parts[1], To: parts[2], Count: count[k]})
	}
	return out, nil
}

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

func (c *CayleyDB) ensureNodeTypeLocked(name string, nodeType string, tx *cgraph.Transaction) error {
	_, ok, err := c.getNodeTypeLocked(name)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	tx.AddQuad(quad.Make(quad.String(name), quad.String("type"), quad.String(nodeType), nil))
	return nil
}

func (c *CayleyDB) getNodeTypeLocked(name string) (string, bool, error) {
	p := cayley.StartPath(c.store, quad.String(name)).Out(quad.String("type"))
	var out string
	ctx := context.Background()
	it := p.Iterate(ctx)
	err := it.EachValue(c.store, func(v quad.Value) {
		if out != "" {
			return
		}
		if s, ok := quad.NativeOf(v).(string); ok {
			out = strings.TrimSpace(s)
		}
	})
	if err != nil {
		return "", false, err
	}
	if out == "" {
		return "", false, nil
	}
	return out, true, nil
}

func (c *CayleyDB) subjectsByTypeLocked(nodeType string) ([]string, error) {
	it := c.store.QuadIterator(quad.Object, c.store.ValueOf(quad.String(nodeType)))
	defer it.Close()
	ctx := context.Background()
	seen := map[string]bool{}
	out := make([]string, 0, 32)
	for it.Next(ctx) {
		q := c.store.Quad(it.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp != "type" {
			continue
		}
		ss, _ := quad.NativeOf(q.Subject).(string)
		if ss == "" || seen[ss] {
			continue
		}
		seen[ss] = true
		out = append(out, ss)
	}
	return out, nil
}

func (c *CayleyDB) typeMapLocked() (map[string]string, error) {
	it := c.store.QuadIterator(quad.Predicate, c.store.ValueOf(quad.String("type")))
	defer it.Close()
	ctx := context.Background()
	out := map[string]string{}
	for it.Next(ctx) {
		q := c.store.Quad(it.Result())
		ss, _ := quad.NativeOf(q.Subject).(string)
		oo, _ := quad.NativeOf(q.Object).(string)
		if ss == "" || oo == "" {
			continue
		}
		if _, exists := out[ss]; !exists {
			out[ss] = oo
		}
	}
	return out, nil
}

func (c *CayleyDB) neighborNamesLocked(seed string, relType string, limit int) ([]string, error) {
	seen := map[string]bool{}
	names := make([]string, 0, limit)

	ctx := context.Background()

	outIt := c.store.QuadIterator(quad.Subject, c.store.ValueOf(quad.String(seed)))
	for outIt.Next(ctx) {
		q := c.store.Quad(outIt.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		if relType != "" && pp != relType {
			continue
		}
		oo, _ := quad.NativeOf(q.Object).(string)
		if oo == "" || seen[oo] {
			continue
		}
		seen[oo] = true
		names = append(names, oo)
		if len(names) >= limit {
			break
		}
	}
	outIt.Close()

	if len(names) < limit {
		inIt := c.store.QuadIterator(quad.Object, c.store.ValueOf(quad.String(seed)))
		for inIt.Next(ctx) {
			q := c.store.Quad(inIt.Result())
			pp, _ := quad.NativeOf(q.Predicate).(string)
			if pp == "" || pp == "type" {
				continue
			}
			if relType != "" && pp != relType {
				continue
			}
			ss, _ := quad.NativeOf(q.Subject).(string)
			if ss == "" || seen[ss] {
				continue
			}
			seen[ss] = true
			names = append(names, ss)
			if len(names) >= limit {
				break
			}
		}
		inIt.Close()
	}

	sort.Strings(names)
	if len(names) > limit {
		names = names[:limit]
	}
	return names, nil
}

func (c *CayleyDB) neighborEdgesLocked(seed string, relTypes map[string]bool) ([]schema.Edge, error) {
	edges := make([]schema.Edge, 0, 64)
	ctx := context.Background()

	outIt := c.store.QuadIterator(quad.Subject, c.store.ValueOf(quad.String(seed)))
	for outIt.Next(ctx) {
		q := c.store.Quad(outIt.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		if len(relTypes) > 0 && !relTypes[pp] {
			continue
		}
		oo, _ := quad.NativeOf(q.Object).(string)
		if oo == "" {
			continue
		}
		edge := schema.Edge{
			Type: pp,
			Attr: map[string]any{},
			Subject: &schema.Node{
				Label: seed,
				Attr:  map[string]any{"name": seed},
			},
			Object: &schema.Node{
				Label: oo,
				Attr:  map[string]any{"name": oo},
			},
		}
		if q.Label != nil {
			if l, ok := quad.NativeOf(q.Label).(string); ok && l != "" {
				edge.Attr["link"] = l
			}
		}
		edges = append(edges, edge)
	}
	outIt.Close()

	inIt := c.store.QuadIterator(quad.Object, c.store.ValueOf(quad.String(seed)))
	for inIt.Next(ctx) {
		q := c.store.Quad(inIt.Result())
		pp, _ := quad.NativeOf(q.Predicate).(string)
		if pp == "" || pp == "type" {
			continue
		}
		if len(relTypes) > 0 && !relTypes[pp] {
			continue
		}
		ss, _ := quad.NativeOf(q.Subject).(string)
		if ss == "" {
			continue
		}
		edge := schema.Edge{
			Type: pp,
			Attr: map[string]any{},
			Subject: &schema.Node{
				Label: seed,
				Attr:  map[string]any{"name": seed},
			},
			Object: &schema.Node{
				Label: ss,
				Attr:  map[string]any{"name": ss},
			},
		}
		if q.Label != nil {
			if l, ok := quad.NativeOf(q.Label).(string); ok && l != "" {
				edge.Attr["link"] = l
			}
		}
		edges = append(edges, edge)
	}
	inIt.Close()

	return edges, nil
}
