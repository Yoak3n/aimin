package decision

import "dna/fsm"

func NewIntrospectionNode() *fsm.CompositeState {
	return fsm.NewCompositeState(Introspection, Introspection, nil, nil)
}
