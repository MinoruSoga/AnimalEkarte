package infra

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalUploader_UploadReturnsKeyAndWritesFile(t *testing.T) {
	baseDir := t.TempDir()
	u := NewLocalUploader(baseDir, "/uploads")
	key := "medical-records/5/photo.png"

	got, err := u.Upload(context.Background(), key, strings.NewReader("payload"), "image/png")
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if got != key {
		t.Fatalf("Upload returned %q, want object key %q", got, key)
	}
	if got == "/uploads/"+key {
		t.Fatal("Upload must return the object key, not /uploads/{key}")
	}

	written, err := os.ReadFile(filepath.Join(baseDir, key))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(written) != "payload" {
		t.Fatalf("uploaded file contents = %q, want payload", written)
	}
}

func TestLocalUploader_GetSignedURLReturnsBaseURLPrefixedPath(t *testing.T) {
	u := NewLocalUploader("/unused", "/uploads")
	key := "medical-records/5/photo.png"

	gotZeroTTL, err := u.GetSignedURL(context.Background(), key, 0)
	if err != nil {
		t.Fatalf("GetSignedURL(ttl=0): %v", err)
	}
	gotWithTTL, err := u.GetSignedURL(context.Background(), key, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetSignedURL(ttl=15m): %v", err)
	}

	want := "/uploads/" + key
	if gotZeroTTL != want {
		t.Fatalf("GetSignedURL(ttl=0) = %q, want %q", gotZeroTTL, want)
	}
	if gotWithTTL != want {
		t.Fatalf("GetSignedURL(ttl=15m) = %q, want %q (local TTL is unused)", gotWithTTL, want)
	}
	if gotZeroTTL == key {
		t.Fatal("GetSignedURL must return a URL distinct from the object key")
	}
}

func TestLocalUploader_DeleteByKey(t *testing.T) {
	baseDir := t.TempDir()
	u := NewLocalUploader(baseDir, "/uploads")
	key := "medical-records/5/photo.png"

	if _, err := u.Upload(context.Background(), key, strings.NewReader("payload"), "image/png"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := u.Delete(context.Background(), key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, key)); !os.IsNotExist(err) {
		t.Fatalf("stat after Delete: %v, want not exist", err)
	}
}
