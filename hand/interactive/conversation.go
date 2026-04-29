package interactive

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/agent"
	"github.com/Yoak3n/aimin/blood/agent/workspace"
	"github.com/Yoak3n/aimin/blood/schema"
	schemaws "github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/nerve"
)

var WSReplyBroadcast func(clientID string, message []byte)

var interruptMu sync.Mutex
var interruptCtxByClient = map[string]context.Context{}
var interruptRoundIDByClient = map[string]string{}
var interruptCancelByClient = map[string]context.CancelFunc{}
var interruptClearQueueByClient = map[string]bool{}

func BeginInterruptibleRound(clientID string) (string, context.Context, context.CancelFunc) {
	if clientID == "" {
		return "", nil, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	roundID := fmt.Sprintf("ir_%d", time.Now().UnixNano())
	interruptMu.Lock()
	interruptCtxByClient[clientID] = ctx
	interruptRoundIDByClient[clientID] = roundID
	interruptCancelByClient[clientID] = cancel
	interruptMu.Unlock()
	return roundID, ctx, cancel
}

func EndInterruptibleRound(clientID string, roundID string) {
	if clientID == "" || roundID == "" {
		return
	}
	interruptMu.Lock()
	currentID := interruptRoundIDByClient[clientID]
	if currentID == roundID {
		cancel := interruptCancelByClient[clientID]
		if cancel != nil {
			cancel()
		}
		delete(interruptCancelByClient, clientID)
		delete(interruptCtxByClient, clientID)
		delete(interruptRoundIDByClient, clientID)
	}
	interruptMu.Unlock()
}

func CurrentInterruptErr(clientID string) error {
	if clientID == "" {
		return nil
	}
	interruptMu.Lock()
	ctx := interruptCtxByClient[clientID]
	interruptMu.Unlock()
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func HasInterruptibleRound(clientID string) bool {
	if clientID == "" {
		return false
	}
	interruptMu.Lock()
	cancel := interruptCancelByClient[clientID]
	interruptMu.Unlock()
	return cancel != nil
}

func RequestInterrupt(clientID string) bool {
	if clientID == "" {
		return false
	}
	interruptMu.Lock()
	cancel := interruptCancelByClient[clientID]
	interruptClearQueueByClient[clientID] = true
	interruptMu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func ConsumeInterruptQueueClear(clientID string) bool {
	if clientID == "" {
		return false
	}
	interruptMu.Lock()
	v := interruptClearQueueByClient[clientID]
	if v {
		delete(interruptClearQueueByClient, clientID)
	}
	interruptMu.Unlock()
	return v
}

func NewConversationTask(id, from string) *agent.ConversationAgent {
	base := agent.NewAgent(workspace.PromptPurposeReAct)
	conv := agent.NewConversationAgent(base)
	conv.SetMaxTurns(8)
	chunkIdx := 1

	base.RegisterAssistantDeltaHandler(func(delta string) error {
		if err := CurrentInterruptErr(from); err != nil {
			return err
		}
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

	base.RegisterToolResultHandler(func(toolCallID string, action string, result string, err error) {
		if CurrentInterruptErr(from) != nil {
			return
		}
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		msg := schemaws.NewToolResultMessage(id, toolCallID, action, result, errMsg)
		buf, _ := json.Marshal(msg)
		if WSReplyBroadcast != nil {
			WSReplyBroadcast(from, buf)
		}
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
