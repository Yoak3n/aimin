package componet

import (
	"sync"
	"time"

	"github.com/Yoak3n/aimin/dna/decision"
	"github.com/Yoak3n/aimin/dna/fsm"
)

type Component struct {
	fsm  *fsm.FSM
	log  chan string
}

var component *Component
var once sync.Once

func GetGlobalComponent() *Component {
	once.Do(func() {
		component = &Component{
			fsm: decision.NewStateTree(),
			log: make(chan string, 100),
		}
	})
	return component
}

func (c *Component) Start() {
	c.fsm.Start(decision.Root)
	for {
		c.fsm.Update()
		time.Sleep(time.Microsecond * 100)
	}
}

func (c *Component) FSM() *fsm.FSM {
	return c.fsm
}

func (c *Component) AddTask(data fsm.TaskData) {
	task := decision.NewTaskState()
	c.fsm.UpdateContext(decision.TaskDataKey, data)
	c.fsm.AddTask(task)
}
