// Package infra はインフラストラクチャ層の実装（S3、ローカルファイルシステム等）を提供する。
package infra

import (
	"context"
	"io"
	"time"
)

// FileUploader はファイルアップロード・削除・署名付きURL生成のインターフェース。
// 実装は LocalUploader（ローカル開発）と S3Uploader（staging/production）。
type FileUploader interface {
	// Upload はファイルをストレージに保存し、オブジェクト key を返す。
	Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error)
	// Delete はストレージからファイルを削除する。
	Delete(ctx context.Context, key string) error
	// Exists は key のオブジェクトが存在するかを返す。
	// 存在しないオブジェクトは (false, nil)。存在を確認できないインフラ
	// エラー（ネットワーク・権限等）は error を返し、呼び出し元は
	// 存在不明として fail-closed に扱う。
	Exists(ctx context.Context, key string) (bool, error)
	// GetSignedURL は指定 TTL で有効な取得用 URL を生成する。
	GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
