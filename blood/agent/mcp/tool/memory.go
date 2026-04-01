package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/service/retrieval"
)

func ManageMemory(ctx *Context) string {
	raw := strings.TrimSpace(ctx.GetPayload())
	if raw == "" {
		return "ERROR: args is empty"
	}
	args := parseArgs(raw)

	action := strings.ToLower(strings.TrimSpace(args["action"]))
	if action == "" {
		action = strings.ToLower(strings.TrimSpace(args["_0"]))
	}
	if action == "" {
		return "ERROR: missing action"
	}

	switch action {
	case "read_long_term":
		return manageMemoryReadLongTerm()
	case "write_long_term":
		content := args["content"]
		if strings.TrimSpace(content) == "" {
			content = args["_1"]
		}
		return manageMemoryWriteLongTerm(content)
	case "read_daily":
		date := strings.TrimSpace(args["date"])
		if date == "" {
			date = strings.TrimSpace(args["_1"])
		}
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		return manageMemoryReadDaily(date)
	case "write_daily":
		date := strings.TrimSpace(args["date"])
		content := args["content"]
		if date == "" && strings.TrimSpace(args["_1"]) != "" && strings.TrimSpace(args["_2"]) != "" {
			date = strings.TrimSpace(args["_1"])
			content = args["_2"]
		} else if strings.TrimSpace(content) == "" {
			content = args["_1"]
		}
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		return manageMemoryWriteDaily(date, content)
	case "search":
		query := strings.TrimSpace(args["query"])
		if query == "" {
			query = strings.TrimSpace(args["_1"])
		}
		if query == "" {
			return "ERROR: missing query"
		}
		limit := parseLimit(args["limit"], 5)
		return manageMemorySearchConversationSummaries(query, limit)
	case "vector_search":
		query := strings.TrimSpace(args["query"])
		if query == "" {
			query = strings.TrimSpace(args["_1"])
		}
		if query == "" {
			return "ERROR: missing query"
		}
		limit := parseLimit(args["limit"], 5)
		return manageMemoryVectorSearchConversationSummaries(query, limit)
	case "get_conversation":
		id := strings.TrimSpace(firstNonEmpty(args["id"], args["conversation_id"], args["_1"]))
		if id == "" {
			return "ERROR: missing id"
		}
		return manageMemoryGetConversationByID(id)
	case "recent_conversations":
		limit := parseLimit(args["limit"], 10)
		return manageMemoryRecentConversations(limit)
	case "graph_get_node":
		nodeType := strings.TrimSpace(firstNonEmpty(args["node_type"], args["type"], args["_1"]))
		name := strings.TrimSpace(firstNonEmpty(args["name"], args["_2"]))
		if nodeType == "" || name == "" {
			return "ERROR: missing node_type/type or name"
		}
		return manageMemoryGraphGetNode(nodeType, name)
	case "graph_neighbors", "graph_related":
		nodeType := strings.TrimSpace(firstNonEmpty(args["node_type"], args["type"], args["_1"]))
		name := strings.TrimSpace(firstNonEmpty(args["name"], args["_2"]))
		if nodeType == "" || name == "" {
			return "ERROR: missing node_type/type or name"
		}
		relType := strings.TrimSpace(firstNonEmpty(args["rel_type"], args["rel"]))
		limit := parseLimit(args["limit"], 20)
		return manageMemoryGraphNeighbors(nodeType, name, relType, limit)
	case "graph_search_nodes":
		keyword := strings.TrimSpace(firstNonEmpty(args["keyword"], args["query"], args["_1"]))
		if keyword == "" {
			return "ERROR: missing keyword/query"
		}
		nodeType := strings.TrimSpace(firstNonEmpty(args["node_type"], args["type"], args["_2"]))
		limit := parseLimit(args["limit"], 20)
		return manageMemoryGraphSearchNodes(nodeType, keyword, limit)
	case "graph_relations_by_link":
		link := strings.TrimSpace(firstNonEmpty(args["link"], args["_1"]))
		if link == "" {
			return "ERROR: missing link"
		}
		limit := parseLimit(args["limit"], 50)
		return manageMemoryGraphRelationsByLink(link, limit)
	case "graph_schema_summary":
		refresh := strings.TrimSpace(firstNonEmpty(args["refresh"], args["force"])) != ""
		return manageMemoryGraphSchemaSummary(refresh)
	case "graph_subgraph":
		nodeType := strings.TrimSpace(firstNonEmpty(args["node_type"], args["type"], args["_1"]))
		name := strings.TrimSpace(firstNonEmpty(args["name"], args["_2"]))
		keyword := strings.TrimSpace(firstNonEmpty(args["keyword"], args["query"]))
		if nodeType == "" && name == "" && keyword == "" {
			return "ERROR: missing node_type/name or keyword"
		}
		hops := parseLimit(args["hops"], 2)
		if hops < 1 {
			hops = 1
		}
		if hops > 2 {
			hops = 2
		}
		limit1 := parseLimit(args["limit1"], 30)
		limit2 := parseLimit(args["limit2"], 10)
		maxTriples := parseLimit(args["max_triples"], 60)
		relTypes := strings.TrimSpace(firstNonEmpty(args["rel_types"], args["rels"]))
		return manageMemoryGraphSubgraph(nodeType, name, keyword, hops, limit1, limit2, maxTriples, relTypes)
	case "graph_add_triples":
		content := strings.TrimSpace(firstNonEmpty(args["triples"], args["content"], args["_1"]))
		if content == "" {
			return "ERROR: missing triples/content"
		}
		link := strings.TrimSpace(firstNonEmpty(args["link"], args["source"]))
		return manageMemoryGraphAddTriples(content, link)
	case "graph_seed_demo":
		link := strings.TrimSpace(firstNonEmpty(args["link"], args["source"]))
		if link == "" {
			link = "seed_demo"
		}
		return manageMemoryGraphSeedDemo(link)
	default:
		return fmt.Sprintf("ERROR: unsupported action: %s", action)
	}
}

