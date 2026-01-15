package router

import "github.com/gin-gonic/gin"

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(Cors())
	RegisterRouter(r)
	return r
}

func RegisterRouter(r *gin.Engine) {

}
