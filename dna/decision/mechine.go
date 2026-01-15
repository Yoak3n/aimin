package decision

import (
	"math/rand"

	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewStateTree() *fsm.FSM {
	tree := fsm.NewFSM()
	tree.RegisterState(newRootNode())
	tree.RegisterState(NewTaskState())
	return tree
}

func newRootNode() *fsm.CompositeState {
	root := fsm.NewCompositeState(Root, Root, []fsm.State{NewIdleNode(), NewSleepNode()}, nil)
	root.SetSelect(func(ctx *fsm.Context, states []fsm.State) int {
		idx := rand.Intn(len(states))
		ctx.Data[root.RouterKey] = states[idx].ID()
		return idx
	})
	return root
}