func manageMemoryGraphSchemaSummary(refresh bool) string {
	path, err := resolveWorkspaceFile(filepath.ToSlash(filepath.Join("memory", "GRAPH_SCHEMA.md")))
	if err != nil {
		return "ERROR: " + err.Error()
	}

	if !refresh {
		if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
			return string(b)
		}
	}

	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}

	labels, err := n4.GetTopLabels(30)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	rels, err := n4.GetTopRelationshipTypes(30)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	patterns, err := n4.GetTopPatterns(30)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	labelProps := make([]string, 0, 10)
	for i, lc := range labels {
		if i >= 10 {
			break
		}
		props, err := n4.SampleTopPropsByLabel(lc.Label, 200, 5)
		if err != nil {
			continue
		}
		keys := make([]string, 0, len(props))
		for _, p := range props {
			if p.Key == "" {
				continue
			}
			keys = append(keys, p.Key)
		}
		if len(keys) == 0 {
			continue
		}
		labelProps = append(labelProps, fmt.Sprintf("%s{%s}", lc.Label, strings.Join(keys, ",")))
	}

	sb := strings.Builder{}
	sb.WriteString("# GRAPH_SCHEMA (cached)\n\n")
	sb.WriteString("HowToUse:\n")
	sb.WriteString("- Read this once as a global overview (types/relations/patterns).\n")
	sb.WriteString("- Then query a small local subgraph per question: manage_memory(action=\"graph_subgraph\",keyword=\"...\",hops=1|2,max_triples=40).\n")
	sb.WriteString("- Prefer triples + small props; avoid dumping full nodes.\n\n")

	var totalNodes int64
	for _, lc := range labels {
		totalNodes += lc.Count
	}
	var totalRels int64
	for _, rc := range rels {
		totalRels += rc.Count
	}
	fmt.Fprintf(&sb, "Stats:\n- nodes=%d\n- rels=%d\n\n", totalNodes, totalRels)

	sb.WriteString("NodeTypes:\n")
	for _, lc := range labels {
		fmt.Fprintf(&sb, "- %s(%d)\n", lc.Label, lc.Count)
	}
	sb.WriteString("\nRelTypes:\n")
	for _, rc := range rels {
		fmt.Fprintf(&sb, "- %s(%d)\n", rc.Type, rc.Count)
	}
	sb.WriteString("\nTopPatterns:\n")
	for _, pc := range patterns {
		fmt.Fprintf(&sb, "- (%s)-%s->(%s) (%d)\n", pc.From, pc.Rel, pc.To, pc.Count)
	}
	if len(labelProps) > 0 {
		sb.WriteString("\nKeyProps:\n")
		for _, lp := range labelProps {
			fmt.Fprintf(&sb, "- %s\n", lp)
		}
	}
	if totalNodes < 10 {
		sb.WriteString("\nNextStep:\n")
		sb.WriteString("- Graph is sparse. Add more entities/relations, then refresh this file.\n")
		sb.WriteString("- Options:\n")
		sb.WriteString("  - manage_memory(action=\"graph_add_triples\",triples=\"人物|张三|工作于|地点|上海;人物|张三|具有|特性|擅长Go\",link=\"...optional...\")\n")
		sb.WriteString("  - manage_memory(action=\"graph_seed_demo\")  (demo data)\n")
		sb.WriteString("  - manage_memory(action=\"graph_schema_summary\",refresh=1)\n")
	}
	sb.WriteString("\n")

	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, []byte(sb.String()), 0644)
	return sb.String()
}

