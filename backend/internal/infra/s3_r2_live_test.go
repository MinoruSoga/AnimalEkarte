package infra

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TestR2S3Live は Cloudflare R2 の S3 互換 API 経路（P2-3）を検証する。
// 実行: R2_LIVE_TEST=1 と R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY / S3_ENDPOINT /
// S3_SHARED_BUCKET（または R2_BUCKET_NAME）を環境変数で供給する。
func TestR2S3Live(t *testing.T) {
	if os.Getenv("R2_LIVE_TEST") != "1" {
		t.Skip("R2_LIVE_TEST=1 が未設定のためスキップ")
	}

	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("S3_SHARED_BUCKET")
	if bucket == "" {
		bucket = os.Getenv("R2_BUCKET_NAME")
	}
	if accessKey == "" || secretKey == "" || endpoint == "" || bucket == "" {
		t.Fatal("R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, S3_ENDPOINT, S3_SHARED_BUCKET(or R2_BUCKET_NAME) が必要")
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		t.Fatalf("LoadDefaultConfig: %v", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		applyS3EndpointOverride(o, endpoint)
	})
	storage := &S3FileStorage{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  bucket,
	}

	key := "_smoke/p2-3-live-test.txt"
	body := strings.NewReader("p2-3-r2-s3-live")

	if err := storage.Upload(ctx, key, body, "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	t.Cleanup(func() {
		_ = storage.Delete(context.Background(), key)
	})

	signedURL, err := storage.GetSignedURL(ctx, key, 5*time.Minute)
	if err != nil {
		t.Fatalf("GetSignedURL: %v", err)
	}

	resp, err := http.Get(signedURL)
	if err != nil {
		t.Fatalf("GET presigned URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET presigned URL status = %d, want 200", resp.StatusCode)
	}
	gotBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read presigned response: %v", err)
	}
	if string(gotBody) != "p2-3-r2-s3-live" {
		t.Fatalf("presigned body = %q, want %q", gotBody, "p2-3-r2-s3-live")
	}

	if err := storage.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Fatal("HeadObject after delete: expected error, got nil")
	}
}
