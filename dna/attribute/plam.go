package attribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/util"
)

type MinAttribute struct {
	Curiosity  float64 `json:"curiosity"`
	Energy     float64 `json:"energy"`
	Openness   float64 `json:"openness"`
	Lifespan   float64 `json:"lifespan"`
	LastUpdate int64   `json:"last_update"`
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
	m.Curiosity = v
	m.SaveCache()
}

func (m *MinAttribute) SetEnergy(v float64) {
	m.Energy = v
	m.SaveCache()
}

func (m *MinAttribute) SetOpenness(v float64) {
	m.Openness = v
	m.SaveCache()
}

func (m *MinAttribute) SaveCache() {
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
			err := json.Unmarshal(buf, &a)
			if err != nil {
				return defaultAttribute()
			}
			return a
		}
		return defaultAttribute()
	} else if errors.Is(err, os.ErrNotExist) {
		e := util.CreateDirNotExists(attributePath)
		a := defaultAttribute()
		if e != nil {
			return a
		}
		buf, e := json.Marshal(&a)
		if e != nil {
			return a
		}
		fp, e := os.Create(fmt.Sprintf("%s/attribute.json", attributePath))
		defer fp.Close()
		if e != nil {
			return a
		}
		_, e = fp.Write(buf)
		if e != nil {
			return a
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