func manageMemoryGraphAddTriples(content string, link string) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return "ERROR: empty triples"
	}

	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	content = strings.ReplaceAll(content, ";", "\n")

	lines := strings.Split(content, "\n")
	chains := make([]schema.EntityTable, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			return "ERROR: invalid triple line: " + line
		}
		subType := strings.TrimSpace(parts[0])
		sub := strings.TrimSpace(parts[1])
		pred := strings.TrimSpace(parts[2])
		objType := strings.TrimSpace(parts[3])
		obj := strings.TrimSpace(parts[4])
		lk := link
		if len(parts) >= 6 && strings.TrimSpace(parts[5]) != "" {
			lk = strings.TrimSpace(parts[5])
		}
		chains = append(chains, schema.EntityTable{
			Subject:     sub,
			SubjectType: subType,
			Predicate:   pred,
			Object:      obj,
			ObjectType:  objType,
			Link:        lk,
		})
	}

	if len(chains) == 0 {
		return "ERROR: no triples"
	}
	if err := n4.CreateNode(chains); err != nil {
		return "ERROR: " + err.Error()
	}
	return fmt.Sprintf("OK: inserted %d triples", len(chains))
}

func manageMemoryGraphSeedDemo(link string) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}

	chains := []schema.EntityTable{
		{Subject: "张三", SubjectType: "人物", Predicate: "是", Object: "工程师", ObjectType: "概念", Link: link},
		{Subject: "张三", SubjectType: "人物", Predicate: "具有", Object: "擅长Go", ObjectType: "特性", Link: link},
		{Subject: "张三", SubjectType: "人物", Predicate: "工作于", Object: "上海", ObjectType: "地点", Link: link},
		{Subject: "李雷", SubjectType: "人物", Predicate: "是", Object: "产品经理", ObjectType: "概念", Link: link},
		{Subject: "李雷", SubjectType: "人物", Predicate: "具有", Object: "重视用户体验", ObjectType: "特性", Link: link},
		{Subject: "李雷", SubjectType: "人物", Predicate: "工作于", Object: "北京", ObjectType: "地点", Link: link},
		{Subject: "韩梅梅", SubjectType: "人物", Predicate: "是", Object: "研究员", ObjectType: "概念", Link: link},
		{Subject: "韩梅梅", SubjectType: "人物", Predicate: "具有", Object: "喜欢阅读", ObjectType: "特性", Link: link},
		{Subject: "韩梅梅", SubjectType: "人物", Predicate: "工作于", Object: "深圳", ObjectType: "地点", Link: link},
	}

	if err := n4.CreateNode(chains); err != nil {
		return "ERROR: " + err.Error()
	}
	return fmt.Sprintf("OK: seeded %d triples (link=%s)", len(chains), link)
}

