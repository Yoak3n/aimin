package state

import (
	"context"
)

type Flag uint

// State 状态接口
type State interface {
	Type() Flag
	Label() string
	Execute(ctx *Context) (Result, error)
	AddChild(state State)
	Children() []State
	SaveState() ([]byte, error)
	CanInterrupt() bool
}

type Context struct {
	context.Context
	Caller     State        // 调用者
	ReturnTo   State        // 回归目标（由调用者设置）
	Data       map[Flag]any // 共享数据
	IsComplete bool
	Save       map[Flag]any
	Interrupt  chan struct{}
}

func NewStateContext(ctx context.Context) *Context {
	return &Context{
		Context:  ctx,
		ReturnTo: nil,
		Caller:   nil,
		Save:     map[Flag]any{},
		Data:     make(map[Flag]any),
	}
}

// Result 执行结果
type Result struct {
	Data       any
	NextState  State
	IsReturn   bool // 是否为回归结果
	ShouldSave bool // 是否需要保存结果
	IsComplete bool
}
