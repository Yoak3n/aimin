package mcp

import (
	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
)

func ShellCommandTool() *Tool {
	return &Tool{
		Name:   "ShellCommand",
		Desc:   "Execute a shell command. args: os_type,command. eg: ShellCommand(windows,dir)",
		Action: tool.ShellCommand,
	}
}

func FileOperationTool() *Tool {
	return &Tool{
		Name:   "FileOperation",
		Desc:   "Read, write or append file, args: Read,path | Write,path,content | Append,path,content. eg: FileOperation(Read,/etc/hosts)",
		Action: tool.FileOperation,
	}
}

func SkillTool() *Tool {
	return &Tool{
		Name:   "Skill",
		Desc:   "Use skill, args: task. eg: Skill(list all skills)",
		Action: tool.ComplexTaskForSkill,
	}
}

func GlobTool() *Tool {
	return &Tool{
		Name:   "Glob",
		Desc:   "Find files by glob pattern. args: pattern[,root]. eg: Glob(**/*.go,./)",
		Action: tool.Glob,
	}
}

func GrepTool() *Tool {
	return &Tool{
		Name:   "Grep",
		Desc:   "Search file contents by regex. args: pattern[,root][,file_glob][,max_matches]. eg: Grep(TODO,./,**/*.go,100)",
		Action: tool.Grep,
	}
}
