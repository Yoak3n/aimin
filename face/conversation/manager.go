package conversation

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/dao/controller"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/nerve/reason"

	"gorm.io/gorm"
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

func (m *Manager) EntryConversation(conversationId string, question string) {
	if !m.running {
		m.running = true
		go m.ConversationLoop()
	}
	m.data <- Input{
		question: question,
		id:       conversationId,
	}
	m.setTimeout()

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
	conversationId := data.id
	question := data.question
	questionMessage := schema.OpenAIMessage{Role: "user", Content: question}
	if _, exist := m.conversationMap[conversationId]; !exist {
		conversationRecord, err := helper.UseDB().GetConversationRecord(conversationId)
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			delete(m.conversationMap, m.current)
			systemPrompt := reason.GenConversationSystemPrompt()
			m.conversationMap[conversationId] = &Conversation{
				Id:           conversationId,
				Messages:     []schema.OpenAIMessage{questionMessage},
				systemPrompt: systemPrompt,
			}
		} else {
			// 旧有对话从数据库中读取
			records, err := controller.GetDialoguesWithConversation(conversationId)
			if err != nil {
				return
			}
			var openAIMessages []schema.OpenAIMessage
			for idx, record := range records {
				if idx == len(records)-1 && record.Role == "user" {
					continue
				}
				openAIMessages = append(openAIMessages, schema.OpenAIMessage{
					Role:    record.Role,
					Content: record.Content,
				})
			}
			m.conversationMap[conversationId] = &Conversation{
				Id:           conversationId,
				Messages:     openAIMessages,
				systemPrompt: conversationRecord.SystemPrompt,
			}
		}
		// 切换当前对话到新对话，不让map继续扩充减少内存占用
	} else {
		m.conversationMap[conversationId].Messages = append(m.conversationMap[conversationId].Messages, questionMessage)
	}
	err := controller.CreateDialogueWithConversation(questionMessage, conversationId)
	if err != nil {
		return
	}
	m.current = conversationId

	// 之后再包一层思考层
	answer, err := helper.UseLLM().Chat(m.conversationMap[conversationId].Messages, m.conversationMap[conversationId].systemPrompt)
	if err != nil {
		return
	}
	answerMessage := schema.OpenAIMessage{Role: "assistant", Content: answer}
	m.conversationMap[conversationId].Messages = append(m.conversationMap[conversationId].Messages, answerMessage)

	err = controller.CreateDialogueWithConversation(answerMessage, conversationId)
	if err != nil {
		return
	}
}

func (m *Manager) exitConversation() {
	m.timer.Stop()
	m.timer = nil
	m.running = false
}
