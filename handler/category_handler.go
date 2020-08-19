package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/util/log"
	mygin "github.com/wmsx/pkg/gin"
	categoryProto "github.com/wmsx/post_svc/proto/category"
	postProto "github.com/wmsx/post_svc/proto/post"
	"strconv"
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
		createCategoryRes *categoryProto.CreateCategoryResponse
	)
	app := mygin.Gin{C: c}

	if err = c.ShouldBindJSON(&categoryParam); err != nil {
		app.LogicErrorResponse("参数错误")
		return
	}

	mengerId, err := strconv.ParseInt(c.GetHeader("uid"), 10, 64)
	if err != nil {
		log.Error("获取用户id失败 err:  ", err)
		app.ServerErrorResponse()
		return
	}

	createCategoryRequest := &categoryProto.CreateCategoryRequest{
		Name:     categoryParam.Name,
		ShowName: categoryParam.ShowName,
		MengerId: mengerId,
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

func (h *CategoryHandler) GetCategoryList(c *gin.Context) {
	var (
		err               error
		getAllCategoryRequest *categoryProto.GetAllCategoryRequest
		getAllCategoryResp *categoryProto.GetAllCategoryResponse
	)
	app := mygin.Gin{C: c}

	getAllCategoryRequest = &categoryProto.GetAllCategoryRequest{}

	getAllCategoryResp, err =  h.categoryClient.GetAllCategory(c, getAllCategoryRequest)
	if err != nil {
		log.Error("调用【post_svc】【获取分类列表】失败 err:  ", err)
		app.ServerErrorResponse()
		return
	}
	if getAllCategoryResp.ErrorMsg != "" {
		app.LogicErrorResponse(getAllCategoryResp.ErrorMsg)
		return
	}

	categoryInfo := getAllCategoryResp.CategoryInfos
	app.Response(categoryInfo)
	return
}
