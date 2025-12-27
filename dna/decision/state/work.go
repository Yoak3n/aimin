package state

import (
	"blood/pkg/logger"
	"time"
)

type WorkState struct {
	name      Flag
	children  []State
	Save      any
	Interrupt bool
	action    func(ctx *Context) (Result, error)
	conds     []func(ctx *Context) bool
}

type WorkResult struct {
	ResultData
}

func (w *WorkState) CheckConditions(ctx *Context) bool {
	if len(w.conds) < 1 {
		return true
	}
	for _, cond := range w.conds {
		if cond != nil && cond(ctx) {
			return true
		}
	}
	return false
}

func (w *WorkState) AddCondition(cond func(ctx *Context) bool) {
	w.conds = append(w.conds, cond)
}

func (w *WorkState) Label() string {
	return NameMap[uint(w.name)]
}

func NewWorkState(name Flag, action func(ctx *Context) (Result, error)) *WorkState {
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
	prepared := WorkResult{
		ResultData: ResultData{
			Status: Running,
			From:   w,
		},
	}

	data, err := w.action(ctx)
	if err != nil {
		return prepared, err
	}
	switch data.GetStatus() {
	case Interrupted:
		w.Save = ctx.Save
		prepared.Status = Interrupted
		prepared.Data = data.GetData()
		return prepared, nil
	case ToReturn:
		prepared.Status = ToReturn
		prepared.Data = data.GetData()
		prepared.NextState = data.GetNextState()
		return prepared, nil
	default:
		prepared.Data = data.GetData()
		prepared.NextState = w.getNextState(ctx)
	}
	time.Sleep(time.Second)
	// 也许应该另外准备一条数据日志，每次执行工作都记录带有id的数据，
	// ctx.Data[w.Type()] = data
	return prepared, nil
}

func (w *WorkState) getNextState(ctx *Context) State {
	if len(w.children) < 1 {
		return nil
	}
	for _, child := range w.children {
		if child.CheckConditions(ctx) {
			return child
		}
	}
	return w.children[len(w.children)-1]
}

func (w *WorkState) AddChild(state State) {
	w.children = append(w.children, state)
}

func (w *WorkState) Children() []State {
	return w.children
}

func (w *WorkState) CanInterrupt() bool {
	return w.Interrupt
}
