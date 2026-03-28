package schema

const (
	OpenAIMessageRoleUser      = "user"
	OpenAIMessageRoleAssistant = "assistant"
	OpenAIMessageRoleSystem    = "system"
)

type OpenAIMessageRole string

type OpenAIMessage struct {
	Role    OpenAIMessageRole `json:"role"`
	Content string            `json:"content"`
}

type ConversationData struct {
	From string `json:"from"`
	Id   string `json:"id"`
}
