package agent

import (
	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/schema"
)

type ExecPlanAgent struct {
	Hooks *AgentHooks
}

func (e *ExecPlanAgent) ensureHooks() *AgentHooks {
	if e.Hooks == nil {
		e.Hooks = NewAgentHooks()
	}
	return e.Hooks
}

func (e *ExecPlanAgent) RegisterThoughtHandler(h func(string)) {
	e.ensureHooks().AddThoughtHandler(h)
}

func (e *ExecPlanAgent) RegisterActionHandler(h func(string)) {
	e.ensureHooks().AddActionHandler(h)
}

func (e *ExecPlanAgent) RegisterToolResultHandler(h func(action string, result string, err error)) {
	e.ensureHooks().AddToolResultHandler(h)
}

func (e *ExecPlanAgent) RegisterFinalAnswerHandler(h func(string)) {
	e.ensureHooks().AddFinalAnswerHandler(h)
}

func (e *ExecPlanAgent) RegisterAssistantDeltaHandler(h func(string) error) {
	e.ensureHooks().AddAssistantDeltaHandler(h)
}

func (e *ExecPlanAgent) RegisterLLMResponseHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, response string)) {
	e.ensureHooks().AddLLMResponseHandler(h)
}

func (e *ExecPlanAgent) Run(input string) {
	if input == "" {
		return
	}
}

type ExecPlanSubAgent struct {
	Hooks *AgentHooks
	Mcp   *mcp.McpHUB
}

func (es *ExecPlanSubAgent) ensureHooks() *AgentHooks {
	if es.Hooks == nil {
		es.Hooks = NewAgentHooks()
	}
	return es.Hooks
}

func (es *ExecPlanSubAgent) RegisterThoughtHandler(h func(string)) {
	es.ensureHooks().AddThoughtHandler(h)
}

func (es *ExecPlanSubAgent) RegisterActionHandler(h func(string)) {
	es.ensureHooks().AddActionHandler(h)
}

func (es *ExecPlanSubAgent) RegisterToolResultHandler(h func(action string, result string, err error)) {
	es.ensureHooks().AddToolResultHandler(h)
}

func (es *ExecPlanSubAgent) RegisterFinalAnswerHandler(h func(string)) {
	es.ensureHooks().AddFinalAnswerHandler(h)
}

func (es *ExecPlanSubAgent) RegisterAssistantDeltaHandler(h func(string) error) {
	es.ensureHooks().AddAssistantDeltaHandler(h)
}

func (es *ExecPlanSubAgent) RegisterLLMResponseHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, response string)) {
	es.ensureHooks().AddLLMResponseHandler(h)
}

func (es *ExecPlanSubAgent) Run(input string) {
	if input == "" {
		return
	}
}
