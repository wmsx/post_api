module github.com/wmsx/post_api

go 1.14

replace (
	github.com/wmsx/menger_svc => /Users/zengqiang96/codespace/sx/menger_svc
	github.com/wmsx/pkg => /Users/zengqiang96/codespace/sx/pkg
	github.com/wmsx/post_svc => /Users/zengqiang96/codespace/sx/post_svc
	github.com/wmsx/store_svc => /Users/zengqiang96/codespace/sx/store_svc
	google.golang.org/grpc => google.golang.org/grpc v1.26.0

)

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/micro/go-micro/v2 v2.9.1
	github.com/minio/minio-go/v6 v6.0.57
	github.com/wmsx/menger_svc v0.0.0-00010101000000-000000000000
	github.com/wmsx/pkg v0.0.0-20200710124640-b827730961c0
	github.com/wmsx/post_svc v0.0.0-00010101000000-000000000000
	github.com/wmsx/store_api v0.0.0-20200716042549-4215a3f76a87
	github.com/wmsx/store_svc v0.0.0-00010101000000-000000000000
)
