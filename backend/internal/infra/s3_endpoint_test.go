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
		name     string
		endpoint string
		bucket   string
		region   string
		key      string
		want     string
	}{
		{
			name:     "endpoint空はAWSバーチャルホスト形式",
			endpoint: "",
			bucket:   "my-bucket",
			region:   "ap-northeast-1",
			key:      "clinics/1/foo.png",
			want:     "https://my-bucket.s3.ap-northeast-1.amazonaws.com/clinics/1/foo.png",
		},
		{
			name:     "カスタムendpointはpath-style",
			endpoint: "https://example.r2.cloudflarestorage.com",
			bucket:   "my-bucket",
			region:   "auto",
			key:      "clinics/1/foo.png",
			want:     "https://example.r2.cloudflarestorage.com/my-bucket/clinics/1/foo.png",
		},
		{
			name:     "endpoint末尾スラッシュの重複を防ぐ",
			endpoint: "https://example.r2.cloudflarestorage.com/",
			bucket:   "my-bucket",
			region:   "auto",
			key:      "foo.png",
			want:     "https://example.r2.cloudflarestorage.com/my-bucket/foo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildS3ObjectURL(tt.endpoint, tt.bucket, tt.region, tt.key)
			if got != tt.want {
				t.Errorf("buildS3ObjectURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
