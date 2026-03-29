package nerve

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/service/retrieval"
	"github.com/Yoak3n/aimin/nerve/controller"
	"github.com/Yoak3n/aimin/nerve/memory"
)

type Nerve struct {
	Hippocampus *memory.Hippocampus
}

func ResponseHook(systemPrompt, answer string, messages []schema.OpenAIMessage) {
	cid := util.RandomIdWithPrefix("con")
	question, thoughts := extractQuestionAndThoughts(messages)
	go controller.InsertConversation(cid, systemPrompt, question, thoughts, answer)
	go controller.InsertSummaryMemory(systemPrompt, question, thoughts, answer, cid)
}

func QueryReleventConversations(input string) (string, error) {
	conversations, err := retrieval.VectorSearchConversations(input, 5)
	if err != nil {
		return "", err
	}
	if len(conversations) == 0 {
		return "no matches", nil
	}

	out := strings.Builder{}
	out.WriteString("<vector_search_results>\n")
	fmt.Fprintf(&out, "<query>%s</query>\n", compactOneLine(input, 240))
	for _, c := range conversations {
		sum := ""
		if s, err := helper.UseDB().GetSummaryMemoryTableRecordByLink(c.Id); err == nil {
			sum = strings.TrimSpace(s.Content)
		}
		q := compactOneLine(c.Question, 240)
		fmt.Fprintf(&out, "<conversation_summary id=%q>\n", c.Id)
		if sum != "" {
			fmt.Fprintf(&out, "<summary>%s</summary>\n", compactOneLine(sum, 600))
		} else {
			a := compactOneLine(c.Answer, 360)
			fmt.Fprintf(&out, "<summary>%s</summary>\n", compactOneLine(q+" / "+a, 600))
		}
		fmt.Fprintf(&out, "<question>%s</question>\n", q)
		out.WriteString("</conversation_summary>\n")
	}
	out.WriteString("</vector_search_results>")
	return out.String(), nil
}

func GetConversationByID(id string) (string, error) {
	rec, err := helper.UseDB().GetConversationByID(id)
	if err != nil {
		return "", err
	}

	q := compactOneLine(rec.Question, 240)
	a := compactOneLine(rec.Answer, 0)
	sys := compactOneLine(rec.System, 200)

	detail := strings.Builder{}
	fmt.Fprintf(&detail, "<conversation id=%q>\n", rec.Id)
	if sys != "" {
		fmt.Fprintf(&detail, "<system>%s</system>\n", sys)
	}
	fmt.Fprintf(&detail, "<question>%s</question>\n", q)
	fmt.Fprintf(&detail, "<answer>%s</answer>\n", a)
	fmt.Fprint(&detail, "</conversation>")
	return detail.String(), nil
}

func compactOneLine(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.Join(strings.Fields(s), " ")
	if maxRunes <= 0 {
		return s
	}
	rs := []rune(s)
	if len(rs) <= maxRunes {
		return s
	}
	return string(rs[:maxRunes]) + "..."
}
