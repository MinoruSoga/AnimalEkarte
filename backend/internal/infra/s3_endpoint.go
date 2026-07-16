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

// buildS3ObjectURL はアップロード後のオブジェクト URL を組み立てる。
// URL 組み立ての優先順位は次のとおり。
//
//  1. publicBaseURL（env: S3_PUBLIC_BASE_URL）が指定されていれば最優先で使う。
//     ブラウザ向け公開ホスト（R2 custom domain / *.r2.dev / CloudFront 等）は
//     S3 API endpoint（*.r2.cloudflarestorage.com）とは別ホストであり、
//     API endpoint を公開 URL に流用すると認証が必要でブラウザから参照できない。
//     R2 の公開ドメインはドメイン→バケットが1:1対応するためバケット名はパスに含めない。
//  2. publicBaseURL が空 かつ endpoint も空 = AWS S3 既定。バーチャルホスト形式
//     （bucket.s3.region.amazonaws.com）はブラウザから直接参照できるため維持する。
//  3. publicBaseURL が空 かつ endpoint（R2 等）指定時は、推測公開ドメインを
//     捏造せず従来の path-style endpoint URL を暫定で返す。この URL は API ホストを
//     指すためブラウザ参照不可であり、S3_PUBLIC_BASE_URL 投入までの一時挙動。
//     この誤設定は起動時（cmd/api/main.go）に警告ログで通知する。
func buildS3ObjectURL(publicBaseURL, endpoint, bucket, region, key string) string {
	if publicBaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(publicBaseURL, "/"), key)
	}
	if endpoint == "" {
		return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(endpoint, "/"), bucket, key)
}
