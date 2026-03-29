package controller

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/util"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/tidwall/gjson"
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

const temporarySummaryConfidenceThreshold = 0.6

func upsertSummaryByLink(link string, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("summary content is empty")
	}
	now := time.Now()
	existing, err := helper.UseDB().GetSummaryMemoryTableRecordByLink(link)
	if err == nil && existing.Id != "" {
		existing.Content = content
		existing.LastSimulatedTime = now
		return helper.UseDB().UpdateSummaryMemoryTableRecord(&existing)
	}
	rec := &schema.SummaryMemoryTable{
		Id:                util.RandomIdWithPrefix("sum"),
		Link:              link,
		Content:           content,
		LastSimulatedTime: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return helper.UseDB().CreateSummaryMemoryTableRecord(rec)
}

func InsertSummaryMemory(system, question, thoughts, answer, cid string) {
	strategy := fetchMemoryStrategy(system, question, thoughts, answer)
	extractStrategy(strategy, cid)
}

func fetchMemoryStrategy(system, question, thoughts, answer string) string {
	systemPrompt := `你是一个“记忆分流器 + 摘要器”。

你会收到一次对话的三部分内容：
- Question（用户问题）
- Thoughts（助手思考过程，可能很长、含试错）
- Answer（助手最终答复）

你的任务分两步：
1) 先根据 Question/Thoughts/Answer 生成一段“对未来最有用”的摘要（1-3句，总字数<200），忽略寒暄与无价值推理，不要复述系统提示词原文。
2) 再基于该摘要，判断是否需要长期记忆：
   - enduring：稳定事实、长期偏好、承诺、关系、持久身份信息、后续有参考价值的事件/时间
   - temporary：一次性问题/临时上下文/无长期价值

输出要求：
- 仅输出一段合法 JSON（不要任何解释性文本、不要 Markdown code fence）
- 必含字段：
  - decision: "enduring" 或 "temporary"
  - confidence: 0.0 到 1.0
- 当 decision="temporary"：必须输出 temporary 对象：
  - topic: 一句话主题
  - content: 摘要（1-3句，总字数<200）
  - reason: 判定为短期的简要理由
- 当 decision="enduring"：必须输出 enduring 对象，并且才做实体/三元组抽取：
  - topic: 一句话主题
  - triples: 至少 1 条三元组（必须是数组），字段包括 subject、subject_type、predicate、object、object_type，尽量围绕主题
  - content: 稳定事实/长期偏好等的自然语言描述（可与摘要不同，偏“可长期保存”的稳定信息）
  - reason: 判定为长期的简要理由

取值规范：
- subject/object：自然语言实体名（如“Yoa”“Aimin”“项目A”“上海”）
- subject_type/object_type：人物/组织/地点/物品/事件/概念 等
- predicate：动词或关系短语（如“喜欢/位于/属于/承诺/认识/工作于/发生于/是/偏好”）
- 如果能判断为 enduring 但难以抽取具体三元组，也必须给出 1 条“弱三元组”（例如用“用户/对话主题”等作为实体）以满足结构要求。`

	userPrompt := schema.OpenAIMessage{
		Role:    schema.OpenAIMessageRoleUser,
		Content: fmt.Sprintf("- System prompt:%s\n- Question:%s\n- Thoughts:%s\n- Answer:%s", system, question, thoughts, answer),
	}
	resp, err := helper.UseLLM().Chat([]schema.OpenAIMessage{userPrompt}, systemPrompt)
	if err != nil {
		return ""
	}
	return resp
}

func extractStrategy(s, cid string) error {
	trimmedRespMsg := strings.TrimSpace(s)
	trimmedRespMsg = strings.TrimPrefix(trimmedRespMsg, "```json")
	trimmedRespMsg = strings.TrimSuffix(trimmedRespMsg, "```")
	result := gjson.Parse(trimmedRespMsg)
	decision := result.Get("decision").String()
	confidence := result.Get("confidence").Float()
	switch decision {
	case "enduring":
		enduring := result.Get("enduring")
		if !enduring.Exists() {
			return errors.New("enduring not exists")
		}
		enduringSummary := strings.TrimSpace(enduring.Get("content").String())
		if enduringSummary == "" {
			enduringSummary = strings.TrimSpace(enduring.Get("topic").String())
		}
		if err := upsertSummaryByLink(cid, enduringSummary); err != nil {
			return err
		}
		triples := enduring.Get("triples")
		if !triples.Exists() {
			return errors.New("triples not exists")
		}
		entitiesMap := make([]schema.EntityTable, 0)

		triples.ForEach(func(key, value gjson.Result) bool {
			memoryMap := &schema.EntityTable{
				Link:        cid,
				Subject:     value.Get("subject").String(),
				Predicate:   value.Get("predicate").String(),
				Object:      value.Get("object").String(),
				SubjectType: value.Get("subject_type").String(),
				ObjectType:  value.Get("object_type").String(),
			}
			entitiesMap = append(entitiesMap, *memoryMap)
			return true
		})
		if len(entitiesMap) == 0 {
			return errors.New("triples is empty")
		}
		if err := helper.UseDB().GetPostgresSQL().Create(&entitiesMap).Error; err != nil {
			return err
		}
		if err := helper.UseDB().GetNeuroDB().CreateNode(entitiesMap); err != nil {
			return err
		}
	case "temporary":
		if confidence < temporarySummaryConfidenceThreshold {
			return nil
		}
		temporary := result.Get("temporary")
		if !temporary.Exists() {
			return errors.New("temporary not exists")
		}
		content := temporary.Get("content").String()
		return upsertSummaryByLink(cid, content)
	default:
		log.Println("unknown decision: ", decision)
	}
	return nil
}
