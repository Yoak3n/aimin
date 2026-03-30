package decision

import "github.com/Yoak3n/aimin/dna/fsm"

func NewWatchNode(check func(ctx *fsm.Context) bool) *fsm.WorkState {
	node := fsm.NewWorkState(Watch, Watch, makeWatchAction(), check)
	node.SetDoneHook(Watch, Introspection)
	return node
}

func makeWatchAction() fsm.WorkAction {
	return func(ctx *fsm.Context) string {
		ctx.Attr.AddEnergy(-1)
		ctx.Attr.AddOpenness(3)
		return fsm.Done
	}
}
