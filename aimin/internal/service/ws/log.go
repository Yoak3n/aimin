package ws

import (
	"encoding/json"

	"github.com/Yoak3n/aimin/blood/schema/ws"
)

func (wh *WebSocketHub) BroadcastLog(content string) {
	logItem := ws.NewLogMessage(content)
	buf, _ := json.Marshal(logItem)
	wh.Broadcast(buf)
}
