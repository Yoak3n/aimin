package decision

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/service/llm"
	"github.com/Yoak3n/aimin/dna/action"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/Yoak3n/aimin/dna/persist"
)

func NewExploreNode(check func(ctx *fsm.Context) bool) *fsm.WorkState {
	node := fsm.NewWorkState(Explore, Explore, makeExploreAction(), check)
	node.SetDoneHook(Explore, Sleep, Introspection)
	return node
}

func makeExploreAction() fsm.WorkAction {
	progress := 1
	question := ""
	answer := make([]string, 0)
	chosenType := ""
	chosenName := ""
	var chosenDegree int64
	return func(ctx *fsm.Context) string {
		ctx.Attr.AddEnergy(-2)
		for i := progress; i < 4; i++ {
			switch i {
			case 1:
				question, chosenType, chosenName, chosenDegree = createExploreQuestionFromGraph(ctx, 10)
				progress++
			case 2:
				answer = askForAnswer(question)
				progress++
			case 3:
				handleExploreAnswer(answer)
				if ctx != nil && ctx.Persist != nil {
					ctx.Persist.Append("explore", map[string]any{
						"node_type": chosenType,
						"node_name": chosenName,
						"degree":    chosenDegree,
						"question":  question,
						"answer":    compactOneLine(strings.Join(answer, " | "), 400),
					})
				}
				return fsm.Done
			}
		}
		return fsm.Interrupt
	}
}

func createExploreQuestionFromGraph(ctx *fsm.Context, candidateLimit int) (string, string, string, int64) {
	if candidateLimit <= 0 {
		candidateLimit = 10
	}
	n4 := helper.UseDB().GetNeuroDB()
	if n4 == nil {
		return "请探索：从图数据库中选取一个连接边最少的节点（当前数据库不可用）", "", "", 0
	}

	candidates, err := n4.FindLeastConnectedNodes("", candidateLimit)
	if err != nil || len(candidates) == 0 {
		return "请探索：从图数据库中选取一个连接边最少的节点（未获取到候选节点）", "", "", 0
	}

	chosen := candidates[rand.IntN(len(candidates))]
	noise := ""
	if ctx != nil && ctx.Persist != nil {
		noise = exploreNoise(ctx.Persist, 1200, 18)
	}
	question := generateExploreQuestionByLLM(chosen.Type, chosen.Name, chosen.Degree, noise)
	if strings.TrimSpace(question) == "" {
		question = fmt.Sprintf("请探索图谱节点：%s:%s（连接度=%d）", chosen.Type, chosen.Name, chosen.Degree)
	}
	logger.Logger.Infof("[Explore] Question: %s", question)
	return question, chosen.Type, chosen.Name, chosen.Degree
}

func generateExploreQuestionByLLM(nodeType string, nodeName string, degree int64, noise string) string {
	user := fmt.Sprintf("目标节点：%s:%s；连接度：%d。\n噪声上下文（最近探索记录压缩）：\n%s", nodeType, nodeName, degree, strings.TrimSpace(noise))
	systemPrompt := "你是一个智能体的探索问题生成器。请基于给定目标节点生成一个中文探索问题/任务指令，要求：1) 只输出一句话，不要换行；2) 风格有随机性（可在提问角度/任务形式上变化）；3) 可参考噪声上下文调整角度，但不要复述上下文；4) 不要输出任何额外解释。"
	out, err := llm.Chat([]schema.OpenAIMessage{{Role: schema.OpenAIMessageRoleUser, Content: user}}, systemPrompt)
	if err != nil {
		return ""
	}
	out = strings.TrimSpace(out)
	out = strings.Trim(out, "\"'“”")
	out = strings.ReplaceAll(out, "\n", " ")
	out = strings.Join(strings.Fields(out), " ")
	return out
}

func exploreNoise(p *persist.PersistStore, maxChars int, recentLimit int) string {
	if p == nil {
		return ""
	}
	if maxChars <= 0 {
		maxChars = 1200
	}
	if recentLimit <= 0 {
		recentLimit = 18
	}
	cs, ok := p.GetChannel("explore")
	if !ok {
		return ""
	}
	sb := strings.Builder{}
	if strings.TrimSpace(cs.Summary) != "" {
		sb.WriteString(strings.TrimSpace(cs.Summary))
		sb.WriteString("\n")
	}
	if len(cs.Records) > 0 {
		sb.WriteString("recent:\n")
		start := 0
		if len(cs.Records) > recentLimit {
			start = len(cs.Records) - recentLimit
		}
		for i := start; i < len(cs.Records); i++ {
			r := cs.Records[i]
			line := exploreOneLineJSON(r.Data, 220)
			sb.WriteString("- ")
			if strings.TrimSpace(r.At) != "" {
				sb.WriteString(r.At)
				sb.WriteString(" ")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
			if sb.Len() >= maxChars {
				break
			}
		}
	}
	out := strings.TrimSpace(sb.String())
	if len(out) > maxChars {
		out = out[:maxChars]
	}
	return strings.TrimSpace(out)
}

func exploreOneLineJSON(v any, max int) string {
	buf, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(buf))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}

func compactOneLine(s string, max int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}

func askForAnswer(question string) []string {
	choice := rand.IntN(3)
	if choice >= 2 {
		return action.ProactiveAsk(question)
	}
	// 网络搜索
	return []string{"回答1", "回答2", "回答3"}
}

func handleExploreAnswer(answer []string) bool {
	return true
}
