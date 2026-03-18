package action

import "fmt"

const ReActPrompt = "你需要解决一个任务。" +
	"为此，你需要将任务分解为多个步骤。对于每个步骤，首先使用<thought>标签来思考这个步骤需要做什么。" +
	"然后使用<action>标签调用一个工具，工具的执行结果会通过<observation>标签返回。" +
	"根据工具的执行结果，你可以决定是否继续执行下一个步骤，直到你有足够的信息来提供<final_answer>。" +
	"所有步骤请严格使用以下XML标签格式输出：" +
	"<task>用户提出的任务</task>" +
	"<thought>你思考的内容</thought>" +
	"<action>工具调用的名称</action>" +
	"<observation>工具调用或环境返回的结果</observation>" +
	"<final_answer>任务的最终答案</final_answer>"

func genCompletedPrompt(task string) string {
	return fmt.Sprintf(ReActPrompt, task)
}

func getMcpTool() string {
	return "McpTool"
}
