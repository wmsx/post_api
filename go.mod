module github.com/wmsx/post_api

go 1.14

require (
	github.com/deckarep/golang-set v1.7.1
	github.com/gin-gonic/gin v1.6.3
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.5.0
	github.com/minio/minio-go/v7 v7.0.1
	github.com/wmsx/menger_svc v0.0.0-20200721170201-dd47044981ba
	github.com/wmsx/pkg v0.0.0-20200722160831-4cb77a04c806
	github.com/wmsx/post_svc v0.0.0-20200722121416-f735c9f55907
	github.com/wmsx/store_svc v0.0.0-20200721165626-2634000225dd
	github.com/wmsx/xconf v0.0.0-20200721142237-75926266fd1a
)

// 本地测试使用
//github.com/wmsx/post_svc => /Users/zengqiang96/codespace/sx/post_svc
//github.com/wmsx/store_svc => /Users/zengqiang96/codespace/sx/store_svc

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
