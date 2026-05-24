package lstep

import (
	"context"
	"fmt"
	"net/http"
)

// getUserTagsResponse はタグ一覧APIレスポンス
type getUserTagsResponse struct {
	Tags []string `json:"tags"`
}

// AddTag は指定LINE UserにLステップタグを付与する。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
//
// Temporarily disabled: L-step write operations are paused by policy.
func (c *httpLstepClient) AddTag(_ context.Context, lineUserID, _ string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	// [DISABLED] HTTP call to POST /contacts/{id}/tags is suppressed.
	// To re-enable, restore the original implementation from git history.
	return nil
}

// RemoveTag は指定LINE UserからLステップタグを解除する。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
//
// Temporarily disabled: L-step write operations are paused by policy.
func (c *httpLstepClient) RemoveTag(_ context.Context, lineUserID, _ string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	// [DISABLED] HTTP call to DELETE /contacts/{id}/tags is suppressed.
	// To re-enable, restore the original implementation from git history.
	return nil
}

// GetUserTags は指定LINE Userに付与されているLステップタグの一覧を返す。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
func (c *httpLstepClient) GetUserTags(ctx context.Context, lineUserID string) ([]string, error) {
	if lineUserID == "" {
		return nil, fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodGet,
			fmt.Sprintf("/contacts/%s/tags", lineUserID), nil)
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("lstep GetUserTags: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lineUserID=%s: %w", lineUserID, ErrUserNotFound)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("lstep GetUserTags: status=%d", resp.StatusCode)
	}
	var result getUserTagsResponse
	if err := decodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("lstep GetUserTags: %w", err)
	}
	return result.Tags, nil
}

// AddTagBulk は複数LINE Userに同一タグを一括付与する。
// 空の lineUserIDs は即座にnilを返す（APIを呼ばない）。
//
// Temporarily disabled: L-step write operations are paused by policy.
func (c *httpLstepClient) AddTagBulk(_ context.Context, _ []string, _ string) error {
	// [DISABLED] HTTP call to POST /contacts/tags/bulk is suppressed.
	// To re-enable, restore the original implementation from git history.
	return nil
}
