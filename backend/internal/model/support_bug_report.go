package model

import (
	"gorm.io/gorm"
	"time"
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
	ID              uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64         `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	ReporterStaffID uint64         `gorm:"column:reporter_staff_id;type:bigint;not null" json:"reporter_staff_id"`
	Title           string         `gorm:"column:title;type:text;not null" json:"title"`
	Detail          string         `gorm:"column:detail;type:text;not null;default:''" json:"detail"`
	PageURL         string         `gorm:"column:page_url;type:text;not null;default:''" json:"page_url"`
	RoutePath       string         `gorm:"column:route_path;type:text;not null;default:''" json:"route_path"`
	UserAgent       string         `gorm:"column:user_agent;type:text;not null;default:''" json:"user_agent"`
	Viewport        string         `gorm:"column:viewport;type:text;not null;default:''" json:"viewport"`
	AppVersion      string         `gorm:"column:app_version;type:text;not null;default:''" json:"app_version"`
	ScreenshotKey   *string        `gorm:"column:screenshot_key;type:text" json:"screenshot_key,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(20);not null;default:'open'" json:"status"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
	Version         int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
}

// TableName は SupportBugReport のテーブル名
func (SupportBugReport) TableName() string { return "support_bug_reports" }
