package mcp

import "fmt"

type Tool struct {
	Name   string                    `json:"name"`
	Desc   string                    `json:"desc"`
	Action func(ctx *Context) string `json:"action"`
}

func (m *Tool) String() string {
	return fmt.Sprintf("%s: %s", m.Name, m.Desc)
}

func GetMcpTool(dir string) []*Tool {
	var tools []*Tool
	for _, tool := range GlobalMcpHUB().tools {
		tools = append(tools, tool)
	}
	return tools
}
