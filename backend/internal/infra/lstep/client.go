package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 10 * time.Second
	maxRetries       = 3
	retryInitialWait = time.Second
)

// Client はLステップAPIクライアントのインターフェース。
// DI可能にすることでテスト時のモック差し替えを可能にする。
type Client interface {
	// AddTag は指定LINE UserにLステップタグを付与する。
	AddTag(ctx context.Context, lineUserID, tagName string) error
	// RemoveTag は指定LINE UserからLステップタグを解除する。
	RemoveTag(ctx context.Context, lineUserID, tagName string) error
	// GetUserTags は指定LINE Userの全タグを返す。
	GetUserTags(ctx context.Context, lineUserID string) ([]string, error)
	// AddTagBulk は複数LINE Userに同一タグを一括付与する（健診対象者一括用）。
	AddTagBulk(ctx context.Context, lineUserIDs []string, tagName string) error
	// GetUser はLステップ登録情報（タグ・プロパティ含む）を返す。
	GetUser(ctx context.Context, lineUserID string) (*UserInfo, error)
	// SetProperty は指定LINE Userのカスタムプロパティを設定する。
	SetProperty(ctx context.Context, lineUserID, key, value string) error
}

// httpLstepClient はHTTPベースのLステップAPIクライアント実装。
type httpLstepClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewClient はLステップAPIクライアントを生成する。
// apiKey: LステップAPIキー（BE-000設定サービス経由で解決済みの値を渡す）
// baseURL: LステップAPIベースURL（例: https://api.lstep.jp/api/v1）
func NewClient(apiKey, baseURL string) Client {
	return &httpLstepClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}

// doWithRetry はレート制限時に指数バックオフで最大 maxRetries 回リトライする。
func (c *httpLstepClient) doWithRetry(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	wait := retryInitialWait
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := fn()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		// 429: レート制限 — レスポンスボディを破棄してリトライ
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		lastErr = ErrRateLimit
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("lstep client: %w", ctx.Err())
			case <-time.After(wait):
				wait *= 2
			}
		}
	}
	return nil, fmt.Errorf("%w: exhausted %d retries", lastErr, maxRetries)
}

// newRequest はLステップAPI向けのHTTPリクエストを生成する。
func (c *httpLstepClient) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create lstep request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// checkResponse はLステップAPIレスポンスのステータスコードを検査する。
// 404は ErrUserNotFound を返す。それ以外のエラーは汎用エラーとして返す。
func checkResponse(resp *http.Response, lineUserID string) error {
	if resp.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("lineUserID=%s: %w", lineUserID, ErrUserNotFound)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("lstep API error: status=%d body=%s", resp.StatusCode, body)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// decodeJSON はレスポンスボディをJSONデコードする。
func decodeJSON(resp *http.Response, dst any) error {
	defer func() { _, _ = io.Copy(io.Discard, resp.Body) }()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("lstep: decode response: %w", err)
	}
	return nil
}

// IsUserNotFound はエラーが ErrUserNotFound かどうかを判定する。
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