func manageMemoryGraphSubgraph(nodeType string, name string, keyword string, hops int, limit1 int, limit2 int, maxTriples int, relTypes string) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}

	if strings.TrimSpace(name) == "" {
		k := strings.TrimSpace(keyword)
		if k == "" {
			k = strings.TrimSpace(name)
		}
		if k == "" {
			return "ERROR: missing seed name/keyword"
		}
		cands, err := n4.FindNodesByNameContains(nodeType, k, 5)
		if err != nil {
			return "ERROR: " + err.Error()
		}
		if len(cands) == 0 {
			return "no matches"
		}
		nodeType = cands[0].Type
		name = cands[0].Label
	}

	rt := splitRelTypes(relTypes)

	edges1, err := n4.FindNeighborEdges(nodeType, name, rt, limit1)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	edges := make([]schema.Edge, 0, len(edges1)+(len(edges1)*limit2))
	edges = append(edges, edges1...)

	if hops >= 2 && limit2 > 0 {
		seen := map[string]bool{}
		neighbors := make([]schema.Node, 0, len(edges1))
		for _, e := range edges1 {
			if e.Subject != nil && e.Subject.Type == nodeType && e.Subject.Label == name {
				if e.Object != nil {
					key := e.Object.Type + "|" + e.Object.Label
					if !seen[key] {
						seen[key] = true
						neighbors = append(neighbors, *e.Object)
					}
				}
			} else if e.Object != nil && e.Object.Type == nodeType && e.Object.Label == name {
				if e.Subject != nil {
					key := e.Subject.Type + "|" + e.Subject.Label
					if !seen[key] {
						seen[key] = true
						neighbors = append(neighbors, *e.Subject)
					}
				}
			} else {
				if e.Subject != nil {
					key := e.Subject.Type + "|" + e.Subject.Label
					if !seen[key] {
						seen[key] = true
						neighbors = append(neighbors, *e.Subject)
					}
				}
				if e.Object != nil {
					key := e.Object.Type + "|" + e.Object.Label
					if !seen[key] {
						seen[key] = true
						neighbors = append(neighbors, *e.Object)
					}
				}
			}
		}

		for i := range neighbors {
			if len(edges) >= maxTriples {
				break
			}
			nn := neighbors[i]
			if nn.Type == "" || nn.Label == "" {
				continue
			}
			e2, err := n4.FindNeighborEdges(nn.Type, nn.Label, rt, limit2)
			if err != nil {
				continue
			}
			for _, e := range e2 {
				edges = append(edges, e)
				if len(edges) >= maxTriples {
					break
				}
			}
		}
	}

	if maxTriples > 0 && len(edges) > maxTriples {
		edges = edges[:maxTriples]
	}

	seedNode, _ := n4.GetNode(nodeType, name)
	sb := strings.Builder{}
	sb.WriteString("<graph_subgraph>\n")
	sb.WriteString("<seed>\n")
	fmt.Fprintf(&sb, "<type>%s</type>\n", compactOneLine(nodeType, 80))
	fmt.Fprintf(&sb, "<name>%s</name>\n", compactOneLine(name, 240))
	sb.WriteString("</seed>\n")
	if seedNode != nil {
		sb.WriteString(formatProps(limitProps(seedNode.Attr, 3), 10))
	}
	sb.WriteString(formatTriples(edges, maxTriples))
	sb.WriteString(formatNodeNotes(edges, 10, 3))
	sb.WriteString("</graph_subgraph>")
	return sb.String()
}

func splitRelTypes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func limitProps(props map[string]any, max int) map[string]any {
	if len(props) == 0 || max <= 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if len(out) >= max {
			break
		}
		out[k] = props[k]
	}
	return out
}

func formatTriples(edges []schema.Edge, max int) string {
	sb := strings.Builder{}
	sb.WriteString("<triples>\n")
	n := len(edges)
	if max > 0 && n > max {
		n = max
	}
	for i := 0; i < n; i++ {
		e := edges[i]
		s := nodeRef(e.Subject)
		o := nodeRef(e.Object)
		rel := compactOneLine(e.Type, 80)
		fmt.Fprintf(&sb, "<t>(%s)-[%s]->(%s)</t>\n", s, rel, o)
	}
	sb.WriteString("</triples>\n")
	return sb.String()
}

func nodeRef(n *schema.Node) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%s|%s", compactOneLine(n.Type, 60), compactOneLine(n.Label, 180))
}

func formatNodeNotes(edges []schema.Edge, maxNodes int, maxProps int) string {
	nodes := map[string]*schema.Node{}
	for _, e := range edges {
		if e.Subject != nil {
			key := e.Subject.Type + "|" + e.Subject.Label
			nodes[key] = e.Subject
		}
		if e.Object != nil {
			key := e.Object.Type + "|" + e.Object.Label
			nodes[key] = e.Object
		}
	}
	if len(nodes) == 0 {
		return ""
	}
	keys := make([]string, 0, len(nodes))
	for k := range nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if maxNodes > 0 && len(keys) > maxNodes {
		keys = keys[:maxNodes]
	}
	sb := strings.Builder{}
	sb.WriteString("<node_notes>\n")
	for _, k := range keys {
		n := nodes[k]
		sb.WriteString("<node>\n")
		fmt.Fprintf(&sb, "<id>%s</id>\n", compactOneLine(k, 260))
		sb.WriteString(formatProps(limitProps(n.Attr, maxProps), maxProps))
		sb.WriteString("</node>\n")
	}
	sb.WriteString("</node_notes>\n")
	return sb.String()
}

