package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
)

type Storage interface {
	GetPresignedUrl(
		ctx context.Context, objectKey string, lifetimeSecs int64) (string, error)
}

type s3Storage struct {
	s3Client   *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

func NewStorage(ctx context.Context, cfg config.Storage) (Storage, error) {
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("garage"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")))
	if err != nil {
		return nil, fmt.Errorf("loading storage: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = &cfg.S3Api
		options.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(s3Client)

	return &s3Storage{
		s3Client:   s3Client,
		presigner:  presignClient,
		bucketName: cfg.S3Bucket,
	}, nil

}

func (s *s3Storage) GetPresignedUrl(
	ctx context.Context, objectKey string, lifetimeSecs int64) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return "", fmt.Errorf("creating presigned url for object %s: %w", objectKey, err)
	}
	return request.URL, nil
}
