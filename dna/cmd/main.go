package main

import (
	"dna/decision"
	"dna/pkg/logger"
)

func init() {
	logger.Init()
}

func main() {
	decision.BuildStateTree()
}
