package mcp

import "context"

type Context struct {
	Ctx     context.Context
	Payload string
}

func NewMcpContext() *Context {
	ctx := context.Background()
	return &Context{
		Ctx: ctx,
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
