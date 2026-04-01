package decision

import (
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewSleepNode() *fsm.WorkState {
	node := fsm.NewWorkState(Sleep, Sleep, makeSleepNode(), func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[rootRouterKey]
		return ok && v == Sleep
	})
	node.SetDoneHook(Sleep, Explore, Watch)

	return node
}

func makeSleepNode() fsm.WorkAction {
	progress := 1
	return func(ctx *fsm.Context) string {
		for i := progress; i < 6; i++ {
			switch i {
			case 1:
				time.Sleep(time.Second)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 2:
				time.Sleep(time.Second)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 3:
				time.Sleep(time.Second)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 4:
				time.Sleep(time.Second)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 5:
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(10)
				logger.Logger.Println("Sleep:", i)
				progress = 1
				return fsm.Done
			}
		}
		return fsm.Interrupt
	}
}
