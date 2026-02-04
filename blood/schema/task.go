package schema

type TaskData struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type ConversationContinuePayload struct {
	ConversationId string `json:"conversation_id"`
	Question       string `json:"question"`
}
