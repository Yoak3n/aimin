package fsm

type TaskData struct {
	ID      string `json:"id"`
	Type    int    `json:"type"`
	Payload any    `json:"payload"`
	From    string `json:"from"`
}
