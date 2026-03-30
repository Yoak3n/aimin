package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
)

type PersistRecord struct {
	At   string         `json:"at"`
	Data map[string]any `json:"data,omitempty"`
}

type PersistChannelState struct {
	Summary string          `json:"summary,omitempty"`
	Records []PersistRecord `json:"records,omitempty"`
}

type PersistState struct {
	Channels map[string]PersistChannelState `json:"channels,omitempty"`
}

type PersistStore struct {
	mu   sync.Mutex
	path string
}

func NewPersistStore() *PersistStore {
	return &PersistStore{
		path: defaultPersistPath(),
	}
}

func (p *PersistStore) Append(channel string, data map[string]any) {
	p.AppendWithLimits(channel, data, 40, 2000)
}

func (p *PersistStore) AppendWithLimits(channel string, data map[string]any, keep int, maxSummaryChars int) {
	if p == nil {
		return
	}
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	st, _ := p.loadLocked()
	if st.Channels == nil {
		st.Channels = make(map[string]PersistChannelState)
	}
	cs := st.Channels[channel]
	cs.Records = append(cs.Records, PersistRecord{
		At:   time.Now().UTC().Format(time.RFC3339),
		Data: data,
	})
	cs = compressChannel(cs, keep, maxSummaryChars)
	st.Channels[channel] = cs
	_ = p.saveLocked(st)
}

func (p *PersistStore) GetChannel(channel string) (PersistChannelState, bool) {
	if p == nil {
		return PersistChannelState{}, false
	}
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return PersistChannelState{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	st, err := p.loadLocked()
	if err != nil {
		return PersistChannelState{}, false
	}
	cs, ok := st.Channels[channel]
	if !ok {
		return PersistChannelState{}, false
	}
	out := PersistChannelState{Summary: cs.Summary}
	if len(cs.Records) > 0 {
		out.Records = append([]PersistRecord(nil), cs.Records...)
	}
	return out, true
}

func compressChannel(cs PersistChannelState, keep int, maxSummaryChars int) PersistChannelState {
	if keep <= 0 {
		keep = 40
	}
	if maxSummaryChars <= 0 {
		maxSummaryChars = 2000
	}
	if len(cs.Records) <= keep {
		return cs
	}

	excess := len(cs.Records) - keep
	old := cs.Records[:excess]
	cs.Records = cs.Records[excess:]

	sb := strings.Builder{}
	if strings.TrimSpace(cs.Summary) != "" {
		sb.WriteString(strings.TrimSpace(cs.Summary))
		sb.WriteString("\n")
	}
	sb.WriteString("archive:\n")
	for _, r := range old {
		sb.WriteString("- ")
		if strings.TrimSpace(r.At) != "" {
			sb.WriteString(r.At)
			sb.WriteString(" ")
		}
		sb.WriteString(oneLineJSON(r.Data, 200))
		sb.WriteString("\n")
		if sb.Len() >= maxSummaryChars {
			break
		}
	}

	out := strings.TrimSpace(sb.String())
	if len(out) > maxSummaryChars {
		out = out[:maxSummaryChars]
	}
	cs.Summary = strings.TrimSpace(out)
	return cs
}

func oneLineJSON(v any, max int) string {
	buf, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(buf))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}

func defaultPersistPath() string {
	ws := ""
	cfg := config.GlobalConfiguration()
	if cfg != nil && cfg.Workspace != nil {
		ws = strings.TrimSpace(cfg.Workspace.Path)
	}
	if ws == "" {
		ws = "./default_workspace"
	}
	return filepath.Join(ws, "fsm-state.json")
}

func (p *PersistStore) loadLocked() (PersistState, error) {
	path := strings.TrimSpace(p.path)
	if path == "" {
		path = defaultPersistPath()
		p.path = path
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PersistState{Channels: map[string]PersistChannelState{}}, nil
		}
		return PersistState{}, err
	}
	if len(buf) == 0 {
		return PersistState{Channels: map[string]PersistChannelState{}}, nil
	}
	var st PersistState
	if err := json.Unmarshal(buf, &st); err != nil {
		return PersistState{}, err
	}
	if st.Channels == nil {
		st.Channels = map[string]PersistChannelState{}
	}
	return st, nil
}

func (p *PersistStore) saveLocked(st PersistState) error {
	path := strings.TrimSpace(p.path)
	if path == "" {
		path = defaultPersistPath()
		p.path = path
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}
