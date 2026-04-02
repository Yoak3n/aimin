package interactive

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/agent"
	"github.com/Yoak3n/aimin/blood/agent/workspace"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
)

// 当状态机通过网络搜索或主动提问得到一个探索的粗略回答后
// 该回答会被作为探索的初始状态，使用ReActAgent进行深入探索
// 1. 当通过网络搜索获得一个搜索页的内容时，agent根据搜索页的内容或链接继续探索，直到给出一个能回答最初提问的答案
// 2. 当通过主动提问获得一个回答时，agent根据回答继续探索，直到给出一个能回答最初提问的答案
func ExploreRun(question string, answer string, strategy string) (string, error, []schema.OpenAIMessage) {
	a := agent.NewAgent(workspace.PromptPurposeReAct)
	a.SetContextChoice(workspace.Single)
	a.RegisterThoughtHandler(func(thought string) {
		logger.Logger.Println("🧩Thought:", thought)
	})

	msgs := make([]schema.OpenAIMessage, 0)
	msgs = append(msgs, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleUser,
		Content: fmt.Sprintf("<question>agent提出的问题：%s</question>", question),
	})
	actionText := ""
	if strategy == "web_search" {
		actionText = fmt.Sprintf("Web(search,query=\"%s\")", question)
	} else {
		actionText = fmt.Sprintf("ask_user(\"%s\")", question)
	}
	msgs = append(msgs, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleAssistant,
		Content: fmt.Sprintf("<thought>根据agent提出的问题，我采取“%s”的策略，希望得到一个粗略的回答，之后我将根据这个回答决定是否继续探索，直到能够完整清晰地帮忙补充（以原本回答者的角度）回答好最初的提问。</thought>\n<action>%s</action>", strategy, actionText),
	})
	msgs = append(msgs, schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleUser,
		Content: fmt.Sprintf("<observation>%s</observation>", answer),
	})
	result, err := a.RunWithMessages(msgs)
	if err != nil {
		return "", err, nil
	}
	return result.FinalAnswer, nil, msgs
}
