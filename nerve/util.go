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
		if candidate == "" || isSyntheticRetryMessage(candidate) {
			continue
		}
		q = candidate
		lastQuestionIdx = i
	}

	ts := make([]string, 0)
	// TODO 因为think模型的reasoning总是存在，所以这里只提取reasoning作为thought，可能其他模型会有不同
	for i := lastQuestionIdx + 1; i >= 0 && i < len(messages); i++ {
		m := messages[i]
		if m.Role != schema.OpenAIMessageRoleAssistant {
			continue
		}
		t := extractThought(m)
		if t != "" {
			ts = append(ts, t)
		}
	}
	tsj = strings.Join(ts, "\n")
	return q, tsj
}

func isSyntheticRetryMessage(s string) bool {
	if strings.Contains(s, "请继续输出完整答复") {
		return true
	}
	if strings.Contains(s, "请重新发起该工具的 tool call") {
		return true
	}
	return false
}

func extractQuestion(content string) string {
	return strings.TrimSpace(content)
}

func extractThought(m schema.OpenAIMessage) string {
	if s := strings.TrimSpace(m.Reasoning); s != "" && s != "null" {
		return s
	}
	return ""
}
