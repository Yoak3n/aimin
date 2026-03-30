package decision

import (
	"math/rand"

	"github.com/Yoak3n/aimin/dna/fsm"
)

const idleChoiceKey = "idle_choice"

func NewIdleNode() *fsm.CompositeState {
	exploreCheck := func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[idleChoiceKey]
		return ok && v == Explore
	}
	watchCheck := func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[idleChoiceKey]
		return ok && v == Watch
	}
	introspectionCheck := func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[idleChoiceKey]
		return ok && v == Introspection
	}
	children := []fsm.State{
		NewExploreNode(exploreCheck),
		NewWatchNode(watchCheck),
		NewIntrospectionNode(introspectionCheck),
	}
	check := func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[rootRouterKey]
		return ok && v == Idle
	}
	selection := func(ctx *fsm.Context, states []fsm.State) int {
		energy := clamp01(ctx.Attr.Energy / 100)
		curiosity := clamp01(ctx.Attr.Curiosity / 100)
		openness := clamp01(ctx.Attr.Openness / 100)

		exploreWeight := 0.05 + (energy * curiosity)
		watchWeight := 0.05 + (energy * openness)
		introspectionWeight := 0.05 + ((1 - energy) * curiosity)

		exploreWeight *= 1 + ctx.GetStateBias(Explore)
		watchWeight *= 1 + ctx.GetStateBias(Watch)
		introspectionWeight *= 1 + ctx.GetStateBias(Introspection)

		switch ctx.LastDoneState() {
		case Explore:
			exploreWeight *= 0.6
			introspectionWeight *= 1.1
		case Watch:
			watchWeight *= 0.6
			introspectionWeight *= 1.1
		case Introspection:
			introspectionWeight *= 0.6
			exploreWeight *= 1.1
		}

		weights := []float64{exploreWeight, watchWeight, introspectionWeight}
		idx := fsm.WeightedIndex(weights)
		if idx < 0 || idx >= len(states) {
			idx = rand.Intn(len(states))
		}
		ctx.Data[idleChoiceKey] = states[idx].ID()
		return idx
	}
	e := fsm.NewCompositeState(Idle, Idle, children, check)
	e.SetRouterKey(idleChoiceKey)
	e.SetSelect(selection)
	return e
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
