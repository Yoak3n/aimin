package nerve

import (
	"blood/pkg/helper"
	"blood/pkg/util"
	"blood/schema"
	"errors"
	"log"
	nc "nerve/controller"
	"nerve/memory"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

type Nerve struct {
	Hippocampus *memory.Hippocampus
}

func AddMemory(messages []schema.DialogueRecord) error {
	ids := make([]string, 0, len(messages))
	resources := make([]schema.OpenAIMessage, 0, len(messages))
	for _, message := range messages {
		resources = append(resources, schema.OpenAIMessage{
			Role:    message.Role,
			Content: message.Content,
		})
		ids = append(ids, message.Id)
	}
	strategy := nc.FetchMemoryStrategy(resources)

	if strategy == "" {
		return errors.New("failed to fetch memory strategy")
	}
	return extractStrategy(strategy, ids)
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
	helper.UseDB().AddEntityTableRecord(entitiesMap)
	e := &schema.EnduringMemoryTable{
		Id:                m.Id,
		Content:           m.Content,
		Topic:             m.Topic,
		LastSimulatedTime: m.LastSimulated,
	}
	embeddings, err := helper.UseLLM().Embedding([]string{m.Content})
	if err != nil {
		log.Println("Embedding error: ", err)
		return
	}
	if len(embeddings) > 0 {
		err = helper.UseDB().CreateEnduringTableRecord(*e, embeddings[0])
		if err != nil {
			log.Println("CreateEnduringTableRecord error: ", err)
		}
		_ = helper.UseDB().UpdatedDialogueRecordsLinks(source, m.Id)
	}
}

func addTemporary(m *memory.Memory, source []string) {

}
