package mcp

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
)

type Tool struct {
	Name   string                         `json:"name"`
	Desc   string                         `json:"desc"`
	Action func(ctx *tool.Context) string `json:"action"`
}

func (m *Tool) String() string {
	return fmt.Sprintf("- %s: %s", m.Name, m.Desc)
}

func GetMcpTools() []*Tool {
	var tools []*Tool
	for _, tool := range GlobalMcpHUB().tools {
		tools = append(tools, tool)
	}
	return tools
}
