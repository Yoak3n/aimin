package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/config"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
)

type WorkspaceContext struct {
	prompt              string
	previousTaskResults string
}

func NewWorkspaceContext() *WorkspaceContext {
	return &WorkspaceContext{
		prompt: ReActPromptTmpl,
	}
}

type PromptPurpose int

const (
	PromptPurposeReAct PromptPurpose = iota
	PromptPurposeExecPlanPrimary
	PromptPurposeExecPlanSecondary
)

func NewWorkspaceContextForPurpose(purpose PromptPurpose) *WorkspaceContext {
	prompt := ReActPromptTmpl
	switch purpose {
	case PromptPurposeExecPlanPrimary:
		prompt = ExecPlanPrimaryPromptTmpl
	case PromptPurposeExecPlanSecondary:
		prompt = ExecPlanSecondaryPromptTmpl
	}
	return &WorkspaceContext{
		prompt: prompt,
	}
}

func (wc *WorkspaceContext) WithPreviousTaskResults(results string) *WorkspaceContext {
	wc.previousTaskResults = results
	return wc
}

func (wc *WorkspaceContext) String(choose ...ContextChoice) string {
	if len(choose) == 0 {
		choose = append(choose, Normal)
	}
	wc.BuildPreviousTaskResults().BuildEnvInfo().BuildSkillInfo().BuildWorkspaceRoots().BuildWorkspaceContext(choose[0])
	return wc.prompt
}

func (wc *WorkspaceContext) BuildPreviousTaskResults() *WorkspaceContext {
	wc.prompt = strings.Replace(wc.prompt, "{previous_task_results}", wc.previousTaskResults, 1)
	return wc
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
	cwdFiles = fmt.Sprintf("<cwd_files>\n%s</cwd_files>", cwdFiles)
	out := strings.Replace(wc.prompt, "{local_time}", localTime, 1)
	out = strings.Replace(out, "{os_info}", osInfo, 1)
	out = strings.Replace(out, "{cwd}", cwd, 1)
	out = strings.Replace(out, "{cwd_files}", cwdFiles, 1)
	wc.prompt = out
	return wc
}

func (wc *WorkspaceContext) BuildSkillInfo() *WorkspaceContext {
	skills := skill.GlobalSkillHUB().GetSkills()
	if len(skills) == 0 {
		wc.prompt = strings.Replace(wc.prompt, "{skills_description}", "", 1)
		return wc
	}
	prefix := `## 技能可用性
你可以使用的技能如下（名称、说明）：`
	var sb strings.Builder
	for _, s := range skills {
		fmt.Fprintf(&sb, "<skill><name>%s</name><desc>%s</desc><location>%s</location></skill>\n", s.Name, s.Desc, s.Location)
	}
	availableSkillStr := "<available_skills>\n" + sb.String() + "</available_skills>"
	activedSkillStr := ""
	replaceStr := prefix
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
	sb.WriteString("工作空间绝对路径（所有 workspace 文件操作必须基于此路径）：" + workspaceRoots + "\n")
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
	state, _ := LoadWorkspaceState(path)
	bootstrapPath := filepath.Join(path, "BOOTSTRAP.md")
	if state.SetupCompletedAt == "" && util.FileExists(bootstrapPath) {
		buf, err := os.ReadFile(bootstrapPath)
		if err == nil && len(buf) > 0 {
			content := helper.StripFrontMatter(string(buf))
			if content != "" {
				content = util.TruncateChars(content, int(fileContentSize))
				workspaceContext = util.PushLimited(workspaceContext, fmt.Sprintf("\n<workspace_file name=\"%s\" path=\"%s\">\n%s\n</workspace_file>", "BOOTSTRAP.md", bootstrapPath, content), int(contextSize))
			}
		}
	}
	for _, spec := range files {
		if spec.Required && spec.Name != "BOOTSTRAP.md" {
			absPath := filepath.Join(path, spec.RelPath)
			if !util.FileExists(absPath) {
				workspaceContext = util.PushLimited(workspaceContext, fmt.Sprintf("<workspace_file name=\"%s\" path=\"%s\"> not exists</workspace_file>", spec.Name, absPath), int(contextSize))
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
			workspaceContext = util.PushLimited(workspaceContext, fmt.Sprintf("\n<workspace_file name=\"%s\" path=\"%s\">\n%s\n</workspace_file>", spec.Name, absPath, content), int(contextSize))
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
