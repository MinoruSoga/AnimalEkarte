package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// planeResponseMaxBytes は Plane API レスポンスの読み取り上限（異常に大きな応答からの防衛）
const planeResponseMaxBytes = 1 << 20

// planeErrorBodyDiscardBytes は非 2xx 時にボディを捨てるための読み取り上限。
// エラーボディはログ・DB へ書かない（上流のエラー詳細に機微情報が混入しうるため、
// ステータスコードだけを記録する）。
const planeErrorBodyDiscardBytes = 64 << 10

// planeExternalSource は Plane 側でこのアプリ発のワークアイテムを識別する固定値。
// external_id（= ローカル報告 ID）と組み合わせて Plane 側から辿れるようにする。
const planeExternalSource = "animalekarte-bug-report"

// planeIssueNameMaxRunes は Plane の name フィールド上限(255)に対する安全側の上限。
const planeIssueNameMaxRunes = 200

// TicketCreator は外部チケットシステムへのバグ報告起票を抽象化する。
// nil の場合は連携無効（報告はローカル保存のみ・手動起票は 501 を返す）。
type TicketCreator interface {
	CreateBugReportIssue(ctx context.Context, report *model.SupportBugReport) (*PlaneIssue, error)
}

// TicketStateReader は起票済みチケットの対応状況を外部から読み取る抽象化。
// TicketCreator とは別 interface に分け、実装が存在する場合のみ
// scheduled sync が有効になる（mock/nil 実装は読み取り経路を持たない）。
type TicketStateReader interface {
	FetchIssueStateGroup(ctx context.Context, issueID string) (string, error)
}

// PlaneIssue は Plane に作成されたワークアイテムの参照情報。
type PlaneIssue struct {
	ID  string // Plane work item の UUID
	URL string // 表示用 URL（<webBase>/<workspace>/browse/<IDENT>/）。identifier が判明しない場合は空
}

// planeUpstreamError は Plane API の非 2xx 応答を表す。
type planeUpstreamError struct {
	status int
}

func (e *planeUpstreamError) Error() string {
	return fmt.Sprintf("plane upstream returned status %d", e.status)
}

// planeSyncErrorMessage は Plane 起票失敗を DB/レスポンスへ安全に記録する短い文字列に変換する。
// 上流ボディやネットワーク層の詳細はそのまま保存しない（内部実装の露出防止）。
func planeSyncErrorMessage(err error) string {
	var upstream *planeUpstreamError
	if errors.As(err, &upstream) {
		return fmt.Sprintf("plane api error (status %d)", upstream.status)
	}
	return "plane request failed"
}

// planeClient は Plane REST API v1 の work item 作成クライアント。
// POST {baseURL}/v1/workspaces/{slug}/projects/{projectID}/issues/ に
// X-API-Key 認証で投げる。projectIdentifier はレスポンスに identifier が無い
// 場合の表示 URL 組立 fallback（例 "EMR"）。
type planeClient struct {
	baseURL           string
	webBaseURL        string
	workspaceSlug     string
	projectID         string
	projectIdentifier string
	apiKey            string
	frontendURL       string
	httpClient        *http.Client
}

// NewPlaneTicketCreator は Plane クライアントを初期化する。
// apiKey / workspaceSlug / projectID のいずれかが空なら nil を返す（連携無効）。
func NewPlaneTicketCreator(
	baseURL, webBaseURL, apiKey, workspaceSlug, projectID, projectIdentifier, frontendURL string,
	timeout time.Duration,
) TicketCreator {
	if strings.TrimSpace(apiKey) == "" ||
		strings.TrimSpace(workspaceSlug) == "" ||
		strings.TrimSpace(projectID) == "" {
		return nil
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.plane.so/api"
	}
	if strings.TrimSpace(webBaseURL) == "" {
		webBaseURL = "https://app.plane.so"
	}
	return &planeClient{
		baseURL:           strings.TrimRight(baseURL, "/"),
		webBaseURL:        strings.TrimRight(webBaseURL, "/"),
		workspaceSlug:     workspaceSlug,
		projectID:         projectID,
		projectIdentifier: projectIdentifier,
		apiKey:            apiKey,
		frontendURL:       strings.TrimRight(frontendURL, "/"),
		httpClient:        &http.Client{Timeout: timeout},
	}
}

// planeIssueResponse は Plane issues API の作成応答で使う項目だけを抜き出す。
// identifier は API バージョン/設定により返らないことがあるため、
// project_detail.identifier + sequence_id からの合成もフォールバックとして持つ。
type planeIssueResponse struct {
	ID            string `json:"id"`
	SequenceID    int64  `json:"sequence_id"`
	Identifier    string `json:"identifier"`
	ProjectDetail *struct {
		Identifier string `json:"identifier"`
	} `json:"project_detail"`
}

// planeIssueName はチケット件名を組み立てる（[バグ報告] 接頭辞 + 報告件名）。
func planeIssueName(report *model.SupportBugReport) string {
	name := "[バグ報告] " + strings.TrimSpace(report.Title)
	if r := []rune(name); len(r) > planeIssueNameMaxRunes {
		name = string(r[:planeIssueNameMaxRunes-1]) + "…"
	}
	return name
}

