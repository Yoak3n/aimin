package node

import "dna/state"

const (
	Root state.Flag = iota
	Idle
	Work
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

	node.AddChild(IdleNode())
	// node.AddChild(WorkNode())
	node.AddChild(SleepNode())

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

func WatchEnvironmentNode(action func(ctx *state.Context) (any, error)) state.State {
	return state.NewWorkState(WatchEnvironment, action)
}
