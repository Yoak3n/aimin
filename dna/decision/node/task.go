package node

import (
	"dna/decision/state"
)

func TaskNode() state.State {
	node := state.NewWorkState(Task, func(ctx *state.Context) (state.Result, error) {
		ctx.Status = state.TaskState
		return nil, nil
	})
	node.AddChild(AnswerNode())
	checkHasTask := func(ctx *state.Context) bool {
		return ctx.CheckTask()
	}
	node.AddCondition(checkHasTask)
	node.AddChild(AnswerNode())
	return node
}

func AnswerNode() state.State {
	return state.NewVirtualState(Answer)
}
