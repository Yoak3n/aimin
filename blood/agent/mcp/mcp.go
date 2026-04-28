package mcp

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/hand/sandbox"
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
	return m.ExecuteTool(name, payload)
}

func (m *McpHUB) ExecuteTool(name string, payload string) (string, error) {
	return m.ExecuteToolWithMeta(name, payload, "", "", "", nil)
}

func (m *McpHUB) ExecuteToolWithMeta(name string, payload string, runID string, toolCallID string, action string, onProgress func(string)) (string, error) {
	t, ok := m.tools[name]
	if !ok {
		return "未找到对应的工具", fmt.Errorf("未找到对应的工具")
	}
	baseCtx := m.Ctx
	if baseCtx == nil {
		baseCtx = tool.NewMcpContext()
		m.Ctx = baseCtx
	}
	if baseCtx.Sandbox == nil {
		baseCtx.Sandbox = sandbox.NewManager()
	}
	callCtx := &tool.Context{
		Ctx:        baseCtx.Ctx,
		Payload:    payload,
		RunID:      runID,
		ToolCallID: toolCallID,
		Action:     action,
		OnProgress: onProgress,
		Sandbox:    baseCtx.Sandbox,
	}

	res := t.Action(callCtx)
	if after, ok0 := strings.CutPrefix(res, "ERROR:"); ok0 {
		return res, errors.New(strings.TrimSpace(after))
	}
	return res, nil
}

func (m *McpHUB) CleanupRun(runID string) int {
	if m == nil || m.Ctx == nil || m.Ctx.Sandbox == nil {
		return 0
	}
	return m.Ctx.Sandbox.KillRun(runID)
}
