package nerve

import (
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

func extractQuestionAndThoughts(messages []schema.OpenAIMessage) (q string, tsj string) {
	ts := make([]string, 0)
	qed := false
	for _, m := range messages {
		switch m.Role {
		case schema.OpenAIMessageRoleUser:
			if qed {
				continue
			}
			q = strings.TrimSpace(helper.ExtractContentByTag(m.Content, "question"))
			qed = true
		case schema.OpenAIMessageRoleAssistant:
			ts = append(ts, strings.TrimSpace(helper.ExtractContentByTag(m.Content, "thought")))
		}
	}
	tsj = strings.Join(ts, "")
	return q, tsj
}
