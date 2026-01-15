package app

import (
	"sync"

	"github.com/Yoak3n/aimin/aimin/internal/service/router"
	"github.com/Yoak3n/aimin/dna/decision"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/gin-gonic/gin"
)

type Component struct {
	router *gin.Engine
	fsm    *fsm.FSM
}

var component *Component
var once *sync.Once

func GetGlobalComponent() *Component {
	once.Do(func() {
		component = &Component{
			router: router.InitRouter(),
			fsm:    decision.NewStateTree(),
		}
	})
	return component
}

func (c *Component) Router() *gin.Engine {
	return c.router
}

func (c *Component) FSM() *fsm.FSM {
	return c.fsm
}
