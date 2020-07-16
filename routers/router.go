package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/wmsx/post_api/handler"
)

/**
 * 初始化路由信息
 */
func InitRouter(c client.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())

	postHandler := handler.NewPostHandler(c)
	postRouter := r.Group("/post")

	postRouter.POST("/list/", postHandler.GetPostList)

	return r
}
