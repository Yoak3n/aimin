package llm

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

type Chatter interface {
	Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error)
}

type StreamChatter interface {
	ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error)
}

type defaultChat struct{}

func (defaultChat) Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	return helper.UseLLM().Chat(userMessages, systemPrompt)
}

func (defaultChat) ChatStream(userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error) {
	return helper.UseLLM().ChatStream(userMessages, onDelta, systemPrompt...)
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
	return ChatStreamWith(defaultChat{}, userMessages, onDelta, systemPrompt...)
}

func ChatStreamWith(chatter StreamChatter, userMessages []schema.OpenAIMessage, onDelta func(string) error, systemPrompt ...string) (string, error) {
	if chatter == nil {
		return "", fmt.Errorf("stream chatter is nil")
	}
	return chatter.ChatStream(userMessages, onDelta, systemPrompt...)
}
