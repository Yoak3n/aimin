package main

import (
	"github.com/Yoak3n/aimin/aimin/cmd/app/net"
)

func main() {
	s := net.UseService()
	err := s.Start(8080)
	if err != nil {
		panic(err)
	}
}
