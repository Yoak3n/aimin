package nerve

import (
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

func extractQuestionAndThoughts(messages []schema.OpenAIMessage) (q string, tsj string) {
	lastQuestionIdx := -1
	for i, m := range messages {
		if m.Role != schema.OpenAIMessageRoleUser {
			continue
		}
		candidate := strings.TrimSpace(helper.ExtractContentByTag(m.Content, "question"))
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
		t := strings.TrimSpace(helper.ExtractContentByTag(m.Content, "thought"))
		if t != "" {
			ts = append(ts, t)
		}
	}
	tsj = strings.Join(ts, "\n")
	return q, tsj
}
