package adapter

import (
	"encoding/json"

	"github.com/Yoak3n/aimin/blood/schema"
)

type ToolCallDelta struct {
	Index    int    `json:"index"`
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Function struct {
		Name      string `json:"name,omitempty"`
		Arguments string `json:"arguments,omitempty"`
	} `json:"function,omitempty"`
}

type Chunk struct {
	Choices []struct {
		Delta struct {
			Content   string          `json:"content"`
			Reasoning json.RawMessage `json:"reasoning_content"`
			ToolCalls []ToolCallDelta `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string               `json:"finish_reason"`
		Message      schema.OpenAIMessage `json:"message"`
		Text         string               `json:"text"`
	} `json:"choices"`
}
