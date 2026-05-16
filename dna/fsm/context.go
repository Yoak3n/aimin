package fsm

import (
	"math"

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
	contextHasTaskKey  = "__fsm_has_pending_task__"
)

func NewContext() *Context {
	return &Context{
		Data:    make(map[string]any),
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
		c.Data = make(map[string]any)
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
		c.Data = make(map[string]any)
	}
	c.Data[contextLastDoneKey] = stateID
	// bias: 惩罚当前状态、鼓励对立状态。
	// 稳态惩罚 ≈ -0.15/(1-0.6) = -0.375 → 权重乘数 0.625×
	// 稳态奖励 ≈  0.10/(1-0.6) =  0.250 → 权重乘数 1.250×
	// 最大偏置比 ≈ 2:1，属性影响力保持主导地位 (修复前 ≈ 19:1)
	c.DecayStateBias(0.6)
	if stateID != "" {
		c.AddStateBias(stateID, -0.15)
	}
	for _, other := range oppositeStateIDs {
		if other == "" || other == stateID {
			continue
		}
		c.AddStateBias(other, 0.1)
	}
}

func (c *Context) SetHasPendingTask(v bool) {
	if c == nil {
		return
	}
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	if v {
		c.Data[contextHasTaskKey] = true
		return
	}
	delete(c.Data, contextHasTaskKey)
}

func (c *Context) HasPendingTask() bool {
	if c == nil || c.Data == nil {
		return false
	}
	v, ok := c.Data[contextHasTaskKey]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

type AttrFactor struct {
	Weight   float64
	Inverted bool
}

type AttrBiasProfile struct {
	Base                 float64
	Energy               AttrFactor
	Curiosity            AttrFactor
	Openness             AttrFactor
	EnergyCuriosity      float64
	EnergyOpenness       float64
	CuriosityOpenness    float64
	InvEnergyCuriosity   float64
	InvEnergyOpenness    float64
	InvCuriosityOpenness float64
}

func clampAttr01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func attrContrib(norm float64, f AttrFactor) float64 {
	if f.Inverted {
		return f.Weight * (1 - norm)
	}
	return f.Weight * norm
}

func (c *Context) CalcAttrBias(profile AttrBiasProfile) float64 {
	if c == nil || c.Attr == nil {
		return math.Max(0, profile.Base)
	}
	e := clampAttr01(c.Attr.Energy / 100)
	cur := clampAttr01(c.Attr.Curiosity / 100)
	o := clampAttr01(c.Attr.Openness / 100)
	w := profile.Base
	w += attrContrib(e, profile.Energy)
	w += attrContrib(cur, profile.Curiosity)
	w += attrContrib(o, profile.Openness)
	w += profile.EnergyCuriosity * e * cur
	w += profile.EnergyOpenness * e * o
	w += profile.CuriosityOpenness * cur * o
	w += profile.InvEnergyCuriosity * (1 - e) * cur
	w += profile.InvEnergyOpenness * (1 - e) * o
	w += profile.InvCuriosityOpenness * (1 - cur) * o

	if w < 0 {
		w = 0
	}
	return w
}

func (c *Context) CalcStateWeight(stateID string, profile AttrBiasProfile, repeatPenalty float64) float64 {
	w := c.CalcAttrBias(profile)
	w *= 1 + c.GetStateBias(stateID)
	if repeatPenalty > 0 && repeatPenalty < 1 && c.LastDoneState() == stateID {
		w *= repeatPenalty
	}
	if w < 0 {
		w = 0
	}
	return w
}