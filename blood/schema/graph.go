package schema

import "fmt"

type GraphElement interface {
	Info() string
	Connect(e GraphElement) string
}

type Node struct {
	Label string
	Type  string
	Attr  map[string]any
}

func (n Node) Info() string {
	ret := fmt.Sprintf("(%s:%s", n.Label, n.Type)
	if len(n.Attr) > 0 {
		ret += " {"
		for k, v := range n.Attr {
			ret += fmt.Sprintf("%s: %v, ", k, v)
		}
		ret = ret[:len(ret)-2]
		ret += "}"
	}
	ret += ")"
	return ret
}

func (n Node) Connect(e GraphElement) string {
	edge, ok := e.(*Edge)
	if ok {
		edge.Subject = &n
	}
	return n.Info() + "-" + e.Info()
}

type Edge struct {
	Label   string
	Type    string
	Attr    map[string]any
	Subject *Node
	Object  *Node
}

func (e Edge) Info() string {
	ret := fmt.Sprintf("[%s:%s", e.Label, e.Type)
	if len(e.Attr) > 0 {
		ret += " {"
		for k, v := range e.Attr {
			ret += fmt.Sprintf("%s: %s, ", k, v)
		}
		ret = ret[:len(ret)-2]
		ret += "}"
	}
	ret += "]"
	return ret
}

func (e Edge) Connect(ele GraphElement) string {
	o, ok := ele.(Node)
	if ok {
		e.Object = &o
	}
	return e.Info() + "->" + e.Info()
}

func (e Edge) Draw() string {
	if e.Subject != nil && e.Object != nil {
		return fmt.Sprintf("%s - %s -> %s", e.Subject.Info(), e.Info(), e.Object.Info())
	}
	return e.Info()
}
