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
