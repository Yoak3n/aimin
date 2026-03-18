package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Yoak3n/aimin/blood/agent/skill"
	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/schema"
)

func ShellCommand(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	ps := strings.SplitN(p, ",", 2)
	if len(ps) != 2 {
		return fmt.Sprintf("invalid args format for ShellCommand: %s", p)
	}

	osType := strings.ToLower(strings.TrimSpace(ps[0]))
	commandStr := strings.TrimSpace(ps[1])

	var cmd *exec.Cmd

	switch osType {
	case "windows":
		cmd = exec.Command("cmd", "/C", commandStr)
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", commandStr)
	default:
		return fmt.Sprintf("unsupported os type: %s", osType)
	}

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// 如果不是 ExitError，说明是其他系统错误（如找不到命令）
			return fmt.Sprintf("command execution failed: %s\nOutput: %s", err, string(output))
		}
		return fmt.Sprintf("command execution finished with error\nExit Code: %d\nOutput: %s", exitCode, string(output))
	}

	return fmt.Sprintf("command execution success\nExit Code: %d\nOutput: %s", exitCode, string(output))
}

func ShellCommandTool() *Tool {
	return &Tool{
		Name:   "ShellCommand",
		Desc:   "Execute a shell command. args: os_type,command. eg: ShellCommand(windows,dir)",
		Action: ShellCommand,
	}
}

func FileOperationTool() *Tool {
	return &Tool{
		Name:   "FileOperation",
		Desc:   "Read or write file, args: Read,path | Write,path,content. eg: FileOperation(Read,/etc/hosts)",
		Action: FileOperation,
	}
}

func FileOperation(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}
	ps := strings.Split(p, ",")
	if len(ps) >= 2 {
		op := strings.ToLower(strings.TrimSpace(ps[0]))
		switch op {
		case "read":
			return ReadFile(ps[1])
		case "write":
			if len(ps) != 3 {
				return fmt.Sprintf("args %s is invalid", p)
			}
			return WriteFile(ps[1], ps[2])
		default:
			return fmt.Sprintf("file operation %s not found", ps[0])
		}
	} else {
		return fmt.Sprintf("args %s is invalid", p)
	}
}

func ReadFile(path string) string {
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("read file failed with error: %s", err.Error())
	}
	return string(buf)
}

func WriteFile(path, content string) string {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Sprintf("write file failed with error: %s", err.Error())
	}
	return "write file success"
}

func SkillTool() *Tool {
	return &Tool{
		Name:   "Skill",
		Desc:   "Use skill, args: task. eg: Skill(list all skills)",
		Action: ComplexTaskForSkill,
	}
}

func ComplexTaskForSkill(ctx *Context) string {
	p := ctx.GetPayload()
	if p == "" {
		return "args is empty"
	}

	return UseSkill(p)
}

func UseSkill(args string) string {
	hub := skill.NewSkillHUB()
	hub.ScanSkills("./skills")
	sl := hub.RenderSkillsList()
	res, err := helper.UseLLM().Chat([]schema.OpenAIMessage{
		{
			Role:    schema.OpenAIMessageRoleUser,
			Content: fmt.Sprintf("请根据技能列表，为了完成该任务，选择需要使用的技能 %s，仅返回技能名称", args),
		},
	}, skill.FormatPrompt(sl))
	if err != nil {
		return fmt.Sprintf("use skill failed with error: %s", err.Error())
	}
	skillContents := hub.LoadSkill(res)
	if skillContents == "" {
		return fmt.Sprintf("skill %s not found", res)
	}
	return fmt.Sprintf("已加载技能【%s】及其说明\n%s", res, skillContents)
}
