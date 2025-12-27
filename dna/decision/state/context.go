package state

import (
	"blood/pkg/logger"
	"context"
	"fmt"
)

// Context 还得是Context，还不行的话上外部全局变量
type Context struct {
	context.Context
	Caller       State        // 调用者
	ReturnTo     State        // 回归目标（由调用者设置）
	Data         map[Flag]any // 共享数据
	Save         map[Flag]any
	Interrupt    chan struct{}
	TaskEntrance chan Task
	TaskQueue    chan Task
	resume       chan State
	Status       CtxStatus
}

func NewStateContext(ctx context.Context) *Context {
	return &Context{
		Context:      ctx,
		ReturnTo:     nil,
		Caller:       nil,
		Save:         map[Flag]any{},
		Data:         make(map[Flag]any),
		Interrupt:    make(chan struct{}, 1),
		TaskEntrance: make(chan Task),
		TaskQueue:    make(chan Task, 10),
	}
}

func (c *Context) CheckTask() bool {
	select {
	case <-c.Done():
		return false
	case task := <-c.TaskEntrance:
		logger.Logger.Println("task received")
		c.queueTask(task)
		if task.Priority > 5 {
			err := c.SubmitTask()
			if err != nil {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (c *Context) SubmitTask() error {
	select {
	case c.Interrupt <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("interrupt channel full")
	}
}

func (c *Context) queueTask(task Task) {
	c.TaskQueue <- task
}
