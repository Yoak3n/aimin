package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

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
		logger.Logger.Infof("第一次运行，初始化工作空间")
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

func (a *ReActAgent) RegisterToolResultHandler(h func(toolCallID string, action string, result string, err error)) {
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
	chater, err := llm.NewPinnedStreamChatter()
	if err != nil {
		logger.Logger.Error("LLM适配器选择失败", err)
		return RunResult{}, err
	}
	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	cleaned := false
	defer func() {
		if cleaned {
			return
		}
		a.Mcp.CleanupRun(runID)
	}()

	wc := workspace.NewWorkspaceContextForPurpose(a.purpose)
	thoughts := make([]string, 0, 8)
	tools := append(mcp.ToOpenAITools(), schema.OpenAITool{
		Type: "function",
		Function: schema.OpenAIFunctionToolSpec{
			Name:        "final_answer",
			Description: "结束对话并返回最终答复。只在确认无需继续调用其他任何工具时调用。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"final_answer": map[string]any{
						"type":        "string",
						"description": "给用户的最终答复内容。",
					},
				},
				"required":             []string{"final_answer"},
				"additionalProperties": false,
			},
		},
	})
	consecutiveNoToolCalls := 0
	steps := 0
	consecutiveToolErrors := 0
	lastToolErrorKey := ""
	for {
		steps++
		if steps > 24 {
			return RunResult{}, fmt.Errorf("超过最大推理步数（%d），疑似陷入循环：请检查工具调用参数或提示词约束", steps-1)
		}
		sp := wc.String(a.choice)
		_ = os.WriteFile("sp.md", []byte(sp), 0644)
		var onDelta func(string) error
		if len(hooks.AssistantDeltaHandlers) > 0 {
			onDelta = hooks.EmitAssistantDelta
		}
		msg, err := llm.ChatStreamWith(chater, messages, tools, onDelta, sp)
		if err != nil {
			logger.Logger.Error("LLM调用失败", err)
			return RunResult{}, err
		}
		if len(hooks.LLMResponseHandlers) > 0 {
			msgSnapshot := append([]schema.OpenAIMessage(nil), messages...)
			raw, _ := json.Marshal(msg)
			hooks.EmitLLMResponse(sp, msgSnapshot, string(raw))
		}
		if msg.Role == "" {
			msg.Role = schema.OpenAIMessageRoleAssistant
		}
		messages = append(messages, msg)

		// 检测thought标签并提取thought内容
		thoughtContent := helper.ExtractContentByTag(msg.Content, "thought")
		if thoughtContent != "" {
			thoughts = append(thoughts, strings.TrimSpace(thoughtContent))
			if len(hooks.ThoughtHandlers) > 0 {
				hooks.EmitThought(thoughtContent)
			} else if noHooks {
				fmt.Println("☁Thought:", thoughtContent)
			}
		}

		if len(msg.ToolCalls) == 0 {
			consecutiveNoToolCalls++
			if consecutiveNoToolCalls >= 2 {
				return RunResult{}, fmt.Errorf("assistant 未返回 tool_calls（需要调用工具或 final_answer）: %s", strings.TrimSpace(msg.Content))
			}
			messages = append(messages, schema.OpenAIMessage{
				Role:    schema.OpenAIMessageRoleUser,
				Content: "请通过调用工具输出下一步：需要工具就调用对应工具；完成任务就调用 final_answer(final_answer=...)。不要直接输出最终答案文本。",
			})
			continue
		}
		consecutiveNoToolCalls = 0

		for _, tc := range msg.ToolCalls {
			toolName := strings.TrimSpace(tc.Function.Name)
			if toolName == "" {
				continue
			}

			if toolName == "final_answer" {
				finalText, parseErr := parseFinalAnswerArgs(tc.Function.Arguments)
				if parseErr != nil {
					messages = append(messages, schema.OpenAIMessage{
						Role:       schema.OpenAIMessageRoleTool,
						ToolCallID: tc.ID,
						Content:    "ERROR: final_answer 参数解析失败: " + parseErr.Error(),
					})
					messages = append(messages, schema.OpenAIMessage{
						Role:    schema.OpenAIMessageRoleUser,
						Content: `final_answer 工具调用失败：arguments 必须是合法 JSON 对象，例如 {"final_answer":"..."}。请重新调用 final_answer，并保证 JSON 可解析且字段名为 final_answer。`,
					})
					continue
				}
				if len(extractEmbeddedActions(finalText)) > 0 {
					messages = append(messages, schema.OpenAIMessage{
						Role:       schema.OpenAIMessageRoleTool,
						ToolCallID: tc.ID,
						Content:    "ERROR: final_answer 里检测到 <action> 标签。请不要把 action 写进 final_answer；需要工具请直接发起 tool call。",
					})
					messages = append(messages, schema.OpenAIMessage{
						Role:    schema.OpenAIMessageRoleUser,
						Content: `final_answer 只允许输出最终答复文本。需要继续用工具请直接发起对应 tool call；不要把 <action> 写进 final_answer。请重新调用 final_answer。`,
					})
					continue
				}

				finalText = strings.TrimSpace(finalText)
				if finalText == "" {
					finalText = "NO_REPLY"
				}
				if len(hooks.FinalAnswerHandlers) > 0 {
					msgSnapshot := append([]schema.OpenAIMessage(nil), messages...)
					hooks.EmitFinalAnswer(sp, msgSnapshot, finalText)
				} else if noHooks {
					fmt.Println("✅Final Answer:", finalText)
				}
				a.Mcp.CleanupRun(runID)
				cleaned = true
				return RunResult{
					Thought:     strings.Join(thoughts, "\n"),
					FinalAnswer: finalText,
				}, nil
			}

			payload, payloadErr := toolArgsToPayload(tc.Function.Arguments)
			actionBody := toolName + "(" + payload + ")"
			if payloadErr != nil {
				actionBody = toolName + "(" + tc.Function.Arguments + ")"
			}

			if len(hooks.ActionHandlers) > 0 {
				hooks.EmitAction(actionBody)
			} else if noHooks {
				fmt.Println("👍Action:", actionBody)
			}

			actionRes := ""
			var execErr error
			if payloadErr != nil {
				actionRes = "ERROR: tool args 解析失败: " + payloadErr.Error()
				execErr = payloadErr
			} else {
				actionRes, execErr = a.Mcp.ExecuteToolWithMeta(
					toolName,
					payload,
					runID,
					tc.ID,
					actionBody,
					func(p string) {
						if len(hooks.ToolResultHandlers) > 0 {
							hooks.EmitToolResult(tc.ID, actionBody, p, nil)
						}
					},
				)
			}

			if len(hooks.ToolResultHandlers) > 0 {
				hooks.EmitToolResult(tc.ID, actionBody, actionRes, execErr)
			}

			messages = append(messages, schema.OpenAIMessage{
				Role:       schema.OpenAIMessageRoleTool,
				ToolCallID: tc.ID,
				Content:    actionRes,
			})

			if execErr != nil {
				errKey := toolName + "|" + strings.TrimSpace(tc.Function.Arguments)
				if errKey == lastToolErrorKey {
					consecutiveToolErrors++
				} else {
					lastToolErrorKey = errKey
					consecutiveToolErrors = 1
				}
				if consecutiveToolErrors >= 5 {
					return RunResult{}, fmt.Errorf("工具调用连续失败（%s），疑似陷入循环：%s", toolName, execErr.Error())
				}
				messages = append(messages, schema.OpenAIMessage{
					Role:    schema.OpenAIMessageRoleUser,
					Content: "上一次工具调用失败。请重新发起该工具的 tool call，并确保 arguments 是严格的 JSON 对象且符合 tools 参数给出的 schema（不要传多余字段；字符串要加双引号）。",
				})
			} else {
				lastToolErrorKey = ""
				consecutiveToolErrors = 0
			}
		}
	}
}

func parseFinalAnswerArgs(arguments string) (string, error) {
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		return "", nil
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(arguments), &obj); err != nil {
		return "", err
	}
	if v, ok := obj["final_answer"]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	return "", nil
}

func toolArgsToPayload(arguments string) (string, error) {
	arguments = strings.TrimSpace(arguments)
	if arguments == "" {
		return "", nil
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(arguments), &obj); err != nil {
		return "", err
	}
	if v, ok := obj["payload"]; ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	if len(obj) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := obj[k]
		parts = append(parts, strings.ToLower(strings.TrimSpace(k))+"="+formatPayloadValue(v))
	}
	return strings.Join(parts, ","), nil
}

func formatPayloadValue(v any) string {
	switch x := v.(type) {
	case string:
		return quotePayloadString(x)
	case float64, bool, int, int64, uint64:
		return fmt.Sprintf("%v", x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return quotePayloadString(fmt.Sprintf("%v", x))
		}
		return quotePayloadString(string(b))
	}
}

func quotePayloadString(s string) string {
	needQuote := strings.ContainsAny(s, ",\n\r\t\"\\") || strings.Contains(s, "=")
	if !needQuote {
		return s
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
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
