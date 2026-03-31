package interactive

import (
	"encoding/json"

	"github.com/Yoak3n/aimin/blood/agent"
	"github.com/Yoak3n/aimin/blood/schema"
	schemaws "github.com/Yoak3n/aimin/blood/schema/ws"
)

var WSReplyBroadcast func(clientID string, message []byte)

func NewConversationTask(id, from string) *agent.ConversationAgent {
	base := agent.NewAgent()
	conv := agent.NewConversationAgent(base)
	conv.SetMaxTurns(10)
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

	base.RegisterFinalAnswerHandler(func(_ string, _ []schema.OpenAIMessage, finalAnswer string) {
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
	})

	return conv
}
