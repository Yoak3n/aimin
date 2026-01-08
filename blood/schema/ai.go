package schema

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationData struct {
	From string `json:"from"`
	Id   string `json:"id"`
}
