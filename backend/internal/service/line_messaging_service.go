package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const lineMessagingAPIURL = "https://api.line.me/v2/bot/message/push"

// LineMessagingService はLINE Messaging APIを使ったプッシュ通知クライアント。
// channelToken が空の場合はすべての送信をスキップする（設定不要時に安全に無効化）。
type LineMessagingService struct {
	channelToken string
	httpClient   *http.Client
}

// NewLineMessagingService はLINE Messaging Serviceを初期化して返す。
// channelToken が空の場合は noop インスタンスを返す。
func NewLineMessagingService(channelToken string) *LineMessagingService {
	return &LineMessagingService{
		channelToken: channelToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type lineTextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type linePushPayload struct {
	To       string            `json:"to"`
	Messages []lineTextMessage `json:"messages"`
}

// PushText は指定したLINEユーザーIDにテキストメッセージを送信する。
// channelToken が未設定の場合は何もしない。
// エラーは呼び出し元に返すが、予約フロー自体を止めてはならない。
func (s *LineMessagingService) PushText(ctx context.Context, lineUserID, text string) error {
	if s.channelToken == "" {
		return nil
	}

	payload := linePushPayload{
		To: lineUserID,
		Messages: []lineTextMessage{
			{Type: "text", Text: text},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal line payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineMessagingAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create line request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.channelToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send line push: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("line push returned status %d", resp.StatusCode)
	}

	slog.InfoContext(ctx, "LINE push sent", "to", lineUserID)
	return nil
}
