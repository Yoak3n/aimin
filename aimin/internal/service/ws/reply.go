package ws

import (
	"encoding/json"
	"log"

	"github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/gorilla/websocket"
)

func (wh *WebSocketHub) SendReply(id string, content string) {
	msg := ws.WebsocketMessage{
		Action: ws.ReplyMessage,
		Data: map[string]string{
			"conversation_id": id,
			"content":         content,
		},
	}
	buf, _ := json.Marshal(msg)
	// Broadcast to all clients
	for k, client := range wh.clients {
		if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			log.Println("Error sending reply to client:", err)
			client.conn.Close()
			delete(wh.clients, k)
		}
	}
}
