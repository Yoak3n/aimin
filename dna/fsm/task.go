package fsm

type TaskData struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Payload  any    `json:"payload"`
}
