package node

import (
	"dna/state"
	"math/rand"
	"strings"
)

type ExploreResult struct {
	Question string
	Answer   string
	NeedAsk  bool
}

type ExploreSave struct {
	Step         int
	SearchResult []string
	Question     string
}

func ExploreNode() state.State {
	node := state.NewVirtualState(
		Explore,
	)
	node.AddChild(ExploreConceptNode())
	node.AddChild(ExploreBehaviorNode())
	node.AddChild(ExploreCharacterNode())
	return node
}

func ExploreConceptNode() state.State {
	node := state.NewConditionalWorkState(
		ExploreConcept,
		ExploreAction,
		func(data any) bool {
			if d, ok := data.(ExploreResult); ok {
				return d.NeedAsk
			}
			return false
		},
	)
	node.AddChild(AskNode())
	return node
}

func ExploreBehaviorNode() state.State {
	node := state.NewConditionalWorkState(
		ExploreBehavior,
		ExploreAction,
		func(data any) bool {
			if d, ok := data.(ExploreResult); ok {
				return d.NeedAsk
			}
			return false
		},
	)
	node.AddChild(AskNode())
	return node
}

func ExploreCharacterNode() state.State {
	node := state.NewConditionalWorkState(
		ExploreCharacter,
		ExploreAction,
		func(data any) bool {
			if d, ok := data.(ExploreResult); ok {
				return d.NeedAsk
			}
			return false
		},
	)
	node.AddChild(AskNode())
	return node
}

func createConceptQuestion(t uint) string {
	// 模拟从知识库中获取概念图谱
	question := ""
	switch t {
	case uint(ExploreConcept):
		key := "随机一个概念节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	case uint(ExploreBehavior):
		key := "随机一个行为节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	case uint(ExploreCharacter):
		key := "随机一个人物节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	}
	return question
}

func ExploreAction(ctx *state.Context) (any, error) {
	// 读取Ask节点的结果
	if v, exists := ctx.Data[Ask]; exists {
		if ad, ok := v.(AskData); ok && ad.Answer != "" {
			// ret := fmt.Sprintf("%s %s", ad.Question, ad.Answer)
			// 7.处理答案，更新知识库
			// fmt.Println(ret)
			a := HandleExploreResult([]string{ad.Answer})
			ctx.Data[Ask] = nil // 清除Ask数据
			return ExploreResult{
				Question: ad.Question,
				Answer:   a,
				NeedAsk:  false,
			}, nil
		}
	}
	initSave := ExploreSave{
		Step:         1,
		Question:     "",
		SearchResult: make([]string, 0),
	}
	caller := ctx.Caller.Type()
	if save, exists := ctx.Save[caller]; exists {
		if es, ok := save.(ExploreSave); ok {
			initSave.Step = es.Step
			initSave.SearchResult = es.SearchResult
			initSave.Question = es.Question
		}
	}

	for i := initSave.Step; i <= 3; i++ {
		// 按步骤更新，指向的步骤为将要进行的步骤
		initSave.Step = i
		switch i {
		case 1:
			initSave.Question = createConceptQuestion(uint(caller))
			initSave.Step = 2
			if exploreCheckpoint(ctx, initSave) {
				break
			}
		case 2:
			choice := rand.Intn(3) != 1
			if choice {
				// 5. 从网络上搜寻答案，
				initSave.SearchResult = []string{"随机一个概念节点", "随机一个行为节点", "随机一个人物节点", initSave.Question}
				initSave.Step = 3
				if exploreCheckpoint(ctx, initSave) {
					break
				}
			}
			// 将data交给Ask节点
			ctx.Data[Ask] = AskData{Question: initSave.Question}
			return ExploreResult{
				Question: initSave.Question,
				NeedAsk:  true,
			}, nil
		case 3:
			// 继续处理从网络上搜寻的答案
			// 7.处理答案，更新知识库

			answer := HandleExploreResult(initSave.SearchResult)
			ctx.Save[caller] = nil // 清除Explore数据
			return ExploreResult{
				Question: initSave.Question,
				Answer:   answer,
				NeedAsk:  false,
			}, nil
		}
	}
	// 从步骤中跳出，说明是未完成状态
	return ExploreResult{
		Question: initSave.Question,
		NeedAsk:  false,
	}, state.ErrInterrupted
}

func exploreCheckpoint(ctx *state.Context, save ExploreSave) bool {
	caller := ctx.Caller.Type()
	ctx.Save[caller] = save
	select {
	case <-ctx.Done():
		return false
	case <-ctx.Interrupt:
		return true
	default:
		return false
	}
}

func HandleExploreResult(results []string) string {

	return ""
}
