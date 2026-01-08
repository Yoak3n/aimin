package reason

import (
	"fmt"
	"nerve/controller"
)

// GenConversationSystemPrompt 生成对话系统的提示信息
func GenConversationSystemPrompt() string {
	// 获取当前记录
	records, err := controller.GetCurrentMetacognitionRecords()
	if err != nil {
		return ""
	}
	ms := make([]string, 0)
	for _, record := range records {
		ms = append(ms, fmt.Sprintf("%s:%s\n", record.InsightType, record.InsightType))
	}

	// 返回一个系统提示字符串，用于指导AI如何判断对话是否需要长期记忆
	return fmt.Sprintf("请根据以下元认知记录生成对话系统提示信息：\n%s", ms)
}
