package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationScheduleRepository はスタッフ個人スケジュール（shift_entries + shift_entry_breaks）のデータアクセスインターフェース
type ReservationScheduleRepository interface {
	FindAllByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]model.ShiftEntry, error)
	// FindAllByStaffIDsAndDateRange は複数スタッフの指定期間内シフトエントリを一括取得する(G7-1: 日付ループN+1回避のプリフェッチ用)。
	FindAllByStaffIDsAndDateRange(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error)
	// clinicID correlates breaks to parent shift_entries.clinic_id (RSV-08 / SEC-SWEEP-02).
	FindAllBreaksByEntryIDs(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
	FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	FindAllBreaksByEntryID(ctx context.Context, clinicID, entryID uint64) ([]model.ShiftEntryBreak, error)
	Save(
		ctx context.Context,
		clinicID uint64,
		entry *model.ShiftEntry,
		breaks []model.ShiftEntryBreak,
	) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
	Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

type reservationScheduleRepository struct {
	db *gorm.DB
	// entries は shift_entries の唯一の書き込み者（staff domain・ADR-006 論点#1 案A）の
	// consumer-side view。read は対象外（裁定条件iii）。具象は repository facade が注入する。
	entries shiftWriter
}

// shiftWriter は staff domain（shiftentry）の予約スケジュール用途 write の最小 view。
type shiftWriter interface {
	SaveByStaffDate(
		ctx context.Context,
		clinicID uint64,
		entry *model.ShiftEntry,
		breaks []model.ShiftEntryBreak,
	) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error)
	DeleteByStaffDate(ctx context.Context, clinicID, staffID uint64, date time.Time) error
}

func NewReservationScheduleRepository(db *gorm.DB, entries shiftWriter) ReservationScheduleRepository {
	return &reservationScheduleRepository{db: db, entries: entries}
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
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where("staff_id IN ? AND date >= ? AND date < ?",
			staffIDs, from.Format(time.DateOnly), to.Format(time.DateOnly)).
		Order("date ASC").
		Find(&entries).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	return entries, nil
}

func (r *reservationScheduleRepository) FindAllBreaksByEntryIDs(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	if len(entryIDs) == 0 {
		return map[uint64][]model.ShiftEntryBreak{}, nil
	}
	var breaks []model.ShiftEntryBreak
	// RSV-08: JOIN string must be a string literal so grandchild lint can see parent clinic correlation.
	if err := r.db.WithContext(ctx).
		Model(&model.ShiftEntryBreak{}).
		Joins("JOIN shift_entries ON shift_entries.id = shift_entry_breaks.shift_entry_id AND shift_entries.clinic_id = ?", clinicID).
		Where("shift_entry_breaks.shift_entry_id IN ?", entryIDs).
		Find(&breaks).Error; err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	result := make(map[uint64][]model.ShiftEntryBreak, len(entryIDs))
	for _, b := range breaks {
		result[b.ShiftEntryID] = append(result[b.ShiftEntryID], b)
	}
	return result, nil
}

func (r *reservationScheduleRepository) FindAllBreaksByEntryID(ctx context.Context, clinicID, entryID uint64) ([]model.ShiftEntryBreak, error) {
	var breaks []model.ShiftEntryBreak
	if err := r.db.WithContext(ctx).
		Model(&model.ShiftEntryBreak{}).
		Joins("JOIN shift_entries ON shift_entries.id = shift_entry_breaks.shift_entry_id AND shift_entries.clinic_id = ?", clinicID).
		Where("shift_entry_breaks.shift_entry_id = ?", entryID).
		Find(&breaks).Error; err != nil {
		return nil, apperrors.FromGORM(err, "schedule_entry", "")
	}
	return breaks, nil
}

func (r *reservationScheduleRepository) FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	var entry model.ShiftEntry
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("staff_id = ? AND date = ?",
			staffID, date.Format(time.DateOnly)).
		First(&entry).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_entry", fmt.Sprintf("staff=%d date=%s", staffID, date.Format(time.DateOnly)))
	}
	return &entry, nil
}

// Save は shiftentry.SaveByStaffDate へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationScheduleRepository) Save(
	ctx context.Context,
	clinicID uint64,
	entry *model.ShiftEntry,
	breaks []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return r.entries.SaveByStaffDate(ctx, clinicID, entry, breaks)
}

// Delete は shiftentry.DeleteByStaffDate へ delegate する（ADR-006 論点#1 案A: 実装は staff domain 側）。
func (r *reservationScheduleRepository) Delete(ctx context.Context, clinicID, staffID uint64, date time.Time) error {
	return r.entries.DeleteByStaffDate(ctx, clinicID, staffID, date)
}
