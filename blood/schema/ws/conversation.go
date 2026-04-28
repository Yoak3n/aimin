package ws

type ReplyStatus int

const (
	ReplyStatusProcessing ReplyStatus = iota
	ReplyStatusFinish
)

func NewReplyMessage(status ReplyStatus, taskID, content string) WebsocketMessage {
	switch status {
	case ReplyStatusProcessing:
		msg := WebsocketMessage{
			Action: ReplyMessage,
			Data: &ReplyMessageData{
				TaskID: taskID,
				Status: status,
				Chunk: &ReplyChunkData{
					TaskID:  taskID,
					Content: content,
				},
			},
		}
		return msg

	case ReplyStatusFinish:
		msg := WebsocketMessage{
			Action: ReplyMessage,
			Data: &ReplyMessageData{
				TaskID: taskID,
				Status: status,
				Result: &ReplyFinishData{
					TaskID:  taskID,
					Content: content,
				},
			},
		}
		return msg
	default:
		return NewLogMessage("Unknown")
	}
}

type ReplyMessageData struct {
	TaskID string           `json:"task_id"`
	Status ReplyStatus      `json:"status"`
	Chunk  *ReplyChunkData  `json:"chunk,omitempty"`
	Result *ReplyFinishData `json:"result,omitempty"`
}

type ReplyChunkData struct {
	TaskID   string `json:"task_id"`
	ChunkIdx int    `json:"chunk_idx"`
	Content  string `json:"content"`
}

type ReplyFinishData struct {
	TaskID  string `json:"task_id"`
	Content string `json:"content"`
}

type ToolResultMessageData struct {
	TaskID     string `json:"task_id"`
	ToolCallID string `json:"tool_call_id"`
	Action     string `json:"action"`
	Result     string `json:"result"`
	Error      string `json:"error,omitempty"`
	HasError   bool   `json:"has_error"`
	FinishedAt int64  `json:"finished_at,omitempty"`
}

func NewToolResultMessage(taskID, toolCallID, action, result, errMsg string) WebsocketMessage {
	return WebsocketMessage{
		Action: ToolResultMessage,
		Data: &ToolResultMessageData{
			TaskID:     taskID,
			ToolCallID: toolCallID,
			Action:     action,
			Result:     result,
			Error:      errMsg,
			HasError:   errMsg != "",
		},
	}
}