func manageMemoryGraphGetNode(nodeType string, name string) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}
	node, err := n4.GetNode(nodeType, name)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	if node == nil {
		return "no matches"
	}
	sb := strings.Builder{}
	sb.WriteString("<graph_node>\n")
	fmt.Fprintf(&sb, "<type>%s</type>\n", compactOneLine(node.Type, 80))
	fmt.Fprintf(&sb, "<name>%s</name>\n", compactOneLine(node.Label, 240))
	sb.WriteString(formatProps(node.Attr, 30))
	sb.WriteString("</graph_node>")
	return sb.String()
}

func manageMemoryGraphNeighbors(nodeType string, name string, relType string, limit int) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}
	nodes, err := n4.FindNeighbors(nodeType, name, relType, limit)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	if len(nodes) == 0 {
		return "no matches"
	}
	sb := strings.Builder{}
	sb.WriteString("<graph_neighbors>\n")
	fmt.Fprintf(&sb, "<center>\n<type>%s</type>\n<name>%s</name>\n</center>\n", compactOneLine(nodeType, 80), compactOneLine(name, 240))
	if strings.TrimSpace(relType) != "" {
		fmt.Fprintf(&sb, "<rel_type>%s</rel_type>\n", compactOneLine(relType, 80))
	}
	sb.WriteString("<neighbors>\n")
	for _, n := range nodes {
		fmt.Fprintf(&sb, "<node>\n<type>%s</type>\n<name>%s</name>\n</node>\n", compactOneLine(n.Type, 80), compactOneLine(n.Label, 240))
	}
	sb.WriteString("</neighbors>\n</graph_neighbors>")
	return sb.String()
}

func manageMemoryGraphSearchNodes(nodeType string, keyword string, limit int) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}
	nodes, err := n4.FindNodesByNameContains(nodeType, keyword, limit)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	if len(nodes) == 0 {
		return "no matches"
	}
	sb := strings.Builder{}
	sb.WriteString("<graph_search_results>\n")
	fmt.Fprintf(&sb, "<keyword>%s</keyword>\n", compactOneLine(keyword, 200))
	if strings.TrimSpace(nodeType) != "" {
		fmt.Fprintf(&sb, "<node_type>%s</node_type>\n", compactOneLine(nodeType, 80))
	}
	sb.WriteString("<nodes>\n")
	for _, n := range nodes {
		fmt.Fprintf(&sb, "<node>\n<type>%s</type>\n<name>%s</name>\n</node>\n", compactOneLine(n.Type, 80), compactOneLine(n.Label, 240))
	}
	sb.WriteString("</nodes>\n</graph_search_results>")
	return sb.String()
}

func manageMemoryGraphRelationsByLink(link string, limit int) string {
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "ERROR: neuro db is nil"
	}
	edges, err := n4.FindRelationshipsByLink(link, limit)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	if len(edges) == 0 {
		return "no matches"
	}
	sb := strings.Builder{}
	sb.WriteString("<graph_edges>\n")
	fmt.Fprintf(&sb, "<link>%s</link>\n", compactOneLine(link, 240))
	sb.WriteString("<edges>\n")
	for _, e := range edges {
		sb.WriteString("<edge>\n")
		fmt.Fprintf(&sb, "<type>%s</type>\n", compactOneLine(e.Type, 80))
		if e.Subject != nil {
			fmt.Fprintf(&sb, "<subject>\n<type>%s</type>\n<name>%s</name>\n</subject>\n", compactOneLine(e.Subject.Type, 80), compactOneLine(e.Subject.Label, 240))
		}
		if e.Object != nil {
			fmt.Fprintf(&sb, "<object>\n<type>%s</type>\n<name>%s</name>\n</object>\n", compactOneLine(e.Object.Type, 80), compactOneLine(e.Object.Label, 240))
		}
		sb.WriteString("</edge>\n")
	}
	sb.WriteString("</edges>\n</graph_edges>")
	return sb.String()
}

