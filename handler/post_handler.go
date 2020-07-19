package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	mengerProto "github.com/wmsx/menger_svc/proto/menger"
	mygin "github.com/wmsx/pkg/gin"
	postProto "github.com/wmsx/post_svc/proto/post"
	storeProto "github.com/wmsx/store_svc/proto/store"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type PostHandler struct {
	postClient   postProto.PostService
	storeClient  storeProto.StoreService
	mengerClient mengerProto.MengerService
}

func NewPostHandler(c client.Client) *PostHandler {
	return &PostHandler{
		postClient:   postProto.NewPostService(postSvcName, c),
		storeClient:  storeProto.NewStoreService(storeSvcName, c),
		mengerClient: mengerProto.NewMengerService(mengerSvcName, c),
	}
}

// 获取post列表内容
func (s *PostHandler) GetPostList(c *gin.Context) {
	var (
		err               error
		getPostListRes    *postProto.GetPostListResponse
		getByObjectIdsRes *storeProto.GetByObjectIdsResponse
		getMengerListRes  *mengerProto.GetMengerListResponse
	)
	app := mygin.Gin{C: c}

	var categoryListParam CategoryListParam
	if err = c.ShouldBindQuery(&categoryListParam); err != nil {
		app.LogicErrorResponse("参数错误")
		return
	}

	getPostListReq := &postProto.GetPostListRequest{
		CategoryId: categoryListParam.CategoryId,
		LastId:     categoryListParam.LastId,
	}
	if getPostListRes, err = s.postClient.GetPostList(c, getPostListReq); err != nil {
		app.ServerErrorResponse()
		return
	}

	objectIds := make([]int64, 0)
	mengerIds := make([]int64, 0)
	mengerTemp := make(map[int64]struct{})
	for _, postInfo := range getPostListRes.PostInfos {
		if _, ok := mengerTemp[postInfo.MengerId]; !ok {
			mengerTemp[postInfo.MengerId] = struct{}{}
			mengerIds = append(mengerIds, postInfo.MengerId)
		}

		for _, item := range postInfo.Item {
			objectIds = append(objectIds, item.ObjectId)
		}
	}

	getByObjectIdsRequest := &storeProto.GetByObjectIdsRequest{ObjectIds: objectIds}
	if getByObjectIdsRes, err = s.storeClient.GetByObjectIds(c, getByObjectIdsRequest); err != nil {
		c.String(http.StatusInternalServerError, "服务器异常")
		return
	}

	objectInfoMap := make(map[int64]*storeProto.ObjectInfo)
	for _, objectInfo := range getByObjectIdsRes.ObjectInfos {
		objectInfoMap[objectInfo.Id] = objectInfo
	}

	getMengerListRequest := &mengerProto.GetMengerListRequest{MengerIds: mengerIds}
	if getMengerListRes, err = s.mengerClient.GetMengerList(c, getMengerListRequest); err != nil {
		app.ServerErrorResponse()
		return
	}

	mengerInfoMap := make(map[int64]*MengerInfo)
	for _, protoMengerInfo := range getMengerListRes.MengerInfos {
		mengerInfoMap[protoMengerInfo.MengerId] = &MengerInfo{
			MengerId: protoMengerInfo.MengerId,
			Name:     protoMengerInfo.Name,
			Email:    protoMengerInfo.Email,
			Avatar:   protoMengerInfo.Avatar,
		}
	}

	postInfos := make([]*PostInfo, 0)
	for _, protoPostInfo := range getPostListRes.PostInfos {
		mengerInfo, _ := mengerInfoMap[protoPostInfo.MengerId]
		protoPostItems := protoPostInfo.Item

		sort.SliceStable(protoPostItems, func(i, j int) bool {
			return protoPostItems[i].Index < protoPostItems[j].Index
		})

		postItems := make([]*PostItem, 0)
		for _, protoPostItem := range protoPostItems {
			objectInfo, _ := objectInfoMap[protoPostItem.ObjectId]

			url, err := PresignedGetObject(objectInfo.Bulk, objectInfo.ObjectName, 10*time.Minute)
			if err != nil {
				app.ServerErrorResponse()
				return
			}

			postItem := &PostItem{
				Url: url,
			}
			postItems = append(postItems, postItem)
		}

		postInfo := &PostInfo{
			Id:          protoPostInfo.Id,
			Type:        protoPostInfo.Type,
			Title:       protoPostInfo.Title,
			Description: protoPostInfo.Description,
			MengerInfo:  mengerInfo,
			Item:        postItems,
		}
		postInfos = append(postInfos, postInfo)
	}
	app.Response(postInfos)
	return
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var (
		createPostParam CreatePostParam
		err             error
		savePostRes     *postProto.SavePostResponse
	)
	app := mygin.Gin{C: c}
	if err = c.ShouldBindJSON(createPostParam); err != nil {
		app.LogicErrorResponse("参数错误")
		return
	}

	mengerId, err := strconv.ParseInt(c.GetHeader("id"), 10, 64)
	if err != nil {
		app.ServerErrorResponse()
		return
	}

	protoPostItems := make([]*postProto.PostItem, 0)
	for _, item := range createPostParam.PostItems {
		protoPostItem := &postProto.PostItem{
			ObjectId: item.ObjectId,
			Index:    item.Index,
		}
		protoPostItems = append(protoPostItems, protoPostItem)
	}

	savePostRequest := &postProto.SavePostRequest{
		Type:        0,
		Title:       createPostParam.Title,
		Description: createPostParam.Description,
		MengerId:    mengerId,
		CategoryId:  createPostParam.CategoryId,
		Items:       protoPostItems,
	}

	if savePostRes, err = h.postClient.SavePost(c, savePostRequest); err != nil {
		app.ServerErrorResponse()
		return
	}
	if savePostRes.ErrorMsg != nil {
		app.LogicErrorResponse(savePostRes.ErrorMsg.Msg)
		return
	}
	app.Response(nil)
	return
}
