package handler

import (
	"github.com/minio/minio-go/v6"
	"net/url"
	"time"
)

var (
	minioClient *minio.Client
)

func PresignedGetObject(bucketName string, objectName string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := minioClient.PresignedGetObject(bucketName, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
