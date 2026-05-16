package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/hand/interactive"
	"github.com/gorilla/websocket"
)

type QuestionRequest struct {
	ID       string
	Content  string
	AnswerCh chan QuestionResult
	Ctx      context.Context
}

type QuestionResult struct {
	Answer  string
	Skipped bool
}

type AnswerPayload struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Skip    bool   `json:"skip"`
}

func (wh *WebSocketHub) sendQuestionToClient(req *QuestionRequest, client *Client) {
	msg := ws.WebsocketMessage{
		Action: ws.AskMessage,
		Data: map[string]any{
			"id":      req.ID,
			"content": req.Content,
		},
	}
	buf, _ := json.Marshal(msg)
	client.mu.Lock()
	err := client.conn.WriteMessage(websocket.TextMessage, buf)
	client.mu.Unlock()
	if err != nil {
		log.Println("Error sending question to new client:", err)
	}
}

func (wh *WebSocketHub) SendToClient(id string, message []byte) {
	if id == "" {
		return
	}
	wh.clientsMu.RLock()
	client, ok := wh.clients[id]
	wh.clientsMu.RUnlock()
	if !ok || client == nil {
		return
	}
	client.mu.Lock()
	err := client.conn.WriteMessage(websocket.TextMessage, message)
	client.mu.Unlock()
	if err != nil {
		client.conn.Close()
		wh.clientsMu.Lock()
		delete(wh.clients, id)
		wh.clientsMu.Unlock()
		interactive.RequestInterrupt(id)
	}
}



func (wh *WebSocketHub) AskClient(ctx context.Context, question string) []string {
	wh.clientsMu.RLock()
	hasClient := len(wh.clients) > 0
	wh.clientsMu.RUnlock()
	if !hasClient {
		return []string{"[AskUser][无客户端] 当前没有连接的 WebSocket 客户端，无法向用户提问。"}
	}

	answerCh := make(chan QuestionResult, 1)
	req := &QuestionRequest{
		ID:       util.RandomIdWithPrefix("qst"),
		Content:  question,
		AnswerCh: answerCh,
		Ctx:      ctx,
	}
	wh.AskChan <- req

	select {
	case res := <-answerCh:
		if res.Skipped {
			return nil
		}
		if strings.TrimSpace(res.Answer) == "" {
			return nil
		}
		return []string{res.Answer}
	case <-ctx.Done():
		select {
		case wh.CancelAsk <- req.ID:
		default:
		}
		return nil
	}
}
