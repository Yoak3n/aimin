package decision

import (
	"math/rand"

	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/fsm"
)

const idleChoiceKey = "idle_choice"

var (
	// explore: 好奇驱动，能量充足时加成交互。curiosity 高 → 倾向探索
	exploreProfile = fsm.AttrBiasProfile{
		Base:            0.05,
		Curiosity:       fsm.AttrFactor{Weight: 0.3},
		EnergyCuriosity: 0.7,
	}
	// watch: 开放驱动，能量充足时加成交互。openness 高 → 倾向观察
	watchProfile = fsm.AttrBiasProfile{
		Base:           0.05,
		Openness:       fsm.AttrFactor{Weight: 0.3},
		EnergyOpenness: 0.7,
	}
	// introspection: 能量低时倾向反思，好奇心低时也倾向内省
	introspectionProfile = fsm.AttrBiasProfile{
		Base:               0.05,
		Energy:             fsm.AttrFactor{Weight: 0.2, Inverted: true},
		Curiosity:          fsm.AttrFactor{Weight: 0.1, Inverted: true},
		InvEnergyCuriosity: 0.7,
	}
)

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
		exploreWeight := ctx.CalcStateWeight(Explore, exploreProfile, 0.6)
		watchWeight := ctx.CalcStateWeight(Watch, watchProfile, 0.6)
		introspectionWeight := ctx.CalcStateWeight(Introspection, introspectionProfile, 0.6)

		// energy 加权：高能量→更多向外探索，低能量→更多反思/观察
		// explore 耗能最多(-5)，低能量时应被抑制；introspect(-3)/watch(-4) 耗能更少
		if ctx != nil && ctx.Attr != nil {
			e := ctx.Attr.Energy / 100
			// normalizedBoost: 0-1 范围，energy=0 时 explore 乘 0.8，energy=100 时乘 1.0
			exploreBoost := 0.8 + 0.2*e
			// lowEnergyBoost: 能量越低 introspect 越高，1.0→0.9
			introspectBoost := 1.0 - 0.1*(1-e)
			watchBoost := 1.0 - 0.05*(1-e)

			exploreWeight *= exploreBoost
			introspectionWeight *= introspectBoost
			watchWeight *= watchBoost
		}

		weights := []float64{exploreWeight, watchWeight, introspectionWeight}
		logger.Logger.Printf("exploreWeight: %f, watchWeight: %f, introspectionWeight: %f", exploreWeight, watchWeight, introspectionWeight)
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