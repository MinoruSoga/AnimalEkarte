package infra

import (
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
