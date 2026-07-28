package lstep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/line"
)

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
		return apperrors.Wrap(err, "marshal line payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, line.PushEndpoint, bytes.NewReader(body))
	if err != nil {
		return apperrors.Wrap(err, "create line request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.channelToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err, "send line push")
	}
	defer resp.Body.Close() //nolint:errcheck // レスポンスボディのクローズ失敗は復旧不可のため無視

	if resp.StatusCode != http.StatusOK {
		return apperrors.WrapInternalServerError(fmt.Sprintf("line push returned status %d", resp.StatusCode))
	}

	// Do not log LINE user IDs (owner identifiers). CODING_RULES.md:68 / LSA-07.
	slog.InfoContext(ctx, "LINE push sent")
	return nil
}
