package node

import "dna/fsm"

func NewExploreNode() *fsm.VirtualState {
	return fsm.NewVirtualState(Explore, Explore, []string{ExploreCharacter, ExploreBehavior, ExploreConcept})
}
