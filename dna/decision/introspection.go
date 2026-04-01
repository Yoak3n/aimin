package decision

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewIntrospectionNode(check func(ctx *fsm.Context) bool) *fsm.WorkState {
	node := fsm.NewWorkState(Introspection, Introspection, makeIntrospectionAction(), check)
	node.SetDoneHook(Introspection, Explore, Watch)
	return node
}

func makeIntrospectionAction() fsm.WorkAction {
	return func(ctx *fsm.Context) string {
		ctx.Attr.AddEnergy(-1)
		ctx.Attr.AddCuriosity(-2)
		ctx.Attr.AddOpenness(1)
		logger.Logger.Println("Introspection Action")
		time.Sleep(time.Second)
		return fsm.Done
	}
}
