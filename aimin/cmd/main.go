package main

import (
	"github.com/Yoak3n/aimin/aimin/cmd/app/componet"
	"github.com/Yoak3n/aimin/aimin/internal/service/router"
	"github.com/Yoak3n/aimin/aimin/internal/service/ws"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/action"
	"github.com/Yoak3n/aimin/face/conversation"
)

func init() {
	hub := ws.UseWebSocketHub()
	go hub.Run()

	logger.Init()
	logger.SetExternalHandler(hub.BroadcastLog)
	action.RemoteAsk = hub.Ask

	conversation.GetManager().SetReplyHandler(hub.SendReply)

	c := componet.GetGlobalComponent()
	c.FSM().SetOnStateChange(hub.BroadcastState)
	go c.Start()
}

func main() {
	r := router.InitRouter()
	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
