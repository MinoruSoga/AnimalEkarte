package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ShiftEntryFilter はシフト一覧取得のフィルタ条件
type ShiftEntryFilter struct {
	YearMonth string // "YYYY-MM" 形式
	StaffID   *uint64
}

// ShiftEntryRepository はシフト管理のデータアクセスインターフェース
type ShiftEntryRepository interface {
	FindAll(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	Create(ctx context.Context, entry *model.ShiftEntry) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error)
	ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error
}

type shiftEntryRepository struct{ db *gorm.DB }

// NewShiftEntryRepository はShiftEntryRepositoryを初期化して返す
func NewShiftEntryRepository(db *gorm.DB) ShiftEntryRepository {
	return &shiftEntryRepository{db: db}
}

func (r *shiftEntryRepository) FindAll(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error) {
	q := r.db.WithContext(ctx).
		Preload("Staff").
		Preload("Breaks").
		Scopes(clinicScope(clinicID)).
		Order("date ASC, staff_id ASC")

	if filter.YearMonth != "" {
		// YYYY-MM → start/end dates
		t, err := time.Parse("2006-01", filter.YearMonth)
		if err == nil {
			start := t
			end := t.AddDate(0, 1, 0)
			q = q.Where("date >= ? AND date < ?", start.Format("2006-01-02"), end.Format("2006-01-02"))
		}
	}
	if filter.StaffID != nil {
		q = q.Where("staff_id = ?", *filter.StaffID)
	}

	entries := make([]model.ShiftEntry, 0)
	if err := q.Find(&entries).Error; err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", "")
	}
	return entries, nil
}

func (r *shiftEntryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := r.db.WithContext(ctx).
		Preload("Staff").
		Preload("Breaks").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", strconv.FormatUint(id, 10))
	}
	return &entry, nil
}

func (r *shiftEntryRepository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		// PostgreSQL UNIQUE違反 (23505)
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("shift_entry",
				fmt.Sprintf("staff_id=%d date=%s", entry.StaffID, entry.Date.Format("2006-01-02")))
		}
		return apperrors.FromGORM(err, "shift_entry", "")
	}
	return nil
}

func (r *shiftEntryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ShiftEntry{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_entry", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftEntryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftEntry{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_entry", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_entry", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftEntryRepository) ReplaceBreaks(ctx context.Context, shiftEntryID uint64, breaks []model.ShiftEntryBreak) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("shift_entry_id = ?", shiftEntryID).Delete(&model.ShiftEntryBreak{}).Error; err != nil {
			return apperrors.FromGORM(err, "shift_entry_break", fmt.Sprintf("%d", shiftEntryID))
		}
		if len(breaks) == 0 {
			return nil
		}
		for i := range breaks {
			breaks[i].ShiftEntryID = shiftEntryID
		}
		if err := tx.Create(&breaks).Error; err != nil {
			return apperrors.FromGORM(err, "shift_entry_break", "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace shift entry breaks")
	}
	return nil
}

func (r *shiftEntryRepository) ExistsByStaffID(ctx context.Context, staffID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ShiftEntry{}).
		Where("staff_id = ?", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "shift_entry", "")
	}
	return count > 0, nil
}
