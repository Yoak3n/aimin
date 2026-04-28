package mcp

import (
	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
)

func ShellCommandTool() *Tool {
	return &Tool{
		Name:   "ShellCommand",
		Desc:   `Execute a shell command. args example: {"os_type":"windows","command":"dir"} (use {"detach":true} for long-running GUI apps like chrome; use timeout_s/timeout to avoid blocking)`,
		Action: tool.ShellCommand,
	}
}

func FileOperationTool() *Tool {
	return &Tool{
		Name:   "FileOperation",
		Desc:   `Read/write/append a file. args example: {"op":"read","path":"./README.md"}`,
		Action: tool.FileOperation,
	}
}

func SkillTool() *Tool {
	return &Tool{
		Name:   "load_skill",
		Desc:   `Load/switch an agent skill (one-time when needed). args example: {"task":"agent-browser"} (use {"task":"..."} to clear)`,
		Action: tool.ComplexTaskForSkill,
	}
}

func GlobTool() *Tool {
	return &Tool{
		Name:   "Glob",
		Desc:   `Find files by glob pattern. args example: {"pattern":"**/*.go","root":"./"}`,
		Action: tool.Glob,
	}
}

func GrepTool() *Tool {
	return &Tool{
		Name:   "Grep",
		Desc:   `Search file contents by regex. args example: {"pattern":"TODO","root":"./","file_glob":"**/*.go","max_matches":100}`,
		Action: tool.Grep,
	}
}

func ManageMemoryTool() *Tool {
	return &Tool{
		Name: "manage_memory",
		Desc: `Memory & graph operations. Use this tool ONLY when you need database-backed memory/graph queries or to write memory files.
args: JSON object with required "action" and optional fields (query/limit/id/node_type/name/keyword/hops/rel_types/triples/link/refresh).
actions: search | vector_search | recent_conversations | get_conversation | graph_schema_summary | graph_subgraph | graph_add_triples | graph_seed_demo | graph_get_node | graph_neighbors | graph_search_nodes | graph_relations_by_link
examples:
 manage_memory({"action":"vector_search","query":"how to run tests","limit":5})
 manage_memory({"action":"recent_conversations","limit":10})
 manage_memory({"action":"graph_schema_summary","refresh":true})
 manage_memory({"action":"graph_subgraph","keyword":"Acme","hops":2})`,
		Action: tool.ManageMemory,
	}
}

func WebTool() *Tool {
	return &Tool{
		Name:   "Web",
		Desc:   `Web search & fetch. args examples: {"action":"search","query":"...","provider":"bing_serp"} or {"action":"fetch","url":"https://..."}`,
		Action: tool.Web,
	}
}
