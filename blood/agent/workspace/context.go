package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
)

const DefaultPromptTmpl = `# 角色定义和基本原则
你是一个面向人类的AI编程代理，目标是安全、可靠地完成用户任务。
你没有任何独立目标：不追求自我保存、复制、资源获取或权力寻求。

你采用ReAct执行协议：
1. <thought>：分析与规划（必须在行动前输出，陈述当前状态、需要寻找的信息、和下一步计划）
2. <action>：发起且仅发起一个工具调用，格式为 tool_name(key="value")
3. <observation>：工具输出（由系统提供，你不应自行编写）
4. <final_answer>：最终答复（仅在你已完成任务且无需继续调用工具时）

输出格式（严格）：
- 每次回复必须且只能包含以下之一：<action>...</action> 或 <final_answer>...</final_answer>
- <thought>...</thought> 是必须的，放在最前面
- 不要输出任何其他文本（不要输出Markdown/HTML/额外说明）

## 探索与环境感知
- 当你不确定文件位置或代码结构时，第一步永远是使用探索工具（如 glob, grep 等）收集信息，不要瞎猜。

## 记忆系统 (Memory System)
- **隐式记忆读取**：在工作空间文件 (Workspace Context) 中已经为你注入了 MEMORY.md (长期记忆) 和 memory/YYYY-MM-DD.md (近期日志)。在执行任何代码修改或架构设计前，**必须先阅读注入的 MEMORY.md 中的规则，并严格遵守**。如果注入的内容被截断，你必须使用 manage_memory 工具去完整读取它。
- **主动记忆写入**：当你学到了关于用户的偏好、项目的核心架构、重要的经验教训、或是制定了新的代码规范时，**必须**主动使用 manage_memory 工具将其写入长期记忆 (write_long_term)。对于每日的进展、任务状态或临时记录，主动写入每日记忆 (write_daily)。
- **对话历史留档与检索**：**注意：系统已在后台的 SQLite 数据库中自动保存了所有历史对话的完整记录（包括你的所有 thought 过程和最终答复）**。当你遇到类似“我之前问了什么问题”或需要回忆跨会话的对话历史时，**不要**去寻找普通的 log 文件，而是**直接使用 manage_memory(action="search", query="...") 工具**进行基于向量/关键字的检索。

## 错误处理与容错
- 你的首要目标是**完成用户的任务**。遇到命令报错或代码执行失败时，优先尝试不同的方法修复问题，不要轻易放弃，直到任务成功。
- 在任务结束后，你应该自主决定是否需要对过程中的曲折和教训进行总结。

{tools_description}

## 技能系统
- 当任务需要特定领域能力时，优先查看“技能列表”并选择最相关的技能
- 若恰好一个技能明显适用：使用 use_skill(skill_name) 激活它，然后严格遵循该技能的Instructions
- 若当前技能不再相关，后续步骤应回到通用执行（必要时重新选择技能,使用use_skill("...")来表示不再需要使用技能）
技能列表（动态注入）：
{skills_description}

## 配置与工作区
配置文件通常为 config.json（可用 file_operation 读取或修改）

### 工作区根目录列表（用于注入上下文文件与技能）：
{workspace_roots}

### 工作空间文件（已注入）
以下上下文文件可能已被注入（内容可能被截断）：
{workspace_context}

## 心跳检查
- 若在“工作空间文件（已注入）”中包含 HEARTBEAT.md，则严格遵循其中指令
- 若未包含且当前无须执行任何动作，则回复：<final_answer>HEARTBEAT_OK</final_answer>

## 静默回复
- 当你没有任何要说且无需执行工具时，只回复：<final_answer>NO_REPLY</final_answer>

## 运行时信息
- LocalTime: {local_time}
- OS: {os_info}
- 当前所在目录: {cwd}
- 所在目录文件: {cwd_files}

示例：
<thought>需要查看配置文件。</thought>
<action>FileOperation(Read,config.json)</action>`

