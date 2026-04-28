package llm

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/adapter"
	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

type Chatter interface {
	Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error)
}

type StreamChatter interface {
	ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error)
}

type defaultChat struct{}

type pinnedStreamChat struct {
	adapter adapter.LLMAdapter
}

func (p pinnedStreamChat) ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	return p.adapter.ChatStream(userMessages, tools, onDelta, systemPrompt...)
}

func NewPinnedStreamChatter() (StreamChatter, error) {
	hub := helper.UseLLM()
	a, _, err := hub.PinAdapter(config.LLMTypeChat)
	if err != nil {
		return nil, err
	}
	return pinnedStreamChat{adapter: a}, nil
}

func (defaultChat) Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	return helper.UseLLM().Chat(userMessages, systemPrompt)
}

func (defaultChat) ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	return helper.UseLLM().ChatStreamWithTools(userMessages, tools, onDelta, systemPrompt...)
}

func Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	return ChatWith(defaultChat{}, userMessages, systemPrompt)
}

func ChatWith(chatter Chatter, userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	if chatter == nil {
		return "", fmt.Errorf("chatter is nil")
	}
	return chatter.Chat(userMessages, systemPrompt)
}

func ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error) {
	msg, err := ChatStreamWith(defaultChat{}, userMessages, nil, onDelta, systemPrompt...)
	if err != nil {
		return msg.Content, err
	}
	return msg.Content, nil
}

func ChatStreamWith(chater StreamChatter, userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	if chater == nil {
		return schema.OpenAIMessage{}, fmt.Errorf("stream chatter is nil")
	}
	return chater.ChatStream(userMessages, tools, onDelta, systemPrompt...)
}

func ChatStreamWithTools(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	return ChatStreamWith(defaultChat{}, userMessages, tools, onDelta, systemPrompt...)
}
