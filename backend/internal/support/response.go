package support

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// BugReportResponse はバグ報告の HTTP レスポンス
type BugReportResponse struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	PageURL         string    `json:"page_url"`
	RoutePath       string    `json:"route_path"`
	UserAgent       string    `json:"user_agent"`
	Viewport        string    `json:"viewport"`
	AppVersion      string    `json:"app_version"`
	Status          string    `json:"status"`
	ReporterStaffID uint64    `json:"reporter_staff_id"`
	ReporterName    string    `json:"reporter_name"`
	ScreenshotURL   string    `json:"screenshot_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BugReportListResponse はバグ報告一覧の HTTP レスポンス
type BugReportListResponse struct {
	Data []BugReportResponse `json:"data"`
}

func toBugReportResponse(report *model.SupportBugReport, reporterName string) BugReportResponse {
	return BugReportResponse{
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
		CreatedAt:       httpapi.LocalTime(report.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(report.UpdatedAt),
	}
}
