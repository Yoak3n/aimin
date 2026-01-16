package net

import (
	"fmt"
	"sync"

	"github.com/Yoak3n/aimin/aimin/internal/service/router"
	"github.com/Yoak3n/aimin/aimin/internal/service/ws"
	"github.com/gin-gonic/gin"
)

type Service struct {
	router *gin.Engine
	hub    *ws.WebSocketHub
}

var Srv *Service
var once sync.Once

func UseService() *Service {
	once.Do(func() {
		Srv = &Service{
			router: router.InitRouter(),
			hub:    ws.NewWebSocketHub(),
		}

	})
	return Srv
}

func (S *Service) Start(port int) error {
	go S.hub.Run()
	return S.router.Run(fmt.Sprintf(":%d", port))
}

func (S *Service) Router() *gin.Engine {
	return S.router
}

func (S *Service) Hub() *ws.WebSocketHub {
	return S.hub
}
