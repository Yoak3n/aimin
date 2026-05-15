package nerve

import (
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
	if s := strings.TrimSpace(m.Reasoning); s != "" && s != "null" {
		return s
	}
	return strings.TrimSpace(m.Content)
}
