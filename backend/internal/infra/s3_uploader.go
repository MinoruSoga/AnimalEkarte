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
	client        *s3.Client
	bucket        string
	region        string
	endpoint      string
	publicBaseURL string
}

// NewS3Uploader は S3Uploader を生成する。
// bucket と region は環境変数から受け取る。
// endpoint は S3 API 接続先（空文字=AWS 既定、非空=Cloudflare R2 等の S3 互換ストレージ、
// この場合 path-style でアクセスする）。
// publicBaseURL（env: S3_PUBLIC_BASE_URL）はブラウザ向けオブジェクト公開 URL の base。
// S3 API endpoint とは別ホストであり、設定時はアップロード後 URL の組み立てに優先使用する。
func NewS3Uploader(ctx context.Context, bucket, region, endpoint, publicBaseURL string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		applyS3EndpointOverride(o, endpoint)
	})
	return &S3Uploader{
		client:        client,
		bucket:        bucket,
		region:        region,
		endpoint:      endpoint,
		publicBaseURL: publicBaseURL,
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
	return buildS3ObjectURL(u.publicBaseURL, u.endpoint, u.bucket, u.region, key), nil
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
