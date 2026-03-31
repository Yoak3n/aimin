package ws

import (
	"log"

	"github.com/Yoak3n/aimin/blood/schema/ws"
)

func (wh *WebSocketHub) registerClient(client *Client) {
	wh.clientsMu.Lock()
	wh.clients[client.id] = client
	clientCount := len(wh.clients)
	wh.clientsMu.Unlock()
	log.Printf("New connection with %d clients\n", clientCount)

}

func (wh *WebSocketHub) unregisterClient(id string) {
	wh.clientsMu.Lock()
	client, ok := wh.clients[id]
	if ok {
		delete(wh.clients, id)
	}
	clientCount := len(wh.clients)
	wh.clientsMu.Unlock()
	if ok {
		cl := ws.WebsocketMessage{
			Action: ws.CloseMessage,
			Data:   ws.CloseMessage,
		}
		client.mu.Lock()
		client.conn.WriteJSON(cl)
		client.conn.Close()
		client.mu.Unlock()
		log.Printf("Client disconnected with %d clients\n", clientCount)
	}
}
