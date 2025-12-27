package node

import (
	"dna/decision/state"
)

const (
	Root state.Flag = iota
	Idle
	Task
	Sleep
	Watch
	Explore
	Introspection
	Answer
	Memory
	Ask
	WatchEnvironment
	WatchRoom
	ExploreConcept
	ExploreBehavior
	ExploreCharacter
	OrganizeMetacognition
)

func RootNodeAsTree() state.State {
	node := state.NewVirtualState(Root)

	node.AddChild(TaskNode())
	node.AddChild(SleepNode())
	// 默认进入最后一个状态
	node.AddChild(IdleNode())
	return node
}

func IdleNode() state.State {
	node := state.NewVirtualState(Idle)
	node.AddChild(ExploreNode())
	node.AddChild(WatchNode())
	node.AddChild(IntrospectionNode())
	return node
}

func SleepNode() state.State {
	return state.NewVirtualState(Sleep)
}

func IntrospectionNode() state.State {
	return state.NewVirtualState(Introspection)
}

func MemoryNode() state.State {
	return state.NewReturnState(Memory, nil)
}

func NewNodeByFlag(flag uint) state.State {
	name, exists := state.NameMap[flag]
	if !exists {
		return nil
	}

	switch name {
	case "Root":
		return RootNodeAsTree()
	case "Idle":
		return IdleNode()
	case "Task":
		return TaskNode()
	case "Sleep":
		return SleepNode()
	case "Watch":
		return WatchNode()
	case "Explore":
		return ExploreNode()
	case "Introspection":
		return IntrospectionNode()
	case "Answer":
		return AnswerNode()
	case "Memory":
		return MemoryNode()
	case "Ask":
		return AskNode()
	case "WatchEnvironment":
		return WatchEnvironmentNode()
	case "WatchRoom":
		return WatchRoomNode()
	case "ExploreConcept":
		return ExploreConceptNode()
	case "ExploreBehavior":
		return ExploreBehaviorNode()
	case "ExploreCharacter":
		return ExploreCharacterNode()
	// case "OrganizeMetacognition":
	// 	return OrganizeMetacognitionNode()
	default:
		return nil
	}
}
