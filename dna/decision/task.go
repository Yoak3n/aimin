package decision

import (
	"blood/pkg/util"
	"dna/action"
	"dna/fsm"
)

const TaskDataKey = "task_data"

type TaskType string

const (
	Conversation TaskType = "conversation"
)

func NewTaskState() fsm.State {
	return fsm.NewTaskState(Task, Task, executeTask)
}

func executeTask(ctx *fsm.Context) {
	if d, exist := ctx.Data[TaskDataKey]; exist {
		data, ok := d.(fsm.TaskData)
		if ok {
			switch data.Type {
			case string(Conversation):
				conversationId := util.RandomIdWithPrefix("conversation-")
				action.EntryConversationTask(data.Payload.(string), conversationId)
			}
		}
	}
}