// planeDescriptionHTML は報告の内容・画面文脈を HTML に組み立てる。
// すべてのユーザー入力は html.EscapeString でエスケープしてから埋め込む。
// スクショ画像の URL は直接貼らない（署名付き URL は短命であり、チケット内に
// 固定リンクを残さないため）。「あり」の場合はアプリ内管理画面で確認する旨だけ記す。
func (c *planeClient) planeDescriptionHTML(report *model.SupportBugReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<p>%s</p>", strings.ReplaceAll(html.EscapeString(report.Detail), "\n", "<br>"))

	b.WriteString("<h3>環境</h3><ul>")
	writeListItem := func(label, value string) {
		if value == "" {
			value = "―"
		}
		fmt.Fprintf(&b, "<li>%s: %s</li>", label, html.EscapeString(value))
	}
	writeListItem("画面", report.RoutePath)
	writeListItem("ページURL", report.PageURL)
	writeListItem("表示領域", report.Viewport)
	writeListItem("User-Agent", report.UserAgent)
	writeListItem("アプリバージョン", report.AppVersion)
	b.WriteString("</ul>")

	screenshot := "なし"
	if report.ScreenshotKey != nil {
		screenshot = "あり（アプリのバグ報告詳細で確認。個人情報を含む可能性あり）"
	}
	fmt.Fprintf(&b, "<p>報告ID: #%d / 医院ID: %d / 報告者: staff#%d / スクリーンショット: %s</p>",
		report.ID, report.ClinicID, report.ReporterStaffID, screenshot)
	if !report.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "<p>報告日時: %s</p>", report.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	}
	if c.frontendURL != "" {
		fmt.Fprintf(&b, "<p>アプリ内管理画面: <a href=\"%s/settings/bug-reports\">%s/settings/bug-reports</a></p>",
			html.EscapeString(c.frontendURL), html.EscapeString(c.frontendURL))
	}
	return b.String()
}

// CreateBugReportIssue は報告内容を Plane ワークアイテムとして作成する。
// 非 2xx は *planeUpstreamError で返す（エラーボディは読み捨てのみ）。
func (c *planeClient) CreateBugReportIssue(ctx context.Context, report *model.SupportBugReport) (*PlaneIssue, error) {
	payload, err := json.Marshal(map[string]any{
		"name":             planeIssueName(report),
		"description_html": c.planeDescriptionHTML(report),
		"external_source":  planeExternalSource,
		"external_id":      strconv.FormatUint(report.ID, 10),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plane request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/workspaces/%s/projects/%s/issues/", c.baseURL, c.workspaceSlug, c.projectID),
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build plane request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("plane request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // クローズ失敗は復旧不可

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, planeErrorBodyDiscardBytes))
		return nil, &planeUpstreamError{status: resp.StatusCode}
	}

	var parsed planeIssueResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, planeResponseMaxBytes)).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to decode plane response: %w", err)
	}
	if parsed.ID == "" {
		return nil, fmt.Errorf("plane response missing issue id")
	}
	return &PlaneIssue{ID: parsed.ID, URL: c.issueURL(&parsed)}, nil
}

// planeIssueStateResponse は GET issues/{id}/ 応答から対応状況グループだけを抜き出す。
// Plane の状態グループは backlog/unstarted/started/completed/cancelled —
// グループが取れない応答は「不明」として空文字を返す（resolved 側へ誤判定しないため）。
type planeIssueStateResponse struct {
	StateDetail *struct {
		Group string `json:"group"`
	} `json:"state_detail"`
}

// FetchIssueStateGroup は GET /v1/workspaces/{slug}/projects/{id}/issues/{issueID}/ の
// state_detail.group を返す。非 2xx は *planeUpstreamError、group 欠落は空文字。
func (c *planeClient) FetchIssueStateGroup(ctx context.Context, issueID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/workspaces/%s/projects/%s/issues/%s/",
		c.baseURL, c.workspaceSlug, c.projectID, url.PathEscape(issueID)), http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to build plane request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("plane request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // クローズ失敗は復旧不可

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, planeErrorBodyDiscardBytes))
		return "", &planeUpstreamError{status: resp.StatusCode}
	}

	var parsed planeIssueStateResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, planeResponseMaxBytes)).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode plane response: %w", err)
	}
	if parsed.StateDetail == nil {
		return "", nil
	}
	return parsed.StateDetail.Group, nil
}

// issueURL は app.plane.so/<workspace>/browse/<IDENT>/ 形式の表示 URL を組み立てる。
// identifier を決定できない場合は空文字を返す（チケット自体は作成済みのため起票は失敗扱いにしない）。
func (c *planeClient) issueURL(resp *planeIssueResponse) string {
	identifier := resp.Identifier
	if identifier == "" && resp.ProjectDetail != nil && resp.SequenceID > 0 {
		identifier = fmt.Sprintf("%s-%d", resp.ProjectDetail.Identifier, resp.SequenceID)
	}
	if identifier == "" && c.projectIdentifier != "" && resp.SequenceID > 0 {
		identifier = fmt.Sprintf("%s-%d", c.projectIdentifier, resp.SequenceID)
	}
	if identifier == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/browse/%s/", c.webBaseURL, c.workspaceSlug, identifier)
}
