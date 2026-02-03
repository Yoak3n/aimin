package response

import "github.com/gin-gonic/gin"

func Response(c *gin.Context, status int, code int, data any, msg ...string) {
	ret := map[string]any{
		"code": code,
		"data": data,
	}
	if len(msg) > 0 {
		ret["msg"] = msg[0]
	}
	c.JSON(status, ret)
}

func Success(c *gin.Context, data any) {
	Response(c, 200, 0, data, "success")
}

func Error(c *gin.Context, status int, msg string) {
	Response(c, status, -1, nil, msg)
}

func ErrorWithCode(c *gin.Context, status int, code int, msg string) {
	Response(c, status, code, nil, msg)
}
