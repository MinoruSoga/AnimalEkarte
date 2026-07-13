package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/animal-ekarte/backend/internal/infra/httpx"
)

const (
	defaultTimeout   = 10 * time.Second
	maxRetries       = 3
	retryInitialWait = time.Second
)

// sharedHTTPClient はLステップAPI呼出全体で共有するhttp.Client。
// 操作毎にNewClientを呼ぶたびに新規Transportを生成するとTCP/TLS接続が再利用されない
// （BE-refactor.md B-3）。資格情報はリクエストヘッダ渡しのためクリニック間で共有して問題ない。
var sharedHTTPClient = &http.Client{Timeout: defaultTimeout}

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
		http:    sharedHTTPClient,
	}
}

// doWithRetry はレート制限時に指数バックオフで最大 maxRetries 回リトライする
// （共通ロジックは internal/infra/httpx.DoWithRetry に集約。BE-refactor.md C-2）。
func (c *httpLstepClient) doWithRetry(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	resp, err := httpx.DoWithRetry(ctx, c.http, maxRetries, retryInitialWait, fn)
	if err != nil {
		if errors.Is(err, httpx.ErrRateLimit) {
			return nil, fmt.Errorf("%w: exhausted %d retries", ErrRateLimit, maxRetries)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("lstep client: %w", err)
		}
		return nil, err
	}
	return resp, nil
}

// newRequest はLステップAPI向けのHTTPリクエストを生成する
// （共通ロジックは internal/infra/httpx.NewBearerRequest に集約。BE-refactor.md C-2）。
func (c *httpLstepClient) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := httpx.NewBearerRequest(ctx, method, c.baseURL+path, c.apiKey, body)
	if err != nil {
		return nil, fmt.Errorf("create lstep request: %w", err)
	}
	return req, nil
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
