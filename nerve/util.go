package nerve

import (
	"encoding/json"
	"strings"

	"github.com/Yoak3n/aimin/blood/schema"
)

func extractQuestionAndThoughts(messages []schema.OpenAIMessage) (q string, tsj string) {
	lastQuestionIdx := -1
	for i, m := range messages {
		if m.Role != schema.OpenAIMessageRoleUser {
			continue
		}
		candidate := extractQuestion(m.Content)
		if candidate != "" {
			q = candidate
			lastQuestionIdx = i
		}
	}

	ts := make([]string, 0)
	for i := lastQuestionIdx + 1; i >= 0 && i < len(messages); i++ {
		m := messages[i]
		if m.Role != schema.OpenAIMessageRoleAssistant {
			continue
		}
		t := extractThought(m)
		if t != "" {
			ts = append(ts, t)
			continue
		}
		c := strings.TrimSpace(m.Content)
		if c != "" {
			ts = append(ts, c)
		}
	}
	tsj = strings.Join(ts, "\n")
	return q, tsj
}

func extractQuestion(content string) string {
	return strings.TrimSpace(content)
}

func extractThought(m schema.OpenAIMessage) string {
	if len(m.Reasoning) > 0 {
		var s string
		if err := json.Unmarshal(m.Reasoning, &s); err == nil {
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
		raw := strings.TrimSpace(string(m.Reasoning))
		if raw != "" && raw != "null" {
			return raw
		}
	}
	return strings.TrimSpace(m.Content)
}
