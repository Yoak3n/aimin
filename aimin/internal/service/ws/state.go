package ws

import (
	"encoding/json"
	"log"

	"github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/gorilla/websocket"
)

func (wh *WebSocketHub) BroadcastState(state string) {
	select {
	case wh.State <- state:
	default:
		log.Println("State channel full or no receiver, skipping state update:", state)
	}
}

func (wh *WebSocketHub) broadcastState(state string) {
	msg := ws.WebsocketMessage{
		Action: ws.StateMessage,
		Data:   state,
	}
	buf, _ := json.Marshal(msg)
	wh.clientsMu.RLock()
	ids := make([]string, 0, len(wh.clients))
	clients := make([]*Client, 0, len(wh.clients))
	for k, client := range wh.clients {
		ids = append(ids, k)
		clients = append(clients, client)
	}
	wh.clientsMu.RUnlock()
	for i, client := range clients {
		k := ids[i]
		client.mu.Lock()
		if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			log.Println("Error broadcasting state to client:", err)
			client.conn.Close()
			wh.clientsMu.Lock()
			delete(wh.clients, k)
			wh.clientsMu.Unlock()
		}
		client.mu.Unlock()
	}
}
