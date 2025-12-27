package schema

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
