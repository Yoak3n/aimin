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
		shouldInterrupt := func() bool {
			return ctx != nil && ctx.HasPendingTask()
		}
		for i := progress; i < 6; i++ {
			switch i {
			case 1:
				if shouldInterrupt() {
					return fsm.Interrupt
				}
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(2)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 2:
				if shouldInterrupt() {
					return fsm.Interrupt
				}
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(2)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 3:
				if shouldInterrupt() {
					return fsm.Interrupt
				}
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(2)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 4:
				if shouldInterrupt() {
					return fsm.Interrupt
				}
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(2)
				logger.Logger.Println("Sleep:", i)
				progress++
			case 5:
				if shouldInterrupt() {
					return fsm.Interrupt
				}
				time.Sleep(time.Second)
				ctx.Attr.AddEnergy(2)
				logger.Logger.Println("Sleep:", i)
				progress = 1
				return fsm.Done
			}
		}
		return fsm.Interrupt
	}
}
