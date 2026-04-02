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
		Name: "manage_memory",
		Desc: `Memory & graph operations. Use this tool ONLY when you need database-backed memory/graph queries or to write memory files.
args: action[,key=value...]
actions:
 search(query=...,limit=5) | vector_search(query=...,limit=5) // search past conversations in DB
 recent_conversations(limit=10) | get_conversation(id=...)     // DB conversation history
 graph_schema_summary(refresh=1)                               // (re)build cached overview -> memory/GRAPH_SCHEMA.md
 graph_subgraph(node_type=...,name=...,keyword=...,hops=1|2,limit1=30,limit2=10,max_triples=60,rel_types="A,B")
 graph_add_triples(triples="Type|Name|REL|Type|Name;...",link="optional")          // add data
 graph_seed_demo(link="seed_demo")                                                 // demo data
 graph_get_node(node_type=...,name=...)
 graph_neighbors(node_type=...,name=...,rel_type=...,limit=20)
 graph_search_nodes(node_type=...,keyword=...,limit=20)
 graph_relations_by_link(link=...,limit=50)
examples:
 manage_memory(action="vector_search",query="how to run tests",limit=5)
 manage_memory(action="recent_conversations",limit=10)
 manage_memory(action="graph_schema_summary",refresh=1)
 manage_memory(action="graph_subgraph",keyword="Acme",hops=2,max_triples=50)`,
		Action: tool.ManageMemory,
	}
}

func WebTool() *Tool {
	return &Tool{
		Name:   "Web",
		Desc:   `Web search & fetch. args: action=search|fetch. search: Web(search,query="...",limit=5). fetch: Web(fetch,url="...",js=false,timeout_s=20,pdf_max_pages=0). action can be omitted: Web("https://...") or Web("query")`,
		Action: tool.Web,
	}
}
