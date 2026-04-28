package line

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// pushRequest はLINE push APIリクエストボディ
type pushRequest struct {
	To       string    `json:"to"`
	Messages []message `json:"messages"`
}

// message はLINEメッセージ型（type に応じてフィールドが変わる）
type message map[string]any

// PushText はテキストメッセージを指定LINE Userにプッシュ送信する。
func (c *httpLineClient) PushText(ctx context.Context, lineUserID, text string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrInvalidRecipient)
	}
	payload := pushRequest{
		To: lineUserID,
		Messages: []message{
			{"type": "text", "text": text},
		},
	}
	return c.sendPush(ctx, payload)
}

// PushImageURL は画像URLをプッシュ送信する。
// originalURL: 元画像URL, previewURL: プレビュー画像URL（省略時はoriginalURLと同じ）
func (c *httpLineClient) PushImageURL(ctx context.Context, lineUserID, originalURL, previewURL string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrInvalidRecipient)
	}
	if previewURL == "" {
		previewURL = originalURL
	}
	payload := pushRequest{
		To: lineUserID,
		Messages: []message{
			{
				"type":               "image",
				"originalContentUrl": originalURL,
				"previewImageUrl":    previewURL,
			},
		},
	}
	return c.sendPush(ctx, payload)
}

// PushFileURL はFlex MessageでPDFファイルリンクをプッシュ送信する。
// altText: Flex Message の代替テキスト（通知バナー等に表示）
func (c *httpLineClient) PushFileURL(ctx context.Context, lineUserID, fileURL, altText string) error {
	if lineUserID == "" {
		return fmt.Errorf("lineUserID is empty: %w", ErrInvalidRecipient)
	}
	flexMsg := buildFileFlexMessage(altText, fileURL)
	payload := pushRequest{
		To:       lineUserID,
		Messages: []message{flexMsg},
	}
	return c.sendPush(ctx, payload)
}

// buildFileFlexMessage はPDFダウンロード用のFlex Messageを構築する。
func buildFileFlexMessage(altText, fileURL string) message {
	return message{
		"type":    "flex",
		"altText": altText,
		"contents": map[string]any{
			"type": "bubble",
			"body": map[string]any{
				"type":   "box",
				"layout": "vertical",
				"contents": []any{
					map[string]any{
						"type":   "text",
						"text":   altText,
						"wrap":   true,
						"weight": "bold",
						"size":   "md",
					},
				},
			},
			"footer": map[string]any{
				"type":   "box",
				"layout": "vertical",
				"contents": []any{
					map[string]any{
						"type":  "button",
						"style": "primary",
						"action": map[string]any{
							"type":  "uri",
							"label": "PDFを開く",
							"uri":   fileURL,
						},
					},
				},
			},
		},
	}
}

// sendPush はpushRequestをLINE Messaging APIに送信する共通処理。
func (c *httpLineClient) sendPush(ctx context.Context, payload pushRequest) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("line sendPush marshal: %w", err)
	}
	resp, err := c.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := c.newRequest(ctx, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		return c.http.Do(req)
	})
	if err != nil {
		return fmt.Errorf("line sendPush: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("line sendPush: %w", err)
	}
	return nil
}