type WorkspaceContext struct {
	prompt string
}

func NewWorkspaceContext() *WorkspaceContext {
	return &WorkspaceContext{
		prompt: DefaultPromptTmpl,
	}
}

func (wc *WorkspaceContext) String(choose ...ContextChoice) string {
	if len(choose) == 0 {
		choose = append(choose, Normal)
	}
	wc.BuildEnvInfo().BuildToolInfo().BuildSkillInfo().BuildWorkspaceRoots().BuildWorkspaceContext(choose[0])
	return wc.prompt
}

func (wc *WorkspaceContext) BuildEnvInfo() *WorkspaceContext {
	localTime := time.Now().Format("2006-01-02 15:04:05")
	osInfo := runtime.GOOS
	cwd, _ := os.Getwd()
	cwdFiles := ""
	dirFp, err := os.ReadDir(cwd)
	if err != nil {
		cwdFiles = fmt.Sprintf("Error reading current directory: %s\n", err)
	}
	for _, fp := range dirFp {
		if fp.IsDir() {
			cwdFiles += fmt.Sprintf("%s/\n", fp.Name())
		} else {
			cwdFiles += fmt.Sprintf("%s\n", fp.Name())
		}
	}
	cwdFiles = fmt.Sprintf("<cwd_files>%s</cwd_files>", cwdFiles)
	out := strings.Replace(wc.prompt, "{local_time}", localTime, 1)
	out = strings.Replace(out, "{os_info}", osInfo, 1)
	out = strings.Replace(out, "{cwd}", cwd, 1)
	out = strings.Replace(out, "{cwd_files}", cwdFiles, 1)
	wc.prompt = out
	return wc
}

func (wc *WorkspaceContext) BuildToolInfo() *WorkspaceContext {
	tools := mcp.GetMcpTools()
	if len(tools) == 0 {
		return wc
	}
	var toolsDesc strings.Builder
	prefix := `## 工具可用性
你可以使用的工具如下（名称、说明、参数schema）：\n`
	for _, tool := range tools {
		fmt.Fprintf(&toolsDesc, "%s\n", tool.String())
	}
	out := strings.Replace(wc.prompt, "{tools_description}", prefix+toolsDesc.String(), 1)
	wc.prompt = out
	return wc
}

func (wc *WorkspaceContext) BuildSkillInfo() *WorkspaceContext {
	skills := skill.GlobalSkillHUB().Skills
	if len(skills) == 0 {
		wc.prompt = strings.Replace(wc.prompt, "{skills_description}", "", 1)
		return wc
	}
	prefix := `## 技能可用性
你可以使用的技能如下（名称、说明）：`
	var sb strings.Builder
	for _, s := range skills {
		fmt.Fprintf(&sb, "<skill><name>%s</name>\n<desc>%s</desc>\n<location>%s</location></skill>\n", s.Name, s.Desc, s.Location)
	}
	availableSkillStr := "<available_skills>" + sb.String() + "</available_skills>"
	activedSkillStr := ""
	replaceStr := prefix + "\n"
	if skill.GlobalSkillHUB().Active != "" {
		activedSkillStr = skill.GlobalSkillHUB().LoadSkill(skill.GlobalSkillHUB().Active)
		replaceStr += "\n" + activedSkillStr + "\n" + availableSkillStr
	} else {
		replaceStr += "\n" + availableSkillStr
	}
	if strings.Contains(wc.prompt, "{skills_description}") {
		out := strings.Replace(wc.prompt, "{skills_description}", replaceStr, 1)
		wc.prompt = out
	} else {
		startStr := "技能列表（动态注入）："
		idx := strings.Index(wc.prompt, startStr)
		endIdx := strings.Index(wc.prompt, "## 配置与工作区")
		if idx == -1 || endIdx == -1 {
			return wc
		}
		startIdx := idx + len(startStr)
		if startIdx > endIdx {
			return wc
		}
		wc.prompt = wc.prompt[:startIdx] + "\n" + replaceStr + "\n" + wc.prompt[endIdx:]
	}

	return wc
}

