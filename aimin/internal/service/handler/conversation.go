package handler

import (
	"github.com/Yoak3n/aimin/aimin/internal/util/response"
	"github.com/Yoak3n/aimin/blood/dao/controller"
	"github.com/gin-gonic/gin"
)

func ConversationListHandler(c *gin.Context) {
	records, err := controller.GetAllConversations()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, records)
}

func ConversationDetailHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, 400, "id is required")
		return
	}
	records, err := controller.GetDialoguesWithConversation(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, records)
}
