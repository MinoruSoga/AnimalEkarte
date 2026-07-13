package line

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
	defaultTimeout   = 15 * time.Second
	pushEndpoint     = "https://api.line.me/v2/bot/message/push"
	maxRetries       = 3
	retryInitialWait = time.Second
)

// sharedHTTPClient はLINE Messaging API呼出全体で共有するhttp.Client。
// 通知のたびにNewMessagingClientを呼ぶたびに新規Transportを生成するとTCP/TLS接続が
// 再利用されない（BE-refactor.md B-3）。資格情報はリクエストヘッダ渡しのため
// クリニック間で共有して問題ない。
var sharedHTTPClient = &http.Client{Timeout: defaultTimeout}

// MessagingClient はLINE Messaging APIクライアントのインターフェース。
// DI可能にすることでテスト時のモック差し替えを可能にする。
type MessagingClient interface {
	// PushText はテキストメッセージを指定LINE Userにプッシュ送信する。
	PushText(ctx context.Context, lineUserID, text string) error
	// PushImageURL は画像URLをプッシュ送信する（originalContentUrl / previewImageUrl）。
	PushImageURL(ctx context.Context, lineUserID, originalURL, previewURL string) error
	// PushFileURL はFlex MessageでPDFファイルリンクをプッシュ送信する。
	PushFileURL(ctx context.Context, lineUserID, fileURL, altText string) error
}

// httpLineClient はHTTPベースのLINE Messaging APIクライアント実装。
type httpLineClient struct {
	channelAccessToken string
	http               *http.Client
}

// NewMessagingClient はLINE Messaging APIクライアントを生成する。
// channelAccessToken: LINE Messaging APIチャネルアクセストークン
func NewMessagingClient(channelAccessToken string) MessagingClient {
	return &httpLineClient{
		channelAccessToken: channelAccessToken,
		http:               sharedHTTPClient,
	}
}

// newRequest はLINE Messaging API向けのHTTPリクエストを生成する
// （共通ロジックは internal/infra/httpx.NewBearerRequest に集約。BE-refactor.md C-2）。
func (c *httpLineClient) newRequest(ctx context.Context, body io.Reader) (*http.Request, error) {
	req, err := httpx.NewBearerRequest(ctx, http.MethodPost, pushEndpoint, c.channelAccessToken, body)
	if err != nil {
		return nil, fmt.Errorf("create line request: %w", err)
	}
	return req, nil
}

// doWithRetry はレート制限時に指数バックオフで最大 maxRetries 回リトライする
// （共通ロジックは internal/infra/httpx.DoWithRetry に集約。BE-refactor.md C-2）。
func (c *httpLineClient) doWithRetry(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	resp, err := httpx.DoWithRetry(ctx, c.http, maxRetries, retryInitialWait, fn)
	if err != nil {
		if errors.Is(err, httpx.ErrRateLimit) {
			return nil, fmt.Errorf("%w: exhausted %d retries", ErrRateLimit, maxRetries)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("line client: %w", err)
		}
		return nil, err
	}
	return resp, nil
}

// lineErrorResponse はLINE Messaging API エラーレスポンス
type lineErrorResponse struct {
	Message string `json:"message"`
	Details []struct {
		Message  string `json:"message"`
		Property string `json:"property"`
	} `json:"details"`
}

// checkResponse はLINE Messaging APIレスポンスのステータスコードを検査する。
// 400/403 は ErrInvalidRecipient (ブロック済み・未追加) として処理継続扱い。
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	var lineErr lineErrorResponse
	if jsonErr := json.Unmarshal(body, &lineErr); jsonErr == nil {
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest {
			return fmt.Errorf("line messaging: %w (status=%d msg=%s)", ErrInvalidRecipient, resp.StatusCode, lineErr.Message)
		}
	}
	return fmt.Errorf("line messaging API error: status=%d body=%s", resp.StatusCode, body)
}
