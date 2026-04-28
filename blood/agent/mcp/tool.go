package mcp

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/agent/mcp/tool"
	"github.com/Yoak3n/aimin/blood/schema"
)

type Tool struct {
	Name   string                         `json:"name"`
	Desc   string                         `json:"desc"`
	Action func(ctx *tool.Context) string `json:"action"`
}

func (m *Tool) String() string {
	return fmt.Sprintf("<tool><name>%s</name><desc>%s</desc></tool>", m.Name, m.Desc)
}

func GetMcpTools() []*Tool {
	var tools []*Tool
	for _, tool := range GlobalMcpHUB().tools {
		tools = append(tools, tool)
	}
	return tools
}

func ToOpenAITools() []schema.OpenAITool {
	tools := GetMcpTools()
	out := make([]schema.OpenAITool, 0, len(tools))
	for _, t := range tools {
		if t == nil {
			continue
		}
		out = append(out, toOpenAITool(t))
	}
	return out
}

func toOpenAITool(t *Tool) schema.OpenAITool {
	switch t.Name {
	case "ShellCommand":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"os_type": map[string]any{
							"type":        "string",
							"description": "操作系统类型，例如 windows。",
						},
						"command": map[string]any{
							"type":        "string",
							"description": "要执行的命令。",
						},
						"timeout_s": map[string]any{
							"type":        "number",
							"description": "超时秒数（可选）。也可传如 1m/30s 这样的 duration 字符串到 timeout 字段。",
						},
						"timeout": map[string]any{
							"type":        "string",
							"description": "超时（可选），如 30s/1m。",
						},
						"detach": map[string]any{
							"type":        "boolean",
							"description": "是否后台启动（Windows 打开浏览器/GUI 程序建议 true，避免阻塞后续对话）。",
						},
					},
					"required":             []string{"os_type", "command"},
					"additionalProperties": false,
				},
			},
		}
	case "FileOperation":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"op": map[string]any{
							"type":        "string",
							"description": "操作类型：read/write/append（不区分大小写）。",
						},
						"path": map[string]any{
							"type":        "string",
							"description": "文件路径。",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "写入或追加内容（op=write/append 时使用）。",
						},
					},
					"required":             []string{"op", "path"},
					"additionalProperties": false,
				},
			},
		}
	case "Glob":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pattern": map[string]any{
							"type":        "string",
							"description": "glob 模式，例如 **/*.go。",
						},
						"root": map[string]any{
							"type":        "string",
							"description": "根目录（可选，默认当前目录）。",
						},
					},
					"required":             []string{"pattern"},
					"additionalProperties": false,
				},
			},
		}
	case "Grep":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"pattern": map[string]any{
							"type":        "string",
							"description": "正则表达式模式。",
						},
						"root": map[string]any{
							"type":        "string",
							"description": "根目录（可选，默认当前目录）。",
						},
						"file_glob": map[string]any{
							"type":        "string",
							"description": "文件过滤 glob（可选），例如 **/*.go。",
						},
						"max_matches": map[string]any{
							"type":        "integer",
							"description": "最多匹配条数（可选）。",
						},
					},
					"required":             []string{"pattern"},
					"additionalProperties": false,
				},
			},
		}
	case "Skill", "load_skill":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"task": map[string]any{
							"type":        "string",
							"description": "技能名称或技能相关任务描述；传入 ... 清除当前技能。",
						},
					},
					"required":             []string{"task"},
					"additionalProperties": false,
				},
			},
		}
	case "Web":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"action": map[string]any{
							"type":        "string",
							"description": "fetch/search（可选；若不填会根据 url/query 自动判断）。",
						},
						"provider": map[string]any{
							"type":        "string",
							"description": "search provider（action=search 时可选）：bing_serp（默认）/bilibili/duckduckgo。",
							"enum":        []string{"bing_serp", "bilibili", "duckduckgo"},
						},
						"url": map[string]any{
							"type":        "string",
							"description": "抓取的 URL（action=fetch 时）。",
						},
						"query": map[string]any{
							"type":        "string",
							"description": "搜索关键词（action=search 时）。",
						},
						"timeout_s": map[string]any{
							"type":        "integer",
							"description": "超时时间秒数（可选）。",
						},
						"js": map[string]any{
							"type":        "boolean",
							"description": "是否启用 JS 回退（可选）。",
						},
						"pdf_max_pages": map[string]any{
							"type":        "integer",
							"description": "PDF 最大页数（可选）。",
						},
						"out_dir": map[string]any{
							"type":        "string",
							"description": "输出目录（可选）。",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "搜索条数（action=search 时可选）。",
						},
					},
					"required":             []string{},
					"additionalProperties": false,
				},
			},
		}
	case "manage_memory":
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"action": map[string]any{
							"type":        "string",
							"description": "操作名，例如 vector_search/recent_conversations/graph_subgraph 等。",
						},
						"query": map[string]any{
							"type":        "string",
							"description": "检索查询（可选）。",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "数量限制（可选）。",
						},
						"id": map[string]any{
							"type":        "string",
							"description": "会话 ID / 节点 ID（可选）。",
						},
						"node_type": map[string]any{
							"type":        "string",
							"description": "图谱节点类型（可选）。",
						},
						"name": map[string]any{
							"type":        "string",
							"description": "图谱节点名称（可选）。",
						},
						"keyword": map[string]any{
							"type":        "string",
							"description": "图谱检索关键词（可选）。",
						},
						"hops": map[string]any{
							"type":        "integer",
							"description": "子图跳数（可选）。",
						},
						"rel_types": map[string]any{
							"type":        "string",
							"description": "关系类型列表（可选，逗号分隔）。",
						},
						"triples": map[string]any{
							"type":        "string",
							"description": "三元组内容（可选）。",
						},
						"link": map[string]any{
							"type":        "string",
							"description": "link/source（可选）。",
						},
						"refresh": map[string]any{
							"type":        "boolean",
							"description": "是否强制刷新（可选）。",
						},
					},
					"required":             []string{"action"},
					"additionalProperties": false,
				},
			},
		}
	default:
		return schema.OpenAITool{
			Type: "function",
			Function: schema.OpenAIFunctionToolSpec{
				Name:        t.Name,
				Description: t.Desc,
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"payload": map[string]any{
							"type":        "string",
							"description": "工具参数（字符串）。可用形式：逗号分隔位置参数，或 key=value 对（必要时用双引号包裹）。",
						},
					},
					"required":             []string{"payload"},
					"additionalProperties": false,
				},
			},
		}
	}
}
