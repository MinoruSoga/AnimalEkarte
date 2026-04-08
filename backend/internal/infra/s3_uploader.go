package infra

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Uploader は AWS S3 にファイルをアップロードする FileUploader 実装。
type S3Uploader struct {
	client *s3.Client
	bucket string
	region string
}

// NewS3Uploader は S3Uploader を生成する。
// bucket と region は環境変数から受け取る。
func NewS3Uploader(ctx context.Context, bucket, region string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return &S3Uploader{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		region: region,
	}, nil
}

// Upload はファイルを S3 にアップロードし、公開 URL を返す。
func (u *S3Uploader) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload failed: %w", err)
	}
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucket, u.region, key)
	return url, nil
}

// Delete は S3 からオブジェクトを削除する。
func (u *S3Uploader) Delete(ctx context.Context, key string) error {
	_, err := u.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete failed: %w", err)
	}
	return nil
}
