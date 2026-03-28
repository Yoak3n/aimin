package conversation

import (
	"context"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/schema"
)

type Input struct {
	question string
	id       string
}

type Conversation struct {
	Id           string                 `json:"id"`
	Messages     []schema.OpenAIMessage `json:"messages"`
	systemPrompt string
}

type Manager struct {
	current         string
	conversationMap map[string]*Conversation
	data            chan Input
	timer           *time.Timer
	ctx             context.Context
	running         bool
	replyHandler    func(id string, content string)
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

func (m *Manager) SetReplyHandler(h func(id string, content string)) {
	m.replyHandler = h
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

func (m *Manager) EntryConversation(conversationId string, question string) {
	if !m.running {
		m.running = true
		m.setTimeout()
		go m.ConversationLoop()
	}
	m.data <- Input{
		question: question,
		id:       conversationId,
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

}

func (m *Manager) exitConversation() {
	m.timer.Stop()
	m.timer = nil
	m.running = false
}
