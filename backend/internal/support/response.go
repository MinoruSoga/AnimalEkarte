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
	// PlaneIssueURL は起票済み Plane チケットの表示 URL（未起票なら省略）
	PlaneIssueURL string `json:"plane_issue_url,omitempty"`
	// PlaneSyncError は直近の Plane 起票失敗理由（成功・未試行なら省略）
	PlaneSyncError string    `json:"plane_sync_error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

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
	if report.PlaneIssueURL != nil {
		resp.PlaneIssueURL = *report.PlaneIssueURL
	}
	if report.PlaneSyncError != nil {
		resp.PlaneSyncError = *report.PlaneSyncError
	}
	return resp
}
