package ws

type ActionType string

const (
	ConnectedSuccess ActionType = "Connected"
	LogMessage       ActionType = "Log"

	CloseMessage ActionType = "Close"
	PingMessage  ActionType = "Ping"
	PongMessage  ActionType = "Pong"
	// agent主动提问通道
	AskMessage ActionType = "Ask"
	// agent主动提问获得回答通道
	AnswerMessage ActionType = "Answer"
	// 用户添加任务（发送消息）通道
	AddTaskMessage ActionType = "Task"
	// 用户添加任务（接收消息）通道
	ReplyMessage ActionType = "Reply"
	StateMessage ActionType = "State"
)

func NewWebsocketMessage(action ActionType, data any) WebsocketMessage {
	return WebsocketMessage{
		Action: action,
		Data:   data,
	}
}
