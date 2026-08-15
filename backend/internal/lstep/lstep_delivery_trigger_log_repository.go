package lstep

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// jstHalfOpenDay returns the half-open [dayStart, dayEnd) window for the JST
// calendar day that contains t, plus the YYYY-MM-DD key for advisory locks.
// Matches migration 002 expression index:
// ((scheduled_at AT TIME ZONE 'Asia/Tokyo')::date).
func jstHalfOpenDay(t time.Time) (dayStart, dayEnd time.Time, dayKey string) {
	jst := t.In(config.JST)
	dayStart = time.Date(jst.Year(), jst.Month(), jst.Day(), 0, 0, 0, 0, config.JST)
	dayEnd = dayStart.AddDate(0, 0, 1)
	dayKey = dayStart.Format(time.DateOnly)
	return dayStart, dayEnd, dayKey
}

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

// VisitConversionRow はトリガー種別ごとの配信後来院集計行。
type VisitConversionRow struct {
	TriggerType    string `gorm:"column:trigger_type"    json:"trigger_type"`
	DeliveredCount int64  `gorm:"column:delivered_count" json:"delivered_count"`
	VisitedCount   int64  `gorm:"column:visited_count"   json:"visited_count"`
}

// LstepDeliveryTriggerLogRepository は lstep_delivery_trigger_log テーブルの永続化インターフェース。
type LstepDeliveryTriggerLogRepository interface {
	// Create は新規トリガーログを作成する。
	Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error
	// CreateIfAbsentToday serializes check+create under a day-scoped advisory lock (LSA-15).
	// Returns created=false when another row already claims the same clinic/owner/type/day slot.
	// UNIQUE index (step ④) remains the final DB defense; this closes the app-level race first.
	CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (created bool, err error)
	// ExistsTodayByOwnerAndType は当日同一飼い主・同一トリガー種別のログが存在するか返す（二重発火防止）。
	ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error)
	// UpdateStatus はログのステータス・fired_at・excluded_reason を更新する。
	UpdateStatus(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error
	// CountByStatusAndDateRange はクリニック・期間でステータス別件数マップを返す。triggerType が空なら全種別。
	CountByStatusAndDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error)
	// CountExcludedReasonByDateRange は除外ログを除外理由別に集計する。triggerType が空なら全種別。
	CountExcludedReasonByDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error)
	// CountSuppressedByPriorityDateRange は優先順位により抑制されたログ件数を返す。triggerType が空なら全種別。
	CountSuppressedByPriorityDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (int64, error)
	// FindByDateRangeWithFilters は飼い主名 JOIN 付きでページングログ一覧と総件数を返す。
	FindByDateRangeWithFilters(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error)
	// CountByTypeAndStatus はクリニック単位で期間内トリガー種別 × ステータス別集計を返す。
	CountByTypeAndStatus(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error)
	// CountVisitConversionsByType は期間内の fired ログについてトリガー種別ごとの来院転換数を返す。
	CountVisitConversionsByType(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error)
	// FindByOwnerAndDate は同日同 owner_id の既存ログを返す (Q23 suppressed check 用)。
	FindByOwnerAndDate(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error)
	// UpdateSuppressed は既存ログを suppressed_by_priority=true に更新する (Q23 降格処理用)。
	UpdateSuppressed(ctx context.Context, clinicID, logID uint64, reason string) error
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

func (r *lstepDeliveryTriggerLogRepository) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if log == nil {
		return false, apperrors.WrapInvalidInput("delivery trigger log is required")
	}
	dayStart, dayEnd, dayKey := jstHalfOpenDay(log.ScheduledAt)
	lockKey := fmt.Sprintf("lstep_delivery_trigger:%d:%d:%s:%s", log.ClinicID, log.OwnerID, log.TriggerType, dayKey)

	var created bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize concurrent claim of the same JST day/type slot (LSA-15 / X-05).
		// Fail-closed: lock acquisition failure aborts the write.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).Error; err != nil {
			return apperrors.Wrap(err, "failed to acquire delivery trigger day lock")
		}
		var count int64
		if err := tx.Model(&model.LstepDeliveryTriggerLog{}).
			Where("clinic_id = ? AND owner_id = ? AND trigger_type = ? AND scheduled_at >= ? AND scheduled_at < ?",
				log.ClinicID, log.OwnerID, log.TriggerType, dayStart, dayEnd).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "lstep_delivery_trigger_log", "exists_today_locked")
		}
		if count > 0 {
			created = false
			return nil
		}
		if err := tx.Create(log).Error; err != nil {
			return apperrors.FromGORM(err, "lstep_delivery_trigger_log", "")
		}
		created = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return created, nil
}

