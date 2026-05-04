package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// DeliveryTriggerLogRow は FindByDateRangeWithFilters の飼い主名 JOIN 結合行。
type DeliveryTriggerLogRow struct {
	model.LstepDeliveryTriggerLog
	OwnerName string `gorm:"column:owner_name"`
}

// DeliveryStatsRow はトリガー種別 × ステータス別集計行。
type DeliveryStatsRow struct {
	TriggerType string `gorm:"column:trigger_type" json:"trigger_type"`
	Status      string `gorm:"column:status"       json:"status"`
	Count       int64  `gorm:"column:count"        json:"count"`
}

// LstepDeliveryTriggerLogRepository は lstep_delivery_trigger_log テーブルの永続化インターフェース。
type LstepDeliveryTriggerLogRepository interface {
	// Create は新規トリガーログを作成する。
	Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error
	// ExistsTodayByOwnerAndType は当日同一飼い主・同一トリガー種別のログが存在するか返す（二重発火防止）。
	ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error)
	// UpdateStatus はログのステータス・fired_at・excluded_reason を更新する。
	UpdateStatus(ctx context.Context, id uint64, status string, firedAt *time.Time, excludedReason *string) error
	// FindByClinicAndDate はクリニック・日付でトリガーログ一覧を返す（管理確認用）。
	FindByClinicAndDate(ctx context.Context, clinicID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error)
	// CountByStatusAndDateRange はクリニック・期間でステータス別件数マップを返す。triggerType が空なら全種別。
	CountByStatusAndDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error)
	// CountExcludedReasonByDateRange は除外ログを除外理由別に集計する。triggerType が空なら全種別。
	CountExcludedReasonByDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error)
	// FindByDateRangeWithFilters は飼い主名 JOIN 付きでページングログ一覧と総件数を返す。
	FindByDateRangeWithFilters(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error)
	// ListByOwnerAndDateRange はクリニック・飼主単位で期間内トリガーログ一覧を返す。
	ListByOwnerAndDateRange(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error)
	// CountByTypeAndStatus はクリニック・飼主単位で期間内トリガー種別 × ステータス別集計を返す。
	CountByTypeAndStatus(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]DeliveryStatsRow, error)
}

type lstepDeliveryTriggerLogRepository struct{ db *gorm.DB }

// NewLstepDeliveryTriggerLogRepository は LstepDeliveryTriggerLogRepository を初期化して返す。
func NewLstepDeliveryTriggerLogRepository(db *gorm.DB) LstepDeliveryTriggerLogRepository {
	return &lstepDeliveryTriggerLogRepository{db: db}
}

func (r *lstepDeliveryTriggerLogRepository) Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_delivery_trigger_log", "")
	}
	return nil
}

func (r *lstepDeliveryTriggerLogRepository) ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	var count int64
	err := r.db.WithContext(ctx).Model(&model.LstepDeliveryTriggerLog{}).
		Where("clinic_id = ? AND owner_id = ? AND trigger_type = ? AND scheduled_at >= ? AND scheduled_at < ?",
			clinicID, ownerID, triggerType, dayStart, dayEnd).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "exists_today")
	}
	return count > 0, nil
}

func (r *lstepDeliveryTriggerLogRepository) UpdateStatus(ctx context.Context, id uint64, status string, firedAt *time.Time, excludedReason *string) error {
	fields := map[string]any{"status": status}
	if firedAt != nil {
		fields["fired_at"] = firedAt
	}
	if excludedReason != nil {
		fields["excluded_reason"] = excludedReason
	}
	result := r.db.WithContext(ctx).Model(&model.LstepDeliveryTriggerLog{}).
		Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_delivery_trigger_log", "update_status")
	}
	return nil
}

func (r *lstepDeliveryTriggerLogRepository) FindByClinicAndDate(ctx context.Context, clinicID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	var logs []model.LstepDeliveryTriggerLog
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, dayStart, dayEnd).
		Order("scheduled_at ASC").
		Find(&logs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "find_by_clinic_date")
	}
	return logs, nil
}

func (r *lstepDeliveryTriggerLogRepository) CountByStatusAndDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error) {
	type row struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	query := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_logs").
		Select("status, COUNT(*) AS count").
		Where("clinic_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, from, to).
		Group("status")
	if triggerType != "" {
		query = query.Where("trigger_type = ?", triggerType)
	}
	var rows []row
	if err := query.Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_by_status")
	}
	result := make(map[string]int64)
	for _, rw := range rows {
		result[rw.Status] = rw.Count
	}
	return result, nil
}

func (r *lstepDeliveryTriggerLogRepository) CountExcludedReasonByDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error) {
	type row struct {
		ExcludedReason string `gorm:"column:excluded_reason"`
		Count          int64  `gorm:"column:count"`
	}
	query := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_logs").
		Select("COALESCE(excluded_reason, '') AS excluded_reason, COUNT(*) AS count").
		Where("clinic_id = ? AND scheduled_at >= ? AND scheduled_at < ? AND status = ?", clinicID, from, to, model.TriggerStatusExcluded).
		Group("excluded_reason")
	if triggerType != "" {
		query = query.Where("trigger_type = ?", triggerType)
	}
	var rows []row
	if err := query.Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_excluded_reason")
	}
	result := make(map[string]int64)
	for _, rw := range rows {
		result[rw.ExcludedReason] = rw.Count
	}
	return result, nil
}

func (r *lstepDeliveryTriggerLogRepository) FindByDateRangeWithFilters(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error) {
	// count without join for efficiency
	countQ := r.db.WithContext(ctx).
		Model(&model.LstepDeliveryTriggerLog{}).
		Where("clinic_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, from, to)
	if triggerType != "" {
		countQ = countQ.Where("trigger_type = ?", triggerType)
	}
	if status != "" {
		countQ = countQ.Where("status = ?", status)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_find")
	}

	// data query with owner name join
	findQ := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_logs l").
		Select("l.*, COALESCE(o.name, '') AS owner_name").
		Joins("LEFT JOIN owners o ON o.id = l.owner_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ?", clinicID, from, to)
	if triggerType != "" {
		findQ = findQ.Where("l.trigger_type = ?", triggerType)
	}
	if status != "" {
		findQ = findQ.Where("l.status = ?", status)
	}
	var rows []DeliveryTriggerLogRow
	if err := findQ.
		Order("l.scheduled_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "find_with_filters")
	}
	return rows, total, nil
}

func (r *lstepDeliveryTriggerLogRepository) ListByOwnerAndDateRange(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	var logs []model.LstepDeliveryTriggerLog
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, ownerID, from, to).
		Order("scheduled_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "list_by_owner")
	}
	return logs, nil
}

func (r *lstepDeliveryTriggerLogRepository) CountByTypeAndStatus(ctx context.Context, clinicID, ownerID uint64, from, to time.Time) ([]DeliveryStatsRow, error) {
	var rows []DeliveryStatsRow
	err := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_logs").
		Select("trigger_type, status, COUNT(*) AS count").
		Where("clinic_id = ? AND owner_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, ownerID, from, to).
		Group("trigger_type, status").
		Order("trigger_type, status").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_by_type_status")
	}
	return rows, nil
}
