package decision

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Yoak3n/aimin/dna/action"
	"github.com/Yoak3n/aimin/dna/fsm"
)

const ExploreChoice = "explore_choice"

func NewExploreNode() *fsm.CompositeState {
	e := fsm.NewCompositeState(Explore, Explore, []fsm.State{NewExploreCharacterNode(), NewExploreBehaviorNode(), NewExploreConceptNode()}, nil)
	e.SetRouterKey(ExploreChoice)
	e.SetSelect(func(ctx *fsm.Context, children []fsm.State) int {
		c := rand.IntN(3)
		ctx.Data[ExploreChoice] = children[c].ID()
		return c
	})
	return e
}

func NewExploreCharacterNode() *fsm.WorkState {
	check := func(ctx *fsm.Context) bool {
		c, ok := ctx.Data[ExploreChoice]
		return ok && c == ExploreCharacter
	}
	return fsm.NewWorkState(ExploreCharacter, ExploreCharacter, makeExploreNode(ExploreCharacter), check)
}

func NewExploreBehaviorNode() *fsm.WorkState {
	check := func(ctx *fsm.Context) bool {
		c, ok := ctx.Data[ExploreChoice]
		return ok && c == ExploreBehavior
	}
	return fsm.NewWorkState(ExploreBehavior, ExploreBehavior, makeExploreNode(ExploreBehavior), check)
}

func NewExploreConceptNode() *fsm.WorkState {
	check := func(ctx *fsm.Context) bool {
		c, ok := ctx.Data[ExploreChoice]
		return ok && c == ExploreConcept
	}
	return fsm.NewWorkState(ExploreConcept, ExploreConcept, makeExploreNode(ExploreConcept), check)
}

func createConceptQuestion(t string) string {
	// 模拟从知识库中获取概念图谱
	question := ""
	switch t {
	case ExploreConcept:
		key := "随机一个概念节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	case ExploreBehavior:
		key := "随机一个行为节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	case ExploreCharacter:
		key := "随机一个人物节点"
		chain := make([]string, 0)
		graph := append(chain, key)
		question = strings.Join(graph, " -> ")
	}
	fmt.Println("[Explore] Question: ", question)
	return question
}

func makeExploreNode(t string) fsm.WorkAction {
	progress := 1
	question := ""
	answer := make([]string, 0)
	return func(ctx *fsm.Context) string {
		ctx.Data[ExploreChoice] = ""
		for i := progress; i < 4; i++ {
			switch i {
			case 1:
				question = createConceptQuestion(t)
				progress++
			case 2:
				answer = askForAnswer(question)
				progress++
			case 3:
				handleExploreAnswer(answer)
				return fsm.Done
			}
		}
		return fsm.Interrupt
	}
}

func askForAnswer(question string) []string {
	choice := rand.IntN(3)
	if choice >= 2 {
		return action.ProactiveAsk(question)
	}
	// 网络搜索或主动提问
	return []string{"回答1", "回答2", "回答3"}
}

func handleExploreAnswer(answer []string) bool {
	return true
}
