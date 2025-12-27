package memory

import (
	"time"
)

type Hippocampus struct {
	Temporary    map[string]Temporary `json:"temporary"`
	temporaryMap map[string]string
	Enduring     map[string]Enduring `json:"enduring"`
}

func NewHippocampus() *Hippocampus {
	h := &Hippocampus{
		Temporary: make(map[string]Temporary),
		Enduring:  make(map[string]Enduring),
	}
	return h
}

// AddTemporary adds a temporary memory to the hippocampus.
func (h *Hippocampus) AddTemporary(m *Memory, expired time.Time) {
	h.Temporary[m.Id] = Temporary{
		Memory:  m,
		Expired: expired,
	}
	h.temporaryMap[m.Topic] = m.Id
}

// GetTemporary returns a temporary memory by id.
func (h *Hippocampus) GetTemporary(id string) (Temporary, bool) {
	temporary, ok := h.Temporary[id]
	return temporary, ok
}

// AddEnduring adds an enduring memory to the hippocampus.
func (h *Hippocampus) AddEnduring(e *Enduring) {
	h.Enduring[e.Id] = Enduring{
		Memory: e.Memory,
	}
}

// GetEnduring returns an enduring memory by id.
func (h *Hippocampus) GetEnduring(id string) (Enduring, bool) {
	enduring, ok := h.Enduring[id]
	return enduring, ok
}

// GetEntityOfEnduring returns all enduring memories by entity id.
func (h *Hippocampus) GetEntityOfEnduring(entityId string) ([]Enduring, bool) {
	var endurings []Enduring
	//for _, enduring := range h.Enduring {
	//	if enduring.EntityId == entityId {
	//		endurings = append(endurings, enduring)
	//	}
	//}
	return endurings, len(endurings) > 0
}
