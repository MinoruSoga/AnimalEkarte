package model

import (
	"time"

	"gorm.io/gorm"
)

// SupportBugReportStatus はバグ報告の対応状況
type SupportBugReportStatus string

const (
	SupportBugReportStatusOpen     SupportBugReportStatus = "open"
	SupportBugReportStatusResolved SupportBugReportStatus = "resolved"
)

// IsValidSupportBugReportStatus は SupportBugReportStatus が有効値か判定
func IsValidSupportBugReportStatus(s string) bool {
	return s == string(SupportBugReportStatusOpen) || s == string(SupportBugReportStatusResolved)
}

// SupportBugReport はアプリ内サポートウィジェットから送信されたバグ報告。
// スクショ画像本体は FileUploader に保存し、DB にはオブジェクト key のみ保持する。
type SupportBugReport struct {
	ID              uint64                 `gorm:"primaryKey;autoIncrement"            json:"id"`
	ClinicID        uint64                 `gorm:"not null;index"                    json:"clinic_id"`
	ReporterStaffID uint64                 `gorm:"not null"                          json:"reporter_staff_id"`
	Title           string                 `gorm:"not null;type:text"                json:"title"`
	Detail          string                 `gorm:"not null;type:text;default:''"     json:"detail"`
	PageURL         string                 `gorm:"not null;type:text;default:''"     json:"page_url"`
	RoutePath       string                 `gorm:"not null;default:''"               json:"route_path"`
	UserAgent       string                 `gorm:"not null;default:''"               json:"user_agent"`
	Viewport        string                 `gorm:"not null;default:''"               json:"viewport"`
	AppVersion      string                 `gorm:"not null;default:''"               json:"app_version"`
	ScreenshotKey   *string                `gorm:"type:text"                         json:"screenshot_key,omitempty"`
	Status          SupportBugReportStatus `gorm:"type:varchar(20);not null;default:'open'" json:"status"`
	CreatedAt       time.Time              `gorm:"autoCreateTime"                    json:"created_at"`
	UpdatedAt       time.Time              `gorm:"autoUpdateTime"                    json:"updated_at"`
	DeletedAt       gorm.DeletedAt         `gorm:"index"                             json:"deleted_at,omitempty"`
}

// TableName は SupportBugReport のテーブル名
func (SupportBugReport) TableName() string { return "support_bug_reports" }
