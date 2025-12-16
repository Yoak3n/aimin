package state

import (
	"dna/pkg/logger"
	"math/rand/v2"
)

type VirtualState struct {
	name     Flag
	children []State
}

func (v *VirtualState) Label() string {
	return NameMap[uint(v.name)]
}

func NewVirtualState(name Flag) *VirtualState {
	return &VirtualState{
		name:     name,
		children: make([]State, 0),
	}
}

func (v *VirtualState) Type() Flag {
	return v.name
}

func (v *VirtualState) Execute(ctx *Context) (Result, error) {
	logger.Logger.Debugln("Entering Virtual State:", NameMap[uint(v.name)])
	// 虚状态本身不执行具体操作，只是传递到子状态
	if len(v.children) > 0 {
		// 虚状态不读取操作数据，所以随机选择子状态
		choice := rand.IntN(len(v.children))
		return Result{
			NextState: v.children[choice],
		}, nil
	}
	return Result{}, nil
}

func (v *VirtualState) AddChild(state State) {
	v.children = append(v.children, state)
}

func (v *VirtualState) Children() []State {
	return v.children
}

func (v *VirtualState) SaveState() ([]byte, error) {
	// 实现状态序列化逻辑
	return nil, nil
}

func (v *VirtualState) CanInterrupt() bool {
	return false
}