func manageMemoryReadLongTerm() string {
	path, err := resolveWorkspaceFile("MEMORY.md")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return "ERROR: " + err.Error()
	}
	return string(b)
}

func manageMemoryWriteLongTerm(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "ERROR: content is empty"
	}
	path, err := resolveWorkspaceFile("MEMORY.md")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	existing, _ := os.ReadFile(path)
	out := strings.Builder{}
	out.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		out.WriteString("\n")
	}
	if len(existing) > 0 {
		out.WriteString("\n")
	}
	out.WriteString(content)
	out.WriteString("\n")

	if err := os.WriteFile(path, []byte(out.String()), 0644); err != nil {
		return "ERROR: " + err.Error()
	}
	return "ok"
}

func manageMemoryReadDaily(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return "ERROR: date is empty"
	}
	path, err := resolveWorkspaceFile(filepath.ToSlash(filepath.Join("memory", date+".md")))
	if err != nil {
		return "ERROR: " + err.Error()
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return "ERROR: " + err.Error()
	}
	return string(b)
}

func manageMemoryWriteDaily(date string, content string) string {
	date = strings.TrimSpace(date)
	content = strings.TrimSpace(content)
	if date == "" {
		return "ERROR: date is empty"
	}
	if content == "" {
		return "ERROR: content is empty"
	}
	path, err := resolveWorkspaceFile(filepath.ToSlash(filepath.Join("memory", date+".md")))
	if err != nil {
		return "ERROR: " + err.Error()
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	existing, _ := os.ReadFile(path)
	out := strings.Builder{}
	out.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		out.WriteString("\n")
	}
	if len(existing) > 0 {
		out.WriteString("\n")
	}
	out.WriteString(content)
	out.WriteString("\n")

	if err := os.WriteFile(path, []byte(out.String()), 0644); err != nil {
		return "ERROR: " + err.Error()
	}
	return "ok"
}

func manageMemorySearchConversationSummaries(query string, limit int) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return "ERROR: query is empty"
	}
	if limit <= 0 {
		limit = 5
	}

	emb, err := helper.UseLLM().Embedding([]string{query})
	var records []schema.ConversationRecord
	if err == nil && len(emb) > 0 {
		records, err = helper.UseDB().GetReleventConversationRecords(emb[0], limit)
	}
	if err != nil || len(records) == 0 {
		db := helper.UseDB().GetPostgresSQL()
		if db == nil {
			return "ERROR: database is nil"
		}
		like := "%" + query + "%"
		records = make([]schema.ConversationRecord, 0, limit)
		_ = db.
			Where("question ILIKE ? OR answer ILIKE ?", like, like).
			Order("updated_at desc").
			Limit(limit).
			Find(&records).Error
	}

	if len(records) == 0 {
		return "no matches"
	}

	sb := strings.Builder{}
	sb.WriteString("<search_results>\n")
	for _, r := range records {
		sum := ""
		if s, err := helper.UseDB().GetSummaryMemoryTableRecordByLink(r.Id); err == nil {
			sum = strings.TrimSpace(s.Content)
		}
		q := compactOneLine(r.Question, 240)
		fmt.Fprintf(&sb, "<conversation_summary id=%q>\n", r.Id)
		if sum != "" {
			fmt.Fprintf(&sb, "<summary>%s</summary>\n", compactOneLine(sum, 600))
		} else {
			a := compactOneLine(r.Answer, 360)
			fmt.Fprintf(&sb, "<summary>%s</summary>\n", compactOneLine(q+" / "+a, 600))
		}
		fmt.Fprintf(&sb, "<question>%s</question>\n", q)
		sb.WriteString("</conversation_summary>\n")
	}
	sb.WriteString("</search_results>")
	return sb.String()
}

func manageMemoryVectorSearchConversationSummaries(query string, limit int) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return "ERROR: query is empty"
	}
	records, err := retrieval.VectorSearchConversations(query, limit)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	if len(records) == 0 {
		return "no matches"
	}

	sb := strings.Builder{}
	sb.WriteString("<vector_search_results>\n")
	fmt.Fprintf(&sb, "<query>%s</query>\n", compactOneLine(query, 240))
	for _, r := range records {
		sum := ""
		if s, err := helper.UseDB().GetSummaryMemoryTableRecordByLink(r.Id); err == nil {
			sum = strings.TrimSpace(s.Content)
		}
		q := compactOneLine(r.Question, 240)
		fmt.Fprintf(&sb, "<conversation_summary id=%q>\n", r.Id)
		if sum != "" {
			fmt.Fprintf(&sb, "<summary>%s</summary>\n", compactOneLine(sum, 600))
		} else {
			a := compactOneLine(r.Answer, 360)
			fmt.Fprintf(&sb, "<summary>%s</summary>\n", compactOneLine(q+" / "+a, 600))
		}
		fmt.Fprintf(&sb, "<question>%s</question>\n", q)
		sb.WriteString("</conversation_summary>\n")
	}
	sb.WriteString("</vector_search_results>")
	return sb.String()
}

