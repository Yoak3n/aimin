package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/aimin/cmd/app/componet"
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
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	broadcast  chan []byte
	Tasks      chan schema.TaskData
	Answer     chan string
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
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan string),
		broadcast:  make(chan []byte),
		Tasks:      make(chan schema.TaskData),
		Answer:     make(chan string),
		AskChan:    make(chan *QuestionRequest),
		State:      make(chan string),
	}
}

func (wh *WebSocketHub) Run() {
	go wh.sendTask()
	var pendingQuestions []*QuestionRequest
	var currentQuestion *QuestionRequest

	for {
		select {
		case client := <-wh.register:
			wh.clients[client.id] = client
			log.Printf("New connection with %d clients\n", len(wh.clients))
			// Attempt to send current or pending question to the new client
			if currentQuestion != nil {
				wh.sendQuestionToClient(currentQuestion, client)
			} else if len(pendingQuestions) > 0 {
				currentQuestion = pendingQuestions[0]
				pendingQuestions = pendingQuestions[1:]
				wh.sendQuestion(currentQuestion)
			}

		case id := <-wh.unregister:
			if client, ok := wh.clients[id]; ok {
				cl := ws.WebsocketMessage{
					Action: ws.CloseMessage,
					Data:   ws.CloseMessage,
				}
				client.mu.Lock()
				client.conn.WriteJSON(cl)
				delete(wh.clients, id)
				client.conn.Close()
				client.mu.Unlock()
				log.Printf("Client disconnected with %d clients\n", len(wh.clients))
			}
		case message := <-wh.broadcast:
			for k, client := range wh.clients {
				if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					client.conn.Close()
					delete(wh.clients, k)
				}
			}
		case req := <-wh.AskChan:
			pendingQuestions = append(pendingQuestions, req)
			if currentQuestion == nil && len(wh.clients) > 0 {
				currentQuestion = pendingQuestions[0]
				pendingQuestions = pendingQuestions[1:]
				wh.sendQuestion(currentQuestion)
			}
		case answer := <-wh.Answer:
			if currentQuestion != nil {
				// Non-blocking send to avoid deadlock if receiver is gone
				select {
				case currentQuestion.AnswerCh <- answer:
				default:
				}
				currentQuestion = nil
				// Send next question if any
				if len(pendingQuestions) > 0 && len(wh.clients) > 0 {
					currentQuestion = pendingQuestions[0]
					pendingQuestions = pendingQuestions[1:]
					wh.sendQuestion(currentQuestion)
				}
			}
		case state := <-wh.State:
			msg := ws.WebsocketMessage{
				Action: ws.StateMessage,
				Data:   state,
			}
			buf, _ := json.Marshal(msg)
			for k, client := range wh.clients {
				client.mu.Lock()
				if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
					log.Println("Error broadcasting state to client:", err)
					client.conn.Close()
					delete(wh.clients, k)
				}
				client.mu.Unlock()
			}
		}
	}
}

func (wh *WebSocketHub) sendQuestion(req *QuestionRequest) {
	msg := ws.WebsocketMessage{
		Action: ws.AskMessage,
		Data:   req.Content,
	}
	buf, _ := json.Marshal(msg)
	// Broadcast to all clients (or specific logic if needed)
	// Currently using broadcast mechanism within Hub logic
	for k, client := range wh.clients {
		if err := client.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			client.conn.Close()
			delete(wh.clients, k)
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
	if k, ok := wh.clients[id]; ok {
		k.conn.Close()
		delete(wh.clients, id)
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
			wh.Tasks <- taskData
		case ws.AnswerMessage:
			if ans, ok := messageData.Data.(string); ok {
				wh.Answer <- ans
			} else {
				// Try to convert if it's not directly a string (e.g. interface{})
				wh.Answer <- fmt.Sprint(messageData.Data)
			}
		}
		client, ok := wh.clients[id]
		if ok {
			client.mu.Lock()
			client.last = time.Now().Unix()
			client.mu.Unlock()
			wh.clients[id] = client
		}
	}
}

func (wh *WebSocketHub) sendTask() {
	for task := range wh.Tasks {
		fsmTask := fsm.TaskData{
			ID:   task.Id,
			Type: task.Type,
			Name: task.Id,
			// 后续看情况处理任务优先级
			Priority: 5,
			// 需要放入任务的负载数据
			Payload: task.Payload,
		}
		componet.GetGlobalComponent().AddTask(fsmTask)
	}
}

func sendLog(conn *websocket.Conn, content string) {
	logItem := ws.NewLogMessage(content)
	conn.WriteJSON(logItem)
}

func (wh *WebSocketHub) Broadcast(message []byte) {
	wh.broadcast <- message
}

func (wh *WebSocketHub) BroadcastLog(content string) {
	logItem := ws.NewLogMessage(content)
	buf, _ := json.Marshal(logItem)
	wh.Broadcast(buf)
}

func (wh *WebSocketHub) Ask(ctx context.Context, question string) []string {
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
		return []string{}
	}
}
