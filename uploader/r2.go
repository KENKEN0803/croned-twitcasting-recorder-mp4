package uploader

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config" // AWS SDK config
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/jzhang046/croned-twitcasting-recorder-mp4/config" // Our app's config
)

type Uploader interface {
	Upload(filePath, remotePath string) error
}

type R2Uploader struct {
	client *s3.Client
	bucket string
}

func NewR2Uploader(cfg *appconfig.R2Config) (*R2Uploader, error) { // Use appconfig.R2Config
	if cfg == nil || !cfg.Enabled {
		return nil, errors.New("R2 configuration is not enabled")
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.Endpoint,
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	return &R2Uploader{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (u *R2Uploader) Upload(filePath, remotePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Printf("Start uploading %s to r2://%s%s", filePath, u.bucket, remotePath)

	_, err = u.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(remotePath),
		Body:   file,
	})

	if err != nil {
		log.Printf("Failed to upload %s to R2: %v", filepath.Base(filePath), err)
		return err
	}

	log.Printf("Completed uploading %s", filepath.Base(filePath))
	return nil
}
