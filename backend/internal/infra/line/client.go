package line

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 15 * time.Second
	pushEndpoint     = "https://api.line.me/v2/bot/message/push"
	maxRetries       = 3
	retryInitialWait = time.Second
)

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
		http:               &http.Client{Timeout: defaultTimeout},
	}
}

// newRequest はLINE Messaging API向けのHTTPリクエストを生成する。
func (c *httpLineClient) newRequest(ctx context.Context, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pushEndpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.channelAccessToken)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// doWithRetry はレート制限時に指数バックオフで最大 maxRetries 回リトライする。
func (c *httpLineClient) doWithRetry(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
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
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		lastErr = ErrRateLimit
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
				wait *= 2
			}
		}
	}
	return nil, fmt.Errorf("%w: exhausted %d retries", lastErr, maxRetries)
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
