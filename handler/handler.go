package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"
	mengerProto "github.com/wmsx/menger_svc/proto/menger"
	"github.com/wmsx/post_api/handler/minio"
	proto "github.com/wmsx/post_svc/proto/post"
	storeProto "github.com/wmsx/store_svc/proto/store"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const (
	storeSvcName  = "wm.sx.svc.store"
	postSvcName   = "wm.sx.svc.post"
	mengerSvcName = "wm.sx.svc.menger"
)

type PostHandler struct {
	postClient   proto.PostService
	storeClient  storeProto.StoreService
	mengerClient mengerProto.MengerService
}

func NewPostHandler(c client.Client) *PostHandler {
	return &PostHandler{
		postClient:   proto.NewPostService(postSvcName, c),
		storeClient:  storeProto.NewStoreService(storeSvcName, c),
		mengerClient: mengerProto.NewMengerService(mengerSvcName, c),
	}
}

func (s *PostHandler) GetPostList(c *gin.Context) {
	var (
		err               error
		categoryId        int64
		lastId            int64
		getPostListRes    *proto.GetPostListResponse
		getByObjectIdsRes *storeProto.GetByObjectIdsResponse
		getMengerListRes  *mengerProto.GetMengerListResponse
	)
	if categoryId, err = strconv.ParseInt(c.Query("categoryId"), 10, 64); err != nil {
		c.String(http.StatusBadRequest, "参数错误")
		return
	}
	if lastId, err = strconv.ParseInt(c.Query("lastId"), 10, 64); err != nil {
		c.String(http.StatusBadRequest, "参数错误")
		return
	}
	getPostListReq := &proto.GetPostListRequest{
		CategoryId: categoryId,
		LastId:     lastId,
	}
	if getPostListRes, err = s.postClient.GetPostList(c, getPostListReq); err != nil {
		c.String(http.StatusInternalServerError, "服务器异常")
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
		c.String(http.StatusInternalServerError, "服务器异常")
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

			url, err := minio.PresignedGetObject(objectInfo.Bulk, objectInfo.ObjectName, 10*time.Minute)
			if err != nil {
				c.String(http.StatusInternalServerError, "服务器异常")
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
	c.JSON(http.StatusOK, postInfos)
	return
}
