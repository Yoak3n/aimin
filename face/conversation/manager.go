package conversation

import (
	"blood/dao/controller"
	"blood/pkg/helper"
	"blood/schema"
	"errors"
	"nerve/reason"
	"sync"

	"gorm.io/gorm"
)

type Conversation struct {
	Id           string                 `json:"id"`
	Messages     []schema.OpenAIMessage `json:"messages"`
	systemPrompt string
}

type Manager struct {
	current         string
	conversationMap map[string]*Conversation
}

var manager *Manager
var once sync.Once

func NewManager() *Manager {
	return &Manager{
		current:         "",
		conversationMap: make(map[string]*Conversation),
	}
}

func GetManager() *Manager {
	once.Do(func() {
		manager = NewManager()
	})
	return manager
}

func (m *Manager) EntryConversation(conversationId string, question string) {
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
