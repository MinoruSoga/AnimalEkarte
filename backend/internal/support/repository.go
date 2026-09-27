// Package support owns the staff-facing support widget boundary: bug reports with
// screenshots and page context, submitted from the in-app floating widget.
// Clinic-scoped (clinic_id) — every read/write is constrained to the selected clinic.
package support

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// maxBugReportsPerClinicList caps the admin list endpoint (newest first).
const maxBugReportsPerClinicList = 200

// Repository は support_bug_reports のデータアクセスインターフェース
type Repository interface {
	Create(ctx context.Context, report *model.SupportBugReport) error
	FindByClinicID(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error
}

// BugReportWithReporter は一覧表示用に報告者氏名を結合した行
type BugReportWithReporter struct {
	model.SupportBugReport
	ReporterName string
}

type repository struct {
	db *gorm.DB
}

// NewRepository は Repository を初期化して返す
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, report *model.SupportBugReport) error {
	if err := r.db.WithContext(ctx).Create(report).Error; err != nil {
		return apperrors.FromGORM(err, "support_bug_report", "")
	}
	return nil
}

func (r *repository) FindByClinicID(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error) {
	reports := make([]BugReportWithReporter, 0)
	if err := r.db.WithContext(ctx).
		Table("support_bug_reports").
		Select("support_bug_reports.*, staffs.name AS reporter_name").
		Joins("LEFT JOIN staffs ON staffs.id = support_bug_reports.reporter_staff_id").
		Where("support_bug_reports.clinic_id = ? AND support_bug_reports.deleted_at IS NULL", clinicID).
		Order("support_bug_reports.created_at DESC, support_bug_reports.id DESC").
		Limit(maxBugReportsPerClinicList).
		Scan(&reports).Error; err != nil {
		return nil, apperrors.FromGORM(err, "support_bug_report", "")
	}
	return reports, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	var report model.SupportBugReport
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&report).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "support_bug_report", uintToString(id))
	}
	return &report, nil
}

func (r *repository) UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error {
	result := r.db.WithContext(ctx).
		Model(&model.SupportBugReport{}).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		Update("status", status)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("support_bug_report", uintToString(id))
	}
	return nil
}
