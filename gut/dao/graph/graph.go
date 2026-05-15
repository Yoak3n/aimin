package graph

import "github.com/Yoak3n/aimin/blood/schema"

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

type DB interface {
	Close() error
	CreateNode(e []schema.EntityTable) error

	GetNode(nodeType string, name string) (*schema.Node, error)
	FindNodesByType(nodeType string, limit int) ([]schema.Node, error)
	FindNodesByNameContains(nodeType string, keyword string, limit int) ([]schema.Node, error)
	FindNeighbors(nodeType string, name string, relType string, limit int) ([]schema.Node, error)
	FindNeighborEdges(nodeType string, name string, relTypes []string, limit int) ([]schema.Edge, error)
	FindRelationshipsByLink(link string, limit int) ([]schema.Edge, error)

	SampleTriples(limit int) ([]schema.Edge, error)
	SampleNodeNamesByLabel(label string, limit int) ([]string, error)
	SampleTopPropsByLabel(label string, sample int, propLimit int) ([]PropCount, error)

	FindLeastConnectedNodes(nodeType string, limit int) ([]NodeDegree, error)
	GetTopLabels(limit int) ([]LabelCount, error)
	GetTopRelationshipTypes(limit int) ([]RelTypeCount, error)
	GetTopPatterns(limit int) ([]PatternCount, error)
}
