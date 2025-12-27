package node

import (
	state "dna/decision/state"
	"math/rand"
	"strings"
)

type ExploreResult struct {
	Question string
	Answer   string
	NeedAsk  bool
	state.ResultData
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
	node.AddCondition(func(ctx *state.Context) bool {
		return true
	})
	return node
}

func ExploreConceptNode() state.State {
	node := state.NewWorkState(
		ExploreConcept,
		ExploreAction,
	)
	// 由于保底决策的存在，不把回归节点加入状态树，后续再看两者哪个更好
	//node.AddChild(AskNode())
	return node
}

func ExploreBehaviorNode() state.State {
	node := state.NewWorkState(
		ExploreBehavior,
		ExploreAction,
	)
	return node
}

func ExploreCharacterNode() state.State {
	node := state.NewWorkState(
		ExploreCharacter,
		ExploreAction,
	)
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

func ExploreAction(ctx *state.Context) (state.Result, error) {
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
				continue
			}
			// 将data交给Ask节点
			ctx.Data[Ask] = AskData{Question: initSave.Question}
			return ExploreResult{
				ResultData: state.ResultData{
					Status:    state.ToReturn,
					From:      ctx.Caller,
					NextState: AskNode(),
					Data:      AskData{Question: initSave.Question},
				},
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
		ResultData: state.ResultData{
			Status: state.Interrupted,
			From:   ctx.Caller,
		},
	}, nil
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
