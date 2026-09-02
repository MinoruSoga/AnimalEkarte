package lstep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/infra/httpx"
)

const (
	DefaultBaseURL   = "https://api.lstep.jp"
	defaultTimeout   = 10 * time.Second
	maxRetries       = 3
	retryInitialWait = time.Second

	// EnvWriteAPIEnabled は Lステップ write API の deploy-level kill switch。
	// 未設定・空・"false"・未知値は全て無効。exact "true" のみ有効。
	// 既存 UI/API/seed/migration からは変更できない（運用者が deploy 環境変数のみ設定）。
	EnvWriteAPIEnabled = "LSTEP_WRITE_API_ENABLED"
)

// sharedHTTPClient はLステップAPI呼出全体で共有するhttp.Client。
// 操作毎にNewClientを呼ぶたびに新規Transportを生成するとTCP/TLS接続が再利用されない
// （BE-refactor.md B-3）。資格情報はリクエストヘッダ渡しのためクリニック間で共有して問題ない。
// Dial は loopback / RFC1918 / CGNAT を拒否し、リダイレクトは追わない。
var sharedHTTPClient = newHardenedHTTPClient(defaultTimeout)

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
//
// Write 系（AddTag / RemoveTag / AddTagBulk / SetProperty）は
// EnvWriteAPIEnabled が exact "true" のときだけ HTTP を送る。
// clinic 別 is_sync_enabled はサービス層 buildClient が gate する（二重 gate）。
func NewClient(apiKey, baseURL string) Client {
	return &httpLstepClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    sharedHTTPClient,
	}
}

// NewInsecureTestClient talks to httptest.Server (loopback). It panics outside
// `go test` so production binaries cannot use it to bypass hardenedDialContext.
func NewInsecureTestClient(apiKey, baseURL string) Client {
	if !testing.Testing() {
		panic("lstep.NewInsecureTestClient is test-only")
	}
	return &httpLstepClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}

// isWriteAPIEnabled は deploy-level kill switch を評価する。
// exact "true" のみ true。unset/empty/false/未知値は全て false。
func isWriteAPIEnabled() bool {
	return os.Getenv(EnvWriteAPIEnabled) == "true"
}

// ensureWriteEnabled は write 系メソッドの共通 gate。
// 無効時は HTTP を送らず ErrWriteDisabled を返す（nil 成功にしない）。
func ensureWriteEnabled() error {
	if !isWriteAPIEnabled() {
		return ErrWriteDisabled
	}
	return nil
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

// checkResponse はLステップ write API レスポンスのステータスコードを検査する。
// 404 は ErrUserNotFound（lineUserID は error 文字列に埋め込まない）。
// それ以外の 4xx/5xx は status のみを含む observable error。
// api key / request body / response body は error にも log にも出さない。
func checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		_, _ = io.Copy(io.Discard, resp.Body)
		return ErrUserNotFound
	}
	if resp.StatusCode >= 400 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("lstep API error: status=%d", resp.StatusCode)
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

// IsWriteDisabled はエラーが ErrWriteDisabled かどうかを判定する。
func IsWriteDisabled(err error) bool {
	return errors.Is(err, ErrWriteDisabled)
}
