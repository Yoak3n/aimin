package decision

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewWatchNode(check func(ctx *fsm.Context) bool) *fsm.WorkState {
	node := fsm.NewWorkState(Watch, Watch, makeWatchAction(), check)
	node.SetDoneHook(Watch, Introspection)
	return node
}

func makeWatchAction() fsm.WorkAction {
	progress := 1
	return func(ctx *fsm.Context) string {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for i := progress; i <= 3; i++ {
			logger.Logger.Println("I'm watching:", ctx.Current, i)
			<-ticker.C
		}
		progress = 1
		ctx.Attr.AddEnergy(-1)
		ctx.Attr.AddCuriosity(-2)
		ctx.Attr.AddOpenness(-3)
		return fsm.Done
	}
}
