package state

import (
	"blood/pkg/logger"
)

type VirtualState struct {
	name     Flag
	children []State
	conds    []func(ctx *Context) bool
}

func (v *VirtualState) CheckConditions(ctx *Context) bool {
	// 没有条件就直接放行
	if len(v.conds) < 0 {
		return true
	}
	// 检查条件，只要满足就放行
	for _, cond := range v.conds {
		if cond != nil && cond(ctx) {
			return true
		}
	}
	return false
}

func (v *VirtualState) AddCondition(cond func(ctx *Context) bool) {
	v.conds = append(v.conds, cond)
}

func (v *VirtualState) Label() string {
	return NameMap[uint(v.name)]
}

func NewVirtualState(name Flag) *VirtualState {
	return &VirtualState{
		name:     name,
		children: make([]State, 0),
		conds:    make([]func(ctx *Context) bool, 0),
	}
}

func (v *VirtualState) Type() Flag {
	return v.name
}

func (v *VirtualState) Execute(ctx *Context) (Result, error) {
	logger.Logger.Infoln("Entering Virtual State:", NameMap[uint(v.name)])
	if len(v.children) > 0 {
		for _, child := range v.children {
			if child.CheckConditions(ctx) {
				return ResultData{
					NextState: child,
					From:      v,
				}, nil
			}
		}
		return ResultData{
			NextState: v.children[len(v.children)-1],
		}, nil
	}
	return ResultData{}, nil
}

func (v *VirtualState) AddChild(state State) {
	v.children = append(v.children, state)
	v.conds = append(v.conds, nil)
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
