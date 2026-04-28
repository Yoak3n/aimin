package sandbox

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Proc struct {
	ID         string
	RunID      string
	ToolCallID string
	Action     string
	PID        int
	StartedAt  time.Time
	Cancel     context.CancelFunc
}

type Manager struct {
	nextID atomic.Uint64
	mu     sync.Mutex
	procs  map[string]*Proc
	byRun  map[string]map[string]struct{}
}

func NewManager() *Manager {
	return &Manager{
		procs: make(map[string]*Proc),
		byRun: make(map[string]map[string]struct{}),
	}
}

func (m *Manager) NewSandboxID() string {
	n := m.nextID.Add(1)
	return fmt.Sprintf("sb_%d_%d", time.Now().UnixNano(), n)
}

func (m *Manager) Register(p *Proc) {
	if m == nil || p == nil || p.ID == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.procs[p.ID] = p
	if p.RunID != "" {
		set, ok := m.byRun[p.RunID]
		if !ok {
			set = make(map[string]struct{})
			m.byRun[p.RunID] = set
		}
		set[p.ID] = struct{}{}
	}
}

func (m *Manager) Unregister(id string) {
	if m == nil || id == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.procs[id]
	if !ok {
		return
	}
	delete(m.procs, id)
	if p != nil && p.RunID != "" {
		if set, ok := m.byRun[p.RunID]; ok {
			delete(set, id)
			if len(set) == 0 {
				delete(m.byRun, p.RunID)
			}
		}
	}
}

func (m *Manager) Kill(id string) (bool, error) {
	if m == nil || id == "" {
		return false, nil
	}
	m.mu.Lock()
	p := m.procs[id]
	m.mu.Unlock()
	if p == nil {
		return false, nil
	}
	defer m.Unregister(id)

	if p.Cancel != nil {
		p.Cancel()
	}
	if p.PID > 0 {
		proc, err := os.FindProcess(p.PID)
		if err == nil && proc != nil {
			_ = proc.Kill()
		}
	}
	return true, nil
}

func (m *Manager) KillRun(runID string) (killed int) {
	if m == nil || runID == "" {
		return 0
	}
	m.mu.Lock()
	set := m.byRun[runID]
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	m.mu.Unlock()

	for _, id := range ids {
		ok, _ := m.Kill(id)
		if ok {
			killed++
		}
	}
	return killed
}
