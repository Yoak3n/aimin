package controller

import (
	"blood/pkg/util"
	"blood/schema"
	"errors"
)

func CreateDialogueWithConversation(message schema.OpenAIMessage, data ...schema.ConversationData) error {
	if len(data) > 0 {
		conversation := schema.ConversationRecord{
			Id:   data[0].Id,
			From: data[0].From,
		}
		err := db.UpdateConversationRecord(conversation)
		if err != nil {
			return err
		}
	}
	dialogue := schema.DialogueRecord{
		Id:      util.RandomIdWithPrefix("dialogue-"),
		Role:    message.Role,
		Content: message.Content,
	}
	return db.CreateDialogueRecord(dialogue)
}

func GetDialoguesWithConversation(id string) ([]schema.DialogueRecord, error) {
	ds := db.QueryDialogueRecords(id)
	if ds == nil {
		return nil, errors.New("no dialogue found")
	}
	return ds, nil
}
