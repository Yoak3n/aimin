package nerve

import (
	"fmt"
	"log"
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
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
	conversations := make([]schema.ConversationRecord, 0)
	embeding, err := helper.UseLLM().Embedding([]string{input})
	if err != nil {
		log.Fatal(err)
	}
	conversations, err = helper.UseDB().GetReleventConversationRecords(embeding[0])
	if err != nil {
		log.Fatal(err)
	}
	cl := strings.Builder{}
	for _, conversation := range conversations {
		fmt.Fprintf(&cl, "- [%s]%s---%s\n", conversation.Id, conversation.Question, conversation.Answer)
	}
	return cl.String(), nil
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
