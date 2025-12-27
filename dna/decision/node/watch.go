package node

import (
	state "dna/decision/state"
)

func WatchNode() state.State {
	node := state.NewVirtualState(Watch)
	node.AddChild(WatchEnvironmentNode())
	return node
}

func WatchEnvironmentNode() state.State {
	return state.NewWorkState(WatchEnvironment, WatchEnvironmentAction)
}
func WatchRoomNode() state.State {
	return state.NewWorkState(WatchRoom, WatchRoomAction)
}

func WatchEnvironmentAction(ctx *state.Context) (state.Result, error) {
	// 模拟观察环境的操作
	observation := "观察到环境中的变化"
	return state.ResultData{
		Data: observation,
	}, nil
}

func WatchRoomAction(ctx *state.Context) (state.Result, error) {
	// 模拟观察房间的操作
	observation := "观察到房间内的物品和布局"
	return state.ResultData{
		Data: observation,
	}, nil
}
