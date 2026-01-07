package main

import (
	"dna/decision"
	"time"
)

func main() {
	m := decision.NewStateTree()
	m.Start(decision.Root)
	for {
		m.Update()
		time.Sleep(time.Second)
	}
}
