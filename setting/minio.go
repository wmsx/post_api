package setting

import (
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/wmsx/xconf/pkg/client/source"
)

const MinIONamespace = "minio"

var (
	MinIOSetting = Minio{}
)

type Minio struct {
	Endpoint        string `json:"endpoint"`
	AccessKey       string `json:"access_key"`
	SecretAccessKey string `json:"secret_access_key"`
}

func setUpMinio(appName, env string) error {
	sourceUrl := XConfURL
	if env == "dev" {
		sourceUrl = DevXConfURL
	}

	s := source.NewSource(appName, env, MinIONamespace, source.WithURL(sourceUrl))

	minioConfig, err := config.NewConfig(
		config.WithSource(s),
	)
	if err != nil {
		log.Error("获取minio配置失败")
		return err
	}

	err = minioConfig.Scan(&MinIOSetting)
	if err != nil {
		log.Error("获取minio配置失败")
		return err
	}

	log.Info("初始化minio配置: ", MinIOSetting)

	minioWatcher, err := minioConfig.Watch()
	if err != nil {
		log.Error("minio配置watch失败")
		return err
	}

	go func() {
		for {
			// 会比较 value，内容不变不会返回
			v, err := minioWatcher.Next()
			if err != nil {
				log.Error("minio配置变更获取失败")
			} else {
				if err := v.Scan(&MinIOSetting); err != nil {
					log.Error("watch获取minio配置失败")
				} else {
					log.Info("watch获取minio配置结果: ", MinIOSetting)
				}
			}
		}
	}()
	return nil
}
