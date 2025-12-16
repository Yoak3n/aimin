package node

import "dna/state"

func WatchNode() state.State {
	node := state.NewVirtualState(Watch)
	node.AddChild(WatchEnviromentNode())
	return node
}

func WatchEnviromentNode() state.State {
	return state.NewWorkState(WatchEnvironment, WatchEnvironmentAction)
}

func WatchEnvironmentAction(ctx *state.Context) (any, error) {
	// 模拟观察环境的操作
	observation := "观察到环境中的变化"
	return observation, nil
}
