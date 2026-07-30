package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// addTagRequest はタグ付与APIリクエストボディ
type addTagRequest struct {
	TagName string `json:"tag_name"`
}

// removeTagRequest はタグ解除APIリクエストボディ
type removeTagRequest struct {
	TagName string `json:"tag_name"`
}

// getUserTagsResponse はタグ一覧APIレスポンス
type getUserTagsResponse struct {
	Tags []string `json:"tags"`
}

// addTagBulkRequest は一括タグ付与APIリクエストボディ
type addTagBulkRequest struct {
	LineUserIDs []string `json:"line_user_ids"`
	TagName     string   `json:"tag_name"`
}

// AddTag は指定LINE UserにLステップタグを付与する。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
// deploy gate 無効時は HTTP を送らず ErrWriteDisabled を返す。
func (c *httpLstepClient) AddTag(ctx context.Context, lineUserID, tagName string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	if err := ensureWriteEnabled(); err != nil {
		return err
	}
	body, err := json.Marshal(addTagRequest{TagName: tagName})
	if err != nil {
		return fmt.Errorf("lstep AddTag marshal: %w", err)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodPost,
			fmt.Sprintf("/contacts/%s/tags", url.PathEscape(lineUserID)), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return fmt.Errorf("lstep AddTag: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("lstep AddTag: %w", err)
	}
	return nil
}

// RemoveTag は指定LINE UserからLステップタグを解除する。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
// deploy gate 無効時は HTTP を送らず ErrWriteDisabled を返す。
func (c *httpLstepClient) RemoveTag(ctx context.Context, lineUserID, tagName string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	if err := ensureWriteEnabled(); err != nil {
		return err
	}
	body, err := json.Marshal(removeTagRequest{TagName: tagName})
	if err != nil {
		return fmt.Errorf("lstep RemoveTag marshal: %w", err)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodDelete,
			fmt.Sprintf("/contacts/%s/tags", url.PathEscape(lineUserID)), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return fmt.Errorf("lstep RemoveTag: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("lstep RemoveTag: %w", err)
	}
	return nil
}

// GetUserTags は指定LINE Userに付与されているLステップタグの一覧を返す。
// lineUserID が空文字の場合は即座に ErrUserNotFound を返す。
// 読み取り系のため deploy write gate の対象外。
func (c *httpLstepClient) GetUserTags(ctx context.Context, lineUserID string) ([]string, error) {
	if lineUserID == "" {
		return nil, fmt.Errorf("lineUserID is empty: %w", ErrUserNotFound)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodGet,
			fmt.Sprintf("/contacts/%s/tags", url.PathEscape(lineUserID)), nil)
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
		// lineUserID は error 文字列に埋め込まない（ログ/監視への PII 漏えい防止）
		return nil, ErrUserNotFound
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
// deploy gate 無効時は HTTP を送らず ErrWriteDisabled を返す。
// 空の lineUserIDs は gate 通過後に即座に nil を返す（APIを呼ばない）。
func (c *httpLstepClient) AddTagBulk(ctx context.Context, lineUserIDs []string, tagName string) error {
	if err := ensureWriteEnabled(); err != nil {
		return err
	}
	if len(lineUserIDs) == 0 {
		return nil
	}
	body, err := json.Marshal(addTagBulkRequest{LineUserIDs: lineUserIDs, TagName: tagName})
	if err != nil {
		return fmt.Errorf("lstep AddTagBulk marshal: %w", err)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, http.MethodPost, "/contacts/tags/bulk", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return fmt.Errorf("lstep AddTagBulk: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("lstep AddTagBulk: status=%d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