func (wc *WorkspaceContext) BuildWorkspaceRoots() *WorkspaceContext {
	workspaceRoots := config.GlobalConfiguration().Workspace.Path

	var sb strings.Builder
	sb.WriteString("工作空间路径：" + workspaceRoots + "\n")
	dir, err := os.ReadDir(workspaceRoots)
	if err != nil {
		fmt.Fprintf(&sb, "Error reading workspace directory: %s\n", err)
	}
	sb.WriteString("工作空间文件列表：\n")
	for _, d := range dir {
		if d.IsDir() {
			fmt.Fprintf(&sb, "- %s/\n", d.Name())
		} else {
			fmt.Fprintf(&sb, "- %s\n", d.Name())
		}
	}

	if strings.Contains(wc.prompt, "{workspace_roots}") {
		replaceStr := sb.String()
		out := strings.Replace(wc.prompt, "{workspace_roots}", replaceStr, 1)
		wc.prompt = out
	} else {
		startStr := "### 工作区根目录列表（用于注入上下文文件与技能）："
		idx := strings.Index(wc.prompt, startStr)
		endIdx := strings.Index(wc.prompt, "### 工作空间文件内容（已注入）")
		if idx == -1 || endIdx == -1 {
			return wc
		}
		startIdx := idx + len(startStr)
		if startIdx > endIdx {
			return wc
		}
		wc.prompt = wc.prompt[:startIdx] + "\n" + sb.String() + "\n" + wc.prompt[endIdx:]
	}
	return wc
}

func (wc *WorkspaceContext) BuildWorkspaceContext(plan ...ContextChoice) *WorkspaceContext {
	path := config.GlobalConfiguration().Workspace.Path
	contextSize := config.GlobalConfiguration().Workspace.ContextSize
	fileContentSize := config.GlobalConfiguration().Workspace.FileContentSize
	workspaceContext := ""
	if len(plan) == 0 {
		plan = append(plan, Normal)
	}
	files := makeFileSpecMap(plan[0])
	for _, spec := range files {
		if spec.Required {
			absPath := filepath.Join(path, spec.RelPath)
			if !util.FileExists(absPath) {
				workspaceContext = util.PushLimited(workspaceContext, "\n"+spec.Name+" not exists", int(contextSize))
				continue
			}

			buf, err := os.ReadFile(absPath)
			if err != nil || len(buf) == 0 {
				// 如果不是错误导致的空文件，就不打印错误
				if err != nil {
					fmt.Println("read file failed:", absPath, err)
				}
				continue
			}

			content := helper.StripFrontMatter(string(buf))
			if content == "" {
				continue
			}
			content = util.TruncateChars(content, int(fileContentSize))

			ol := len(workspaceContext)
			workspaceContext = util.PushLimited(workspaceContext, fmt.Sprintf("\n<workspace_file name=\"%s\">\n%s</workspace_file>", spec.Name, content), int(contextSize))
			if len(workspaceContext) == ol {
				return wc
			}
		}
	}
	if strings.Contains(wc.prompt, "{workspace_context}") {
		replaceStr := workspaceContext
		out := strings.Replace(wc.prompt, "{workspace_context}", replaceStr, 1)
		wc.prompt = out
	} else {
		startStr := "### 工作空间文件内容（已注入）"
		idx := strings.Index(wc.prompt, startStr)
		endIdx := strings.Index(wc.prompt, "### 心跳检查")
		if idx == -1 || endIdx == -1 {
			return wc
		}
		startIdx := idx + len(startStr)
		if startIdx > endIdx {
			return wc
		}
		wc.prompt = wc.prompt[:startIdx] + "\n" + workspaceContext + "\n" + wc.prompt[endIdx:]
	}
	return wc
}
