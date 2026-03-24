package mcp

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
)

type McpHUB struct {
	tools map[string]*Tool
	Ctx   *tool.Context
}

var (
	hub  *McpHUB
	once sync.Once
)

func GlobalMcpHUB() *McpHUB {
	once.Do(func() {
		hub = NewMcpHUB()
	})
	return hub
}

func NewMcpHUB() *McpHUB {
	return &McpHUB{
		tools: make(map[string]*Tool),
		Ctx:   tool.NewMcpContext(),
	}
}

func (m *McpHUB) RegisterTool(tool *Tool) {
	m.tools[tool.Name] = tool
}

func (m *McpHUB) RenderToolsList() string {
	if len(m.tools) == 0 {
		return ""
	}
	var ret strings.Builder
	for _, tool := range m.tools {
		fmt.Fprintf(&ret, "- %s\n", tool.String())
	}
	return ret.String()
}

func (m *McpHUB) Execute(action string) (string, error) {
	name, payload, err := helper.ParseFunctionCall(action)
	if err != nil {
		logger.Logger.Error("解析函数调用失败", err)
		return "工具调用错误: " + err.Error(), err
	}
	tool, ok := m.tools[name]
	if !ok {
		return "未找到对应的工具", fmt.Errorf("未找到对应的工具")
	}
	m.Ctx.SetPayload(payload)
	// TODO 是否需要请求权限
	return tool.Action(m.Ctx), nil
}
