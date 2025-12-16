package state

import (
	"dna/pkg/logger"
	"fmt"
)

type ReturnState struct {
	name   Flag
	action func(ctx *Context) (any, error)
}

func (r *ReturnState) Label() string {
	return NameMap[uint(r.name)]
}

func NewReturnState(name Flag, action func(ctx *Context) (any, error)) *ReturnState {
	return &ReturnState{
		name:   name,
		action: action,
	}
}

func (r *ReturnState) Type() Flag {
	return r.name
}

func (r *ReturnState) Execute(ctx *Context) (Result, error) {
	logger.Logger.Println("Entering Return State:", NameMap[uint(r.name)])
	// 执行回归操作
	data, err := r.action(ctx)
	if err != nil {
		return Result{}, err
	}
	if ctx.ReturnTo == nil {
		return Result{}, fmt.Errorf("return target not set")
	}
	return Result{
		Data:      data,
		NextState: ctx.ReturnTo,
		IsReturn:  true, // 标记为回归结果
	}, nil
}

func (r *ReturnState) AddChild(state State) {
	// 回归状态通常不需要子状态
}

func (r *ReturnState) Children() []State {
	return []State{}
}

func (r *ReturnState) SaveState() ([]byte, error) {
	// 实现状态序列化逻辑
	return nil, nil
}

func (r *ReturnState) CanInterrupt() bool {
	return false
}
