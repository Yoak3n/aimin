package agent

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/schema"
)

type ConversationAgent struct {
	Base     *ReActAgent
	Messages []schema.OpenAIMessage
	MaxTurns int
}

type ConversationTurn struct {
	Thought     string
	FinalAnswer string
}

func NewConversationAgent(base *ReActAgent) *ConversationAgent {
	if base == nil {
		base = NewAgent()
	}
	return &ConversationAgent{
		Base:     base,
		Messages: make([]schema.OpenAIMessage, 0, 16),
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
		h.AddFinalAnswerHandler(func(string) {})
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
	c.Messages = append([]schema.OpenAIMessage(nil), c.Messages[len(c.Messages)-maxMessages:]...)
}

func formatCompactAssistant(thought string, finalAnswer string) string {
	thought = strings.TrimSpace(thought)
	finalAnswer = strings.TrimSpace(finalAnswer)
	if thought == "" && finalAnswer == "" {
		return ""
	}
	if thought == "" {
		return fmt.Sprintf("<final_answer>%s</final_answer>", finalAnswer)
	}
	if finalAnswer == "" {
		return fmt.Sprintf("<thought>%s</thought>", thought)
	}
	return fmt.Sprintf("<thought>%s</thought>\n<final_answer>%s</final_answer>", thought, finalAnswer)
}
