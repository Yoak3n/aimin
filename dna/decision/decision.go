package decision

import (
	"context"
	"dna/decision/node"
	"dna/state"
	"log"
)

func BuildStateTree() {
	ctx := context.Background()
	tree := node.RootNodeAsTree()

	manager := state.NewStackStateManager(ctx, tree)
	if err := manager.Run(); err != nil {
		log.Fatal(err)
	}
}
