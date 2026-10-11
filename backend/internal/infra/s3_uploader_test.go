package infra

import (
	"context"
	"testing"
	"time"
)

var _ FileUploader = (*S3Uploader)(nil)

func TestS3Uploader_GetSignedURL_NilPresignFailClosed(t *testing.T) {
	u := &S3Uploader{}

	got, err := u.GetSignedURL(context.Background(), "medical-records/5/photo.png", 15*time.Minute)
	if err == nil {
		t.Fatalf("GetSignedURL succeeded with nil presign: %q", got)
	}
	if got != "" {
		t.Fatalf("fail-closed GetSignedURL must not return a URL, got %q", got)
	}
}
