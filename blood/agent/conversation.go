package agent

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
)

type ConversationAgent struct {
	Base     *ReActAgent
	Messages []schema.OpenAIMessage
	MaxTurns int

	memoryFlushMu         sync.Mutex
	memoryFlushInProgress bool
	lastMemoryFlushAt     time.Time
}

type ConversationTurn struct {
	Thought     string
	FinalAnswer string
}

func NewConversationAgent(base *ReActAgent) *ConversationAgent {
	if base == nil {
		base = NewAgent()
	}
	recentConversations, err := helper.UseDB().GetRecentConversations(2)
	if err != nil {
		logger.Logger.Errorf("get recent conversations failed: %v", err)
	}
	messages := make([]schema.OpenAIMessage, 0, 16)
	for i := len(recentConversations) - 1; i >= 0; i-- {
		rec := recentConversations[i]
		q := strings.TrimSpace(rec.Question)
		a := strings.TrimSpace(rec.Answer)
		if q == "" {
			continue
		}
		messages = append(messages, schema.OpenAIMessage{
			Role:    schema.OpenAIMessageRoleUser,
			Content: fmt.Sprintf("<question>%s</question>", q),
		})
		if a != "" {
			messages = append(messages, schema.OpenAIMessage{
				Role:    schema.OpenAIMessageRoleAssistant,
				Content: formatCompactAssistant("", a),
			})
		}
	}
	return &ConversationAgent{
		Base:     base,
		Messages: messages,
		MaxTurns: 0,
	}
}

func (c *ConversationAgent) SetMaxTurns(maxTurns int) {
	c.MaxTurns = maxTurns
	c.trimToMaxTurns()
}

func (c *ConversationAgent) Ask(question string) (ConversationTurn, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return ConversationTurn{}, fmt.Errorf("question 不能为空")
	}

	runMessages := make([]schema.OpenAIMessage, 0, len(c.Messages)+1)
	runMessages = append(runMessages, c.Messages...)
	runMessages = append(runMessages, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleUser,
		Content: fmt.Sprintf("<question>%s</question>", question),
	})

	origHooks := c.Base.Hooks
	replacedHooks := false
	if origHooks == nil || origHooks.IsEmpty() {
		h := NewAgentHooks()
		h.AddFinalAnswerHandler(func(string, []schema.OpenAIMessage, string) {})
		c.Base.SetHooks(h)
		replacedHooks = true
	}
	res, err := c.Base.RunWithMessages(runMessages)
	if replacedHooks {
		c.Base.SetHooks(origHooks)
	}
	if err != nil {
		return ConversationTurn{}, err
	}

	c.Messages = append(c.Messages, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleUser,
		Content: fmt.Sprintf("<question>%s</question>", question),
	})
	c.Messages = append(c.Messages, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleAssistant,
		Content: formatCompactAssistant(res.Thought, res.FinalAnswer),
	})
	c.trimToMaxTurns()

	return ConversationTurn{
		Thought:     res.Thought,
		FinalAnswer: res.FinalAnswer,
	}, nil
}

func (c *ConversationAgent) trimToMaxTurns() {
	if c.MaxTurns <= 0 {
		return
	}
	maxMessages := c.MaxTurns * 2
	if len(c.Messages) <= maxMessages {
		return
	}
	dropped := append([]schema.OpenAIMessage(nil), c.Messages[:len(c.Messages)-maxMessages]...)
	c.Messages = append([]schema.OpenAIMessage(nil), c.Messages[len(c.Messages)-maxMessages:]...)

	c.maybeSilentFlushDailyMemory(dropped, maxMessages)
}

func formatCompactAssistant(_ string, finalAnswer string) string {
	finalAnswer = strings.TrimSpace(finalAnswer)
	if finalAnswer == "" {
		return ""
	}
	return fmt.Sprintf("<final_answer>%s</final_answer>", finalAnswer)
}

func (c *ConversationAgent) maybeSilentFlushDailyMemory(dropped []schema.OpenAIMessage, keepMessages int) {
	if c.Base == nil || c.Base.Mcp == nil {
		return
	}
	if len(dropped) < keepMessages {
		return
	}
	if len(dropped) < 6 {
		return
	}

	c.memoryFlushMu.Lock()
	if c.memoryFlushInProgress {
		c.memoryFlushMu.Unlock()
		return
	}
	if !c.lastMemoryFlushAt.IsZero() && time.Since(c.lastMemoryFlushAt) < 30*time.Minute {
		c.memoryFlushMu.Unlock()
		return
	}
	c.memoryFlushInProgress = true
	c.lastMemoryFlushAt = time.Now()
	c.memoryFlushMu.Unlock()

	go func() {
		defer func() {
			c.memoryFlushMu.Lock()
			c.memoryFlushInProgress = false
			c.memoryFlushMu.Unlock()
		}()

		input := buildDroppedConversationDigest(dropped, keepMessages)
		if strings.TrimSpace(input) == "" {
			return
		}

		origHooks := c.Base.Hooks
		h := NewAgentHooks()
		h.AddFinalAnswerHandler(func(string, []schema.OpenAIMessage, string) {})
		c.Base.SetHooks(h)
		_, _ = c.Base.RunWithMessages([]schema.OpenAIMessage{
			{
				Role:    schema.OpenAIMessageRoleUser,
				Content: fmt.Sprintf("<question>%s</question>", input),
			},
		})
		c.Base.SetHooks(origHooks)
	}()
}

func buildDroppedConversationDigest(dropped []schema.OpenAIMessage, keepMessages int) string {
	total := len(dropped)
	if total == 0 {
		return ""
	}
	tail := dropped
	if len(tail) > 20 {
		tail = tail[len(tail)-20:]
	}

	sb := strings.Builder{}
	fmt.Fprintf(&sb, "你正在进行一次静默记忆更新：对话上下文即将被压缩，丢弃 %d 条历史消息（保留窗口=%d）。\n", total, keepMessages)
	sb.WriteString("请从下面历史片段中提炼对未来有用的信息（偏好/长期事实/关键决定/进行中的任务与下一步），然后调用一次工具追加写入今日记忆：\n")
	sb.WriteString(`manage_memory(action="write_daily", content="...")` + "\n")
	sb.WriteString("content 要求：不超过 12 行；每行以 \"- \" 开头；避免复述细节；不要泄露敏感信息。\n")
	sb.WriteString("<dropped_conversation>\n")
	for _, m := range tail {
		role := string(m.Role)
		text := strings.TrimSpace(m.Content)
		text = truncateRunes(text, 800)
		fmt.Fprintf(&sb, "[%s] %s\n", role, text)
	}
	sb.WriteString("</dropped_conversation>")
	return sb.String()
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	return string(rs[:max]) + "..."
}
