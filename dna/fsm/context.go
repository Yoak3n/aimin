package fsm

import "github.com/Yoak3n/aimin/dna/attribute"

// Context 运行上下文，用于传递数据
type Context struct {
	Data          map[string]interface{}
	Current       string
	OnStateChange func(string)
	Attr          *attribute.MinAttribute
}

func NewContext() *Context {
	return &Context{
		Data:    make(map[string]interface{}),
		Current: "",
		Attr:    attribute.NewMinAttribute(),
	}
}
