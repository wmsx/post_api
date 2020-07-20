package handler

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wmsx/post_api/setting"
	"net/url"
	"time"
)

var (
	minioClient *minio.Client
)

func setUpMinio(minioSetting *setting.Minio) (err error) {
	if minioClient, err = minio.New(
		minioSetting.Endpoint,
		&minio.Options{
			Creds:        credentials.NewStaticV4(minioSetting.AccessKey,		minioSetting.SecretAccessKey, ""),
			Secure:       false,
		}); err != nil {
		return
	}
	return nil
}


func PresignedGetObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
