package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationScheduleRepository はスタッフ個人スケジュール（shift_entries + shift_entry_breaks）のデータアクセスインターフェース
type ReservationScheduleRepository interface {
	FindAllByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error)
	// FindAllByStaffIDsAndDateRange は複数スタッフの指定期間内シフトエントリを一括取得する(G7-1: 日付ループN+1回避のプリフェッチ用)。
	FindAllByStaffIDsAndDateRange(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error)
	FindAllBreaksByEntryIDs(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
	FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	FindAllBreaksByEntryID(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error)
	Save(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error
	Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

type reservationScheduleRepository struct{ db *gorm.DB }

func NewReservationScheduleRepository(db *gorm.DB) ReservationScheduleRepository {
	return &reservationScheduleRepository{db: db}
}

func (r *reservationScheduleRepository) FindAllByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error) {
	t, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("month must be YYYY-MM format")
	}
	start := t
	end := t.AddDate(0, 1, 0)

	entries := make([]model.ShiftEntry, 0)
	err = r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("staff_id = ? AND date >= ? AND date < ?",
			staffID, start.Format(time.DateOnly), end.Format(time.DateOnly)).
		Order("date ASC").
		Find(&entries).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}

	return entries, nil
}

// FindAllByStaffIDsAndDateRange は [from, to) 半開区間・複数スタッフのシフトエントリを1クエリで返す(G7-1)。
// staffIDs が空の場合は空スライスを即返す(クエリを発行しない)。
func (r *reservationScheduleRepository) FindAllByStaffIDsAndDateRange(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error) {
	if len(staffIDs) == 0 {
		return []model.ShiftEntry{}, nil
	}
	entries := make([]model.ShiftEntry, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("staff_id IN ? AND date >= ? AND date < ?",
			staffIDs, from.Format(time.DateOnly), to.Format(time.DateOnly)).
		Order("date ASC").
		Find(&entries).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	return entries, nil
}

func (r *reservationScheduleRepository) FindAllBreaksByEntryIDs(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	if len(entryIDs) == 0 {
		return map[uint64][]model.ShiftEntryBreak{}, nil
	}
	var breaks []model.ShiftEntryBreak
	if err := r.db.WithContext(ctx).Where("shift_entry_id IN ?", entryIDs).Find(&breaks).Error; err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	result := make(map[uint64][]model.ShiftEntryBreak, len(entryIDs))
	for _, b := range breaks {
		result[b.ShiftEntryID] = append(result[b.ShiftEntryID], b)
	}
	return result, nil
}

func (r *reservationScheduleRepository) FindAllBreaksByEntryID(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error) {
	var breaks []model.ShiftEntryBreak
	if err := r.db.WithContext(ctx).Where("shift_entry_id = ?", entryID).Find(&breaks).Error; err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	return breaks, nil
}

func (r *reservationScheduleRepository) FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("staff_id = ? AND date = ?",
			staffID, date.Format(time.DateOnly)).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	return &entry, nil
}

// Save は ShiftEntry と ShiftEntryBreaks をトランザクションで upsert する
func (r *reservationScheduleRepository) Save(ctx context.Context, clinicID uint64, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存エントリを検索
		var existing model.ShiftEntry
		err := tx.Scopes(clinicScope(entry.ClinicID)).
			Where("staff_id = ? AND date = ?",
				entry.StaffID, entry.Date.Format(time.DateOnly)).
			First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("staff=%d", entry.StaffID))
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新規作成
			if err2 := tx.Create(entry).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry", "")
			}
		} else {
			// 更新
			entry.ID = existing.ID
			fields := map[string]any{
				"shift_type": entry.ShiftType,
				"start_time": entry.StartTime,
				"end_time":   entry.EndTime,
				"notes":      entry.Notes,
				"updated_at": gorm.Expr("NOW()"),
			}
			if err2 := tx.Scopes(clinicScope(entry.ClinicID)).
				Model(&model.ShiftEntry{}).Where("id = ?", existing.ID).Updates(fields).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry", fmt.Sprintf("%d", existing.ID))
			}
		}

		// 既存のbreaksを削除してから再作成
		if err2 := tx.Where("shift_entry_id = ?", entry.ID).Delete(&model.ShiftEntryBreak{}).Error; err2 != nil {
			return apperrors.FromGORM(err2, "shift_entry_break", fmt.Sprintf("%d", entry.ID))
		}
		if len(breaks) > 0 {
			for i := range breaks {
				breaks[i].ShiftEntryID = entry.ID
			}
			if err2 := tx.Create(&breaks).Error; err2 != nil {
				return apperrors.FromGORM(err2, "shift_entry_break", "")
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to upsert shift entry")
	}
	return nil
}

func (r *reservationScheduleRepository) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("staff_id = ? AND date = ?",
			staffID, date.Format(time.DateOnly)).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "schedule_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	return nil
}
