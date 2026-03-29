package controller

import (
	"log"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

func InsertConversation(cid,system, question, thoughts, answer string) {
	now := time.Now()
	c := &schema.ConversationRecord{
		Id:        cid,
		Question:  question,
		Thoughts:  thoughts,
		Answer:    answer,
		System:    system,
		CreateAt:  now,
		UpdatedAt: now,
	}
	embeding, err := helper.UseLLM().Embedding([]string{question + "---" + answer})
	if err != nil {
		log.Fatal(err)
	}
	helper.UseDB().CreateConversationRecord(c, embeding[0])
}
