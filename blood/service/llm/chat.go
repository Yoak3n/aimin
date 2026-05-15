package llm

import (
	"context"

	"github.com/Yoak3n/aimin/blood/schema"
	inner "github.com/Yoak3n/aimin/lung/llm"
)

type Chatter = inner.Chatter
type StreamChatter = inner.StreamChatter

func NewPinnedStreamChatter() (StreamChatter, error) { return inner.NewPinnedStreamChatter() }

func Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	return inner.Chat(userMessages, systemPrompt)
}

func ChatWith(chatter Chatter, userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	return inner.ChatWith(chatter, userMessages, systemPrompt)
}

func ChatStream(ctx context.Context, userMessages []schema.OpenAIMessage, onDelta func(string, string) error, systemPrompt ...string) (string, error) {
	return inner.ChatStream(ctx, userMessages, onDelta, systemPrompt...)
}

func ChatStreamWithTools(ctx context.Context, userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, onDelta func(string, string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	return inner.ChatStreamWithTools(ctx, userMessages, tools, onDelta, systemPrompt...)
}
