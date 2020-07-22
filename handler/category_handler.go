package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	mygin "github.com/wmsx/pkg/gin"
	categoryProto "github.com/wmsx/post_svc/proto/category"
	postProto "github.com/wmsx/post_svc/proto/post"
)

type CategoryHandler struct {
	postClient     postProto.PostService
	categoryClient categoryProto.CategoryService
}

func NewCategoryHandler(c client.Client) *CategoryHandler {
	return &CategoryHandler{
		postClient:     postProto.NewPostService(postSvcName, c),
		categoryClient: categoryProto.NewCategoryService(postSvcName, c),
	}
}

// 创建分类
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var (
		categoryParam     CategoryParam
		err               error
		s                 *mygin.Session
		createCategoryRes *categoryProto.CreateCategoryResponse
	)
	app := mygin.Gin{C: c}

	if err = c.ShouldBindJSON(&categoryParam); err != nil {
		app.LogicErrorResponse("参数错误")
		return
	}

	if s, err = mygin.NewSession(c); err != nil {
		app.ServerErrorResponse()
		return
	}

	createCategoryRequest := &categoryProto.CreateCategoryRequest{
		Name:     categoryParam.Name,
		ShowName: categoryParam.ShowName,
		MengerId: s.GetMengerId(),
	}
	if createCategoryRes, err = h.categoryClient.CreateCategory(c, createCategoryRequest); err != nil {
		app.ServerErrorResponse()
		return
	}
	if createCategoryRes.ErrorMsg != "" {
		app.LogicErrorResponse(createCategoryRes.ErrorMsg)
		return
	}

	app.Response(nil)
	return
}