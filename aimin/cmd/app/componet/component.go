package componet

import (
	"sync"

	"github.com/Yoak3n/aimin/dna/decision"
	"github.com/Yoak3n/aimin/dna/fsm"
)

type Component struct {
	fsm *fsm.FSM
}

var component *Component
var once *sync.Once

func GetGlobalComponent() *Component {
	once.Do(func() {
		component = &Component{
			fsm: decision.NewStateTree(),
		}
	})
	return component
}

func (c *Component) FSM() *fsm.FSM {
	return c.fsm
}
