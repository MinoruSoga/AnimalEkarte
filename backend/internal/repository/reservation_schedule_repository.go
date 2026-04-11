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
	FindByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error)
	FindBreaksByEntryIDs(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
	FindByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	FindBreaksByEntryID(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error)
	Upsert(ctx context.Context, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error
	DeleteByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

type reservationScheduleRepository struct{ db *gorm.DB }

func NewReservationScheduleRepository(db *gorm.DB) ReservationScheduleRepository {
	return &reservationScheduleRepository{db: db}
}

func (r *reservationScheduleRepository) FindByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("month must be YYYY-MM format")
	}
	start := t
	end := t.AddDate(0, 1, 0)

	entries := make([]model.ShiftEntry, 0)
	err = r.db.WithContext(ctx).
		Where("clinic_id = ? AND staff_id = ? AND date >= ? AND date < ?",
			clinicID, staffID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("date ASC").
		Find(&entries).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}

	return entries, nil
}

func (r *reservationScheduleRepository) FindBreaksByEntryIDs(ctx context.Context, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
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

func (r *reservationScheduleRepository) FindBreaksByEntryID(ctx context.Context, entryID uint64) ([]model.ShiftEntryBreak, error) {
	var breaks []model.ShiftEntryBreak
	if err := r.db.WithContext(ctx).Where("shift_entry_id = ?", entryID).Find(&breaks).Error; err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	return breaks, nil
}

func (r *reservationScheduleRepository) FindByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND staff_id = ? AND date = ?",
			clinicID, staffID, date.Format("2006-01-02")).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format("2006-01-02")))
	}
	return &entry, nil
}

// Upsert は ShiftEntry と ShiftEntryBreaks をトランザクションで upsert する
func (r *reservationScheduleRepository) Upsert(ctx context.Context, entry *model.ShiftEntry, breaks []model.ShiftEntryBreak) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存エントリを検索
		var existing model.ShiftEntry
		err := tx.Where("clinic_id = ? AND staff_id = ? AND date = ?",
			entry.ClinicID, entry.StaffID, entry.Date.Format("2006-01-02")).
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
				"updated_at": entry.UpdatedAt,
			}
			if err2 := tx.Model(&model.ShiftEntry{}).Where("id = ?", existing.ID).Updates(fields).Error; err2 != nil {
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
	})
}

func (r *reservationScheduleRepository) DeleteByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Where("clinic_id = ? AND staff_id = ? AND date = ?",
			clinicID, staffID, date.Format("2006-01-02")).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "schedule_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format("2006-01-02")))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format("2006-01-02")))
	}
	return nil
}
