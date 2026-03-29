package agent

import "github.com/Yoak3n/aimin/blood/schema"

type AgentHooks struct {
	ThoughtHandlers        []func(string)
	ActionHandlers         []func(string)
	ToolResultHandlers     []func(action string, result string, err error)
	FinalAnswerHandlers    []func(systemPrompt string, messages []schema.OpenAIMessage, finalAnswer string)
	AssistantDeltaHandlers []func(string) error
	LLMResponseHandlers    []func(systemPrompt string, messages []schema.OpenAIMessage, response string)
}

func NewAgentHooks() *AgentHooks {
	return &AgentHooks{}
}

func (h *AgentHooks) IsEmpty() bool {
	return len(h.ThoughtHandlers) == 0 &&
		len(h.ActionHandlers) == 0 &&
		len(h.ToolResultHandlers) == 0 &&
		len(h.FinalAnswerHandlers) == 0 &&
		len(h.AssistantDeltaHandlers) == 0 &&
		len(h.LLMResponseHandlers) == 0
}

func (h *AgentHooks) AddThoughtHandler(f func(string)) {
	if f == nil {
		return
	}
	h.ThoughtHandlers = append(h.ThoughtHandlers, f)
}

func (h *AgentHooks) AddActionHandler(f func(string)) {
	if f == nil {
		return
	}
	h.ActionHandlers = append(h.ActionHandlers, f)
}

func (h *AgentHooks) AddToolResultHandler(f func(action string, result string, err error)) {
	if f == nil {
		return
	}
	h.ToolResultHandlers = append(h.ToolResultHandlers, f)
}

func (h *AgentHooks) AddFinalAnswerHandler(f func(systemPrompt string, messages []schema.OpenAIMessage, finalAnswer string)) {
	if f == nil {
		return
	}
	h.FinalAnswerHandlers = append(h.FinalAnswerHandlers, f)
}

func (h *AgentHooks) AddAssistantDeltaHandler(f func(string) error) {
	if f == nil {
		return
	}
	h.AssistantDeltaHandlers = append(h.AssistantDeltaHandlers, f)
}

func (h *AgentHooks) AddLLMResponseHandler(f func(systemPrompt string, messages []schema.OpenAIMessage, response string)) {
	if f == nil {
		return
	}
	h.LLMResponseHandlers = append(h.LLMResponseHandlers, f)
}

func (h *AgentHooks) EmitThought(v string) {
	for _, f := range h.ThoughtHandlers {
		go f(v)
	}
}

func (h *AgentHooks) EmitAction(v string) {
	for _, f := range h.ActionHandlers {
		go f(v)
	}
}

func (h *AgentHooks) EmitToolResult(action string, result string, err error) {
	for _, f := range h.ToolResultHandlers {
		go f(action, result, err)
	}
}

func (h *AgentHooks) EmitFinalAnswer(systemPrompt string, messages []schema.OpenAIMessage, finalAnswer string) {
	for _, f := range h.FinalAnswerHandlers {
		go f(systemPrompt, messages, finalAnswer)
	}
}

func (h *AgentHooks) EmitLLMResponse(systemPrompt string, messages []schema.OpenAIMessage, response string) {
	for _, f := range h.LLMResponseHandlers {
		go f(systemPrompt, messages, response)
	}
}

func (h *AgentHooks) EmitAssistantDelta(v string) error {
	for _, f := range h.AssistantDeltaHandlers {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}
