package reason

import (
	"fmt"

	"github.com/Yoak3n/aimin/nerve/controller"
)

const defaultSystemPrompt = "请根据以下元认知记录生成对话系统提示信息：\n"
const defaultOutput = "你是一个人工智能助手"

// GenConversationSystemPrompt 生成对话系统的提示信息
func GenConversationSystemPrompt() string {
	// 获取当前元认知记录
	records, err := controller.GetCurrentMetacognitionRecords()
	if err != nil || records == nil {
		return defaultOutput
	}

	ms := make([]string, 0, len(records))
	for _, record := range records {
		ms = append(ms, fmt.Sprintf("%s:%s\n", record.InsightType, record.InsightText))
	}

	// 返回一个系统提示字符串，用于指导AI如何判断对话是否需要长期记忆
	return fmt.Sprintf("%s%s", defaultSystemPrompt, ms)
}
