package handler

import (
	"github.com/Yoak3n/aimin/aimin/internal/service/controller"
	"github.com/Yoak3n/aimin/aimin/internal/util/response"
	"github.com/gin-gonic/gin"
)

func StatusHandler(c *gin.Context) {
	s := controller.GetFSMStatus()
	response.Success(c, s)
}
