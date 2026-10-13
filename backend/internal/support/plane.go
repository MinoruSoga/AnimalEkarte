package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
// external_id（= "<環境名>:<ローカル報告 ID>"）と組み合わせて Plane 側から辿れるようにする。
const planeExternalSource = "animalekarte-bug-report"

// planeEnvironmentAllowlist は external_id 名前空間に使える信頼済み環境名
// （APP_ENV 由来。リクエスト値や URL からは決して導出しない）。
var planeEnvironmentAllowlist = map[string]bool{
	"development": true,
	"staging":     true,
	"production":  true,
}

// normalizePlaneEnvironment は allowlist 外・空の値を development に正規化する。
// release での enabled 連携に対する厳格な拒否は config.Validate の責務であり、
// ここでは debug/テスト経路の既定値だけを保証する。
func normalizePlaneEnvironment(environment string) string {
	normalized := strings.ToLower(strings.TrimSpace(environment))
	if planeEnvironmentAllowlist[normalized] {
		return normalized
	}
	return "development"
}

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
	// environment は external_id / 件名 / 説明に使う信頼済み環境名（APP_ENV 由来）。
	environment string
	// labelIDs は起票時に送信する Plane label UUID（source:manual, lane:dev の順固定）。
	labelIDs []string
}

// PlaneClientOption は NewPlaneTicketCreator へのオプション設定を注入する。
type PlaneClientOption func(*planeClient)

// WithPlaneEnvironment は external_id 名前空間の環境名を設定する。
// allowlist 外・空は development に正規化される。
func WithPlaneEnvironment(environment string) PlaneClientOption {
	return func(c *planeClient) { c.environment = normalizePlaneEnvironment(environment) }
}

// WithPlaneLabelIDs は起票時に送信する label UUID を設定する。
// 両方が揃った場合のみ採用する（片方だけのラベル付き起票を防ぐ）。
func WithPlaneLabelIDs(sourceLabelID, laneLabelID string) PlaneClientOption {
	return func(c *planeClient) {
		sourceLabelID = strings.TrimSpace(sourceLabelID)
		laneLabelID = strings.TrimSpace(laneLabelID)
		if sourceLabelID != "" && laneLabelID != "" {
			c.labelIDs = []string{sourceLabelID, laneLabelID}
		}
	}
}

// NewPlaneTicketCreator は Plane クライアントを初期化する。
// apiKey / workspaceSlug / projectID のいずれかが空なら nil を返す（連携無効）。
func NewPlaneTicketCreator(
	baseURL, webBaseURL, apiKey, workspaceSlug, projectID, projectIdentifier, frontendURL string,
	timeout time.Duration,
	opts ...PlaneClientOption,
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
	client := &planeClient{
		baseURL:           strings.TrimRight(baseURL, "/"),
		webBaseURL:        strings.TrimRight(webBaseURL, "/"),
		workspaceSlug:     workspaceSlug,
		projectID:         projectID,
		projectIdentifier: projectIdentifier,
		apiKey:            apiKey,
		frontendURL:       strings.TrimRight(frontendURL, "/"),
		httpClient:        &http.Client{Timeout: timeout},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
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

// planeEnvironment は external_id / 件名 / 説明に使う環境名を返す。
// WithPlaneEnvironment は設定時に正規化するが、直接構築されたクライアントも
// allowlist 外を素通ししないよう送信時にも正規化する。
func (c *planeClient) planeEnvironment() string {
	return normalizePlaneEnvironment(c.environment)
}

// planeExternalID は "<環境名>:<ローカル報告ID>" 形式の external_id。
// STG/prod など環境が違えば同一報告 ID でも衝突しない。
func (c *planeClient) planeExternalID(report *model.SupportBugReport) string {
	return fmt.Sprintf("%s:%d", c.planeEnvironment(), report.ID)
}

// Planeへの送信は固定文とサーバ採番の報告IDに限定する。
// 件名には環境名を含め、環境の異なる同名報告を Plane 側で区別できるようにする。
// HTMLエスケープは機微情報の除去にならないため、任意入力を転記しない。
func (c *planeClient) planeIssueName(report *model.SupportBugReport) string {
	return fmt.Sprintf("[バグ報告][%s] #%d", c.planeEnvironment(), report.ID)
}

// 自動起票・手動再送に共通する外部送信境界。原文はアプリ内で確認する。
func (c *planeClient) planeDescriptionHTML(report *model.SupportBugReport) string {
	return fmt.Sprintf(
		"<p>環境: %s</p><p>報告ID: #%d</p><p>詳細はアプリ内のバグ報告管理画面で確認してください。</p>",
		c.planeEnvironment(),
		report.ID,
	)
}

// CreateBugReportIssue は報告内容を Plane ワークアイテムとして作成する。
// 非 2xx は *planeUpstreamError で返す（エラーボディは読み捨てのみ）。
// 送信 payload は allowlist のみ — name / description_html / external_source /
// external_id / 設定済み labels。label 未設定時は labels キーを送らない。
func (c *planeClient) CreateBugReportIssue(ctx context.Context, report *model.SupportBugReport) (*PlaneIssue, error) {
	body := map[string]any{
		"name":             c.planeIssueName(report),
		"description_html": c.planeDescriptionHTML(report),
		"external_source":  planeExternalSource,
		"external_id":      c.planeExternalID(report),
	}
	if len(c.labelIDs) > 0 {
		body["labels"] = c.labelIDs
	}
	payload, err := json.Marshal(body)
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
	// Plane Cloud の実応答はフラットな state_group を返す（STG で実測確認）。
	// state_detail.group はセルフホスト等の別版互換のフォールバックとして残す。
	StateGroup  string `json:"state_group"`
	StateDetail *struct {
		Group string `json:"group"`
	} `json:"state_detail"`
}

// FetchIssueStateGroup は GET /v1/workspaces/{slug}/projects/{id}/issues/{issueID}/ の
// state_group（欠落時は state_detail.group）を返す。非 2xx は *planeUpstreamError、group 欠落は空文字。
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
	if parsed.StateGroup != "" {
		return parsed.StateGroup, nil
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
