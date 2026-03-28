package controller

import (
	"fmt"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
)

const summarySystemPrompt = `你是一个“时间线摘要器”。你会收到一次对话的四部分内容：
- System prompt（系统提示词）
- Question（用户问题）
- Thoughts（助手思考过程，可能很长、含试错）
- Answer（助手最终答复）

你的任务：提炼出这次对话对未来最有用的信息，生成一条可用于时间线回溯的摘要。

要求：
- 用中文输出
- 尽量短（建议 1-3 句，总字数 < 200），但不要丢失关键细节
- 必须覆盖：用户意图/问题、关键结论/结果、关键动作或决策点（如涉及文件路径/函数名/命令/错误信息，保留最关键的 1-3 个）
- 忽略无关寒暄、重复内容、无价值的推理过程
- 不要复述 system prompt 原文，不要输出解释、步骤编号、JSON、Markdown、代码块
- 如果信息不足，输出你能确定的最小摘要，不要编造`

func SummaryConversation(c *schema.ConversationRecord, cid string) (string, error) {
	o := fmt.Sprintf("- System prompt:%s\n- Question:%s\n- Thoughts:%s\n- Answer:%s", c.System, c.Question, c.Thoughts, c.Answer)
	sp := summarySystemPrompt
	content, err := helper.UseLLM().Chat([]schema.OpenAIMessage{
		{
			Role:    schema.OpenAIMessageRoleUser,
			Content: o,
		},
	}, sp)
	if err != nil {
		return "", err
	}
	s := &schema.SummaryMemoryTable{
		Id:                util.RandomIdWithPrefix("sum-"),
		Link:              cid,
		LastSimulatedTime: c.CreateAt,
		Content:           content,
	}
	err = helper.UseDB().CreateSummaryMemoryTableRecord(s)
	if err != nil {
		return "", err
	}
	return s.Id, nil
}
