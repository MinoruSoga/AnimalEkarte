package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// getUserResponse はユーザー情報取得APIレスポンス
type getUserResponse struct {
	LineUserID  string            `json:"line_user_id"`
	DisplayName string            `json:"display_name"`
	Tags        []string          `json:"tags"`
	Properties  map[string]string `json:"properties"`
}

// setPropertyRequest はユーザープロパティ設定APIリクエストボディ
type setPropertyRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
// 読み取り系のため deploy write gate の対象外。
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
		// lineUserID は error 文字列に埋め込まない（ログ/監視への PII 漏えい防止）
		return nil, ErrUserNotFound
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
// deploy gate 無効時は HTTP を送らず ErrWriteDisabled を返す。
func (c *httpLstepClient) SetProperty(ctx context.Context, lineUserID, key, value string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	if err := ensureWriteEnabled(); err != nil {
		return err
	}
	body, err := json.Marshal(setPropertyRequest{Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("lstep SetProperty marshal: %w", err)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodPost,
			fmt.Sprintf("/contacts/%s/properties", url.PathEscape(lineUserID)), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return fmt.Errorf("lstep SetProperty: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("lstep SetProperty: %w", err)
	}
	return nil
}
