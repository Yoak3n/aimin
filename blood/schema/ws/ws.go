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
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Content: content,
	}
}

type AudioMessageData struct {
	TaskID     string `json:"task_id"`
	Format     string `json:"format"`
	Voice      string `json:"voice"`
	AudioBase64 string `json:"audio_base64"`
	Bytes      int    `json:"bytes"`
}

func NewAudioMessage(taskID, format, voice, audioBase64 string, bytes int) WebsocketMessage {
	return WebsocketMessage{
		Action: AudioMessage,
		Data: &AudioMessageData{
			TaskID:      taskID,
			Format:      format,
			Voice:       voice,
			AudioBase64: audioBase64,
			Bytes:       bytes,
		},
	}
}
