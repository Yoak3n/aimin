package decision

import "dna/fsm"

func NewWatchNode() *fsm.CompositeState {
	return fsm.NewCompositeState(Watch, Watch, nil, nil)
}
