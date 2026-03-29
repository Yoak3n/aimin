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

func ManageMemoryTool() *Tool {
	return &Tool{
		Name:   "manage_memory",
		Desc:   `Manage memory. args: action[,key=value...]. actions: read_long_term, write_long_term, read_daily, write_daily, search, vector_search, get_conversation, recent_conversations, graph_get_node, graph_neighbors, graph_related, graph_search_nodes, graph_relations_by_link. eg: manage_memory(action="vector_search",query="...")`,
		Action: tool.ManageMemory,
	}
}
