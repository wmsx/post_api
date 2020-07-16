package main

import (
	log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/web"
	"github.com/wmsx/post_api/routers"
)

const name = "wm.sx.web.post"

func main() {
	svc := web.NewService(
		web.Name(name),
	)
	if err := svc.Init(); err != nil {
		log.Fatal("初始化失败", err)
	}

	router := routers.InitRouter(svc.Options().Service.Client())
	svc.Handle("/", router)

	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}
