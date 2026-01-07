package decision

import (
	"dna/fsm"
	"math/rand"
)

func NewIdleNode() *fsm.CompositeState {
	children := []fsm.State{
		NewExploreNode(),
		NewWatchNode(),
		NewIntrospectionNode(),
	}
	check := func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[Root]
		return ok && v == Idle
	}
	selection := func(ctx *fsm.Context, states []fsm.State) int {
		return rand.Intn(len(states))
	}
	e := fsm.NewCompositeState(Idle, Idle, children, check)
	e.SetSelect(selection)
	return e
}
