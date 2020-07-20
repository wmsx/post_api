package handler

import (
	"github.com/micro/go-micro/v2/util/log"
	"github.com/wmsx/post_api/setting"
)

const (
	storeSvcName  = "wm.sx.svc.store"
	postSvcName   = "wm.sx.svc.post"
	mengerSvcName = "wm.sx.svc.menger"
)


func SetUp() (err error) {
	if err = setUpMinio(&setting.MinIOSetting); err != nil {
		log.Error("初始化minio失败 err: ", err)
		return err
	}
	return nil
}