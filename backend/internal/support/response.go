package support

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// BugReportResponse はバグ報告の HTTP レスポンス
type BugReportResponse struct {
	ID              uint64 `json:"id"`
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	PageURL         string `json:"page_url"`
	RoutePath       string `json:"route_path"`
	UserAgent       string `json:"user_agent"`
	Viewport        string `json:"viewport"`
	AppVersion      string `json:"app_version"`
	Status          string `json:"status"`
	ReporterStaffID uint64 `json:"reporter_staff_id"`
	ReporterName    string `json:"reporter_name"`
	// ClinicID は報告元医院ID。一覧の自院判定（操作可否のUI制御）と provenance。
	ClinicID uint64 `json:"clinic_id"`
	// ClinicName は報告元医院名。一覧のみ付与（全医院公開のため provenance として必要）。
	ClinicName    string `json:"clinic_name,omitempty"`
	ScreenshotURL string `json:"screenshot_url,omitempty"`
	// PlaneDeliveryState は Plane への配送状態（D23 の support_delivery_results 由来）。
	// 配送結果の reader が未接続の間は常に "unavailable"（結果を読めない）。
	// 旧 plane_issue_url / plane_sync_error は列ごと削除済みで復元しない。
	// リンク・失敗理由の欠落を「未送信」と解釈してはならない。
	PlaneDeliveryState string    `json:"plane_delivery_state"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// planeDeliveryStateUnavailable は配送結果（support_delivery_results）を参照できない状態の固定値。
// 外部 ID・リンクが無いことは未送信の証拠にならない（technical-spec.md 選定済み契約 第 7 項）。
const planeDeliveryStateUnavailable = "unavailable"

// BugReportListResponse はバグ報告一覧の HTTP レスポンス
type BugReportListResponse struct {
	Data []BugReportResponse `json:"data"`
}

func toBugReportResponse(report *model.SupportBugReport, reporterName, clinicName string) BugReportResponse {
	resp := BugReportResponse{
		ID:              report.ID,
		Title:           report.Title,
		Detail:          report.Detail,
		PageURL:         report.PageURL,
		RoutePath:       report.RoutePath,
		UserAgent:       report.UserAgent,
		Viewport:        report.Viewport,
		AppVersion:      report.AppVersion,
		Status:          string(report.Status),
		ReporterStaffID: report.ReporterStaffID,
		ReporterName:    reporterName,
		ClinicID:        report.ClinicID,
		ClinicName:      clinicName,
		CreatedAt:       httpapi.LocalTime(report.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(report.UpdatedAt),
	}
	// 配送結果 reader 未接続のため、Plane への送信有無を推測しない（外部 ID なし = 未送信ではない）。
	resp.PlaneDeliveryState = planeDeliveryStateUnavailable
	return resp
}
