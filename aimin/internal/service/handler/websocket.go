package handler

import (
	"net/http"

	"github.com/Yoak3n/aimin/aimin/internal/service/ws"
	"github.com/Yoak3n/aimin/aimin/internal/util/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(c *gin.Context) {
	hub := ws.UseWebSocketHub()
	id := c.Param("id")
	if id == "" {
		response.Error(c, 400, "id is required")
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	defer conn.Close()
	hub.Register(id, conn)
}
