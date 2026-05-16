package main

import (
	"github.com/Yoak3n/aimin/aimin/app/componet"
	"github.com/Yoak3n/aimin/aimin/internal/service/router"
	"github.com/Yoak3n/aimin/aimin/internal/service/ws"
	"github.com/Yoak3n/aimin/blood/pkg/logger"
	"github.com/Yoak3n/aimin/dna/action"
	"github.com/Yoak3n/aimin/dna/decision"
	"github.com/Yoak3n/aimin/hand/interactive"
	"github.com/Yoak3n/aimin/tongue/conversation"
)

func init() {
	hub := ws.UseWebSocketHub()
	go hub.Run()

	logger.Init()
	logger.SetExternalHandler(hub.BroadcastLog)
	action.RemoteAsk = hub.AskClient
	interactive.WSReplyBroadcast = hub.SendToClient
	decision.TaskExecutor = func(id, from, question string) error {
		return conversation.GetManager().AskDirect(id, from, question)
	}

	c := componet.GetGlobalComponent()
	c.FSM().SetOnStateChange(hub.BroadcastState)
	go c.Start()
}

func main() {
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
