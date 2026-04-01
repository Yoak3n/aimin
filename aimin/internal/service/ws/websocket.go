package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/aimin/app/componet"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/gorilla/websocket"
)

type QuestionRequest struct {
	Content  string
	AnswerCh chan string
	Ctx      context.Context
}

type WebSocketHub struct {
	clientsMu  sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	broadcast  chan []byte
	// 任务发送通道
	Tasks      chan schema.TaskData
	AnswerChan chan string
	AskChan    chan *QuestionRequest
	State      chan string
}

type Client struct {
	id   string
	conn *websocket.Conn
	last int64
	mu   sync.RWMutex
}

var hub *WebSocketHub
var once sync.Once

func InitWebSocketHub() {
	if hub != nil {
		return
	}
	hub = NewWebSocketHub()
}

func UseWebSocketHub() *WebSocketHub {
	once.Do(func() {
		hub = NewWebSocketHub()
	})
	return hub
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clientsMu:  sync.RWMutex{},
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan string),
		broadcast:  make(chan []byte),
		Tasks:      make(chan schema.TaskData, 100),
		AnswerChan: make(chan string),
		AskChan:    make(chan *QuestionRequest),
		State:      make(chan string, 1000),
	}
}

func (wh *WebSocketHub) Run() {
	go wh.sendTask()
	pendingQuestions := make([]*QuestionRequest, 0, 100)
	go wh.handle(pendingQuestions)

}

func (wh *WebSocketHub) sendQuestion(req *QuestionRequest) {
	msg := ws.WebsocketMessage{
		Action: ws.AskMessage,
		Data:   req.Content,
	}
	buf, _ := json.Marshal(msg)
	// Broadcast to all clients (or specific logic if needed)
	// Currently using broadcast mechanism within Hub logic
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
		if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			client.conn.Close()
			wh.clientsMu.Lock()
			delete(wh.clients, k)
			wh.clientsMu.Unlock()
		}
	}
}

func (wh *WebSocketHub) sendQuestionToClient(req *QuestionRequest, client *Client) {
	msg := ws.WebsocketMessage{
		Action: ws.AskMessage,
		Data:   req.Content,
	}
	buf, _ := json.Marshal(msg)
	if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
		// handle error
		log.Println("Error sending question to new client:", err)
	}
}

func (wh *WebSocketHub) Register(id string, conn *websocket.Conn) {
	wh.clientsMu.Lock()
	existing, ok := wh.clients[id]
	if ok {
		delete(wh.clients, id)
	}
	wh.clientsMu.Unlock()
	if ok {
		existing.conn.Close()
	}
	client := &Client{
		id:   id,
		conn: conn,
		mu:   sync.RWMutex{},
		last: time.Now().Unix(),
	}
	wh.register <- client
	sendLog(client.conn, "Connected successfully")
	go wh.healthCheck(client)
	wh.listen(id, conn)
}

func (wh *WebSocketHub) healthCheck(client *Client) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		wh.clientsMu.RLock()
		client.mu.RLock()
		last := client.last
		now := time.Now().Unix()
		if (now - last) >= 180 {
			wh.unregister <- client.id
			return
		}
		pingMessage := ws.WebsocketMessage{
			Action: ws.PingMessage,
			Data:   ws.PingMessage,
		}
		client.conn.WriteJSON(pingMessage)
		client.mu.RUnlock()
		wh.clientsMu.RUnlock()
	}
}

func (wh *WebSocketHub) listen(id string, conn *websocket.Conn) {
	defer func() {
		wh.unregister <- id
	}()
	for {
		t, msg, err := conn.ReadMessage()
		log.Println("T", t, id, string(msg))
		if err != nil || t == -1 {
			break
		}
		messageData := &ws.WebsocketMessage{}
		err = json.Unmarshal(msg, messageData)
		if err != nil {
			break
		}
		switch messageData.Action {
		case ws.CloseMessage:
			return
		case ws.PingMessage:
			pongMessage := ws.WebsocketMessage{
				Action: ws.PongMessage,
				Data:   ws.PongMessage,
			}
			conn.WriteJSON(pongMessage)
		case ws.AddTaskMessage:
			// 无法直接断言为任务数据
			buf, _ := json.Marshal(messageData.Data)
			taskData := schema.TaskData{}
			err := json.Unmarshal(buf, &taskData)
			if err != nil {
				log.Println("task data unmarshal err:", err)
			}
			if taskData.ID == "" {
				taskData.ID = util.RandomIdWithPrefix("tsk")
			}
			logger.Logger.Infof("AddTaskMessage: %v\n", taskData)
			wh.Tasks <- taskData
		case ws.AnswerMessage:
			if ans, ok := messageData.Data.(string); ok {
				wh.AnswerChan <- ans
			} else {
				// Try to convert if it's not directly a string (e.g. interface{})
				wh.AnswerChan <- fmt.Sprint(messageData.Data)
			}
		}
		wh.clientsMu.RLock()
		client, ok := wh.clients[id]
		wh.clientsMu.RUnlock()
		if ok {
			client.mu.Lock()
			client.last = time.Now().Unix()
			client.mu.Unlock()
		}
	}
}

func (wh *WebSocketHub) sendTask() {
	for task := range wh.Tasks {
		if wh.tryHandleConversationTask(task) {
			continue
		}
		fsmTask := fsm.TaskData{
			ID:   task.ID,
			Type: int(task.Type),
			// 需要放入任务的负载数据
			Payload: task.Payload,
			From:    task.From,
		}
		componet.GetGlobalComponent().AddTask(fsmTask)
	}
}

func (wh *WebSocketHub) tryHandleConversationTask(task schema.TaskData) bool {

	question := ""
	if s, ok := task.Payload.(string); ok {
		question = s
	}
	if question == "" {
		wh.BroadcastLog(fmt.Sprintf("[Task][%d] payload 缺少 question", task.Type))
		return true
	}

	return false
}

func sendLog(conn *websocket.Conn, content string) {
	logItem := ws.NewLogMessage(content)
	conn.WriteJSON(logItem)
}

func (wh *WebSocketHub) Broadcast(message []byte) {
	wh.broadcast <- message
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
	}
}

func (wh *WebSocketHub) BroadcastLog(content string) {
	logItem := ws.NewLogMessage(content)
	buf, _ := json.Marshal(logItem)
	wh.Broadcast(buf)
}

func (wh *WebSocketHub) Ask(ctx context.Context, question string) []string {
	wh.clientsMu.RLock()
	hasClient := len(wh.clients) > 0
	wh.clientsMu.RUnlock()
	if !hasClient {
		return []string{"[AskUser][无客户端] 当前没有连接的 WebSocket 客户端，无法向用户提问。"}
	}

	answerCh := make(chan string)
	req := &QuestionRequest{
		Content:  question,
		AnswerCh: answerCh,
		Ctx:      ctx,
	}
	wh.AskChan <- req

	select {
	case answer := <-answerCh:
		return []string{answer}
	case <-ctx.Done():
		return nil
	}
}
