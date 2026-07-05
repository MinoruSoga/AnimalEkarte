package infra

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// applyS3EndpointOverride は S3_ENDPOINT が指定されている場合、カスタムエンドポイント
// （Cloudflare R2 等の S3 互換ストレージ）へ path-style でアクセスするよう s3.Options を上書きする。
// 空文字の場合は AWS S3 の既定挙動（リージョナルエンドポイント・バーチャルホスト形式）を変更しない。
func applyS3EndpointOverride(o *s3.Options, endpoint string) {
	if endpoint == "" {
		return
	}
	o.BaseEndpoint = aws.String(endpoint)
	o.UsePathStyle = true
}

// buildS3ObjectURL はアップロード後の公開 URL を組み立てる。
// endpoint が空の場合は AWS S3 のバーチャルホスト形式 URL を返し（既存挙動を維持）、
// endpoint が指定されている場合は path-style（R2 等の S3 互換ストレージ向け）で組み立てる。
func buildS3ObjectURL(endpoint, bucket, region, key string) string {
	if endpoint == "" {
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(endpoint, "/"), bucket, key)
}
