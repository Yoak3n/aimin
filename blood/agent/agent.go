package agent

import "github.com/Yoak3n/aimin/blood/schema"

type Agent interface {
	Run(input string)
	RegisterThoughtHandler(h func(string))
	RegisterActionHandler(h func(string))
	RegisterToolResultHandler(h func(action string, result string, err error))
	RegisterFinalAnswerHandler(h func(string))
	RegisterAssistantDeltaHandler(h func(string) error)
	RegisterLLMResponseHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, response string))
}
