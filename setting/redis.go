package setting

import (
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/wmsx/xconf/pkg/client/source"
)

const redisNamespace = "redis"

var (
	RedisSetting = Redis{}
)

type Redis struct {
	Addr string `json:"addr"`
	Password string `json:"password"`
}


func setUpRedis(appName, env string) error {
	sourceUrl := XConfURL
	if env == "dev" {
		sourceUrl = DevXConfURL
	}

	s := source.NewSource(appName, env, redisNamespace, source.WithURL(sourceUrl))

	redisConfig, err := config.NewConfig(
		config.WithSource(s),
	)
	if err != nil {
		log.Error("获取redis配置失败")
		return err
	}

	err = redisConfig.Scan(&RedisSetting)
	if err != nil {
		log.Error("获取redis配置失败")
		return err
	}

	log.Info("初始化redis配置: ", RedisSetting)

	redisWatcher, err := redisConfig.Watch()
	if err != nil {
		log.Error("redis配置watch失败")
		return err
	}

	go func() {
		for {
			// 会比较 value，内容不变不会返回
			v, err := redisWatcher.Next()
			if err != nil {
				log.Error("redis配置变更获取失败")
			} else {
				if err := v.Scan(&MinIOSetting); err != nil {
					log.Error("watch获取redis配置失败")
				} else {
					log.Info("watch获取redis配置结果: ", RedisSetting)
				}
			}
		}
	}()
	return nil
}
