package controller

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

func CreateDialogueWithConversation(message schema.OpenAIMessage, id ...string) error {
	var conversationId string
	if len(id) > 0 {
		conversationId = id[0]
		db.UpdateConversationOnlyTime(conversationId, time.Now())
	}
	dialogue := schema.DialogueRecord{
		Id:             util.RandomIdWithPrefix("dialogue-"),
		Role:           message.Role,
		Content:        message.Content,
		ConversationId: conversationId,
	}
	return db.CreateDialogueRecord(dialogue)
}

func GetDialoguesWithConversation(id string) ([]schema.DialogueRecord, error) {
	ds := db.QueryDialogueRecords(id)
	if ds == nil {
		return []schema.DialogueRecord{}, nil
	}
	return ds, nil
}

func GetAllConversations() ([]schema.ConversationRecord, error) {
	return db.GetAllConversations()
}
