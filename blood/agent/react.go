package agent

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/agent/workspace"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
)

type RunResult struct {
	Thought     string
	FinalAnswer string
}

type ReActAgent struct {
	Mcp   *mcp.McpHUB
	Skill *skill.SkillHUB
	Hooks *AgentHooks
}

func NewAgent() *ReActAgent {
	a := &ReActAgent{
		Mcp:   mcp.GlobalMcpHUB(),
		Skill: skill.NewSkillHUB(),
		Hooks: NewAgentHooks(),
	}
	a.RegisterTool(mcp.FileOperationTool())
	a.RegisterTool(mcp.ShellCommandTool())
	a.RegisterTool(mcp.GlobTool())
	a.RegisterTool(mcp.GrepTool())
	a.RegisterTool(mcp.SkillTool())
	if workspace.EnsureWorkspace() {
		fmt.Println("第一次运行，初始化工作空间")

	}
	return a
}

func (a *ReActAgent) RegisterTool(tool *mcp.Tool) {
	a.Mcp.RegisterTool(tool)
}

func (a *ReActAgent) ensureHooks() *AgentHooks {
	if a.Hooks == nil {
		a.Hooks = NewAgentHooks()
	}
	return a.Hooks
}

func (a *ReActAgent) SetHooks(hooks *AgentHooks) {
	if hooks == nil {
		a.Hooks = NewAgentHooks()
		return
	}
	a.Hooks = hooks
}

func (a *ReActAgent) RegisterThoughtHandler(h func(string)) {
	a.ensureHooks().AddThoughtHandler(h)
}

func (a *ReActAgent) RegisterActionHandler(h func(string)) {
	a.ensureHooks().AddActionHandler(h)
}

func (a *ReActAgent) RegisterToolResultHandler(h func(action string, result string, err error)) {
	a.ensureHooks().AddToolResultHandler(h)
}

func (a *ReActAgent) RegisterFinalAnswerHandler(h func(string)) {
	a.ensureHooks().AddFinalAnswerHandler(h)
}

func (a *ReActAgent) RegisterAssistantDeltaHandler(h func(string) error) {
	a.ensureHooks().AddAssistantDeltaHandler(h)
}

func (a *ReActAgent) RegisterLLMResponseHandler(h func(string)) {
	a.ensureHooks().AddLLMResponseHandler(h)
}

func (a *ReActAgent) RunWithMessages(messages []schema.OpenAIMessage) (RunResult, error) {
	hooks := a.ensureHooks()
	noHooks := hooks.IsEmpty()
	if len(messages) == 0 {
		return RunResult{}, errors.New("messages 不能为空")
	}
	wc := workspace.NewWorkspaceContext()
	thoughts := make([]string, 0, 8)
	for {
		sp := wc.String()
		_ = os.WriteFile("sp.md", []byte(sp), 0644)
		var onDelta func(string) error
		if len(hooks.AssistantDeltaHandlers) > 0 {
			onDelta = hooks.EmitAssistantDelta
		}
		res, err := helper.UseLLM().ChatStream(messages, onDelta, sp)
		if err != nil {
			logger.Logger.Error("LLM调用失败", err)
			return RunResult{}, err
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
			thoughts = append(thoughts, strings.TrimSpace(thoughtContent))
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
			return RunResult{
				Thought:     strings.Join(thoughts, "\n"),
				FinalAnswer: finalAnswerContent,
			}, nil
		}

		// 检测action标签并提取action内容
		actionContent := helper.ExtractContentByTag(res, "action")
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
			obsText := ""
			if err != nil {
				if noHooks {
					fmt.Println("执行action失败", err)
				}
				obsText = fmt.Sprintf("<observation>执行action失败：%s\n%s</observation>", err.Error(), actionRes)
			} else {
				if noHooks {
					fmt.Println("执行action成功", actionRes)
				}
				obsText = fmt.Sprintf("<observation>%s</observation>", actionRes)
			}

			messages = append(messages, schema.OpenAIMessage{
				Role:    schema.OpenAIMessageRoleUser,
				Content: obsText,
			})
			continue
		}
		return RunResult{}, fmt.Errorf("assistant 输出缺少 action 或 final_answer: %s", res)
	}
}

func (a *ReActAgent) Run(input string) {
	if input == "" {
		fmt.Println("输入为空")
		return
	}
	_, _ = a.RunWithMessages([]schema.OpenAIMessage{
		{
			Role:    schema.OpenAIMessageRoleUser,
			Content: fmt.Sprintf("<question>%s</question>", input),
		},
	})
}
