// Package support owns the staff-facing support widget boundary: bug reports with
// screenshots and page context, submitted from the in-app floating widget.
// Clinic-scoped (clinic_id) — every read/write is constrained to the selected clinic.
package support

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// maxBugReportsPerClinicList caps the admin list endpoint (newest first).
const maxBugReportsPerClinicList = 200

// Repository は support_bug_reports と support_chat_messages のデータアクセスインターフェース
type Repository interface {
	Create(ctx context.Context, report *model.SupportBugReport) error
	FindByClinicID(ctx context.Context, clinicID uint64) ([]BugReportWithReporter, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error
	// SetPlaneTicket は起票成功を記録し、claim 成功時（= 先に起票済みでない）に true を返す。
	SetPlaneTicket(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error)
	SetPlaneSyncError(ctx context.Context, clinicID, id uint64, syncErr string) error
	SoftDeleteBugReport(ctx context.Context, clinicID, id uint64) error
	CreateChatMessages(ctx context.Context, messages []*model.SupportChatMessage) error
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error
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

// maxChatHistoryPerFetch は履歴取得の上限（新しいものから n 件を取り出して古い順に返す）
const maxChatHistoryPerFetch = 100

func (r *repository) CreateChatMessages(ctx context.Context, messages []*model.SupportChatMessage) error {
	// 質問+回答のペアを1回の batch INSERT で原子保存する
	if err := r.db.WithContext(ctx).Create(&messages).Error; err != nil {
		return apperrors.FromGORM(err, "support_chat_message", "")
	}
	return nil
}

func (r *repository) ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error) {
	messages := make([]model.SupportChatMessage, 0)
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).
		Order("created_at DESC, id DESC").
		Limit(maxChatHistoryPerFetch).
		Find(&messages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "support_chat_message", "")
	}
	// 新しい順で取得したものを古い順（会話の時系列）に反転する
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	// 窓の境界で最古が assistant の場合、対になる user 質問が窓外に残る。
	// 1件だけ繰り上げ取得してペアを維持する。
	if len(messages) > 0 && messages[0].Role == model.SupportChatRoleAssistant {
		var preceding model.SupportChatMessage
		err := r.db.WithContext(ctx).
			Where("clinic_id = ? AND staff_id = ? AND id < ?", clinicID, staffID, messages[0].ID).
			Order("id DESC").
			Take(&preceding).Error
		switch {
		case err == nil && preceding.Role == model.SupportChatRoleUser:
			messages = append([]model.SupportChatMessage{preceding}, messages...)
		case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
			return nil, apperrors.FromGORM(err, "support_chat_message", "")
		}
	}
	return messages, nil
}

func (r *repository) ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error {
	// deleted_at を持つモデルの Delete は soft delete になる
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND staff_id = ?", clinicID, staffID).
		Delete(&model.SupportChatMessage{}).Error; err != nil {
		return apperrors.FromGORM(err, "support_chat_message", "")
	}
	return nil
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

// SetPlaneTicket は Plane 起票成功を記録する。二重起票防止のため
// plane_issue_id IS NULL の行にのみ書き込む（自動起票と手動再送、あるいは
// 手動再送の二重クリックが競合した場合、後着は claim 失敗 = false を返す）。
func (r *repository) SetPlaneTicket(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.SupportBugReport{}).
		Where("clinic_id = ? AND id = ? AND plane_issue_id IS NULL", clinicID, id).
		Updates(map[string]any{
			"plane_issue_id":   issueID,
			"plane_issue_url":  issueURL,
			"plane_sync_error": nil,
		})
	if result.Error != nil {
		return false, apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	return result.RowsAffected > 0, nil
}

// SetPlaneSyncError は直近の Plane 起票失敗理由を記録する（成功時は SetPlaneTicket が NULL に戻す）。
func (r *repository) SetPlaneSyncError(ctx context.Context, clinicID, id uint64, syncErr string) error {
	result := r.db.WithContext(ctx).
		Model(&model.SupportBugReport{}).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		Update("plane_sync_error", syncErr)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("support_bug_report", uintToString(id))
	}
	return nil
}

// SoftDeleteBugReport は報告を論理削除する（deleted_at 設定）。対象なしは NotFound。
// GORM の soft delete により削除済み行は FindByID/一覧から自然に除外される。
func (r *repository) SoftDeleteBugReport(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		Delete(&model.SupportBugReport{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("support_bug_report", uintToString(id))
	}
	return nil
}
