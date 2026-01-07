package decision

import (
	"context"
	"dna/decision/node"
	"dna/decision/state"
	"log"
)

func BuildStateTree() {
	ctx := context.Background()
	tree := node.RootNodeAsTree()

	manager := state.NewStateManager(tree)
	if err := manager.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func BuildFSM() {}
