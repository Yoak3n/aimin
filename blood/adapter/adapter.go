package adapter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/schema"
)

const defaultSystemPrompt = "你是一个智能助手，你的回答必须符合中文语法规范。"

type LLMAdapter interface {
	Chat(userMessages []schema.OpenAIMessage, systemPrompt ...string) (string, error)
	ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error)
	Embedding(text []string) ([][]float32, error)
	GetConfig() *config.LLMConfig
}

type BaseAdapter struct {
	config *config.LLMConfig
	mutex  sync.RWMutex
	client *http.Client
}

type messageBuildOption struct {
	ForceAssistantReasoning bool
}

func thinkingParam(c *config.LLMConfig) (map[string]any, bool) {
	if c == nil {
		return nil, false
	}
	provider := strings.ToLower(strings.TrimSpace(c.Provider))
	if provider != "deepseek" {
		return nil, false
	}
	if c.Type == config.LLMTypeThink {
		return map[string]any{"type": "enabled"}, true
	}
	if c.Type == config.LLMTypeChat {
		return map[string]any{"type": "disabled"}, true
	}
	return nil, false
}

func shouldForceAssistantReasoning(c *config.LLMConfig) bool {
	if c == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(c.Provider), "bigmodel")
}

func buildChatMessages(c *config.LLMConfig, systemPrompt string, userMessages []schema.OpenAIMessage, opt messageBuildOption) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(userMessages)+1)
	out = append(out, map[string]any{
		"role":    string(schema.OpenAIMessageRoleSystem),
		"content": systemPrompt,
	})

	for _, msg := range userMessages {
		role := string(msg.Role)
		m := map[string]any{
			"role":    role,
			"content": msg.Content,
		}

		if msg.ToolCallID != "" {
			m["tool_call_id"] = msg.ToolCallID
		}
		if len(msg.ToolCalls) > 0 {
			m["tool_calls"] = msg.ToolCalls
		}

		if len(msg.Reasoning) > 0 {
			m["reasoning_content"] = msg.Reasoning
		} else if opt.ForceAssistantReasoning && role == string(schema.OpenAIMessageRoleAssistant) && len(msg.ToolCalls) > 0 {
			m["reasoning_content"] = json.RawMessage(`""`)
		}

		if role == string(schema.OpenAIMessageRoleTool) && strings.TrimSpace(msg.ToolCallID) == "" {
			return nil, fmt.Errorf("tool message missing tool_call_id")
		}
		if c != nil {
			provider := strings.ToLower(strings.TrimSpace(c.Provider))
			if provider == "deepseek" && c.Type == config.LLMTypeThink {
				if role == string(schema.OpenAIMessageRoleAssistant) && len(msg.ToolCalls) > 0 && len(msg.Reasoning) == 0 {
					return nil, fmt.Errorf("deepseek think mode: assistant tool-call message missing reasoning_content")
				}
			}
		}
		if role == string(schema.OpenAIMessageRoleAssistant) && len(msg.ToolCalls) > 0 && msg.Content == "" {
			m["content"] = ""
		}
		if role == string(schema.OpenAIMessageRoleTool) && msg.Content == "" {
			m["content"] = ""
		}

		out = append(out, m)
	}

	return out, nil
}

// Chat implements LLMAdapter.
func (b *BaseAdapter) Chat(userMessages []schema.OpenAIMessage, systemMessage ...string) (string, error) {
	systemPrompt := ""
	if len(systemMessage) == 0 {
		systemPrompt = defaultSystemPrompt
	} else {
		systemPrompt = systemMessage[0]
	}

	if len(userMessages) == 0 {
		return "", fmt.Errorf("用户消息不能为空")
	}
	messages, buildErr := buildChatMessages(b.config, systemPrompt, userMessages, messageBuildOption{
		ForceAssistantReasoning: shouldForceAssistantReasoning(b.config),
	})
	if buildErr != nil {
		return "", buildErr
	}

	reqBody := map[string]any{
		"model":    b.config.Model,
		"messages": messages,
		"stream":   false,
	}
	if v, ok := thinkingParam(b.config); ok {
		reqBody["thinking"] = v
	}

	return b.makeRequest(reqBody)
}

