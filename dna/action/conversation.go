package action

import (
	"blood/pkg/helper"
	"blood/schema"
	"face/conversation"
)

func EntryConversationTask(question string, conversationID string) {
	record := schema.ConversationRecord{
		Id:    conversationID,
		From:  "task",
		Topic: question,
	}
	err := helper.UseDB().CreateConversationRecord(record)
	if err != nil {
		return
	}
	m := conversation.GetManager()
	m.EntryConversation(conversationID, question)
}
