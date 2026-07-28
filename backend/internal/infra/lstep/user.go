package lstep

import (
	"context"
	"fmt"
	"net/url"
	"net/http"
)

// getUserResponse はユーザー情報取得APIレスポンス
type getUserResponse struct {
	LineUserID  string            `json:"line_user_id"`
	DisplayName string            `json:"display_name"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
}

// UserInfo はLステップユーザー情報
type UserInfo struct {
	LineUserID  string
	DisplayName string
	Tags        []string
	Properties  map[string]string
}

// GetUser は指定LINE UserのLステップ登録情報を返す。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
func (c *httpLstepClient) GetUser(ctx context.Context, lineUserID string) (*UserInfo, error) {
	if lineUserID == "" {
		return nil, fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodGet,
			fmt.Sprintf("/contacts/%s", url.PathEscape(lineUserID)), nil)
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("lstep GetUser: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lineUserID=%s: %w", lineUserID, ErrUserNotFound)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("lstep GetUser: status=%d", resp.StatusCode)
	}
	var result getUserResponse
	if err := decodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("lstep GetUser: %w", err)
	}
	return &UserInfo{
		LineUserID:  result.LineUserID,
		DisplayName: result.DisplayName,
		Tags:        result.Tags,
		Properties:  result.Properties,
	}, nil
}

// SetProperty は指定LINE Userのカスタムプロパティを設定する。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
//
// Temporarily disabled: L-step write operations are paused by policy.
func (c *httpLstepClient) SetProperty(_ context.Context, lineUserID, _, _ string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	// [DISABLED] HTTP call to POST /contacts/{id}/properties is suppressed.
	// To re-enable, restore the original implementation from git history.
	return nil
}
