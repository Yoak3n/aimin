package decision

import (
	"time"

	"github.com/Yoak3n/aimin/dna/fsm"
)

func NewSleepNode() *fsm.WorkState {
	return fsm.NewWorkState(Sleep, Sleep, func(ctx *fsm.Context) string {
		time.Sleep(time.Second * 5)
		return fsm.Done
	}, func(ctx *fsm.Context) bool {
		v, ok := ctx.Data[Root]
		return ok && v == Sleep
	})
}
