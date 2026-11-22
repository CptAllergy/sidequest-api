package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/cptallergy/sidequest-api/internal/lib/config"
)

// TODO media cleanup strategy: upload new quest images to /temp. After quest/entry is created move it over to final location
// TODO every 24 hours run a cleanup job that deletes all files in /temp older than 24 hours. This will ensure that we don't have orphaned files taking up space in S3.

type Storage interface {
	PresignGetUrl(
		ctx context.Context, objectKey string, lifetimeSecs int64) (string, error)
	PresignPutUrl(
		ctx context.Context,
		objectKey string,
		contentType string,
		metadata map[string]string,
		lifetimeSecs int64) (*v4.PresignedHTTPRequest, error)
	HeadObject(ctx context.Context, objectKey string) (ObjectInfo, error)
	CopyObject(ctx context.Context, objectKey string, objectDest string) error
	DeleteObject(ctx context.Context, objectKey string) error
	DeleteObjectsWithPrefix(ctx context.Context, prefix string) error
	ReadObjectContentType(ctx context.Context, objectKey string) (string, error)
}

type s3Storage struct {
	s3Client   *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

type ObjectInfo struct {
	Key         string
	ContentType string
	Size        int64
	Metadata    map[string]string
}

var ErrObjectNotFound = errors.New("object not found")

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

	// External API used for the presigner
	s3ExternalClient := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = &cfg.S3ExternalApi
		options.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(s3ExternalClient)

	return &s3Storage{
		s3Client:   s3Client,
		presigner:  presignClient,
		bucketName: cfg.S3Bucket,
	}, nil

}

func (s *s3Storage) PresignGetUrl(
	ctx context.Context, objectKey string, lifetimeSecs int64) (string, error) {
	request, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return "", fmt.Errorf("creating presigned get url for object %s: %w", objectKey, err)
	}
	return request.URL, nil
}

func (s *s3Storage) PresignPutUrl(
	ctx context.Context, objectKey string, contentType string, metadata map[string]string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return nil, fmt.Errorf("creating presigned put url for object %s: %w", objectKey, err)
	}
	return request, nil
}

func (s *s3Storage) HeadObject(ctx context.Context, objectKey string) (ObjectInfo, error) {
	out, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		var respErr *awshttp.ResponseError
		if errors.As(err, &respErr) && respErr.HTTPStatusCode() == http.StatusNotFound {
			return ObjectInfo{}, ErrObjectNotFound
		}
		return ObjectInfo{}, fmt.Errorf("head object %s: %w", objectKey, err)
	}

	return ObjectInfo{
		Key:         objectKey,
		ContentType: aws.ToString(out.ContentType),
		Size:        aws.ToInt64(out.ContentLength),
		Metadata:    out.Metadata,
	}, nil
}

func (s *s3Storage) CopyObject(ctx context.Context, objectKey string, objectDest string) error {
	_, err := s.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		CopySource: aws.String(fmt.Sprintf("%v/%v", s.bucketName, objectKey)),
		Key:        aws.String(objectDest),
	})

	if err != nil {
		return fmt.Errorf("copying object %s->%s: %w", objectKey, objectDest, err)
	}
	return nil
}

func (s *s3Storage) DeleteObject(ctx context.Context, objectKey string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	}
	_, err := s.s3Client.DeleteObject(ctx, input)
	if err != nil {
		if noKey, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return fmt.Errorf("object %s does not exist in %s: %w", objectKey, s.bucketName, noKey)
		} else if apiErr, ok := errors.AsType[*smithy.GenericAPIError](err); ok {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				return fmt.Errorf("access denied: cannot delete object %s from %s: %w", objectKey, s.bucketName, apiErr)
			}
		}
		return err
	}

	err = s3.NewObjectNotExistsWaiter(s.s3Client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucketName), Key: aws.String(objectKey)}, time.Minute)
	if err != nil {
		return fmt.Errorf("attempting to wait for object %s in bucket %s to be deleted: %w", objectKey, s.bucketName, err)
	}

	return nil
}

func (s *s3Storage) DeleteObjectsWithPrefix(ctx context.Context, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(s.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing objects with prefix %s: %w", prefix, err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
		}

		out, err := s.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("deleting objects with prefix %s: %w", prefix, err)
		}

		if len(out.Errors) > 0 {
			first := out.Errors[0]
			return fmt.Errorf(
				"deleting %d object(s) with prefix %s failed (first key %q: %s): %s",
				len(out.Errors), prefix, aws.ToString(first.Key), aws.ToString(first.Code), aws.ToString(first.Message),
			)
		}
	}

	return nil
}

func (s *s3Storage) ReadObjectContentType(ctx context.Context, objectKey string) (string, error) {
	out, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
		Range:  aws.String("bytes=0-511"), // Read the first 512 bytes to detect content type
	})

	if err != nil {
		return "", fmt.Errorf("get object content: %w", err)
	}
	defer func() { _ = out.Body.Close() }()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("read object content: %w", err)
	}

	return http.DetectContentType(data), nil
}
