package decision

import "github.com/Yoak3n/aimin/dna/fsm"

func NewWatchNode() *fsm.CompositeState {
	return fsm.NewCompositeState(Watch, Watch, nil, nil)
}
