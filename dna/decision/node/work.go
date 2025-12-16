package node

import "dna/state"

func WorkNode() state.State {
	return state.NewVirtualState(Work)
}

func AnswerNode() state.State {
	return state.NewVirtualState(Answer)
}
