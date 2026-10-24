package staff_test

// shift_entry_schedule_unknown_source_test.go — 予約スケジュール経路（SaveByStaffDate）の source_unresolved 停止
//
// 保護する不変条件: 既存の勤務行またはその休憩が interval_state=unknown（NULL 日時、off / paid_leave に
//   矛盾して残る日時・休憩）のとき、SaveByStaffDate は PATCH と同じ 409 source_unresolved で拒否し、
//   親・休憩のどちらにも書き込まない（第二の writer から unknown 行を黙って上書きさせない）。

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

// makeKnownWorkingShiftEntry は interval_state=known の全日勤務（09:00-18:00・休憩なし）を作る。
func makeKnownWorkingShiftEntry(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time) *model.ShiftEntry {
	t.Helper()
	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
		StartsAt:  shiftTS(t, date, "09:00"),
		EndsAt:    shiftTS(t, date, "18:00"),
	}
	require.NoError(t, db.Create(entry).Error)
	return entry
}

// seedShiftBreak は休憩行を直接作る。日時は NULL を含められる（unknown 行の再現用）。
func seedShiftBreak(t *testing.T, db *gorm.DB, entryID uint64, startsAt, endsAt *time.Time) {
	t.Helper()
	require.NoError(t, db.Create(&model.ShiftEntryBreak{
		ShiftEntryID: entryID,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
	}).Error)
}

// shiftGraphSnapshot は親と休憩（id 昇順）を DB から読み直し、書込有無の比較用文字列にする。
func shiftGraphSnapshot(t *testing.T, db *gorm.DB, entryID uint64) string {
	t.Helper()
	var entry model.ShiftEntry
	require.NoError(t, db.
		Preload("Breaks", func(tx *gorm.DB) *gorm.DB { return tx.Order("id ASC") }).
		First(&entry, entryID).Error)
	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"type=%s notes=%q version=%d updated=%d starts=%q ends=%q",
		entry.ShiftType, entry.Notes, entry.Version, entry.UpdatedAt.UnixNano(),
		shiftClock(entry.StartsAt), shiftClock(entry.EndsAt),
	)
	for _, b := range entry.Breaks {
		fmt.Fprintf(&sb, " break#%d[%q-%q]", b.ID, shiftClock(b.StartsAt), shiftClock(b.EndsAt))
	}
	return sb.String()
}

// TestRepository_SaveByStaffDate_StopsOnUnknownExistingInterval は、既存の親または休憩が unknown の行を
// 新しい形で上書きしようとしても、source_unresolved で拒否し親・休憩とも不変であることを検証する。
func TestRepository_SaveByStaffDate_StopsOnUnknownExistingInterval(t *testing.T) {
	const clinicID = uint64(1)
	date := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		seed func(t *testing.T, db *gorm.DB, staffID uint64) uint64
	}{
		{
			name: "parent with NULL starts_at and ends_at",
			seed: func(t *testing.T, db *gorm.DB, staffID uint64) uint64 {
				entry := &model.ShiftEntry{
					ClinicID:  clinicID,
					StaffID:   staffID,
					Date:      date,
					ShiftType: model.ShiftTypeFull,
					Notes:     "before",
				}
				require.NoError(t, db.Create(entry).Error)
				seedShiftBreak(t, db, entry.ID, shiftTS(t, date, "12:00"), shiftTS(t, date, "13:00"))
				return entry.ID
			},
		},
		{
			name: "break with a NULL ends_at",
			seed: func(t *testing.T, db *gorm.DB, staffID uint64) uint64 {
				entry := makeKnownWorkingShiftEntry(t, db, clinicID, staffID, date)
				seedShiftBreak(t, db, entry.ID, shiftTS(t, date, "12:00"), nil)
				seedShiftBreak(t, db, entry.ID, shiftTS(t, date, "14:00"), shiftTS(t, date, "15:00"))
				return entry.ID
			},
		},
		{
			name: "contradictory off entry with leftover timestamps and break",
			seed: func(t *testing.T, db *gorm.DB, staffID uint64) uint64 {
				entry := &model.ShiftEntry{
					ClinicID:  clinicID,
					StaffID:   staffID,
					Date:      date,
					ShiftType: model.ShiftTypeOff,
					Notes:     "before",
					StartsAt:  shiftTS(t, date, "09:00"),
					EndsAt:    shiftTS(t, date, "18:00"),
				}
				require.NoError(t, db.Create(entry).Error)
				seedShiftBreak(t, db, entry.ID, shiftTS(t, date, "12:00"), shiftTS(t, date, "13:00"))
				return entry.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupShiftEntryTestDB(t)
			repo := NewShiftEntryRepository(db)
			staff := makeShiftEntryDoctor(t, db, clinicID, "unknown source "+tt.name)
			entryID := tt.seed(t, db, staff.ID)
			before := shiftGraphSnapshot(t, db, entryID)

			saved, savedBreaks, created, err := repo.SaveByStaffDate(
				context.Background(),
				clinicID,
				&model.ShiftEntry{
					ClinicID:  clinicID,
					StaffID:   staff.ID,
					Date:      date,
					ShiftType: model.ShiftTypeMorning,
					Notes:     "after",
					StartsAt:  shiftTS(t, date, "09:00"),
					EndsAt:    shiftTS(t, date, "12:00"),
				},
				[]model.ShiftEntryBreak{shiftBreakAt(t, date, "10:00", "10:15")},
			)

			require.Error(t, err)
			assert.True(t, apperrors.IsConflict(err), "unknown 行は 409 で拒否される: %v", err)
			var appErr *apperrors.AppError
			require.True(t, errors.As(err, &appErr))
			assert.Equal(t, "source_unresolved", appErr.Code)
			assert.Nil(t, saved)
			assert.Nil(t, savedBreaks)
			assert.False(t, created)
			assert.Equal(t, before, shiftGraphSnapshot(t, db, entryID), "親・休憩とも書き込まれない")
		})
	}
}
