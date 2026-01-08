package controller

import (
	"blood/pkg/util"
	"blood/schema"
	"errors"
	"time"
)

func CreateDialogueWithConversation(message schema.OpenAIMessage, id ...string) error {
	if len(id) > 0 {
		db.UpdateConversationOnlyTime(id[0], time.Now())
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
