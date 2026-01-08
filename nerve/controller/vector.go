package controller

import (
	"blood/pkg/helper"
	"blood/schema"
)

func FetchMemoryStrategy(messages []schema.OpenAIMessage) string {
	if len(messages) == 0 {
		return ""
	}

	formatMessage := ""
	for _, m := range messages {
		formatMessage += m.Role + ": "
		formatMessage += m.Content + "\n---\n"
	}
	systemPrompt := `现在你是一个智能助手，你需要根据用户的问题，判断用户的问题是否需要长期记忆。
- 判别逻辑
- 当对话包含稳定事实、长期偏好、承诺、关系、持久身份信息、后续有参考价值的事件/时间等，判定为 enduring
- 当对话仅为临时交流、无长期价值或一次性的任务上下文，判定为 temporary
- 输出要求
- 仅输出一段合法 JSON（不要任何解释性文本）
- 必含字段：
  - decision: enduring 或 temporary
  - confidence: 0.0 到 1.0 的可信度
  - enduring: 当决策为长期记忆时填写，包括：
	- topic: 一句话主题
    - triples: 至少一条三元组，字段包括 subject、subject_type、predicate、object、object_type, 尽量围绕主题
    - content: 稳定事实或长期偏好等的自然语言描述
    - reason: 判定为长期的简要理由
  - temporary: 当决策为短期记忆时填写，包括：
    - topic: 一句话主题
    - content: 简要的对话摘要
    - reason: 判定为短期的简要理由
- 取值规范
- subject/object：使用自然语言实体名，如“张三”“公司A”“上海”，subject为动作或关系的发出者，object为动作或关系的接收者
- subject_type/object_type：填写实体类型，如“人物/组织/地点/物品/事件/概念”等
- predicate：使用动词或关系短语，如“喜欢/位于/属于/承诺/认识/工作于/发生于”等
- 若抽取不到三元组但仍属长期记忆，可用 content 填稳定事实陈述`
	userPrompt := schema.OpenAIMessage{
		Role:    "user",
		Content: formatMessage,
	}
	answer, err := helper.UseLLM().Chat([]schema.OpenAIMessage{userPrompt}, systemPrompt)
	if err != nil {
		return ""
	}
	return answer
}
