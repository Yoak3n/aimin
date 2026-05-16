package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/bone/mcp"
	"github.com/Yoak3n/aimin/bone/skill"
	"github.com/Yoak3n/aimin/bone/workspace"
	"github.com/Yoak3n/aimin/lung/llm"
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
	a.RegisterTool(mcp.TTSTool())
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

func (a *ReActAgent) RegisterAssistantDeltaHandler(h func(string, string) error) {
	a.ensureHooks().AddAssistantDeltaHandler(h)
}

func (a *ReActAgent) RegisterLLMResponseHandler(h func(systemPrompt string, messages []schema.OpenAIMessage, response string)) {
	a.ensureHooks().AddLLMResponseHandler(h)
}

func (a *ReActAgent) RegisterAudioHandler(h func(format, voice, audioBase64 string, bytes int)) {
	a.ensureHooks().AddAudioHandler(h)
}

func (a *ReActAgent) RunWithMessages(ctx context.Context, messages []schema.OpenAIMessage) (RunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
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
	userInput := extractUserInput(messages)
	if userInput != "" {
		wc.WithUserInput(userInput)
	}
	thoughts := make([]string, 0, 8)
	tools := mcp.ToOpenAITools()
	consecutiveEmptyAssistant := 0
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
		dumpLLMInput(runID, steps, sp, tools, messages)
		var onDelta func(string, string) error
		if len(hooks.AssistantDeltaHandlers) > 0 {
			onDelta = hooks.EmitAssistantDelta
		}
		msg, err := llm.ChatStreamWith(ctx, chater, messages, tools, onDelta, sp)
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
			finishReason := strings.ToLower(strings.TrimSpace(msg.FinishReason))
			if finishReason == "tool_calls" {
				return RunResult{}, fmt.Errorf("assistant finish_reason=tool_calls 但未返回 tool_calls：%s", strings.TrimSpace(msg.Content))
			}

			if finishReason == "stop" {
				finalAnswer := strings.TrimSpace(msg.Content)
				if len(hooks.FinalAnswerHandlers) > 0 {
					msgSnapshot := append([]schema.OpenAIMessage(nil), messages...)
					hooks.EmitFinalAnswer(sp, msgSnapshot, finalAnswer)
				} else if noHooks {
					logger.Logger.Println("✅Final Answer:", finalAnswer)
				}
				a.Mcp.CleanupRun(runID)
				cleaned = true
				return RunResult{
					Thought:     strings.Join(thoughts, "\n"),
					FinalAnswer: finalAnswer,
				}, nil
			}

			consecutiveEmptyAssistant++
			if consecutiveEmptyAssistant >= 2 {
				return RunResult{}, fmt.Errorf("assistant 返回空内容且无 tool_calls（finish_reason=%q）", strings.TrimSpace(msg.FinishReason))
			}
			messages = append(messages, schema.OpenAIMessage{
				Role:    schema.OpenAIMessageRoleUser,
				Content: "你的上一条回复为空。请继续输出完整答复；如果需要调用工具，请直接发起 tool call。",
			})
			continue
		}
		consecutiveEmptyAssistant = 0

		for _, tc := range msg.ToolCalls {
			toolName := strings.TrimSpace(tc.Function.Name)
			if toolName == "" {
				continue
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
					func(format, voice, audioBase64 string, bytes int) {
						hooks.EmitAudio(format, voice, audioBase64, bytes)
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
	_, _ = a.RunWithMessages(context.Background(), []schema.OpenAIMessage{
		{
			Role:    schema.OpenAIMessageRoleUser,
			Content: fmt.Sprintf("<question>%s</question>", input),
		},
	})
}

type llmInputDump struct {
	RunID        string                 `json:"run_id"`
	Step         int                    `json:"step"`
	AtUnixMS     int64                  `json:"at_unix_ms"`
	SystemPrompt string                 `json:"system_prompt"`
	Tools        []schema.OpenAITool    `json:"tools,omitempty"`
	Messages     []schema.OpenAIMessage `json:"messages"`
}

func dumpLLMInput(runID string, step int, systemPrompt string, tools []schema.OpenAITool, messages []schema.OpenAIMessage) {
	cfg := config.GlobalConfiguration()
	if cfg == nil || cfg.Workspace == nil {
		return
	}
	base := strings.TrimSpace(cfg.Workspace.Path)
	if base == "" {
		return
	}
	dir := filepath.Join(base, "debug_llm_inputs")
	_ = os.MkdirAll(dir, 0755)
	name := fmt.Sprintf("react_%s_step_%02d.json", runID, step)
	path := filepath.Join(dir, name)

	msgs := append([]schema.OpenAIMessage(nil), messages...)
	ts := append([]schema.OpenAITool(nil), tools...)
	d := llmInputDump{
		RunID:        runID,
		Step:         step,
		AtUnixMS:     time.Now().UnixMilli(),
		SystemPrompt: systemPrompt,
		Tools:        ts,
		Messages:     msgs,
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0644)
}

func extractUserInput(messages []schema.OpenAIMessage) string {
	for _, m := range messages {
		if m.Role == schema.OpenAIMessageRoleUser {
			content := strings.TrimSpace(m.Content)
			return strings.TrimSpace(content)
		}
	}
	return ""
}
