package state

import (
	"blood/pkg/logger"
	"fmt"
)

type ReturnState struct {
	name Flag
	// 是否需要标记状态的类型（回归状态）
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

	prepared := ResultData{
		Status:    Returned,
		NextState: ctx.ReturnTo,
		From:      r,
	}
	if ctx.ReturnTo == nil {
		return prepared, fmt.Errorf("return target not set")
	}
	if r.action != nil {
		d, err := r.action(ctx)
		if err != nil {
			return ResultData{}, err
		}
		prepared.Data = d
	}

	return prepared, nil
}

func (r *ReturnState) AddChild(state State) {
	// 回归状态通常不需要子状态
}

func (r *ReturnState) Children() []State {
	// 回归状态通常不需要子状态
	return []State{}
}

func (r *ReturnState) SaveState() ([]byte, error) {
	// 实现状态序列化逻辑
	return nil, nil
}

func (r *ReturnState) CanInterrupt() bool {
	return false
}

func (r *ReturnState) CheckConditions(ctx *Context) bool {
	// 回归状态不需要条件，且只能主动调用
	return false
}

func (r *ReturnState) AddCondition(cond func(ctx *Context) bool) {
	// 回归状态不需要条件
}
