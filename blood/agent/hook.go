package agent

type AgentHooks struct {
	ThoughtHandlers        []func(string)
	ActionHandlers         []func(string)
	ToolResultHandlers     []func(action string, result string, err error)
	FinalAnswerHandlers    []func(string)
	AssistantDeltaHandlers []func(string) error
	LLMResponseHandlers    []func(string)
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

func (h *AgentHooks) AddFinalAnswerHandler(f func(string)) {
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

func (h *AgentHooks) AddLLMResponseHandler(f func(string)) {
	if f == nil {
		return
	}
	h.LLMResponseHandlers = append(h.LLMResponseHandlers, f)
}

func (h *AgentHooks) EmitThought(v string) {
	for _, f := range h.ThoughtHandlers {
		f(v)
	}
}

func (h *AgentHooks) EmitAction(v string) {
	for _, f := range h.ActionHandlers {
		f(v)
	}
}

func (h *AgentHooks) EmitToolResult(action string, result string, err error) {
	for _, f := range h.ToolResultHandlers {
		f(action, result, err)
	}
}

func (h *AgentHooks) EmitFinalAnswer(v string) {
	for _, f := range h.FinalAnswerHandlers {
		f(v)
	}
}

func (h *AgentHooks) EmitLLMResponse(v string) {
	for _, f := range h.LLMResponseHandlers {
		f(v)
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