func manageMemoryGetConversationByID(id string) string {
	rec, err := helper.UseDB().GetConversationByID(id)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	sb := strings.Builder{}
	fmt.Fprintf(&sb, "<conversation id=%q>\n", rec.Id)
	if strings.TrimSpace(rec.System) != "" {
		fmt.Fprintf(&sb, "<system>%s</system>\n", strings.TrimSpace(rec.System))
	}
	fmt.Fprintf(&sb, "<question>%s</question>\n", strings.TrimSpace(rec.Question))
	if strings.TrimSpace(rec.Thoughts) != "" {
		fmt.Fprintf(&sb, "<thoughts>%s</thoughts>\n", strings.TrimSpace(rec.Thoughts))
	}
	fmt.Fprintf(&sb, "<answer>%s</answer>\n", strings.TrimSpace(rec.Answer))
	sb.WriteString("</conversation>")
	return sb.String()
}

func manageMemoryRecentConversations(limit int) string {
	if limit <= 0 {
		limit = 10
	}
	db := helper.UseDB().GetPostgresSQL()
	if db == nil {
		return "ERROR: database is nil"
	}
	records := make([]schema.ConversationRecord, 0, limit)
	if err := db.Order("updated_at desc").Limit(limit).Find(&records).Error; err != nil {
		return "ERROR: " + err.Error()
	}
	if len(records) == 0 {
		return "no conversations"
	}
	sb := strings.Builder{}
	sb.WriteString("<recent_conversations>\n")
	for _, r := range records {
		q := compactOneLine(r.Question, 180)
		a := compactOneLine(r.Answer, 240)
		fmt.Fprintf(&sb, "<conversation id=%q>\n", r.Id)
		fmt.Fprintf(&sb, "<question>%s</question>\n", q)
		fmt.Fprintf(&sb, "<answer>%s</answer>\n", a)
		sb.WriteString("</conversation>\n")
	}
	sb.WriteString("</recent_conversations>")
	return sb.String()
}

func parseLimit(s string, fallback int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return fallback
	}
	if n > 50 {
		return 50
	}
	return n
}

func compactOneLine(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Join(strings.Fields(s), " ")
	if maxRunes <= 0 {
		return s
	}
	rs := []rune(s)
	if len(rs) <= maxRunes {
		return s
	}
	return string(rs[:maxRunes]) + "..."
}

func formatProps(props map[string]any, limit int) string {
	if len(props) == 0 || limit == 0 {
		return ""
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}
	sb := strings.Builder{}
	sb.WriteString("<props>\n")
	for _, k := range keys {
		v := props[k]
		val := compactOneLine(fmt.Sprintf("%v", v), 240)
		fmt.Fprintf(&sb, "<prop key=%q>%s</prop>\n", k, val)
	}
	sb.WriteString("</props>\n")
	return sb.String()
}

func resolveWorkspaceFile(rel string) (string, error) {
	workspaceRoot := strings.TrimSpace(config.GlobalConfiguration().Workspace.Path)
	if workspaceRoot == "" {
		return "", fmt.Errorf("workspace path is empty")
	}
	rel = filepath.FromSlash(strings.TrimSpace(rel))
	if filepath.IsAbs(rel) {
		abs := filepath.Clean(rel)
		r, err := filepath.Rel(workspaceRoot, abs)
		if err != nil {
			return "", err
		}
		r = filepath.Clean(r)
		if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("path escapes workspace")
		}
		return abs, nil
	}

	rel = strings.TrimLeft(rel, `\/`)
	abs := filepath.Clean(filepath.Join(workspaceRoot, rel))
	r, err := filepath.Rel(workspaceRoot, abs)
	if err != nil {
		return "", err
	}
	r = filepath.Clean(r)
	if r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes workspace")
	}
	return abs, nil
}
