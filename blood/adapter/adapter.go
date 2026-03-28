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
	ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error)
	Embedding(text []string) ([][]float32, error)
	GetConfig() *config.LLMConfig
}

type BaseAdapter struct {
	config *config.LLMConfig
	mutex  sync.RWMutex
	client *http.Client
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
	messages := make([]map[string]string, 0, len(userMessages)+1)
	messages = append(messages, map[string]string{
		"role":    string(schema.OpenAIMessageRoleSystem),
		"content": systemPrompt,
	})
	for _, msg := range userMessages {
		messages = append(messages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	reqBody := map[string]any{
		"model":    b.config.Model,
		"messages": messages,
		"stream":   false,
	}

	return b.makeRequest(reqBody)
}

func (b *BaseAdapter) ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemMessage ...string) (string, error) {
	systemPrompt := ""
	if len(systemMessage) == 0 {
		systemPrompt = defaultSystemPrompt
	} else {
		systemPrompt = systemMessage[0]
	}

	if len(userMessages) == 0 {
		return "", fmt.Errorf("用户消息不能为空")
	}

	messages := make([]map[string]string, 0, len(userMessages)+1)
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": systemPrompt,
	})
	for _, msg := range userMessages {
		messages = append(messages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	reqBody := map[string]any{
		"model":    b.config.Model,
		"messages": messages,
		"stream":   true,
	}

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
	req.Header.Set("Accept", "text/event-stream")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
		}
		return "", fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	var ret strings.Builder
	var raw strings.Builder

	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				return ret.String(), readErr
			}
			continue
		}

		if strings.HasPrefix(line, "data:") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				break
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
					Text string `json:"text"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				raw.WriteString(payload)
			} else if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta == "" {
					delta = chunk.Choices[0].Message.Content
				}
				if delta == "" {
					delta = chunk.Choices[0].Text
				}

				if delta != "" {
					ret.WriteString(delta)
					if onDelta != nil {
						if err := onDelta(delta); err != nil {
							return ret.String(), err
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
			return ret.String(), readErr
		}
	}

	if ret.Len() > 0 {
		return ret.String(), nil
	}

	if raw.Len() > 0 {
		var response struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(raw.String()), &response); err == nil {
			if len(response.Choices) == 0 {
				return "", fmt.Errorf("响应中没有选择项")
			}
			content := response.Choices[0].Message.Content
			if onDelta != nil && content != "" {
				if err := onDelta(content); err != nil {
					return content, err
				}
			}
			return content, nil
		}
	}

	return "", fmt.Errorf("stream response is empty")
}

// Embedding implements LLMAdapter.
func (b *BaseAdapter) Embedding(text []string) ([][]float32, error) {
	if b.config.Type != config.LLMTypeEmbedding {
		return nil, fmt.Errorf("适配器类型不是 embedding")
	}
	reqBody := map[string]any{
		"model": b.config.Model,
		"input": text,
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
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("响应中没有向量值")
	}
	embeddings := make([][]float32, 0, len(response.Data))
	for _, item := range response.Data {
		embeddings = append(embeddings, item.Embedding)
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
