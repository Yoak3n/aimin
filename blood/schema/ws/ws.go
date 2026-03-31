package ws

import "time"

type WebsocketMessage struct {
	Action ActionType `json:"action"`
	Data   any        `json:"data"`
}


type LogMessageData struct {
	Time    string `json:"time"`
	Content string `json:"content"`
}

func NewLogMessage(content string) WebsocketMessage {
	return WebsocketMessage{
		Action: LogMessage,
		Data:   NewLogMessageData(content),
	}
}

func NewLogMessageData(content string) LogMessageData {
	return LogMessageData{
		Time:    time.Now().Format("2006-01-02 15:04"),
		Content: content,
	}
}
