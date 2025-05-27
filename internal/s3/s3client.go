package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3Client *s3.Client
}

type FileInfo struct {
	Key          string
	Size         int64
	S3Path       string
	LastModified time.Time
}

type FileReader interface {
	GetReader(ctx context.Context, path string) (io.ReadCloser, error)
}

type S3FileReader struct {
	client *Client
}

func NewClient(s3Region string) (*Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		s3Client: s3.NewFromConfig(awsCfg),
	}, nil
}

func (c *Client) ListFiles(ctx context.Context, bucket, prefix string) ([]FileInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := c.s3Client.ListObjectsV2(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("faild to list objects in bucket %s: %w", bucket, err)
	}

	files := make([]FileInfo, 0, len(result.Contents))

	for _, obj := range result.Contents {
		files = append(files, FileInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			S3Path:       fmt.Sprintf("s3://%s/%s", bucket, *obj.Key),
			LastModified: *obj.LastModified,
		})
	}

	return files, nil
}

func NewS3FileReader(client *Client) FileReader {
	return &S3FileReader{client: client}
}

func (s *S3FileReader) GetReader(ctx context.Context, s3Path string) (io.ReadCloser, error) {
	if !strings.HasPrefix(s3Path, "s3://") {
		return nil, fmt.Errorf("invalid s3 path (missing s3:// prefix): %s", s3Path)
	}

	trimmed := strings.TrimPrefix(s3Path, "s3://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid s3 path format: %s", s3Path)
	}
	bucket, key := parts[0], parts[1]

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from bucket %s: %w, ", key, bucket, err)
	}

	return result.Body, nil
}
