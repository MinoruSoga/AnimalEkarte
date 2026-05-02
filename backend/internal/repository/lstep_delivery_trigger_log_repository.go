package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

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
