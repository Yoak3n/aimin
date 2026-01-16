package response

import "github.com/gin-gonic/gin"

func Response(c *gin.Context, code int, data interface{}, msg ...string) {
	ret := map[string]interface{}{
		"code": code,
		"data": data,
	}
	if len(msg) > 0 {
		ret["msg"] = msg[0]
	}
	c.JSON(code, ret)
}

func Success(c *gin.Context, data interface{}) {
	Response(c, 200, data, "success")
}

func Error(c *gin.Context, code int, msg string) {
	Response(c, code, nil, msg)
}
