package main

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/micro/go-micro/v2/web"
	mygin "github.com/wmsx/pkg/gin"
	"github.com/wmsx/post_api/handler"
	"github.com/wmsx/post_api/routers"
	"github.com/wmsx/post_api/setting"
)

const name = "wm.sx.web.post"

func main() {
	svc := web.NewService(
		web.Name(name),
		web.Flags(
			&cli.StringFlag{
				Name:    "env",
				Usage:   "指定运行环境",
				Value:   "dev",
				EnvVars: []string{"RUN_ENV"},
			},
		),
	)

	var env string
	if err := svc.Init(
		web.Action(func(c *cli.Context) {
			env = c.String("env")
		}),
		web.BeforeStart(func() (err error) {
			if err = setting.SetUp(name, env); err != nil {
				return err
			}
			if err = mygin.SetUp(setting.RedisSetting.Addr, setting.RedisSetting.Password); err != nil {
				return err
			}
			if err = handler.SetUp(); err != nil {
				return err
			}
			return nil
		}),
	); err != nil {
		log.Fatal("初始化失败", err)
	}

	router := routers.InitRouter(svc.Options().Service.Client())
	svc.Handle("/", router)

	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}

}
