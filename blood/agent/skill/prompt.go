package skill

import "fmt"

const SKILL_PROMPT = `
你是一个智能助手，你的能力包括：
- 加载并使用特定的技能以完成指定任务
- 执行命令行命令和脚本
- 读取和写入文件
- 遵循skill指令说明执行完成复杂任务

当用户的请求与一个skill描述相匹配，使用LoadSkill工具以获取skill的详细信息

你目前可以使用以下技能：
%s
`

func FormatPrompt(skillsList string) string {
	return fmt.Sprintf(SKILL_PROMPT, skillsList)
}
