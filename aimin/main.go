package main

import (
	"github.com/Yoak3n/aimin/aimin/cmd/app/componet"
	"github.com/Yoak3n/aimin/aimin/internal/service/router"
	"github.com/Yoak3n/aimin/aimin/internal/service/ws"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
)

func main() {
	ws.InitWebSocketHub()
	hub := ws.UseWebSocketHub()
	go hub.Run()
	logger.Init()
	logger.SetExternalHandler(hub.BroadcastLog)

	c := componet.GetGlobalComponent()
	go c.Start()
	r := router.InitRouter()
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
