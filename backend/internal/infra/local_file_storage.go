package infra

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalFileStorage はローカルファイルシステムをバックエンドとする FileStorage 実装（開発環境用）。
type LocalFileStorage struct {
	baseDir string // 例: "/app/uploads/shared"
	baseURL string // 例: "http://localhost:8080/uploads/shared"
}

// NewLocalFileStorage は LocalFileStorage を生成する。
func NewLocalFileStorage(baseDir, baseURL string) *LocalFileStorage {
	return &LocalFileStorage{baseDir: baseDir, baseURL: baseURL}
}

// Upload はファイルをローカルに保存する。
func (s *LocalFileStorage) Upload(_ context.Context, key string, content io.Reader, _ string) error {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil { //nolint:gosec // path is constructed from configured baseDir + sanitized key
		return fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.Create(fullPath) //nolint:gosec // path is constructed from configured baseDir + sanitized key
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close() //nolint:errcheck // close error on read-only flush is not actionable
	if _, err := io.Copy(f, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// GetSignedURL はローカルファイルの公開 URL を返す（開発環境では有効期限なし）。
func (s *LocalFileStorage) GetSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return fmt.Sprintf("%s/%s", s.baseURL, key), nil
}

// Delete はローカルのファイルを削除する。
func (s *LocalFileStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
