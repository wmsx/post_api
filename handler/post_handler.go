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
		err              error
		getPostListRes   *postProto.GetPostListResponse
		getByStoreIdsRes *storeProto.GetByStoreIdsResponse
		getMengerListRes *mengerProto.GetMengerListResponse
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

	storeIds := make([]int64, 0)
	mengerIds := make([]int64, 0)
	mengerTemp := make(map[int64]struct{})
	for _, postInfo := range getPostListRes.PostInfos {
		if _, ok := mengerTemp[postInfo.MengerId]; !ok {
			mengerTemp[postInfo.MengerId] = struct{}{}
			mengerIds = append(mengerIds, postInfo.MengerId)
		}

		for _, item := range postInfo.Item {
			storeIds = append(storeIds, item.StoreId)
		}
	}

	getByStoreIdsRequest := &storeProto.GetByStoreIdsRequest{StoreIds: storeIds}
	if getByStoreIdsRes, err = h.storeClient.GetByStoreIds(c, getByStoreIdsRequest); err != nil {
		log.Error("【store svc】【GetByObjectIds】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeInfoMap := make(map[int64]*storeProto.StoreInfo)
	for _, storeInfo := range getByStoreIdsRes.StoreInfos {
		storeInfoMap[storeInfo.Id] = storeInfo
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
			storeInfo, _ := storeInfoMap[protoPostItem.StoreId]

			url, err := PresignedGetObject(c, storeInfo.BulkName, storeInfo.ObjectName, 10*time.Minute)
			if err != nil {
				log.Error("【minio】获取object访问连接失败 err: ", err)
				app.ServerErrorResponse()
				return
			}

			postItem := &PostItem{
				Url:  url,
				Type: protoPostItem.Type,
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
			StoreId: item.ObjectId,
			Index:   item.Index,
			Type:    getPostItemType(item.Filename),
		}
		protoPostItems = append(protoPostItems, protoPostItem)
	}

	filenames := make([]string, 0)
	for _, item := range createPostParam.PostItems {
		filenames = append(filenames, item.Filename)
	}

	var postType = getPostType(filenames)
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

func (h *PostHandler) GetMengerPostlist(c *gin.Context) {
	var (
		mengerId             int64
		err                  error
		getMengerPostListRes *postProto.GetMengerPostListResponse
		getByStoreIdsRes     *storeProto.GetByStoreIdsResponse
		getMengerRes         *mengerProto.GetMengerResponse
	)

	app := mygin.Gin{C: c}

	mengerId, err = strconv.ParseInt(c.GetHeader("uid"), 10, 64)
	if err != nil {
		log.Error("获取用户id失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	var pageQueryParam PageQueryParam
	if err = c.ShouldBindJSON(&pageQueryParam); err != nil {
		log.Error("分页参数错误 err: ", err)
		app.ServerErrorResponse()
		return
	}

	getMengerPostListRequest := &postProto.GetMengerPostListRequest{
		MengerId: mengerId,
		PageNum:  pageQueryParam.PageNum,
		PageSize: pageQueryParam.PageSize,
	}

	getMengerPostListRes, err = h.postClient.GetMengerPostList(c, getMengerPostListRequest)
	if err != nil {
		log.Error("【post_svc】【GetMengerPostList】 远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeIds := make([]int64, 0)
	for _, postInfo := range getMengerPostListRes.PostInfos {
		for _, item := range postInfo.Item {
			storeIds = append(storeIds, item.StoreId)
		}
	}

	getByStoreIdsRequest := &storeProto.GetByStoreIdsRequest{StoreIds: storeIds}
	if getByStoreIdsRes, err = h.storeClient.GetByStoreIds(c, getByStoreIdsRequest); err != nil {
		log.Error("【store_svc】【GetByObjectIds】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeInfoMap := make(map[int64]*storeProto.StoreInfo)
	for _, storeInfo := range getByStoreIdsRes.StoreInfos {
		storeInfoMap[storeInfo.Id] = storeInfo
	}

	getMengerRequest := &mengerProto.GetMengerRequest{
		MengerId: mengerId,
	}
	getMengerRes, err = h.mengerClient.GetMenger(c, getMengerRequest)
	if err != nil {
		log.Error("【menger_svc】【GetMenger】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	protoMengerInfo := getMengerRes.MengerInfo
	mengerInfo := &MengerInfo{
		Id:     protoMengerInfo.Id,
		Name:   protoMengerInfo.Name,
		Avatar: protoMengerInfo.Avatar,
	}

	postInfos := make([]*PostInfo, 0)
	for _, protoPostInfo := range getMengerPostListRes.PostInfos {
		protoPostItems := protoPostInfo.Item

		sort.SliceStable(protoPostItems, func(i, j int) bool {
			return protoPostItems[i].Index < protoPostItems[j].Index
		})

		postItems := make([]*PostItem, 0)
		for _, protoPostItem := range protoPostItems {
			storeInfo, _ := storeInfoMap[protoPostItem.StoreId]

			url, err := PresignedGetObject(c, storeInfo.BulkName, storeInfo.ObjectName, 10*time.Minute)
			if err != nil {
				log.Error("【minio】获取object访问连接失败 err: ", err)
				app.ServerErrorResponse()
				return
			}

			postItem := &PostItem{
				Url:  url,
				Type: protoPostItem.Type,
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

func (h *PostHandler) GetMengerFavoritePostlist(c *gin.Context) {
	var (
		mengerId               int64
		err                    error
		getByStoreIdsRes       *storeProto.GetByStoreIdsResponse
		getMengerRes           *mengerProto.GetMengerResponse
		getFavoritePostListRes *mengerProto.GetFavoritePostListResponse
		getPostByIdsRes        *postProto.GetPostByIdsResponse
	)

	app := mygin.Gin{C: c}

	mengerId, err = strconv.ParseInt(c.GetHeader("uid"), 10, 64)
	if err != nil {
		log.Error("获取用户id失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	var pageQueryParam PageQueryParam
	if err = c.ShouldBindJSON(&pageQueryParam); err != nil {
		log.Error("分页参数错误 err: ", err)
		app.ServerErrorResponse()
		return
	}

	getFavoritePostListRequest := &mengerProto.GetFavoritePostListRequest{
		PageNum:  pageQueryParam.PageNum,
		PageSize: pageQueryParam.PageSize,
		MengerId: mengerId,
	}

	getFavoritePostListRes, err = h.mengerClient.GetFavoritePostList(c, getFavoritePostListRequest)
	if err != nil {
		log.Error("【menger_svc】【GetFavoritePostList】 远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	getPostByIdsRequest := &postProto.GetPostByIdsRequest{
		Ids: getFavoritePostListRes.PostIds,
	}

	getPostByIdsRes, err = h.postClient.GetPostByIds(c, getPostByIdsRequest)
	if err != nil {
		log.Error("【post_svc】【GetPostByIds】 远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeIds := make([]int64, 0)
	for _, postInfo := range getPostByIdsRes.PostInfos {
		for _, item := range postInfo.Item {
			storeIds = append(storeIds, item.StoreId)
		}
	}

	getByStoreIdsRequest := &storeProto.GetByStoreIdsRequest{StoreIds: storeIds}
	if getByStoreIdsRes, err = h.storeClient.GetByStoreIds(c, getByStoreIdsRequest); err != nil {
		log.Error("【store svc】【GetByObjectIds】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeInfoMap := make(map[int64]*storeProto.StoreInfo)
	for _, storeInfo := range getByStoreIdsRes.StoreInfos {
		storeInfoMap[storeInfo.Id] = storeInfo
	}

	getMengerRequest := &mengerProto.GetMengerRequest{
		MengerId: mengerId,
	}
	getMengerRes, err = h.mengerClient.GetMenger(c, getMengerRequest)
	if err != nil {
		log.Error("【menger svc】【GetMenger】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	protoMengerInfo := getMengerRes.MengerInfo
	mengerInfo := &MengerInfo{
		Id:     protoMengerInfo.Id,
		Name:   protoMengerInfo.Name,
		Avatar: protoMengerInfo.Avatar,
	}

	postInfos := make([]*PostInfo, 0)
	for _, protoPostInfo := range getPostByIdsRes.PostInfos {
		protoPostItems := protoPostInfo.Item

		sort.SliceStable(protoPostItems, func(i, j int) bool {
			return protoPostItems[i].Index < protoPostItems[j].Index
		})

		postItems := make([]*PostItem, 0)
		for _, protoPostItem := range protoPostItems {
			storeInfo, _ := storeInfoMap[protoPostItem.StoreId]

			url, err := PresignedGetObject(c, storeInfo.BulkName, storeInfo.ObjectName, 10*time.Minute)
			if err != nil {
				log.Error("【minio】获取object访问连接失败 err: ", err)
				app.ServerErrorResponse()
				return
			}

			postItem := &PostItem{
				Url:  url,
				Type: protoPostItem.Type,
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

func (h *PostHandler) GetMengerThumbUpPostlist(c *gin.Context) {
	var (
		mengerId              int64
		err                   error
		getThumbUpPostListRes *mengerProto.GetThumbUpPostListResponse
		getByStoreIdsRes      *storeProto.GetByStoreIdsResponse
		getMengerRes          *mengerProto.GetMengerResponse
		getPostByIdsRes       *postProto.GetPostByIdsResponse
	)

	app := mygin.Gin{C: c}

	mengerId, err = strconv.ParseInt(c.GetHeader("uid"), 10, 64)
	if err != nil {
		log.Error("获取用户id失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	var pageQueryParam PageQueryParam
	if err = c.ShouldBindJSON(&pageQueryParam); err != nil {
		log.Error("分页参数错误 err: ", err)
		app.ServerErrorResponse()
		return
	}

	getThumbUpPostListRequest := &mengerProto.GetThumbUpPostListRequest{
		MengerId: mengerId,
		PageNum:  pageQueryParam.PageNum,
		PageSize: pageQueryParam.PageSize,
	}

	getThumbUpPostListRes, err = h.mengerClient.GetThumbUpPostList(c, getThumbUpPostListRequest)
	if err != nil {
		log.Error("【menger_svc】【GetThumbUpPostList】 远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	getPostByIdsRequest := &postProto.GetPostByIdsRequest{
		Ids: getThumbUpPostListRes.PostIds,
	}

	getPostByIdsRes, err = h.postClient.GetPostByIds(c, getPostByIdsRequest)
	if err != nil {
		log.Error("【post_svc】【GetPostByIds】 远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeIds := make([]int64, 0)
	for _, postInfo := range getPostByIdsRes.PostInfos {
		for _, item := range postInfo.Item {
			storeIds = append(storeIds, item.StoreId)
		}
	}

	getByStoreIdsRequest := &storeProto.GetByStoreIdsRequest{StoreIds: storeIds}
	if getByStoreIdsRes, err = h.storeClient.GetByStoreIds(c, getByStoreIdsRequest); err != nil {
		log.Error("【store svc】【GetByObjectIds】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	storeInfoMap := make(map[int64]*storeProto.StoreInfo)
	for _, storeInfo := range getByStoreIdsRes.StoreInfos {
		storeInfoMap[storeInfo.Id] = storeInfo
	}

	getMengerRequest := &mengerProto.GetMengerRequest{
		MengerId: mengerId,
	}
	getMengerRes, err = h.mengerClient.GetMenger(c, getMengerRequest)
	if err != nil {
		log.Error("【menger svc】【GetMenger】远程调用失败 err: ", err)
		app.ServerErrorResponse()
		return
	}

	protoMengerInfo := getMengerRes.MengerInfo
	mengerInfo := &MengerInfo{
		Id:     protoMengerInfo.Id,
		Name:   protoMengerInfo.Name,
		Avatar: protoMengerInfo.Avatar,
	}

	postInfos := make([]*PostInfo, 0)
	for _, protoPostInfo := range getPostByIdsRes.PostInfos {
		protoPostItems := protoPostInfo.Item

		sort.SliceStable(protoPostItems, func(i, j int) bool {
			return protoPostItems[i].Index < protoPostItems[j].Index
		})

		postItems := make([]*PostItem, 0)
		for _, protoPostItem := range protoPostItems {
			storeInfo, _ := storeInfoMap[protoPostItem.StoreId]

			url, err := PresignedGetObject(c, storeInfo.BulkName, storeInfo.ObjectName, 10*time.Minute)
			if err != nil {
				log.Error("【minio】获取object访问连接失败 err: ", err)
				app.ServerErrorResponse()
				return
			}

			postItem := &PostItem{
				Url:  url,
				Type: protoPostItem.Type,
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

func getPostItemType(filename string) int32 {
	ext := strings.ToLower(path.Ext(filename))
	if imageTypeSet.Contains(ext) {
		return PostItemImage
	}
	if videoTypeSet.Contains(ext) {
		return PostItemVideo
	}
	return PostItemUnknown
}

func getPostType(filenames []string) int32 {
	if len(filenames) > 1 {
		return PostSidecar
	} else {
		filename := filenames[0]
		ext := strings.ToLower(path.Ext(filename))
		if imageTypeSet.Contains(ext) {
			return PostImage
		}
		if videoTypeSet.Contains(ext) {
			return PostVideo
		}
	}
	return PostUnknown
}
