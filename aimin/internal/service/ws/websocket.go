package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/aimin/cmd/app/componet"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan string
	broadcast  chan []byte
	Tasks      chan TaskData
}

type Client struct {
	id   string
	conn *websocket.Conn
	last int64
	mu   sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan string),
		broadcast:  make(chan []byte),
	}
}

func (wh *WebSocketHub) Run() {
	go wh.sendTask()
	for {
		select {
		case client := <-wh.register:
			wh.clients[client.id] = client
			log.Printf("New connection with %d clients\n", len(wh.clients))
		case id := <-wh.unregister:
			if client, ok := wh.clients[id]; ok {
				cl := WebsocketMessage{
					Action: CloseMessage,
					Data:   "Close",
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
		}
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
			log.Println("check failed")
			wh.unregister <- client.id
			return
		}
		log.Println("check successfully")
		pingMessage := WebsocketMessage{
			Action: PingMessage,
			Data:   "ping",
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
		messageData := &WebsocketMessage{}
		err = json.Unmarshal(msg, messageData)
		if err != nil {
			break
		}
		switch messageData.Action {
		case CloseMessage:
			return
		case PingMessage:
			pongMessage := WebsocketMessage{
				Action: PongMessage,
				Data:   "pong",
			}
			conn.WriteJSON(pongMessage)
		case AddTaskMessage:
			// 无法直接断言为任务数据
			buf, _ := json.Marshal(messageData.Data)
			taskData := TaskData{}
			err := json.Unmarshal(buf, &taskData)
			if err != nil {
				log.Println("task data unmarshal err:", err)
			}
			wh.Tasks <- taskData
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
		state := fsm.NewTaskState(task.Id, task.Type, nil)
		componet.GetGlobalComponent().FSM().AddTask(state)
	}
}

func sendLog(conn *websocket.Conn, content string) {
	logItem := NewLogMessage(content)
	conn.WriteJSON(logItem)
}
