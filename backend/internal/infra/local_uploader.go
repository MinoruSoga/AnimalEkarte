package infra

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalUploader はローカルファイルシステムにファイルを保存する FileUploader 実装。
// ローカル開発環境用。
type LocalUploader struct {
	baseDir string // 例: "/app/uploads"
	baseURL string // 例: "/uploads"
}

// NewLocalUploader は LocalUploader を生成する。
func NewLocalUploader(baseDir, baseURL string) *LocalUploader {
	return &LocalUploader{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

// Upload はファイルをローカルに保存し、オブジェクト key を返す。
func (u *LocalUploader) Upload(_ context.Context, key string, body io.Reader, _ string) (string, error) {
	fullPath := filepath.Join(u.baseDir, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // ローカル開発専用: パスはアプリ設定値のみ
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(fullPath) //nolint:gosec // ローカル開発専用: パスはアプリ設定値のみ
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close() //nolint:errcheck // 書き込み後のクローズ失敗は復旧不可のため無視

	if _, err := io.Copy(f, body); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return key, nil
}

// GetSignedURL はローカル開発用の取得 URL を返す（TTL は使用しない）。
func (u *LocalUploader) GetSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return fmt.Sprintf("%s/%s", u.baseURL, key), nil
}

// Delete はローカルのファイルを削除する。
func (u *LocalUploader) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(u.baseDir, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
