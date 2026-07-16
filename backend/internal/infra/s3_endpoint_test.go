package infra

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestApplyS3EndpointOverride_EmptyLeavesDefaultBehavior(t *testing.T) {
	o := &s3.Options{}
	applyS3EndpointOverride(o, "")

	if o.BaseEndpoint != nil {
		t.Errorf("BaseEndpoint = %v, want nil (AWS 既定のリージョナルエンドポイントを維持)", o.BaseEndpoint)
	}
	if o.UsePathStyle {
		t.Errorf("UsePathStyle = true, want false（AWS 既定はバーチャルホスト形式）")
	}
}

func TestApplyS3EndpointOverride_SetsBaseEndpointAndPathStyle(t *testing.T) {
	o := &s3.Options{}
	applyS3EndpointOverride(o, "https://example.r2.cloudflarestorage.com")

	if o.BaseEndpoint == nil || *o.BaseEndpoint != "https://example.r2.cloudflarestorage.com" {
		t.Errorf("BaseEndpoint = %v, want https://example.r2.cloudflarestorage.com", o.BaseEndpoint)
	}
	if !o.UsePathStyle {
		t.Errorf("UsePathStyle = false, want true（R2 等 S3 互換ストレージは path-style が必須）")
	}
}

func TestBuildS3ObjectURL(t *testing.T) {
	tests := []struct {
		name       string
		publicBase string
		endpoint   string
		bucket     string
		region     string
		key        string
		want       string
	}{
		{
			// publicBase/endpoint 共に空 = AWS S3 既定。バーチャルホスト形式 URL は
			// ブラウザから直接参照できるため従来どおり維持する。
			name:       "publicBaseもendpointも空はAWSバーチャルホスト形式",
			publicBase: "",
			endpoint:   "",
			bucket:     "my-bucket",
			region:     "ap-northeast-1",
			key:        "clinics/1/foo.png",
			want:       "https://my-bucket.s3.ap-northeast-1.amazonaws.com/clinics/1/foo.png",
		},
		{
			// 公開 base 設定時は API endpoint(R2 の *.r2.cloudflarestorage.com)ではなく
			// ブラウザ向け公開ホストを使う。R2 の custom domain / r2.dev は
			// ドメイン→バケットが1:1対応するため、バケット名はパスに含めない。
			name:       "publicBase設定時は公開ホストを使いバケットをパスに含めない",
			publicBase: "https://images.example.com",
			endpoint:   "https://example.r2.cloudflarestorage.com",
			bucket:     "my-bucket",
			region:     "auto",
			key:        "clinics/1/foo.png",
			want:       "https://images.example.com/clinics/1/foo.png",
		},
		{
			// publicBase は endpoint より優先される。endpoint が空(AWS)でも
			// CloudFront 等の公開 base を設定すればそちらを使う。
			name:       "publicBaseはendpoint空(AWS)でも優先される",
			publicBase: "https://cdn.example.com",
			endpoint:   "",
			bucket:     "my-bucket",
			region:     "ap-northeast-1",
			key:        "clinics/1/foo.png",
			want:       "https://cdn.example.com/clinics/1/foo.png",
		},
		{
			name:       "publicBase末尾スラッシュの重複を防ぐ",
			publicBase: "https://images.example.com/",
			endpoint:   "https://example.r2.cloudflarestorage.com",
			bucket:     "my-bucket",
			region:     "auto",
			key:        "foo.png",
			want:       "https://images.example.com/foo.png",
		},
		{
			// 【意図的な一時フォールバック】publicBase 未設定 + endpoint(R2)設定時は、
			// 推測公開ドメインを捏造せず、従来の path-style endpoint URL を維持する。
			// この URL は API ホストを指すためブラウザからは参照できない(要認証)。
			// S3_PUBLIC_BASE_URL に R2 公開ドメインが投入されるまでの暫定挙動であり、
			// 起動時に main.go が警告ログを出す(P2-5: 値投入は USER 運用タスク)。
			name:       "publicBase未設定かつendpoint設定時は暫定でendpoint path-style(ブラウザ参照不可)",
			publicBase: "",
			endpoint:   "https://example.r2.cloudflarestorage.com",
			bucket:     "my-bucket",
			region:     "auto",
			key:        "clinics/1/foo.png",
			want:       "https://example.r2.cloudflarestorage.com/my-bucket/clinics/1/foo.png",
		},
		{
			name:       "endpoint末尾スラッシュの重複を防ぐ(フォールバック経路)",
			publicBase: "",
			endpoint:   "https://example.r2.cloudflarestorage.com/",
			bucket:     "my-bucket",
			region:     "auto",
			key:        "foo.png",
			want:       "https://example.r2.cloudflarestorage.com/my-bucket/foo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildS3ObjectURL(tt.publicBase, tt.endpoint, tt.bucket, tt.region, tt.key)
			if got != tt.want {
				t.Errorf("buildS3ObjectURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
