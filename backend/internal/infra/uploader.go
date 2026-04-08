// Package infra はインフラストラクチャ層の実装（S3、ローカルファイルシステム等）を提供する。
package infra

import (
	"context"
	"io"
)

// FileUploader はファイルアップロード・削除のインターフェース。
// 実装は LocalUploader（ローカル開発）と S3Uploader（staging/production）。
type FileUploader interface {
	// Upload はファイルをストレージに保存し、公開URLを返す。
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
	// Delete はストレージからファイルを削除する。
	Delete(ctx context.Context, key string) error
}
