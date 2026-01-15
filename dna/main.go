package main

import (
	"time"

	"github.com/Yoak3n/aimin/dna/decision"
)

func main() {
	m := decision.NewStateTree()
	m.Start(decision.Root)
	for {
		m.Update()
		time.Sleep(time.Second)
	}
}
