package interactive

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Yoak3n/aimin/blood/agent"
	"github.com/Yoak3n/aimin/blood/schema"
	schemaws "github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/nerve"
)

var WSReplyBroadcast func(clientID string, message []byte)

func NewConversationTask(id, from string) *agent.ConversationAgent {
	base := agent.NewAgent()
	conv := agent.NewConversationAgent(base)
	conv.SetMaxTurns(8)
	chunkIdx := 1

	base.RegisterAssistantDeltaHandler(func(delta string) error {
		msg := schemaws.WebsocketMessage{
			Action: schemaws.ReplyMessage,
			Data: &schemaws.ReplyMessageData{
				TaskID: id,
				Status: schemaws.ReplyStatusProcessing,
				Chunk: &schemaws.ReplyChunkData{
					TaskID:   id,
					ChunkIdx: chunkIdx,
					Content:  delta,
				},
			},
		}
		chunkIdx++
		buf, _ := json.Marshal(msg)
		if WSReplyBroadcast != nil {
			WSReplyBroadcast(from, buf)
		}
		return nil
	})

	base.RegisterFinalAnswerHandler(func(system string, msgs []schema.OpenAIMessage, finalAnswer string) {
		msg := schemaws.WebsocketMessage{
			Action: schemaws.ReplyMessage,
			Data: &schemaws.ReplyMessageData{
				TaskID: id,
				Status: schemaws.ReplyStatusFinish,
				Result: &schemaws.ReplyFinishData{
					TaskID:  id,
					Content: finalAnswer,
				},
			},
		}
		chunkIdx = 0
		buf, _ := json.Marshal(msg)
		if WSReplyBroadcast != nil {
			WSReplyBroadcast(from, buf)
		}
		data, err := json.Marshal(msgs)
		if err != nil {
			log.Println("marshal msg err:", err)
			return
		}
		os.WriteFile("chat.json", data, 0644)
		nerve.ResponseHook(system, finalAnswer, msgs)
	})

	return conv
}
