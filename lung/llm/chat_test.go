package llm

import (
	"testing"

	"github.com/Yoak3n/aimin/blood/schema"
)

type fakeChatter struct {
	gotMessages []schema.OpenAIMessage
	gotSystem   string
	out         string
	err         error
}

func (f *fakeChatter) Chat(userMessages []schema.OpenAIMessage, systemPrompt string) (string, error) {
	f.gotMessages = append([]schema.OpenAIMessage(nil), userMessages...)
	f.gotSystem = systemPrompt
	return f.out, f.err
}

type fakeStreamChatter struct {
	gotMessages []schema.OpenAIMessage
	gotTools    []schema.OpenAITool
	gotSystem   []string
	out         string
	err         error
}

func (f *fakeStreamChatter) ChatStream(userMessages []schema.OpenAIMessage, tools []schema.OpenAITool, _ func(string) error, systemPrompt ...string) (schema.OpenAIMessage, error) {
	f.gotMessages = append([]schema.OpenAIMessage(nil), userMessages...)
	f.gotTools = append([]schema.OpenAITool(nil), tools...)
	f.gotSystem = append([]string(nil), systemPrompt...)
	return schema.OpenAIMessage{Role: schema.OpenAIMessageRoleAssistant, Content: f.out}, f.err
}

func TestChatWith_PassesArgs(t *testing.T) {
	c := &fakeChatter{out: "ok"}
	msgs := []schema.OpenAIMessage{{Role: schema.OpenAIMessageRoleUser, Content: "hi"}}
	got, err := ChatWith(c, msgs, "sys")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "ok" {
		t.Fatalf("unexpected out: %q", got)
	}
	if c.gotSystem != "sys" || len(c.gotMessages) != 1 || c.gotMessages[0].Content != "hi" {
		t.Fatalf("unexpected args: system=%q messages=%#v", c.gotSystem, c.gotMessages)
	}
}

func TestChatStreamWith_PassesArgs(t *testing.T) {
	c := &fakeStreamChatter{out: "ok"}
	msgs := []schema.OpenAIMessage{{Role: schema.OpenAIMessageRoleUser, Content: "hi"}}
	got, err := ChatStreamWith(c, msgs, nil, nil, "sys")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Content != "ok" {
		t.Fatalf("unexpected out: %q", got.Content)
	}
	if len(c.gotSystem) != 1 || c.gotSystem[0] != "sys" || len(c.gotMessages) != 1 || c.gotMessages[0].Content != "hi" {
		t.Fatalf("unexpected args: system=%#v messages=%#v", c.gotSystem, c.gotMessages)
	}
}

