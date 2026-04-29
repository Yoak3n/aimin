package decision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Yoak3n/aimin/blood/agent"
	schemaws "github.com/Yoak3n/aimin/blood/schema/ws"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/Yoak3n/aimin/hand/interactive"
)

type TaskType string

const (
	TaskDataKey  = "task_data"
	TaskQueueKey = "task_queue"
)

func NewTaskState() fsm.State {
	return fsm.NewTaskState(Task, Task, makeTaskAction())
}

func makeTaskAction() func(ctx *fsm.Context) string {
	idleTimeout := 5 * time.Minute
	var lastActivity time.Time
	var busy int32
	var started bool
	var stopCh chan struct{}
	var questionCh chan fsm.TaskData

	startWorker := func() {
		if started {
			return
		}
		started = true
		stopCh = make(chan struct{})
		questionCh = make(chan fsm.TaskData, 32)
		lastActivity = time.Now()

		go func() {
			var conv *agent.ConversationAgent
			var convID string
			var convFrom string

			for {
				select {
				case <-stopCh:
					return
				case td, ok := <-questionCh:
					if !ok {
						return
					}
					q, ok := td.Payload.(string)
					q = strings.TrimSpace(q)
					if !ok || q == "" {
						continue
					}

					if conv == nil || convID != td.ID || convFrom != td.From {
						convID = td.ID
						convFrom = td.From
						conv = interactive.NewConversationTask(td.ID, td.From)
					}

					atomic.StoreInt32(&busy, 1)
					roundID, _, _ := interactive.BeginInterruptibleRound(td.From)
					_, err := conv.Ask(q)
					interactive.EndInterruptibleRound(td.From, roundID)
					atomic.StoreInt32(&busy, 0)
					if err != nil {
						if errors.Is(err, context.Canceled) {
							msg := schemaws.NewReplyMessage(schemaws.ReplyStatusFinish, td.ID, "[已打断]")
							b, _ := json.Marshal(msg)
							if interactive.WSReplyBroadcast != nil {
								interactive.WSReplyBroadcast(td.From, b)
							}
							continue
						}
						msg := schemaws.NewReplyMessage(schemaws.ReplyStatusFinish, td.ID, fmt.Sprintf("[错误] %v", err))
						b, _ := json.Marshal(msg)
						if interactive.WSReplyBroadcast != nil {
							interactive.WSReplyBroadcast(td.From, b)
						}
					}
				}
			}
		}()
	}

	appendTask := func(ctx *fsm.Context, td fsm.TaskData) {
		if ctx == nil {
			return
		}
		v := ctx.Data[TaskQueueKey]
		if q, ok := v.([]fsm.TaskData); ok && q != nil {
			ctx.Data[TaskQueueKey] = append(q, td)
			return
		}
		ctx.Data[TaskQueueKey] = []fsm.TaskData{td}
	}

	return func(ctx *fsm.Context) string {
		startWorker()

		if ctx != nil {
			if v, ok := ctx.Data[TaskDataKey]; ok {
				if td, ok := v.(fsm.TaskData); ok {
					appendTask(ctx, td)
				}
				delete(ctx.Data, TaskDataKey)
			}
		}

		var q []fsm.TaskData
		if ctx != nil {
			if v, ok := ctx.Data[TaskQueueKey]; ok {
				if qq, ok := v.([]fsm.TaskData); ok {
					q = qq
				}
			}
		}

		if len(q) > 0 && interactive.ConsumeInterruptQueueClear(q[0].From) {
			q = nil
			if ctx != nil {
				delete(ctx.Data, TaskQueueKey)
			}
		drainQuestionCh:
			for {
				select {
				case <-questionCh:
				default:
					break drainQuestionCh
				}
			}
		}

	drain:
		for len(q) > 0 {
			td := q[0]
			q = q[1:]
			select {
			case questionCh <- td:
				lastActivity = time.Now()
			default:
				q = append([]fsm.TaskData{td}, q...)
				break drain
			}
		}

		if ctx != nil {
			if len(q) > 0 {
				ctx.Data[TaskQueueKey] = q
			} else {
				delete(ctx.Data, TaskQueueKey)
			}
		}

		if atomic.LoadInt32(&busy) == 0 && time.Since(lastActivity) > idleTimeout {
			if stopCh != nil {
				close(stopCh)
			}
			if ctx != nil {
				delete(ctx.Data, TaskDataKey)
				delete(ctx.Data, TaskQueueKey)
			}
			return fsm.Done
		}
		return fsm.Interrupt
	}
}
