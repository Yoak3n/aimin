package router

import (
	"github.com/Yoak3n/aimin/aimin/internal/service/handler"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(Cors())
	RegisterRouter(r)
	return r
}

func RegisterRouter(r *gin.Engine) {
	APIRouter(r)
	// 查看当前智能体状态
	// 与智能体交互接口（包括主动提问和被动对话）-> ws？
}

func APIRouter(r *gin.Engine) {
	api := r.Group("/api")
	registerV1Router(api)
}

func registerV1Router(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	v1.GET("/status", handler.StatusHandler)
}
