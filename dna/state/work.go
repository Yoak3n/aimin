package state

import (
	"dna/pkg/logger"
	"encoding/json"
	"errors"
	"time"
)

type WorkState struct {
	name      Flag
	children  []State
	Save      any
	Interrupt bool
	action    func(ctx *Context) (any, error)
}

func (w *WorkState) Label() string {
	return NameMap[uint(w.name)]
}

type ConditionalWorkState struct {
	WorkState
	conditions []func(data any) bool
}

func (c *ConditionalWorkState) Label() string {
	return NameMap[uint(c.name)]
}

type WorkResult struct {
	Data      any
	NextState State
}

func NewConditionalWorkState(name Flag,
	action func(ctx *Context) (any, error),
	conditions ...func(data any) bool) *ConditionalWorkState {
	return &ConditionalWorkState{
		WorkState:  *NewWorkState(name, action),
		conditions: conditions,
	}
}

func NewWorkState(name Flag, action func(ctx *Context) (any, error)) *WorkState {
	return &WorkState{
		name:     name,
		action:   action,
		children: make([]State, 0),
	}
}

//func (w *WorkState) AddChild(state State) {}

func (w *WorkState) Type() Flag {
	return w.name
}

func (w *WorkState) Execute(ctx *Context) (Result, error) {
	logger.Logger.Println("Entering Work State:", NameMap[uint(w.name)])
	data, err := w.action(ctx)
	if err != nil {
		if errors.Is(err, ErrInterrupted) {
			w.Save = ctx.Save
			return Result{
				IsComplete: false,
			}, nil
		}
		return Result{}, err
	}
	// 也许应该另外准备一条数据日志，每次执行工作都记录带有id的数据，
	// ctx.Data[w.Type()] = data
	ctx.ReturnTo = ctx.Caller
	time.Sleep(time.Second)
	return Result{
		Data:      data,
		NextState: w.getNextState(),
	}, nil
}

func (w *WorkState) getNextState() State {
	if len(w.children) > 0 {
		return w.children[0]
	}
	return nil
}

func (w *WorkState) AddChild(state State) {
	w.children = append(w.children, state)
}

func (w *WorkState) Children() []State {
	return w.children
}

func (w *WorkState) SaveState() ([]byte, error) {
	// 实现状态序列化逻辑
	return json.Marshal(w.Save)
}

func (w *WorkState) CanInterrupt() bool {
	return w.Interrupt
}

func (c *ConditionalWorkState) Execute(ctx *Context) (Result, error) {
	logger.Logger.Println("Entering Conditional Work State:", NameMap[uint(c.name)])
	eResult, err := c.action(ctx)
	if err != nil {
		if errors.Is(err, ErrInterrupted) {
			c.Save = ctx.Save
			return Result{
				IsComplete: false,
			}, nil
		}
		return Result{}, err
	}
	time.Sleep(time.Second)
	// i 对应子节点的索引
	for i, condition := range c.conditions {
		if condition(eResult) {
			if i < len(c.children) {
				next := c.children[i]
				ctx.ReturnTo = ctx.Caller
				result := Result{
					Data:      eResult,
					NextState: next,
				}
				return result, nil
			}
		} else {
			continue
		}
	}

	return Result{Data: eResult}, nil
}
