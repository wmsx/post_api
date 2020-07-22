package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	mygin "github.com/wmsx/pkg/gin"
	"github.com/wmsx/post_api/handler"
)

/**
 * 初始化路由信息
 */
func InitRouter(c client.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	postHandler := handler.NewPostHandler(c)
	categoryHandler := handler.NewCategoryHandler(c)
	postRouter := r.Group("/post")

	postRouter.POST("/list", postHandler.GetPostList)
	postRouter.POST("/create", mygin.AuthWrapper(postHandler.CreatePost))
	postRouter.POST("/category/create", mygin.AuthWrapper(categoryHandler.CreateCategory))
	return r
}
