package decision

import (
	"time"

	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewSleepNode() *fsm.WorkState {
	node := fsm.NewWorkState(Sleep, Sleep, makeSleepNode(), func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[rootRouterKey]
		return ok && v == Sleep
	})

	return node
}

func makeSleepNode() fsm.WorkAction {
	progress := 1
	return func(ctx *fsm.Context) string {
		ctx.Data[ExploreChoice] = ""
		for i := progress; i < 6; i++ {
			switch i {
			case 1:
				time.Sleep(time.Second)
				progress++
			case 2:
				time.Sleep(time.Second)
				progress++
			case 3:
				time.Sleep(time.Second)
				progress++
			case 4:
				time.Sleep(time.Second)
				progress++
			case 5:
				time.Sleep(time.Second)
				ctx.Attr.SetEnergy(ctx.Attr.Energy + 10)
				return fsm.Done
			}
		}
		return fsm.Interrupt
	}
}
