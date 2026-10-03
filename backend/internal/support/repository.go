// Package support owns the staff-facing support widget boundary: bug reports with
// screenshots and page context, submitted from the in-app floating widget.
//
// バグ報告は意図的な全医院公開の共有ボード: 認証済みスタッフなら誰でも全医院の
// 報告を一覧できる（権限ゲート・医院絞りなし）。報告は医院の業務データではなく
// 製品へのフィードバックとして扱う product 決定（2026-10）。スクショに他院の
// 患者情報が写り得る点は仕様上許容済み。clinic_id/reporter_staff_id は絞り込みではなく
// provenance として記録する。
// 変更（2026-10 セキュリティレビュー）: 一覧・作成は全医院公開を維持するが、
// status 更新・Plane 起票・削除は報告元 clinic_id スコープに限定する。
// 他医院の報告への操作は「操作対象として存在しない」= 通常の NotFound と同じ
// 応答を返す。
// チャット履歴: 個人スコープの GET/DELETE /chat/history と、全医院共有ボードの
// GET /chat/exchanges（質問傾向の横断分析目的で意図的に公開 — バグ報告と同じ
// product 決定。質問内容に個人情報が含まれ得る点はユーザー承認済み）。
package support

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// maxBugReportsList caps the shared list endpoint (newest first).
const maxBugReportsList = 200

// maxChatExchangeMessages caps rows fetched for the shared exchange list
// (user+assistant pairs — at most half of them become list rows).
const maxChatExchangeMessages = 1000

// Repository は support_bug_reports と support_chat_messages のデータアクセスインターフェース
type Repository interface {
	Create(ctx context.Context, report *model.SupportBugReport) error
	FindAll(ctx context.Context) ([]BugReportWithReporter, error)
	// FindByIDForClinic は報告元医院スコープの操作（更新・起票・削除）が対象行を
	// 取り出すための取得。他医院の id は未存在と同じ NotFound を返す。
	FindByIDForClinic(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error)
	UpdateStatus(ctx context.Context, clinicID, id uint64, status model.SupportBugReportStatus) error
	// ListOpenWithPlaneTicket は Plane 起票済みで status=open の報告を返す。
	// scheduled sync（Plane 側 completed → 報告 resolved）の対象抽出用で
	// 全医院横断 — 起票済み報告は各報告の clinic_id を保持しているため
	// 返却後の UpdateStatus も報告元 clinic スコープで実行できる。
	ListOpenWithPlaneTicket(ctx context.Context, limit int) ([]model.SupportBugReport, error)
	// SetPlaneTicket は起票成功を記録し、claim 成功時（= 先に起票済みでない）に true を返す。
	SetPlaneTicket(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error)
	SetPlaneSyncError(ctx context.Context, clinicID, id uint64, syncErr string) error
	SoftDeleteBugReport(ctx context.Context, clinicID, id uint64) error
	CreateChatMessages(ctx context.Context, messages []*model.SupportChatMessage) error
	ListChatHistory(ctx context.Context, clinicID, staffID uint64) ([]model.SupportChatMessage, error)
	// ListAllChatMessages は全医院の履歴を新しい順で返す（共有一覧ページ用 — 個人履歴と別契約）
	ListAllChatMessages(ctx context.Context) ([]ChatMessageWithMeta, error)
	ClearChatHistory(ctx context.Context, clinicID, staffID uint64) error
}

// BugReportWithReporter は一覧表示用に報告者氏名・報告元医院名を結合した行。
// 全医院一覧のため、どの医院からの報告か識別できるよう clinic_name を返す。
type BugReportWithReporter struct {
	model.SupportBugReport
	ReporterName string
	ClinicName   string
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

// FindAll は全医院の報告を新しい順で返す（deleted_at 除外・上限 maxBugReportsList）。
func (r *repository) FindAll(ctx context.Context) ([]BugReportWithReporter, error) {
	reports := make([]BugReportWithReporter, 0)
	if err := r.db.WithContext(ctx).
		Table("support_bug_reports").
		Select("support_bug_reports.*, staffs.name AS reporter_name, clinics.name AS clinic_name").
		Joins("LEFT JOIN staffs ON staffs.id = support_bug_reports.reporter_staff_id").
		Joins("LEFT JOIN clinics ON clinics.id = support_bug_reports.clinic_id").
		Where("support_bug_reports.deleted_at IS NULL").
		Order("support_bug_reports.created_at DESC, support_bug_reports.id DESC").
		Limit(maxBugReportsList).
		Scan(&reports).Error; err != nil {
		return nil, apperrors.FromGORM(err, "support_bug_report", "")
	}
	return reports, nil
}

func (r *repository) FindByIDForClinic(ctx context.Context, clinicID, id uint64) (*model.SupportBugReport, error) {
	var report model.SupportBugReport
	err := r.db.WithContext(ctx).
		Where("id = ? AND clinic_id = ?", id, clinicID).
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

// ChatMessageWithMeta は共有一覧表示用にスタッフ氏名・医院名を結合した行。
type ChatMessageWithMeta struct {
	model.SupportChatMessage
	StaffName  string
	ClinicName string
}

// ListAllChatMessages は全医院の履歴を新しい順で返す（deleted_at 除外・上限 maxChatExchangeMessages）。
// 質問+回答ペア化は service 側で行う。
func (r *repository) ListAllChatMessages(ctx context.Context) ([]ChatMessageWithMeta, error) {
	messages := make([]ChatMessageWithMeta, 0)
	if err := r.db.WithContext(ctx).
		Table("support_chat_messages").
		Select("support_chat_messages.*, staffs.name AS staff_name, clinics.name AS clinic_name").
		Joins("LEFT JOIN staffs ON staffs.id = support_chat_messages.staff_id").
		Joins("LEFT JOIN clinics ON clinics.id = support_chat_messages.clinic_id").
		Where("support_chat_messages.deleted_at IS NULL").
		Order("support_chat_messages.created_at DESC, support_chat_messages.id DESC").
		Limit(maxChatExchangeMessages).
		Scan(&messages).Error; err != nil {
		return nil, apperrors.FromGORM(err, "support_chat_message", "")
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
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("status", status)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("support_bug_report", uintToString(id))
	}
	return nil
}

func (r *repository) ListOpenWithPlaneTicket(ctx context.Context, limit int) ([]model.SupportBugReport, error) {
	var reports []model.SupportBugReport
	query := r.db.WithContext(ctx).
		Where("status = ?", model.SupportBugReportStatusOpen).
		Where("plane_issue_id IS NOT NULL").
		Order("id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&reports).Error; err != nil {
		return nil, apperrors.FromGORM(err, "support_bug_report", "")
	}
	return reports, nil
}

// SetPlaneTicket は Plane 起票成功を記録する。二重起票防止のため
// plane_issue_id IS NULL の行にのみ書き込む（自動起票と手動再送、あるいは
// 手動再送の二重クリックが競合した場合、後着は claim 失敗 = false を返す）。
func (r *repository) SetPlaneTicket(ctx context.Context, clinicID, id uint64, issueID, issueURL string) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.SupportBugReport{}).
		Where("id = ? AND clinic_id = ? AND plane_issue_id IS NULL", id, clinicID).
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
		Where("id = ? AND clinic_id = ?", id, clinicID).
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
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Delete(&model.SupportBugReport{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "support_bug_report", uintToString(id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("support_bug_report", uintToString(id))
	}
	return nil
}