func (r *lstepDeliveryTriggerLogRepository) ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error) {
	dayStart, dayEnd, _ := jstHalfOpenDay(date)
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

func (r *lstepDeliveryTriggerLogRepository) UpdateStatus(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error {
	fields := map[string]any{"status": status}
	if firedAt != nil {
		fields["fired_at"] = firedAt
	}
	if excludedReason != nil {
		fields["excluded_reason"] = excludedReason
	}
	result := r.db.WithContext(ctx).Model(&model.LstepDeliveryTriggerLog{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_delivery_trigger_log", "update_status")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_delivery_trigger_log", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *lstepDeliveryTriggerLogRepository) CountByStatusAndDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (map[string]int64, error) {
	type row struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	query := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log AS l").
		Select("l.status, COUNT(*) AS count").
		Joins("INNER JOIN owners AS o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ?", clinicID, from, to).
		Group("l.status")
	if triggerType != "" {
		query = query.Where("l.trigger_type = ?", triggerType)
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
		Table("lstep_delivery_trigger_log AS l").
		Select("COALESCE(l.excluded_reason, '') AS excluded_reason, COUNT(*) AS count").
		Joins("INNER JOIN owners AS o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ? AND l.status = ?", clinicID, from, to, model.TriggerStatusExcluded).
		Group("l.excluded_reason")
	if triggerType != "" {
		query = query.Where("l.trigger_type = ?", triggerType)
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

func (r *lstepDeliveryTriggerLogRepository) CountSuppressedByPriorityDateRange(ctx context.Context, clinicID uint64, from, to time.Time, triggerType string) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log AS l").
		Joins("INNER JOIN owners AS o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ? AND l.suppressed_by_priority = TRUE", clinicID, from, to)
	if triggerType != "" {
		query = query.Where("l.trigger_type = ?", triggerType)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_suppressed_by_priority")
	}
	return count, nil
}

func (r *lstepDeliveryTriggerLogRepository) FindByDateRangeWithFilters(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error) {
	// Count through the same clinic-scoped owner join as the data query so the
	// pagination total cannot include orphaned or cross-clinic owner references.
	countQ := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log l").
		Joins("INNER JOIN owners o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ?", clinicID, from, to)
	if triggerType != "" {
		countQ = countQ.Where("l.trigger_type = ?", triggerType)
	}
	if status != "" {
		countQ = countQ.Where("l.status = ?", status)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_find")
	}

	// Data query with the clinic-scoped owner name join.
	findQ := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log l").
		Select("l.*, COALESCE(o.name, '') AS owner_name").
		Joins("INNER JOIN owners o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
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

func (r *lstepDeliveryTriggerLogRepository) CountByTypeAndStatus(ctx context.Context, clinicID uint64, from, to time.Time) ([]DeliveryStatsRow, error) {
	var rows []DeliveryStatsRow
	err := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log AS l").
		Select("l.trigger_type, l.status, COUNT(*) AS count").
		Joins("INNER JOIN owners AS o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Where("l.clinic_id = ? AND l.scheduled_at >= ? AND l.scheduled_at < ?", clinicID, from, to).
		Group("l.trigger_type, l.status").
		Order("l.trigger_type, l.status").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_by_type_status")
	}
	return rows, nil
}

func (r *lstepDeliveryTriggerLogRepository) CountVisitConversionsByType(ctx context.Context, clinicID uint64, from, to time.Time, days int) ([]VisitConversionRow, error) {
	var rows []VisitConversionRow
	err := r.db.WithContext(ctx).
		Table("lstep_delivery_trigger_log l").
		Joins("INNER JOIN owners AS o ON o.id = l.owner_id AND o.clinic_id = l.clinic_id").
		Select(`
			l.trigger_type,
			COUNT(*) AS delivered_count,
			SUM(
				CASE WHEN EXISTS (
					SELECT 1
					FROM medical_records mr
					WHERE mr.clinic_id = l.clinic_id
					  AND mr.owner_id = l.owner_id
					  AND mr.deleted_at IS NULL
					  AND mr.date >= DATE(l.fired_at AT TIME ZONE 'Asia/Tokyo')
					  AND mr.date <= DATE((l.fired_at AT TIME ZONE 'Asia/Tokyo') + make_interval(days => ?))
				) THEN 1 ELSE 0 END
			) AS visited_count
		`, days).
		Where(`
			l.clinic_id = ?
			AND l.fired_at IS NOT NULL
			AND l.fired_at >= ?
			AND l.fired_at < ?
			AND l.status = ?
			AND l.suppressed_by_priority = FALSE
		`, clinicID, from, to, model.TriggerStatusFired).
		Group("l.trigger_type").
		Order("l.trigger_type").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "count_visit_conversions")
	}
	return rows, nil
}

// FindByOwnerAndDate は同日同 owner_id の既存ログを返す (Q23 suppressed check 用)。
func (r *lstepDeliveryTriggerLogRepository) FindByOwnerAndDate(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	dayStart, dayEnd, _ := jstHalfOpenDay(date)
	var logs []model.LstepDeliveryTriggerLog
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ? AND scheduled_at >= ? AND scheduled_at < ?", clinicID, ownerID, dayStart, dayEnd).
		Find(&logs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_delivery_trigger_log", "find_by_owner_and_date")
	}
	return logs, nil
}

// UpdateSuppressed は既存ログを suppressed_by_priority=true に更新する (Q23 降格処理用)。
func (r *lstepDeliveryTriggerLogRepository) UpdateSuppressed(ctx context.Context, clinicID, logID uint64, reason string) error {
	result := r.db.WithContext(ctx).
		Model(&model.LstepDeliveryTriggerLog{}).
		Where("id = ? AND clinic_id = ?", logID, clinicID).
		Updates(map[string]any{
			"suppressed_by_priority": true,
			"suppression_reason":     reason,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_delivery_trigger_log", fmt.Sprintf("%d", logID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_delivery_trigger_log", fmt.Sprintf("%d", logID))
	}
	return nil
}
