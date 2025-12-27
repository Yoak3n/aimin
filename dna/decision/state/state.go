package state

type Flag uint

// State 状态接口
type State interface {
	Type() Flag
	Label() string
	Execute(ctx *Context) (Result, error)
	AddChild(state State)
	Children() []State
	CanInterrupt() bool
	CheckConditions(ctx *Context) bool
	AddCondition(func(ctx *Context) bool)
}
