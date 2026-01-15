package decision

import "github.com/Yoak3n/aimin/dna/fsm"

func NewIntrospectionNode() *fsm.CompositeState {
	return fsm.NewCompositeState(Introspection, Introspection, nil, nil)
}
