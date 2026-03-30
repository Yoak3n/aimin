package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/Yoak3n/aimin/blood/pkg/helper"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/blood/schema"
	"github.com/Yoak3n/aimin/blood/service/llm"
	"github.com/Yoak3n/aimin/dna/action"
	"github.com/Yoak3n/aimin/dna/fsm"
	"github.com/Yoak3n/aimin/dna/persist"
	"github.com/Yoak3n/aimin/hand/internet/search/duckduckgo"
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
	chosenStrategy := ExploreStrategyWebSearch
	return func(ctx *fsm.Context) string {
		ctx.Attr.AddEnergy(-2)
		for i := progress; i < 4; i++ {
			switch i {
			case 1:
				chosenStrategy = chooseExploreStrategy()
				question, chosenType, chosenName, chosenDegree = createExploreQuestionFromGraph(ctx, 10, chosenStrategy)
				progress++
			case 2:
				answer = askForAnswer(question, chosenStrategy)
				progress++
			case 3:
				handleExploreAnswer(answer)
				if ctx != nil && ctx.Persist != nil {
					ctx.Persist.Append("explore", map[string]any{
						"node_type": chosenType,
						"node_name": chosenName,
						"degree":    chosenDegree,
						"strategy":  string(chosenStrategy),
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

type ExploreStrategy string

const (
	ExploreStrategyWebSearch ExploreStrategy = "web_search"
	ExploreStrategyAskUser   ExploreStrategy = "ask_user"
)

func chooseExploreStrategy() ExploreStrategy {
	choice := rand.IntN(3)
	if choice >= 2 {
		return ExploreStrategyAskUser
	}
	return ExploreStrategyWebSearch
}

func createExploreQuestionFromGraph(ctx *fsm.Context, candidateLimit int, strategy ExploreStrategy) (string, string, string, int64) {
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
	question := generateExploreQuestionByLLM(chosen.Type, chosen.Name, chosen.Degree, noise, strategy)
	if strings.TrimSpace(question) == "" {
		question = fmt.Sprintf("请探索图谱节点：%s:%s（连接度=%d）", chosen.Type, chosen.Name, chosen.Degree)
	}
	logger.Logger.Infof("[Explore] Strategy=%s Question: %s", strategy, question)
	return question, chosen.Type, chosen.Name, chosen.Degree
}

func generateExploreQuestionByLLM(nodeType string, nodeName string, degree int64, noise string, strategy ExploreStrategy) string {
	user := fmt.Sprintf("目标节点：%s:%s；连接度：%d。\n噪声上下文（最近探索记录压缩）：\n%s", nodeType, nodeName, degree, strings.TrimSpace(noise))
	systemPrompt := ""
	switch strategy {
	case ExploreStrategyWebSearch:
		systemPrompt = "你是搜索查询生成器。请基于给定目标节点生成一个适合在搜索引擎直接搜索的中文查询串，要求：1) 只输出一行，不要换行；2) 不要写成问句，不要带任何解释；3) 尽量用关键词短语，必要时用空格分隔；4) 必须包含节点名，尽量包含节点类型或等价领域词；5) 可参考噪声上下文挑选更有信息量的限定词，但不要复述上下文。"
	default:
		systemPrompt = "你是一个智能体的探索提问生成器。请基于给定目标节点生成一句自然得体的中文提问/任务指令，用于向用户请教或请用户补充信息，要求：1) 只输出一句话，不要换行；2) 语气礼貌、明确且易回答；3) 风格有随机性（可在提问角度/任务形式上变化）；4) 可参考噪声上下文调整角度，但不要复述上下文；5) 不要输出任何额外解释。"
	}
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

func askForAnswer(question string, strategy ExploreStrategy) []string {
	if strategy == ExploreStrategyAskUser {
		return action.ProactiveAsk(question)
	}
	// 网络搜索
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	res, err := duckduckgo.Search(ctx, question, &duckduckgo.Options{
		Timeout: 20 * time.Second,
	})
	if err != nil {
		return []string{fmt.Sprintf("[DuckDuckGo][错误] %v", err)}
	}

	out := make([]string, 0, 16)
	if s := strings.TrimSpace(res.Heading); s != "" {
		out = append(out, "[主题] "+s)
	}
	if s := strings.TrimSpace(res.AbstractText); s != "" {
		out = append(out, "[摘要] "+s)
	}
	if s := strings.TrimSpace(res.Answer); s != "" {
		t := strings.TrimSpace(res.AnswerType)
		if t != "" {
			out = append(out, "[答案/"+t+"] "+s)
		} else {
			out = append(out, "[答案] "+s)
		}
	}
	if s := strings.TrimSpace(res.Definition); s != "" {
		out = append(out, "[定义] "+s)
	}

	if len(res.Results) > 0 {
		out = append(out, "[结果]")
		n := 5
		if len(res.Results) < n {
			n = len(res.Results)
		}
		for i := 0; i < n; i++ {
			it := res.Results[i]
			line := strings.TrimSpace(it.Text)
			if line != "" {
				out = append(out, fmt.Sprintf("- %s", line))
			}
		}
	}

	if len(res.RelatedTopics) > 0 {
		out = append(out, "[相关]")
		n := 5
		if len(res.RelatedTopics) < n {
			n = len(res.RelatedTopics)
		}
		for i := 0; i < n; i++ {
			it := res.RelatedTopics[i]
			line := strings.TrimSpace(it.Text)
			if line != "" {
				out = append(out, fmt.Sprintf("- %s", line))
			}
		}
	}

	if len(out) == 0 {
		return []string{"[DuckDuckGo] 未返回可用内容"}
	}
	return out
}

func handleExploreAnswer(answer []string) bool {
	return true
}
