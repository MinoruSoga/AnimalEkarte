// Package httpx は line/lstep クライアント間で重複していた retry/newRequest ロジックを
// 共有化する（BE-refactor.md C-2）。wrap prefix・タイムアウト定数は各クライアントパッケージに
// 残し、本パッケージは 429 レート制限時の指数バックオフリトライと bearer トークン付き
// リクエスト生成という共通の骨格のみを提供する。
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrRateLimit はリトライを使い切ってもレート制限が解消しなかった場合に返す。
var ErrRateLimit = errors.New("httpx: rate limit exceeded")

// DoWithRetry はレート制限(429)時に指数バックオフで最大 maxRetries 回リトライする。
// 429 応答のボディは次のリクエスト前に drain + close する（コネクション再利用のため）。
// client は将来の拡張のためシグネチャに含めるが、do のクロージャが実際の送信を担う。
func DoWithRetry(ctx context.Context, client *http.Client, maxRetries int, initialWait time.Duration, do func() (*http.Response, error)) (*http.Response, error) {
	_ = client
	wait := initialWait
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := do()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		// 429: レート制限 — レスポンスボディを破棄してリトライ
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("httpx DoWithRetry: %w", ctx.Err())
			case <-time.After(wait):
				wait *= 2
			}
		}
	}
	return nil, fmt.Errorf("%w: exhausted %d retries", ErrRateLimit, maxRetries)
}

// NewBearerRequest は Authorization: Bearer <token> ヘッダ付きの JSON リクエストを生成する。
// http.NewRequestWithContext のエラーは "httpx NewBearerRequest:" プレフィックスで %w ラップして返す。
func NewBearerRequest(ctx context.Context, method, url, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("httpx NewBearerRequest: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
