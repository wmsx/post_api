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

	postCategoryRouter := r.Group("/post/category")
	postCategoryRouter.POST("/create", mygin.AuthWrapper(categoryHandler.CreateCategory))
	postCategoryRouter.POST("/list", categoryHandler.GetCategoryList)

	postMengerRouter := r.Group("/post/menger")
	postMengerRouter.POST("/list", mygin.AuthWrapper(postHandler.GetMengerPostlist))
	postMengerRouter.POST("/favorite/list", mygin.AuthWrapper(postHandler.GetMengerFavoritePostlist))
	postMengerRouter.POST("/thumbup/list", mygin.AuthWrapper(postHandler.GetMengerThumbUpPostlist))

	return r
}
