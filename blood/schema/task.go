package schema

import (
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema/ws"
)

type TaskType int

const (
	TaskTypeChat TaskType = iota
	TaskTypeTranslate
)

type TaskData struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	Type    TaskType `json:"type"`
	Payload any      `json:"payload"`
}

func NewTaskMessage(taskType TaskType, payload any, from string) ws.WebsocketMessage {
	id := util.RandomIdWithPrefix("task")
	return ws.NewWebsocketMessage(ws.AddTaskMessage, &TaskData{
		ID:      id,
		Type:    taskType,
		Payload: payload,
		From:    from,
	})
}
