package tool

import (
	"context"

	"github.com/Yoak3n/aimin/hand/sandbox"
)

type Context struct {
	Ctx        context.Context
	Payload    string
	RunID      string
	ToolCallID string
	Action     string
	OnProgress func(string)
	Sandbox    *sandbox.Manager
}

func NewMcpContext() *Context {
	ctx := context.Background()
	return &Context{
		Ctx:     ctx,
		Sandbox: sandbox.NewManager(),
	}
}

func (c *Context) GetPayload() string {
	ret := c.Payload
	c.Payload = ""
	return ret
}

func (c *Context) SetPayload(data string) {
	c.Payload = data
}
