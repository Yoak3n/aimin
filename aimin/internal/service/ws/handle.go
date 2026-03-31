package ws

import (
	"github.com/gorilla/websocket"
)

func (wh *WebSocketHub) handle(pendingQuestions []*QuestionRequest) {
	var currentQuestion *QuestionRequest
	for {
		select {
		// 客户端注册事件
		case client := <-wh.register:
			wh.registerClient(client)
			// Attempt to send current or pending question to the new client
			if currentQuestion != nil {
				wh.sendQuestionToClient(currentQuestion, client)
			} else if len(pendingQuestions) > 0 {
				currentQuestion = pendingQuestions[0]
				pendingQuestions = pendingQuestions[1:]
				wh.sendQuestion(currentQuestion)
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
			pendingQuestions = append(pendingQuestions, req)
			wh.clientsMu.RLock()
			hasClient := len(wh.clients) > 0
			wh.clientsMu.RUnlock()
			if currentQuestion == nil && hasClient {
				currentQuestion = pendingQuestions[0]
				pendingQuestions = pendingQuestions[1:]
				wh.sendQuestion(currentQuestion)
			}
		// 主动交互回复事件
		case answer := <-wh.AnswerChan:
			if currentQuestion != nil {
				// Non-blocking send to avoid deadlock if receiver is gone
				select {
				case currentQuestion.AnswerCh <- answer:
				default:
				}
				currentQuestion = nil
				// Send next question if any
				wh.clientsMu.RLock()
				hasClient := len(wh.clients) > 0
				wh.clientsMu.RUnlock()
				if len(pendingQuestions) > 0 && hasClient {
					currentQuestion = pendingQuestions[0]
					pendingQuestions = pendingQuestions[1:]
					wh.sendQuestion(currentQuestion)
				}
			}
		case state := <-wh.State:
			wh.broadcastState(state)
		}
	}
}
