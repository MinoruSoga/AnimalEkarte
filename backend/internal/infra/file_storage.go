package infra

import (
	"context"
	"io"
	"time"
)

// FileStorage はファイルの保存・署名付きURL生成・削除のインターフェース。
// 実装は S3FileStorage（staging/production）と LocalFileStorage（開発環境）。
// 既存の FileUploader とは異なり、GetSignedURL をサポートする。
type FileStorage interface {
	// Upload はファイルをストレージに保存する。key はS3キーまたはローカルパス。
	Upload(ctx context.Context, key string, content io.Reader, contentType string) error
	// GetSignedURL は指定 TTL で有効な署名付き URL を生成する。
	GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	// Delete はストレージからファイルを削除する。
	Delete(ctx context.Context, key string) error
}