func (b *BaseAdapter) ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemMessage ...string) (schema.OpenAIMessage, error) {
	systemPrompt := ""
	if len(systemMessage) == 0 {
		systemPrompt = defaultSystemPrompt
	} else {
		systemPrompt = systemMessage[0]
	}

	if len(userMessages) == 0 {
		return schema.OpenAIMessage{}, fmt.Errorf("用户消息不能为空")
	}
	messages, buildErr := buildChatMessages(b.config, systemPrompt, userMessages, messageBuildOption{
		ForceAssistantReasoning: shouldForceAssistantReasoning(b.config),
	})
	if buildErr != nil {
		return schema.OpenAIMessage{}, buildErr
	}

	reqBody := map[string]any{
		"model":    b.config.Model,
		"messages": messages,
		"stream":   true,
	}
	if v, ok := thinkingParam(b.config); ok {
		reqBody["thinking"] = v
	}
	if len(tools) > 0 {
		reqBody["tools"] = tools
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return schema.OpenAIMessage{}, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", b.config.APIUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return schema.OpenAIMessage{}, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := b.client.Do(req)
	if err != nil {
		return schema.OpenAIMessage{}, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return schema.OpenAIMessage{}, fmt.Errorf("API 请求失败，状态码: %d, provider=%s, model=%s", resp.StatusCode, b.config.Provider, b.config.Model)
		}
		return schema.OpenAIMessage{}, fmt.Errorf("API 请求失败，状态码: %d, provider=%s, model=%s, 响应: %s", resp.StatusCode, b.config.Provider, b.config.Model, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	var raw strings.Builder
	var content strings.Builder
	var reasoning strings.Builder
	var reasoningRaw json.RawMessage
	toolCallsByIndex := make(map[int]*schema.OpenAIToolCall)
	toolCallEmitted := make(map[int]bool)

	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				return schema.OpenAIMessage{}, readErr
			}
			continue
		}

		if after, ok := strings.CutPrefix(line, "data:"); ok {
			payload := strings.TrimSpace(after)
			if payload == "[DONE]" {
				break
			}

			type toolCallDelta struct {
				Index    int    `json:"index"`
				ID       string `json:"id,omitempty"`
				Type     string `json:"type,omitempty"`
				Function struct {
					Name      string `json:"name,omitempty"`
					Arguments string `json:"arguments,omitempty"`
				} `json:"function,omitempty"`
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content   string          `json:"content"`
						Reasoning json.RawMessage `json:"reasoning_content"`
						ToolCalls []toolCallDelta `json:"tool_calls"`
					} `json:"delta"`
					Message schema.OpenAIMessage `json:"message"`
					Text    string               `json:"text"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				raw.WriteString(payload)
			} else if len(chunk.Choices) > 0 {
				if len(reasoningRaw) == 0 {
					if text, ok := decodeJSONString(chunk.Choices[0].Delta.Reasoning); ok {
						reasoning.WriteString(text)
					} else if len(chunk.Choices[0].Delta.Reasoning) > 0 && reasoning.Len() == 0 {
						reasoningRaw = append([]byte(nil), chunk.Choices[0].Delta.Reasoning...)
					}
				}
				if reasoning.Len() == 0 && len(reasoningRaw) == 0 && len(chunk.Choices[0].Message.Reasoning) > 0 {
					if text, ok := decodeJSONString(chunk.Choices[0].Message.Reasoning); ok {
						reasoning.WriteString(text)
					} else {
						reasoningRaw = append([]byte(nil), chunk.Choices[0].Message.Reasoning...)
					}
				}

				deltaText := chunk.Choices[0].Delta.Content
				if deltaText == "" {
					deltaText = chunk.Choices[0].Message.Content
				}
				if deltaText == "" {
					deltaText = chunk.Choices[0].Text
				}

				if deltaText != "" {
					content.WriteString(deltaText)
					if onDelta != nil {
						if err := onDelta(deltaText); err != nil {
							return schema.OpenAIMessage{}, err
						}
					}
				}

				for _, tc := range chunk.Choices[0].Delta.ToolCalls {
					existing := toolCallsByIndex[tc.Index]
					if existing == nil {
						toolCallsByIndex[tc.Index] = &schema.OpenAIToolCall{
							ID:   tc.ID,
							Type: tc.Type,
							Function: schema.OpenAIFunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						}
					} else {
						if existing.ID == "" && tc.ID != "" {
							existing.ID = tc.ID
						}
						if existing.Type == "" && tc.Type != "" {
							existing.Type = tc.Type
						}
						if existing.Function.Name == "" && tc.Function.Name != "" {
							existing.Function.Name = tc.Function.Name
						}
						if tc.Function.Arguments != "" {
							existing.Function.Arguments += tc.Function.Arguments
						}
					}

					current := toolCallsByIndex[tc.Index]
					if current == nil || onDelta == nil || toolCallEmitted[tc.Index] {
						continue
					}
					if strings.TrimSpace(current.ID) == "" || strings.TrimSpace(current.Function.Name) == "" {
						continue
					}
					if strings.TrimSpace(current.Function.Arguments) == "" {
						continue
					}
					var argsAny any
					if err := json.Unmarshal([]byte(current.Function.Arguments), &argsAny); err != nil {
						continue
					}
					toolCallEmitted[tc.Index] = true
					if err := onDelta(fmt.Sprintf("[tool_call] %s %s %s", current.ID, current.Function.Name, current.Function.Arguments)); err != nil {
						return schema.OpenAIMessage{}, err
					}
				}

				if len(chunk.Choices[0].Message.ToolCalls) > 0 {
					for i := range chunk.Choices[0].Message.ToolCalls {
						c := chunk.Choices[0].Message.ToolCalls[i]
						toolCallsByIndex[i] = &c
						if onDelta != nil && !toolCallEmitted[i] && strings.TrimSpace(c.ID) != "" && strings.TrimSpace(c.Function.Name) != "" && strings.TrimSpace(c.Function.Arguments) != "" {
							var argsAny any
							if err := json.Unmarshal([]byte(c.Function.Arguments), &argsAny); err == nil {
								toolCallEmitted[i] = true
								if err := onDelta(fmt.Sprintf("[tool_call] %s %s %s", c.ID, c.Function.Name, c.Function.Arguments)); err != nil {
									return schema.OpenAIMessage{}, err
								}
							}
						}
					}
				}
			}
		} else if strings.HasPrefix(line, "event:") {
		} else {
			raw.WriteString(line)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return schema.OpenAIMessage{}, readErr
		}
	}

	if content.Len() > 0 || len(toolCallsByIndex) > 0 {
		toolCalls := make([]schema.OpenAIToolCall, 0, len(toolCallsByIndex))
		for i := 0; i < len(toolCallsByIndex); i++ {
			tc := toolCallsByIndex[i]
			if tc == nil {
				continue
			}
			if strings.TrimSpace(tc.Function.Name) == "" {
				continue
			}
			toolCalls = append(toolCalls, *tc)
		}
		var outReasoning json.RawMessage
		if len(reasoningRaw) > 0 {
			outReasoning = reasoningRaw
		} else if reasoning.Len() > 0 {
			if b, err := json.Marshal(reasoning.String()); err == nil {
				outReasoning = b
			}
		}
		return schema.OpenAIMessage{
			Role:      schema.OpenAIMessageRoleAssistant,
			Content:   content.String(),
			Reasoning: outReasoning,
			ToolCalls: toolCalls,
		}, nil
	}

	if raw.Len() > 0 {
		var response struct {
			Choices []struct {
				Message schema.OpenAIMessage `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(raw.String()), &response); err == nil {
			if len(response.Choices) == 0 {
				return schema.OpenAIMessage{}, fmt.Errorf("响应中没有选择项")
			}
			msg := response.Choices[0].Message
			if onDelta != nil && msg.Content != "" {
				if err := onDelta(msg.Content); err != nil {
					return msg, err
				}
			}
			if msg.Role == "" {
				msg.Role = schema.OpenAIMessageRoleAssistant
			}
			return msg, nil
		}
	}

	return schema.OpenAIMessage{}, fmt.Errorf("stream response is empty")
}

