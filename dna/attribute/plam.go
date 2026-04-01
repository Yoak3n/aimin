package attribute

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	buf, err := json.Marshal(m)
	if err != nil {
		return
	}
	filePath := filepath.Join(attributePath, "attribute.json")
	_ = util.CreateDirNotExists(attributePath)
	_ = os.WriteFile(filePath, buf, 0644)
}

func loadFromCache() MinAttribute {
	filePath := filepath.Join(attributePath, "attribute.json")
	_ = util.CreateDirNotExists(attributePath)
	buf, err := os.ReadFile(filePath)
	if err != nil || len(buf) == 0 {
		a := defaultAttribute()
		if b, e := json.Marshal(&a); e == nil {
			_ = util.CreateFileNotExists(filePath, b)
		}
		return a
	}

	var a MinAttribute
	if err := json.Unmarshal(buf, &a); err != nil {
		da := defaultAttribute()
		if b, e := json.Marshal(&da); e == nil {
			_ = os.WriteFile(filePath, b, 0644)
		}
		return da
	}
	return MinAttribute{
		Curiosity:  a.Curiosity,
		Energy:     a.Energy,
		Openness:   a.Openness,
		Lifespan:   a.Lifespan,
		LastUpdate: a.LastUpdate,
	}
}

func defaultAttribute() MinAttribute {
	return MinAttribute{
		Curiosity: 60,
		Energy:    100,
		Openness:  60,
		Lifespan:  0,
	}
}
