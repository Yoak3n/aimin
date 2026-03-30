package fsm

import (
	"github.com/Yoak3n/aimin/dna/attribute"
	"github.com/Yoak3n/aimin/dna/persist"
)

// Context 运行上下文，用于传递数据
type Context struct {
	Data          map[string]interface{}
	Current       string
	OnStateChange func(string)
	Attr          *attribute.MinAttribute
	Persist       *persist.PersistStore
}

const (
	contextBiasKey     = "__fsm_state_bias__"
	contextLastDoneKey = "__fsm_last_done__"
)

func NewContext() *Context {
	return &Context{
		Data:    make(map[string]interface{}),
		Current: "",
		Attr:    attribute.NewMinAttribute(),
		Persist: persist.NewPersistStore(),
	}
}

func (c *Context) LastDoneState() string {
	if c == nil || c.Data == nil {
		return ""
	}
	if v, ok := c.Data[contextLastDoneKey]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Context) GetStateBias(stateID string) float64 {
	if c == nil || c.Data == nil {
		return 0
	}
	v, ok := c.Data[contextBiasKey]
	if !ok {
		return 0
	}
	m, ok := v.(map[string]float64)
	if !ok {
		return 0
	}
	return m[stateID]
}

func (c *Context) AddStateBias(stateID string, delta float64) {
	if c == nil {
		return
	}
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	m, ok := c.Data[contextBiasKey].(map[string]float64)
	if !ok || m == nil {
		m = make(map[string]float64)
		c.Data[contextBiasKey] = m
	}
	cur := m[stateID]
	next := cur + delta
	if next > 0.9 {
		next = 0.9
	} else if next < -0.9 {
		next = -0.9
	}
	m[stateID] = next
}

func (c *Context) DecayStateBias(factor float64) {
	if c == nil || c.Data == nil {
		return
	}
	if factor < 0 {
		factor = 0
	} else if factor > 1 {
		factor = 1
	}
	m, ok := c.Data[contextBiasKey].(map[string]float64)
	if !ok || m == nil {
		return
	}
	for k, v := range m {
		nv := v * factor
		if nv > -0.01 && nv < 0.01 {
			delete(m, k)
			continue
		}
		m[k] = nv
	}
}

func (c *Context) MarkStateDone(stateID string, oppositeStateIDs ...string) {
	if c == nil {
		return
	}
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data[contextLastDoneKey] = stateID
	c.DecayStateBias(0.85)
	if stateID != "" {
		c.AddStateBias(stateID, -0.35)
	}
	for _, other := range oppositeStateIDs {
		if other == "" || other == stateID {
			continue
		}
		c.AddStateBias(other, 0.2)
	}
}
