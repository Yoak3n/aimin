package main

import (
	"github.com/Yoak3n/aimin/aimin/cmd/app"
)

func main() {
	c := app.GetGlobalComponent()
	err := c.Router().Run(":8080")
	if err != nil {
		panic(err)
	}
}
