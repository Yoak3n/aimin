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

const rootRouterKey = "root-router-key"

func newRootNode() *fsm.CompositeState {
	root := fsm.NewCompositeState(Root, Root, []fsm.State{NewIdleNode(), NewSleepNode()}, nil)
	root.SetRouterKey(rootRouterKey)
	root.SetSelect(rootNodeSelector)
	return root
}

func rootNodeSelector(ctx *fsm.Context, states []fsm.State) int {
	energy := clamp01(ctx.Attr.Energy / 100)
	weights := make([]float64, 0, len(states))
	for _, st := range states {
		w := 0.05
		switch st.ID() {
		case Sleep:
			w += clamp01((0.6 - energy) / 0.6)
			w *= 1 + ctx.GetStateBias(Sleep)
			if ctx.LastDoneState() == Sleep {
				w *= 0.6
			}
		case Idle:
			w += energy
			w *= 1 + ctx.GetStateBias(Idle)
			if ctx.LastDoneState() == Idle {
				w *= 0.8
			}
		default:
			w += 0.1
		}
		weights = append(weights, w)
	}
	idx := fsm.WeightedIndex(weights)
	if idx < 0 || idx >= len(states) {
		idx = rand.Intn(len(states))
	}
	ctx.Data[rootRouterKey] = states[idx].ID()
	return idx
}
