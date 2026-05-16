package decision

import (
	"math/rand"

	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewStateTree() *fsm.FSM {
	tree := fsm.NewFSM()
	tree.RegisterState(newRootNode())
	tree.RegisterState(NewTaskState())
	return tree
}

const rootRouterKey = "root-router-key"

var (
	// sleep: (1-e) 线性 + 乘子型偏置
	// energy 从 50→30 时权重从 0.55→0.75；从 50→70 时 0.55→0.35
	// 配合 bias 系统 0.625~1.25× 的调节范围，energy=50 时 sleep 明显高于 idle
	sleepProfile = fsm.AttrBiasProfile{
		Base:   0.05,
		Energy: fsm.AttrFactor{Weight: 1.0, Inverted: true},
	}
	idleProfile = fsm.AttrBiasProfile{
		Base:   0.05,
		Energy: fsm.AttrFactor{Weight: 1.0},
	}
)

func newRootNode() *fsm.CompositeState {
	root := fsm.NewCompositeState(Root, Root, []fsm.State{NewIdleNode(), NewSleepNode()}, nil)
	root.SetRouterKey(rootRouterKey)
	root.SetSelect(rootNodeSelector)
	return root
}

func rootNodeSelector(ctx *fsm.Context, states []fsm.State) int {
	weights := make([]float64, 0, len(states))
	for _, st := range states {
		var w float64
		switch st.ID() {
		case Sleep:
			w = ctx.CalcStateWeight(Sleep, sleepProfile, 0.3)
			// sleep 非线性 boost：energy<55 时 sleep 加速上升
			// e=50→boost=0.125, e=40→0.375, e=30→0.625, e=20→0.875
			if ctx != nil && ctx.Attr != nil {
				x := (55 - ctx.Attr.Energy) / 55
				if x > 0 {
					w += x * x * x
				}
			}
		case Idle:
			w = ctx.CalcStateWeight(Idle, idleProfile, 0.8)
		default:
			w = 0.1
		}
		weights = append(weights, w)
	}
	logger.Logger.Printf("sleepWeight: %f, idleWeight: %f", weights[0], weights[1])
	idx := fsm.WeightedIndex(weights)
	if idx < 0 || idx >= len(states) {
		idx = rand.Intn(len(states))
	}
	ctx.Data[rootRouterKey] = states[idx].ID()
	return idx
}