package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/schema"
	schemaws "github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/cerebrum/agent"
	"github.com/Yoak3n/aimin/hand/interactive"
)

type Input struct {
	question string
	id       string
	from     string
}

type Conversation struct {
	Id       string                 `json:"id"`
	From     string                 `json:"from"`
	Messages []schema.OpenAIMessage `json:"messages"`
	agent    *agent.ConversationAgent
}

type Manager struct {
	current         string
	conversationMap map[string]*Conversation
	data            chan Input
	timer           *time.Timer
	ctx             context.Context
	running         bool
	mu              sync.Mutex
}

var manager *Manager
var once sync.Once

func NewManager() *Manager {
	m := &Manager{
		current:         "",
		conversationMap: make(map[string]*Conversation),
		data:            make(chan Input),
		ctx:             context.Background(),
	}
	return m
}

func GetManager() *Manager {
	once.Do(func() {
		manager = NewManager()
	})
	return manager
}

func (m *Manager) setTimeout() {
	if m.timer != nil && m.timer.Stop() {
		m.timer.Reset(600 * time.Second)
	} else {
		m.timer = time.NewTimer(600 * time.Second)
	}
}

func (m *Manager) EntryConversation(conversationId string, from string, question string) {
	if !m.running {
		m.running = true
		m.setTimeout()
		go m.ConversationLoop()
	}
	m.data <- Input{
		question: question,
		id:       conversationId,
		from:     from,
	}
}

func (m *Manager) ConversationLoop() {
	for {
		select {
		case input := <-m.data:
			m.executeConversation(input)
		case <-m.ctx.Done():
			m.exitConversation()
			return
		case <-m.timer.C:
			m.exitConversation()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *Manager) executeConversation(data Input) {
	from := strings.TrimSpace(data.from)
	id := strings.TrimSpace(data.id)
	q := strings.TrimSpace(data.question)
	if from == "" || id == "" || q == "" {
		return
	}

	m.setTimeout()
	key := from + "|" + id
	m.mu.Lock()
	c := m.conversationMap[key]
	if c == nil || c.agent == nil {
		c = &Conversation{
			Id:    id,
			From:  from,
			agent: interactive.NewConversationTask(id, from),
		}
		m.conversationMap[key] = c
	}
	m.mu.Unlock()

	roundID, roundCtx, _ := interactive.BeginInterruptibleRound(from)
	_, err := c.agent.Ask(roundCtx, q)
	interactive.EndInterruptibleRound(from, roundID)
	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		msg := schemaws.NewReplyMessage(schemaws.ReplyStatusFinish, id, "[已打断]")
		b, _ := json.Marshal(msg)
		if interactive.WSReplyBroadcast != nil {
			interactive.WSReplyBroadcast(from, b)
		}
		return
	}
	msg := schemaws.NewReplyMessage(schemaws.ReplyStatusFinish, id, fmt.Sprintf("[错误] %v", err))
	b, _ := json.Marshal(msg)
	if interactive.WSReplyBroadcast != nil {
		interactive.WSReplyBroadcast(from, b)
	}
}

func (m *Manager) AskDirect(conversationId string, from string, question string) error {
	from = strings.TrimSpace(from)
	id := strings.TrimSpace(conversationId)
	q := strings.TrimSpace(question)
	if from == "" || id == "" || q == "" {
		return nil
	}

	key := from + "|" + id
	m.mu.Lock()
	c := m.conversationMap[key]
	if c == nil || c.agent == nil {
		c = &Conversation{
			Id:    id,
			From:  from,
			agent: interactive.NewConversationTask(id, from),
		}
		m.conversationMap[key] = c
	}
	m.mu.Unlock()

	ctx := interactive.InterruptContext(from)
	_, err := c.agent.Ask(ctx, q)
	return err
}

func (m *Manager) exitConversation() {
	m.timer.Stop()
	m.timer = nil
	m.running = false
}
