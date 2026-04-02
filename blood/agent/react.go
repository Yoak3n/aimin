package agent

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/agent/workspace"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/service/llm"
)

type RunResult struct {
	Thought     string
	FinalAnswer string
}

type ReActAgent struct {
	Mcp     *mcp.McpHUB
	Skill   *skill.SkillHUB
	Hooks   *AgentHooks
	purpose workspace.PromptPurpose
	choice  workspace.ContextChoice
}

func NewAgent(purpose workspace.PromptPurpose) *ReActAgent {
	a := &ReActAgent{
		Mcp:     mcp.GlobalMcpHUB(),
		Skill:   skill.NewSkillHUB(),
		Hooks:   NewAgentHooks(),
		purpose: purpose,
		choice:  workspace.Normal,
	}
	a.RegisterTool(mcp.FileOperationTool())
	a.RegisterTool(mcp.ShellCommandTool())
	a.RegisterTool(mcp.GlobTool())
	a.RegisterTool(mcp.GrepTool())
	a.RegisterTool(mcp.SkillTool())
	a.RegisterTool(mcp.ManageMemoryTool())
	a.RegisterTool(mcp.WebTool())
	if workspace.EnsureWorkspace() {
		fmt.Println("第一次运行，初始化工作空间")
	}
	return a
}

func (a *ReActAgent) SetContextChoice(choice workspace.ContextChoice) {
	a.choice = choice
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

func (a *ReActAgent) RegisterFinalAnswerHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, finalAnswer string)) {
	a.ensureHooks().AddFinalAnswerHandler(h)
}

func (a *ReActAgent) RegisterAssistantDeltaHandler(h func(string) error) {
	a.ensureHooks().AddAssistantDeltaHandler(h)
}

func (a *ReActAgent) RegisterLLMResponseHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, response string)) {
	a.ensureHooks().AddLLMResponseHandler(h)
}

func (a *ReActAgent) RunWithMessages(messages []schema.OpenAIMessage) (RunResult, error) {
	hooks := a.ensureHooks()
	noHooks := hooks.IsEmpty()
	if len(messages) == 0 {
		return RunResult{}, errors.New("messages 不能为空")
	}
	wc := workspace.NewWorkspaceContextForPurpose(a.purpose)
	thoughts := make([]string, 0, 8)
	for {
		sp := wc.String(a.choice)
		_ = os.WriteFile("sp.md", []byte(sp), 0644)
		var onDelta func(string) error
		if len(hooks.AssistantDeltaHandlers) > 0 {
			onDelta = hooks.EmitAssistantDelta
		}
		res, err := llm.ChatStream(messages, onDelta, sp)
		if err != nil {
			logger.Logger.Error("LLM调用失败", err)
			return RunResult{}, err
		}
		if len(hooks.LLMResponseHandlers) > 0 {
			msgSnapshot := append([]schema.OpenAIMessage(nil), messages...)
			hooks.EmitLLMResponse(sp, msgSnapshot, res)
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
			// 兜底处理：检查 final_answer 中是否嵌套了 <action> 标签
			// 如果有，提取并单独执行，而不是静默丢弃导致死锁
			if embeddedActions := extractEmbeddedActions(finalAnswerContent); len(embeddedActions) > 0 {
				// 清理 final_answer 中的 action 标签，提取出思维内容
				cleanedAnswer := stripActionsFromContent(finalAnswerContent)

				// 把清理后的内容作为 thought 累积（LLM 在 final_answer 里夹带 action，
				// 本质是"边想边做"，清理出来的内容就是它的思考过程）
				if strings.TrimSpace(cleanedAnswer) != "" {
					thoughts = append(thoughts, strings.TrimSpace(cleanedAnswer))
					if len(hooks.ThoughtHandlers) > 0 {
						hooks.EmitThought(cleanedAnswer)
					} else if noHooks {
						fmt.Println("☁Thought (从 Final Answer 中提取):", cleanedAnswer)
					}
				}

				// 替换 messages 中最后一条 assistant 消息（含脏 final_answer）
				// 用干净的版本，避免 LLM 下一轮看到自己带 <action> 的病态回复
				cleanedRes := stripFinalAnswerWithActions(res)
				if len(messages) > 0 && messages[len(messages)-1].Role == schema.OpenAIMessageRoleAssistant {
					messages[len(messages)-1].Content = cleanedRes
				}

				// 逐一执行嵌套的 action
				for _, actionBody := range embeddedActions {
					if len(hooks.ActionHandlers) > 0 {
						hooks.EmitAction(actionBody)
					} else if noHooks {
						fmt.Println("👍Action (从 Final Answer 中提取):", actionBody)
					}

					actionRes, err := a.Mcp.Execute(actionBody)
					if len(hooks.ToolResultHandlers) > 0 {
						hooks.EmitToolResult(actionBody, actionRes, err)
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
				}

				// 继续 ReAct 循环，让 LLM 基于 observation 生成新的回复
				continue
			}

			// 没有嵌套 action，正常返回
			if len(hooks.FinalAnswerHandlers) > 0 {
				msgSnapshot := append([]schema.OpenAIMessage(nil), messages...)
				hooks.EmitFinalAnswer(sp, msgSnapshot, finalAnswerContent)
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

// extractEmbeddedActions 从文本中提取所有嵌套的 <action>...</action> 标签内容
func extractEmbeddedActions(content string) []string {
	re := regexp.MustCompile(`(?s)<action>(.*?)</action>`)
	matches := re.FindAllStringSubmatch(content, -1)
	var actions []string
	for _, m := range matches {
		if len(m) > 1 {
			actions = append(actions, strings.TrimSpace(m[1]))
		}
	}
	return actions
}

// stripActionsFromContent 从文本中移除所有 <action>...</action> 标签及其内容
func stripActionsFromContent(content string) string {
	re := regexp.MustCompile(`(?s)<action>.*?</action>`)
	cleaned := re.ReplaceAllString(content, "")
	// 清理可能残留的多余空行
	cleaned = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleaned, "\n\n")
	return strings.TrimSpace(cleaned)
}

// stripFinalAnswerWithActions 从完整 LLM 回复中移除包含 <action> 的 <final_answer> 整块
// 保留 thought 等其他标签，用于替换 messages 中的脏 assistant 消息
func stripFinalAnswerWithActions(res string) string {
	// 匹配 <final_answer>...（其中包含 <action>...</action>）...</final_answer>
	re := regexp.MustCompile(`(?s)<final_answer>(.*?<action>.*?</action>.*?)</final_answer>`)
	cleaned := re.ReplaceAllString(res, "")
	cleaned = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleaned, "\n\n")
	return strings.TrimSpace(cleaned)
}
