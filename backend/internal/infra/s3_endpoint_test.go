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
