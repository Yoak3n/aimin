package nerve

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/nerve/controller"
	"github.com/Yoak3n/aimin/nerve/memory"

	"github.com/tidwall/gjson"
)

type Nerve struct {
	Hippocampus *memory.Hippocampus
}

func AddMemory(messages []schema.OpenAIMessage) error {
	// ids := make([]string, 0, len(messages))
	// resources := make([]schema.OpenAIMessage, 0, len(messages))
	// for _, message := range messages {
	// 	resources = append(resources, schema.OpenAIMessage{
	// 		Role:    schema.OpenAIMessageRole(message.Role),
	// 		Content: message.Content,
	// 	})
	// 	ids = append(ids, message.Id)
	// }
	// strategy := nc.FetchMemoryStrategy(resources)

	// if strategy == "" {
	// 	return errors.New("failed to fetch memory strategy")
	// }
	// return extractStrategy(strategy, ids)
	return nil
}

func ResponseHook(systemPrompt, answer string, messages []schema.OpenAIMessage) {
	question, thoughts := extractQuestionAndThoughts(messages)
	cid := util.RandomIdWithPrefix("con")
	now := time.Now()
	c := &schema.ConversationRecord{
		Id:        cid,
		Question:  question,
		Thoughts:  thoughts,
		Answer:    answer,
		System:    systemPrompt,
		CreateAt:  now,
		UpdatedAt: now,
	}
	embeding, err := helper.UseLLM().Embedding([]string{question + "---" + answer})
	if err != nil {
		log.Fatal(err)
	}
	go helper.UseDB().CreateConversationRecord(c, embeding[0])
	go controller.SummaryConversation(c, cid)
}

func QueryReleventConversations(input string) (string, error) {
	conversations := make([]schema.ConversationRecord, 0)
	embeding, err := helper.UseLLM().Embedding([]string{input})
	if err != nil {
		log.Fatal(err)
	}
	conversations, err = helper.UseDB().GetReleventConversationRecords(embeding[0])
	if err != nil {
		log.Fatal(err)
	}
	cl := strings.Builder{}
	for _, conversation := range conversations {
		fmt.Fprintf(&cl, "- [%s]%s---%s\n", conversation.Id, conversation.Question, conversation.Answer)
	}
	return cl.String(), nil
}

func GetConversationByID(id string) (schema.ConversationRecord, error) {
	return helper.UseDB().GetConversationByID(id)
}

func extractStrategy(s string, source []string) error {
	trimmedRespMsg := strings.TrimSpace(s)
	trimmedRespMsg = strings.TrimPrefix(trimmedRespMsg, "```json")
	trimmedRespMsg = strings.TrimSuffix(trimmedRespMsg, "```")
	result := gjson.Parse(trimmedRespMsg)
	decision := result.Get("decision").String()
	switch decision {
	case "enduring":
		enduring := result.Get("enduring")
		if !enduring.Exists() {
			return errors.New("enduring not exists")
		}
		enduringId := util.RandomIdWithPrefix("enduring-")
		triples := enduring.Get("triples")
		if !triples.Exists() {
			return errors.New("triples not exists")
		}
		entitiesMap := make([]schema.EntityTable, 0)

		triples.ForEach(func(key, value gjson.Result) bool {
			memoryMap := &schema.EntityTable{
				Link:        enduringId,
				Subject:     value.Get("subject").String(),
				Predicate:   value.Get("predicate").String(),
				Object:      value.Get("object").String(),
				SubjectType: value.Get("subject_type").String(),
				ObjectType:  value.Get("object_type").String(),
			}
			entitiesMap = append(entitiesMap, *memoryMap)
			return true
		})

		m := &memory.Memory{
			Id:            enduringId,
			Content:       enduring.Get("content").String(),
			Topic:         enduring.Get("topic").String(),
			Level:         5,
			Create:        time.Now(),
			LastSimulated: time.Now(),
		}
		go addEnduring(m, source, entitiesMap)
	case "temporary":
		temporaryId := util.RandomIdWithPrefix("temporary-")
		content := result.Get("temporary.content").String()
		topic := result.Get("temporary.topic").String()
		m := &memory.Memory{
			Id:            temporaryId,
			Content:       content,
			Topic:         topic,
			Level:         1,
			Create:        time.Now(),
			LastSimulated: time.Now(),
		}
		go addTemporary(m, source)
	default:
		log.Println("unknown decision: ", decision)
	}
	return nil
}

func addEnduring(m *memory.Memory, source []string, entitiesMap []schema.EntityTable) {

}

func addTemporary(m *memory.Memory, source []string) {

}
