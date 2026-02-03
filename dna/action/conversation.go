package action

import (
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/face/conversation"
)

func EntryConversationTask(question string, conversationID string, from string) {
	record := schema.ConversationRecord{
		Id:    conversationID,
		From:  from,
		Topic: question,
	}
	err := helper.UseDB().CreateConversationRecord(record)
	if err != nil {
		return
	}
	m := conversation.GetManager()
	if m != nil {
		m.EntryConversation(conversationID, question)
	}else{
		logger.Logger.Infoln("conversation manager is invalid")
	}
	
}
