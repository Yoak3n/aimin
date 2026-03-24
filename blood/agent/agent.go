package agent

import (
	"fmt"
	"strings"

	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/agent/workspace"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
)

type Agent struct {
	Mcp   *mcp.McpHUB
	Skill *skill.SkillHUB
	Hooks *AgentHooks
}

func NewAgent() *Agent {
	a := &Agent{
		Mcp:   mcp.GlobalMcpHUB(),
		Skill: skill.NewSkillHUB(),
		Hooks: NewAgentHooks(),
	}
	a.RegisterTool(mcp.FileOperationTool())
	a.RegisterTool(mcp.ShellCommandTool())
	a.RegisterTool(mcp.GlobTool())
	a.RegisterTool(mcp.GrepTool())
	a.RegisterTool(mcp.SkillTool())
	workspace.EnsureWorkspace()
	return a
}

func (a *Agent) RegisterTool(tool *mcp.Tool) {
	a.Mcp.RegisterTool(tool)
}

func (a *Agent) ensureHooks() *AgentHooks {
	if a.Hooks == nil {
		a.Hooks = NewAgentHooks()
	}
	return a.Hooks
}

func (a *Agent) SetHooks(hooks *AgentHooks) {
	if hooks == nil {
		a.Hooks = NewAgentHooks()
		return
	}
	a.Hooks = hooks
}

func (a *Agent) RegisterThoughtHandler(h func(string)) {
	a.ensureHooks().AddThoughtHandler(h)
}

func (a *Agent) RegisterActionHandler(h func(string)) {
	a.ensureHooks().AddActionHandler(h)
}

func (a *Agent) RegisterToolResultHandler(h func(action string, result string, err error)) {
	a.ensureHooks().AddToolResultHandler(h)
}

func (a *Agent) RegisterFinalAnswerHandler(h func(string)) {
	a.ensureHooks().AddFinalAnswerHandler(h)
}

func (a *Agent) RegisterAssistantDeltaHandler(h func(string) error) {
	a.ensureHooks().AddAssistantDeltaHandler(h)
}

func (a *Agent) RegisterLLMResponseHandler(h func(string)) {
	a.ensureHooks().AddLLMResponseHandler(h)
}

func (a *Agent) Run(input string) {
	hooks := a.ensureHooks()
	noHooks := hooks.IsEmpty()
	if input == "" {
		fmt.Println("输入为空")
		return
	}
	messages := []schema.OpenAIMessage{
		{
			Role:    schema.OpenAIMessageRoleUser,
			Content: fmt.Sprintf("<question>%s</question>", input),
		},
	}
	// sp := a.RenderSysytemPrompt()
	wc := workspace.NewWorkspaceContext()
	for {
		sp := wc.String()
		var onDelta func(string) error
		if len(hooks.AssistantDeltaHandlers) > 0 {
			onDelta = hooks.EmitAssistantDelta
		}
		res, err := helper.UseLLM().ChatStream(messages, onDelta, sp)
		if err != nil {
			logger.Logger.Error("LLM调用失败", err)
			break
		}
		if len(hooks.LLMResponseHandlers) > 0 {
			hooks.EmitLLMResponse(res)
		}
		messages = append(messages, schema.OpenAIMessage{
			Role:    schema.OpenAIMessageRoleAssistant,
			Content: res,
		})

		// 检测thought标签并提取thought内容
		thoughtContent := helper.ExtractContentByTag(res, "thought")
		if thoughtContent != "" {
			if len(hooks.ThoughtHandlers) > 0 {
				hooks.EmitThought(thoughtContent)
			} else if noHooks {
				fmt.Println("☁Thought:", thoughtContent)
			}
		}
		// 检测final_answer标签并提取final_answer内容
		finalAnswerContent := helper.ExtractContentByTag(res, "final_answer")
		if finalAnswerContent != "" {
			if len(hooks.FinalAnswerHandlers) > 0 {
				hooks.EmitFinalAnswer(finalAnswerContent)
			} else if noHooks {
				fmt.Println("✅Final Answer:", finalAnswerContent)
			}
			var out strings.Builder
			for _, line := range messages {
				fmt.Fprintf(&out, "%s: %s\n", line.Role, line.Content)
			}
			break
		}

		// 检测action标签并提取action内容
		actionContent := helper.ExtractContentByTag(res, "action")
		obsText := ""
		if actionContent != "" {
			if len(hooks.ActionHandlers) > 0 {
				hooks.EmitAction(actionContent)
			} else if noHooks {
				fmt.Println("👍Action:", actionContent)
			}
			// 调用Mcp执行action
			actionRes, err := a.Mcp.Execute(actionContent)
			if len(hooks.ToolResultHandlers) > 0 {
				hooks.EmitToolResult(actionContent, actionRes, err)
			}
			if err != nil {
				if noHooks {
					fmt.Println("执行action失败", err)
				}
				obsText = fmt.Sprintf("<observation>执行action失败：%s</observation>", err.Error())
			} else {
				if noHooks {
					fmt.Println("执行action成功", actionRes)
				}
				obsText = fmt.Sprintf("<observation>%s</observation>", actionRes)
			}
		}

		messages = append(messages, schema.OpenAIMessage{
			Role:    schema.OpenAIMessageRoleUser,
			Content: obsText,
		})
	}
}