func decodeJSONString(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	if raw[0] != '"' {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", false
	}
	return s, true
}

// Embedding implements LLMAdapter.
func (b *BaseAdapter) Embedding(text []string) ([][]float32, error) {
	if b.config.Type != config.LLMTypeEmbedding {
		return nil, fmt.Errorf("适配器类型不是 embedding")
	}
	reqBody := map[string]any{
		"model":      b.config.Model,
		"input":      text,
		"dimensions": config.GlobalConfiguration().Database.Postgres.Dimension,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", b.config.APIUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Embeddings) == 0 {
		return nil, fmt.Errorf("响应中没有向量值")
	}
	embeddings := make([][]float32, 0, len(response.Embeddings))
	for _, item := range response.Embeddings {
		embeddings = append(embeddings, item)
	}
	return embeddings, nil
}

// GetConfig implements LLMAdapter.
func (b *BaseAdapter) GetConfig() *config.LLMConfig {
	return b.config
}

// NewLLMAdapter 如果BaseAdapter不兼容指定的API，可另行使用实现对应的请求功能
func NewLLMAdapter(c *config.LLMConfig) LLMAdapter {
	switch c.Type {
	case config.LLMTypeChat:
		return newChatAdapter(c)
	case config.LLMTypeThink:
		return newChatAdapter(c)
	case config.LLMTypeEmbedding:
		return newEmbeddingAdapter(c)
	default:
		return nil
	}
}

func CustomAdapter(c *config.LLMConfig, f func(c *config.LLMConfig) LLMAdapter) LLMAdapter {
	return f(c)
}

func newChatAdapter(c *config.LLMConfig) *BaseAdapter {
	return &BaseAdapter{
		config: c,
		mutex:  sync.RWMutex{},
		client: &http.Client{},
	}
}

func newEmbeddingAdapter(c *config.LLMConfig) *BaseAdapter {
	return &BaseAdapter{
		config: c,
		mutex:  sync.RWMutex{},
		client: &http.Client{},
	}
}

func (b *BaseAdapter) makeRequest(reqBody map[string]any) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", b.config.APIUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("响应中没有选择项")
	}

	return response.Choices[0].Message.Content, nil
}
