package handler

import (
	"github.com/deckarep/golang-set"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/util/log"
	mengerProto "github.com/wmsx/menger_svc/proto/menger"
	mygin "github.com/wmsx/pkg/gin"
	postProto "github.com/wmsx/post_svc/proto/post"
	storeProto "github.com/wmsx/store_svc/proto/store"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PostType
const (
	PostSidecar = 0
	PostImage   = 1
	PostVideo   = 2
	PostUnknown = 999
)

// PostItem类型
const (
	PostItemImage   = 1
	PostItemVideo   = 2
	PostItemUnknown = 2
)

var (
	imageTypeSet = mapset.NewSet()
	videoTypeSet = mapset.NewSet()
)

func init() {
	imageTypeSet.Add(".jpg")
	imageTypeSet.Add(".jpeg")
	imageTypeSet.Add(".gif")
	imageTypeSet.Add(".png")

	videoTypeSet.Add(".mp4")
	videoTypeSet.Add(".mkv")
}

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
func (h *PostHandler) GetPostList(c *gin.Context) {
	var (
		err               error
		getPostListRes    *postProto.GetPostListResponse
		getByObjectIdsRes *storeProto.GetByObjectIdsResponse
		getMengerListRes  *mengerProto.GetMengerListResponse
	)
	app := mygin.Gin{C: c}

	var postListParam GetPostListParam
	if err = c.ShouldBindJSON(&postListParam); err != nil {
		log.Error("参数解析错误 err: ", err)
		app.LogicErrorResponse("参数错误")
		return
	}

	getPostListReq := &postProto.GetPostListRequest{
		CategoryId: postListParam.CategoryId,
		LastId:     postListParam.LastId,
	}
	if getPostListRes, err = h.postClient.GetPostList(c, getPostListReq); err != nil {
		log.Error("【post svc】【GetPostList】远程调用失败 err: ", err)
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
	if getByObjectIdsRes, err = h.storeClient.GetByObjectIds(c, getByObjectIdsRequest); err != nil {
		log.Error("【store svc】【GetByObjectIds】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	objectInfoMap := make(map[int64]*storeProto.ObjectInfo)
	for _, objectInfo := range getByObjectIdsRes.ObjectInfos {
		objectInfoMap[objectInfo.Id] = objectInfo
	}

	getMengerListRequest := &mengerProto.GetMengerListRequest{MengerIds: mengerIds}
	if getMengerListRes, err = h.mengerClient.GetMengerList(c, getMengerListRequest); err != nil {
		log.Error("【menger svc】【GetMengerList】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	mengerInfoMap := make(map[int64]*MengerInfo)
	for _, protoMengerInfo := range getMengerListRes.MengerInfos {
		mengerInfoMap[protoMengerInfo.Id] = &MengerInfo{
			Id:     protoMengerInfo.Id,
			Name:   protoMengerInfo.Name,
			Email:  protoMengerInfo.Email,
			Avatar: protoMengerInfo.Avatar,
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

			url, err := PresignedGetObject(c, objectInfo.Bulk, objectInfo.ObjectName, 10*time.Minute)
			if err != nil {
				log.Error("【minio】获取object访问连接失败 err: ", err)
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
			Items:       postItems,
			CreateAt:    protoPostInfo.CreateAt,
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
		createPostRes   *postProto.CreatePostResponse
	)
	app := mygin.Gin{C: c}
	if err = c.ShouldBindJSON(&createPostParam); err != nil {
		log.Errorf("参数解析错误 param: %v, err: %v", createPostParam, err)
		app.LogicErrorResponse("参数错误")
		return
	}

	mengerId, err := strconv.ParseInt(c.GetHeader("uid"), 10, 64)
	if err != nil {
		log.Error("获取用户id失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	protoPostItems := make([]*postProto.PostItem, 0)
	for _, item := range createPostParam.PostItems {
		protoPostItem := &postProto.PostItem{
			ObjectId: item.ObjectId,
			Index:    item.Index,
			Type:     getPostItemType(item),
		}
		protoPostItems = append(protoPostItems, protoPostItem)
	}

	var postType = getPostType(createPostParam.PostItems)
	savePostRequest := &postProto.CreatePostRequest{
		Type:        postType,
		Title:       createPostParam.Title,
		Description: createPostParam.Description,
		MengerId:    mengerId,
		CategoryId:  createPostParam.CategoryId,
		Items:       protoPostItems,
	}

	if createPostRes, err = h.postClient.CreatePost(c, savePostRequest); err != nil {
		log.Error("调用postSvc失败 err: ", err)
		app.ServerErrorResponse()
		return
	}
	if createPostRes.ErrorMsg != "" {
		app.LogicErrorResponse(createPostRes.ErrorMsg)
		return
	}
	app.Response(nil)
	return
}

func getPostItemType(item *CreatePostItemParam) int32 {
	ext := strings.ToLower(path.Ext(item.Filename))
	if imageTypeSet.Contains(ext) {
		return PostItemImage
	}
	if videoTypeSet.Contains(ext) {
		return PostItemVideo
	}
	return PostItemUnknown
}

func getPostType(postItems []*CreatePostItemParam) int32 {
	if len(postItems) > 1 {
		return PostSidecar
	} else {
		postItem := postItems[0]
		ext := strings.ToLower(path.Ext(postItem.Filename))
		if imageTypeSet.Contains(ext) {
			return PostImage
		}
		if videoTypeSet.Contains(ext) {
			return PostVideo
		}
	}
	return PostUnknown
}
