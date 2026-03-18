package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Yoak3n/aimin/blood/agent/mcp"
	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

var (
	agent *Agent
	once  sync.Once
)

type Agent struct {
	Mcp   *mcp.McpHUB
	Skill *skill.SkillHUB
}

func NewAgent() *Agent {
	return &Agent{
		Mcp:   mcp.GlobalMcpHUB(),
		Skill: skill.NewSkillHUB(),
	}
}

func GlobalAgent() *Agent {
	once.Do(func() {
		agent = NewAgent()
		// 加载所有技能
		cwd, _ := os.Getwd()
		skillDir := filepath.Join(cwd, "skills")
		agent.RegisterTool(mcp.FileOperationTool())
		agent.RegisterTool(mcp.ShellCommandTool())
		agent.Skill.ScanSkills(skillDir)
		agent.RegisterTool(mcp.SkillTool())
	})
	return agent
}

func (a *Agent) RegisterTool(tool *mcp.Tool) {
	a.Mcp.RegisterTool(tool)
}

func (a *Agent) RenderSysytemPrompt(skill ...string) string {
	toolsList := a.Mcp.RenderToolsList()
	osInfo, err := helper.GetOSInfo()
	if err != nil {
		osInfo = fmt.Sprintf("操作系统信息获取失败：%s", err.Error())
	}
	fileList := ""
	l, err := util.GetFilesInDir("./")
	if err != nil {
		fileList = fmt.Sprintf("文件列表获取失败：%s", err.Error())
	} else {
		fileList = strings.Join(l, "\n")
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = fmt.Sprintf("当前目录获取失败：%s", err.Error())
	}
	if len(skill) > 0 {
		return FormatPromptWithSkills(toolsList, strings.Join(skill, "\n"), osInfo, cwd, fileList)
	}
	return FormatPrompt(toolsList, osInfo, cwd, fileList)
}

func (a *Agent) Run(input string) {
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
	sp := a.RenderSysytemPrompt()
	for {
		res, err := helper.UseLLM().Chat(messages, sp)
		if err != nil {
			logger.Logger.Error("LLM调用失败", err)
			break
		}
		messages = append(messages, schema.OpenAIMessage{
			Role:    schema.OpenAIMessageRoleAssistant,
			Content: res,
		})

		// 检测thought标签并提取thought内容
		thoughtContent := helper.ExtractContentByTag(res, "thought")
		if thoughtContent != "" {
			// 这些是思考，不需要添加到messages中，后面需要传递给前端进行渲染
			fmt.Println("☁Thought:", thoughtContent)
		}
		// 检测final_answer标签并提取final_answer内容
		finalAnswerContent := helper.ExtractContentByTag(res, "final_answer")
		if finalAnswerContent != "" {
			// TODO 后面需要传递给前端进行渲染
			fmt.Println("✅Final Answer:", finalAnswerContent)
			var out strings.Builder
			for _, line := range messages {
				fmt.Fprintf(&out, "%s: %s\n", line.Role, line.Content)
			}
			mcp.WriteFile("chat.log", out.String())
			break
		}

		// 检测action标签并提取action内容
		actionContent := helper.ExtractContentByTag(res, "action")
		obsText := ""
		if actionContent != "" {
			fmt.Println("👍Action:", actionContent)
			// 调用Mcp执行action
			actionRes, err := a.Mcp.Execute(actionContent)
			if err != nil {
				fmt.Println("执行action失败", err)
				obsText = fmt.Sprintf("<observation>执行action失败：%s</observation>", err.Error())
			} else {
				fmt.Println("执行action成功", actionRes)
				re := regexp.MustCompile(`已加载技能【(.*?)】及其说明`)
				if match := re.FindStringSubmatch(actionRes); len(match) > 1 {
					sp = a.RenderSysytemPrompt(match[1])
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
