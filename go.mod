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
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.9.1
	github.com/minio/minio-go/v7 v7.0.1
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/wmsx/menger_svc v0.0.0-00010101000000-000000000000
	github.com/wmsx/pkg v0.0.0-20200710124640-b827730961c0
	github.com/wmsx/post_svc v0.0.0-00010101000000-000000000000
	github.com/wmsx/store_svc v0.0.0-00010101000000-000000000000
	github.com/wmsx/xconf v0.0.0-20200710193800-f97c7e3c9e84
	sigs.k8s.io/yaml v1.2.0 // indirect
)
