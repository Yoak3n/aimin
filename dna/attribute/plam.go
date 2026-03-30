package attribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
)

type MinAttribute struct {
	mu         sync.RWMutex `json:"-"`
	Curiosity  float64      `json:"curiosity"`
	Energy     float64      `json:"energy"`
	Openness   float64      `json:"openness"`
	Lifespan   float64      `json:"lifespan"`
	LastUpdate int64        `json:"last_update"`
}

const attributePath = "data/cache"

func NewMinAttribute() *MinAttribute {
	c := loadFromCache()
	return &MinAttribute{
		Curiosity:  c.Curiosity,
		Energy:     c.Energy,
		Openness:   c.Openness,
		Lifespan:   c.Lifespan,
		LastUpdate: c.LastUpdate,
	}
}

func (m *MinAttribute) SetCuriosity(v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Curiosity = v
	m.saveCacheLocked()
}

func (m *MinAttribute) SetEnergy(v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Energy = v
	m.saveCacheLocked()
}

func (m *MinAttribute) SetOpenness(v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Openness = v
	m.saveCacheLocked()
}

func (m *MinAttribute) AddCuriosity(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Curiosity = clamp01To100(m.Curiosity + delta)
	m.saveCacheLocked()
}

func (m *MinAttribute) AddEnergy(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Energy = clamp01To100(m.Energy + delta)
	m.saveCacheLocked()
}

func (m *MinAttribute) AddOpenness(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Openness = clamp01To100(m.Openness + delta)
	m.saveCacheLocked()
}

func (m *MinAttribute) SaveCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCacheLocked()
}

func clamp01To100(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func (m *MinAttribute) saveCacheLocked() {
	if m.LastUpdate == 0 {
		m.Lifespan = 0
	} else {
		nl := time.Since(time.Unix(m.LastUpdate, 0))
		m.Lifespan += nl.Seconds()
	}
	m.LastUpdate = time.Now().Unix()
	buf, _ := json.Marshal(m)
	fp, err := os.Open(fmt.Sprintf("%s/attribute.json", attributePath))
	if err != nil {
		return
	}
	defer fp.Close()
	_, err = fp.Write(buf)
}

func loadFromCache() MinAttribute {
	_, err := os.ReadDir(attributePath)
	if err != nil {
		buf, e := os.ReadFile(fmt.Sprintf("%s/attribute.json", attributePath))
		if e != nil {
			a := MinAttribute{}
			err = json.Unmarshal(buf, &a)
			if err != nil {
				return defaultAttribute()
			}
			return MinAttribute{
				Curiosity:  a.Curiosity,
				Energy:     a.Energy,
				Openness:   a.Openness,
				Lifespan:   a.Lifespan,
				LastUpdate: a.LastUpdate,
			}
		}
		return defaultAttribute()
	} else if errors.Is(err, os.ErrNotExist) {
		e := util.CreateDirNotExists(attributePath)
		a := defaultAttribute()
		if e != nil {
			return MinAttribute{
				Curiosity:  a.Curiosity,
				Energy:     a.Energy,
				Openness:   a.Openness,
				Lifespan:   a.Lifespan,
				LastUpdate: a.LastUpdate,
			}
		}
		buf, e := json.Marshal(&a)
		if e != nil {
			return MinAttribute{
				Curiosity:  a.Curiosity,
				Energy:     a.Energy,
				Openness:   a.Openness,
				Lifespan:   a.Lifespan,
				LastUpdate: a.LastUpdate,
			}
		}
		fp, e := os.Create(fmt.Sprintf("%s/attribute.json", attributePath))
		if e != nil {
			return MinAttribute{
				Curiosity:  a.Curiosity,
				Energy:     a.Energy,
				Openness:   a.Openness,
				Lifespan:   a.Lifespan,
				LastUpdate: a.LastUpdate,
			}
		}
		defer fp.Close()
		_, e = fp.Write(buf)
		if e != nil {
			return MinAttribute{
				Curiosity:  a.Curiosity,
				Energy:     a.Energy,
				Openness:   a.Openness,
				Lifespan:   a.Lifespan,
				LastUpdate: a.LastUpdate,
			}
		}
	}
	return defaultAttribute()
}

func defaultAttribute() MinAttribute {
	return MinAttribute{
		Curiosity: 60,
		Energy:    100,
		Openness:  60,
		Lifespan:  0,
	}
}
