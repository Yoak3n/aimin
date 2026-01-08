package fsm

type TaskData struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Payload  any    `json:"payload"`
}
