package router

import (
	"MixFound/web/controller"

	"github.com/gin-gonic/gin"
)

func InitBaseRouter(router *gin.RouterGroup) {
	baseRouter := router.Group("")
	{
		baseRouter.POST("query", controller.Query)
	}
}
