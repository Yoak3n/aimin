package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

func (wh *WebSocketHub) handle() {
	pendingMu := sync.Mutex{}
	pending := map[string]*QuestionRequest{}
	for {
		select {
		// 客户端注册事件
		case client := <-wh.register:
			wh.registerClient(client)
			pendingMu.Lock()
			items := make([]*QuestionRequest, 0, len(pending))
			for _, q := range pending {
				items = append(items, q)
			}
			pendingMu.Unlock()
			for _, q := range items {
				wh.sendQuestionToClient(q, client)
			}
		// 客户端注销事件
		case id := <-wh.unregister:
			wh.unregisterClient(id)

		// 广播事件
		case message := <-wh.broadcast:
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
				if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					client.conn.Close()
					wh.clientsMu.Lock()
					delete(wh.clients, k)
					wh.clientsMu.Unlock()
				}
			}
		// 主动交互提问事件
		case req := <-wh.AskChan:
			if req == nil || req.ID == "" {
				continue
			}
			pendingMu.Lock()
			pending[req.ID] = req
			pendingMu.Unlock()
			wh.sendQuestion(req)
		// 主动交互回复事件
		case ans := <-wh.AnswerChan:
			id := ans.ID
			if id == "" {
				continue
			}
			pendingMu.Lock()
			req, ok := pending[id]
			if ok {
				delete(pending, id)
			}
			pendingMu.Unlock()
			if ok && req != nil {
				select {
				case req.AnswerCh <- QuestionResult{Answer: ans.Content, Skipped: ans.Skip}:
				default:
				}
			}
		case id := <-wh.CancelAsk:
			if id == "" {
				continue
			}
			pendingMu.Lock()
			delete(pending, id)
			pendingMu.Unlock()
		case state := <-wh.State:
			wh.broadcastState(state)
		}
	}
}
