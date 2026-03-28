package decision

import (
	"encoding/json"
	"fmt"

	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/dna/fsm"
)

type TaskType string

const (
	TaskDataKey                   = "task_data"
	ConversationCreate   TaskType = "conversation_create"
	ConversationContinue TaskType = "conversation_continue"
)

func NewTaskState() fsm.State {
	return fsm.NewTaskState(Task, Task, executeTask)
}

func executeTask(ctx *fsm.Context) {
	if d, exist := ctx.Data[TaskDataKey]; exist {
		// TODO: 根据任务类型执行不同的任务，需要有个地方传入任务数据
		data, ok := d.(fsm.TaskData)
		if ok {
			switch data.Type {
			case string(ConversationCreate):
				// conversationId := util.RandomIdWithPrefix("conversation")
				// action.EntryConversationTask(data.Payload.(string), conversationId, data.ID)
			case string(ConversationContinue):
				var payload schema.ConversationContinuePayload
				if p, ok := data.Payload.(schema.ConversationContinuePayload); ok {
					payload = p
				} else {
					bytes, err := json.Marshal(data.Payload)
					if err != nil {
						fmt.Println("Error marshaling payload:", err)
						return
					}
					err = json.Unmarshal(bytes, &payload)
					if err != nil {
						fmt.Println("Error unmarshaling payload:", err)
						return
					}
				}
				// action.EntryConversationTask(payload.Question, payload.ConversationId, data.ID)
			}
		}
	}
}
