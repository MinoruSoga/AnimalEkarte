package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// Discord embed の各フィールド上限に対する安全側の上限。
// 公式上限は title 256 / description 4096 / field value 1024 / embed 合計 6000。
const (
	discordEmbedTitleMaxRunes       = 240
	discordEmbedDescriptionMaxRunes = 3900
	discordEmbedFieldValueMaxRunes  = 900
)

// discordResponseMaxBytes は webhook 応答の読み捨て上限（非 2xx 時の詳細は
// ログに書かずステータスコードだけを返す — 上流エラー文の機微混入防止）。
const discordResponseMaxBytes = 64 << 10

// BugReportNotifier はバグ報告の作成時外部通知を抽象化する。
// nil の場合は通知無効。通知は DB 保存前に best-effort で実行され、
// 失敗しても報告作成自体は失敗させない（WARN ログのみ）。
type BugReportNotifier interface {
	NotifyBugReport(ctx context.Context, report *model.SupportBugReport) error
}

// discordNotifier は Discord Incoming Webhook への通知クライアント。
// webhook URL はチャンネル固有の秘密値であり、ログ・エラーに出力しない。
type discordNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewDiscordNotifier は Discord webhook 通知を初期化する。
// webhookURL が空なら nil を返す（通知無効）。
func NewDiscordNotifier(webhookURL string, timeout time.Duration) BugReportNotifier {
	if strings.TrimSpace(webhookURL) == "" {
		return nil
	}
	return &discordNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type discordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []discordEmbedField `json:"fields"`
	Timestamp   string              `json:"timestamp"`
}

type discordWebhookPayload struct {
	Username        string `json:"username"`
	AllowedMentions struct {
		Parse []string `json:"parse"`
	} `json:"allowed_mentions"`
	Embeds []discordEmbed `json:"embeds"`
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// NotifyBugReport は報告原文を Discord チャンネルへ embed として投稿する。
// allowed_mentions を空にして報告文面中の @everyone 等のメンションを抑止する。
// 非 2xx はステータスコードのみのエラーで返す（応答ボディは読み捨て）。
func (n *discordNotifier) NotifyBugReport(ctx context.Context, report *model.SupportBugReport) error {
	screenshot := "なし"
	if report.ScreenshotKey != nil {
		screenshot = "あり"
	}
	empty := "―"
	fieldValue := func(v string) string {
		if strings.TrimSpace(v) == "" {
			return empty
		}
		return truncateRunes(v, discordEmbedFieldValueMaxRunes)
	}

	payload := discordWebhookPayload{
		Username: "AnimalEkarte バグ報告",
		Embeds: []discordEmbed{{
			Title:       truncateRunes("[バグ報告] "+strings.TrimSpace(report.Title), discordEmbedTitleMaxRunes),
			Description: truncateRunes(report.Detail, discordEmbedDescriptionMaxRunes),
			Color:       15158332, // red
			Fields: []discordEmbedField{
				{Name: "医院", Value: fmt.Sprintf("clinic#%d", report.ClinicID), Inline: true},
				{Name: "報告者", Value: fmt.Sprintf("staff#%d", report.ReporterStaffID), Inline: true},
				{Name: "スクリーンショット", Value: screenshot, Inline: true},
				{Name: "画面", Value: fieldValue(report.RoutePath), Inline: true},
				{Name: "ページURL", Value: fieldValue(report.PageURL), Inline: true},
				{Name: "アプリバージョン", Value: fieldValue(report.AppVersion), Inline: true},
				{Name: "表示領域", Value: fieldValue(report.Viewport), Inline: true},
				{Name: "User-Agent", Value: fieldValue(report.UserAgent)},
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}},
	}
	payload.AllowedMentions.Parse = []string{}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discord webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build discord webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord webhook request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // クローズ失敗は復旧不可

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, discordResponseMaxBytes))
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, discordResponseMaxBytes))
	return nil
}
