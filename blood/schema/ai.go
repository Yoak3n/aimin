package schema

import "encoding/json"

const (
	OpenAIMessageRoleUser      = "user"
	OpenAIMessageRoleAssistant = "assistant"
	OpenAIMessageRoleSystem    = "system"
	OpenAIMessageRoleTool      = "tool"
)

type OpenAIMessageRole string

type OpenAIMessage struct {
	Role       OpenAIMessageRole `json:"role"`
	Content    string            `json:"content,omitempty"`
	Reasoning  json.RawMessage   `json:"reasoning_content,omitempty"`
	ToolCalls  []OpenAIToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function OpenAIFunctionCall `json:"function"`
}

type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAITool struct {
	Type     string                 `json:"type"`
	Function OpenAIFunctionToolSpec `json:"function"`
}

type OpenAIFunctionToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type ConversationData struct {
	From string `json:"from"`
	Id   string `json:"id"`
}
