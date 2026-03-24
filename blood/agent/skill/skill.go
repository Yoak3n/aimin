package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	hub  *SkillHUB
	once sync.Once
)

type Skill struct {
	Name     string
	Desc     string
	Content  string
	Location string
}

type SkillHUB struct {
	Skills map[string]*Skill
	Active string
	mu     sync.RWMutex
}

func NewSkillHUB() *SkillHUB {
	h := &SkillHUB{
		Skills: make(map[string]*Skill),
		mu:     sync.RWMutex{},
	}
	return h
}

func GlobalSkillHUB() *SkillHUB {
	once.Do(func() {
		hub = NewSkillHUB()
		hub.ScanSkills("./skills")
	})
	return hub
}

// LoadSkills 从指定目录加载所有包含 SKILL.md 的技能
func (h *SkillHUB) ScanSkills(baseDir string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read skill directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillPath := filepath.Join(baseDir, entry.Name())
		skillFile := filepath.Join(skillPath, "SKILL.md")

		if _, err := os.Stat(skillFile); err == nil {
			content, err := os.ReadFile(skillFile)
			if err != nil {
				continue
			}

			s := parseSkillMetadata(string(content))
			if s != nil {
				s.Location = skillPath
				h.Skills[s.Name] = s
			}
		}
	}

	return nil
}

// parseSkillMetadata 从 SKILL.md 内容中提取元数据
// 约定元数据格式：
// ---
// name: skill_name
// desc: skill description
// ---
func parseSkillMetadata(content string) *Skill {
	parts := strings.Split(content, "---")
	if len(parts) < 3 {
		return nil
	}
	metadata := parts[1]

	s := &Skill{}
	for line := range strings.SplitSeq(metadata, "\n") {
		line = strings.TrimSpace(line)
		if name, ok := strings.CutPrefix(line, "name:"); ok {
			s.Name = strings.TrimSpace(name)
		} else if desc, ok := strings.CutPrefix(line, "desc:"); ok {
			s.Desc = strings.TrimSpace(desc)
		}
	}

	if s.Name == "" {
		return nil
	}
	s.Content = strings.TrimSpace(parts[2])
	return s
}

// GetSkill 根据名称获取技能
func (h *SkillHUB) GetSkill(name string) *Skill {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Skills[name]
}

// RenderSkillsList 渲染所有技能列表
func (h *SkillHUB) RenderSkillsList() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.Skills) == 0 {
		return "无可用技能"
	}

	var sb strings.Builder
	for _, s := range h.Skills {
		fmt.Fprintf(&sb, "- %s: %s\n", s.Name, s.Desc)
	}
	return sb.String()
}

func (h *SkillHUB) GetSkills() []*Skill {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var skills []*Skill
	for _, s := range h.Skills {
		skills = append(skills, s)
	}
	return skills
}

func (h *SkillHUB) LoadSkill(name ...string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(name) == 0 {
		return ""
	}
	var cs strings.Builder
	for _, n := range name {
		if s, ok := h.Skills[n]; ok {
			fmt.Fprintf(&cs, "<name>%s</name><instruction>%s</instruction><location>%s</location>", s.Name, s.Content, s.Location)
		}
	}
	return cs.String()
}
