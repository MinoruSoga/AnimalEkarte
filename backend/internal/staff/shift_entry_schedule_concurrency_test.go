package staff_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func mutateShiftStaffAssignment(
	db *gorm.DB,
	clinicID, staffID uint64,
) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var staff model.Staff
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", staffID).
			First(&staff).Error; err != nil {
			return err
		}
		result := tx.
			Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
			Delete(&model.StaffClinicAssignment{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("assignment mutation changed %d rows, want 1", result.RowsAffected)
		}
		return nil
	})
}

func TestRepository_SaveByStaffDate_SerializesAssignmentRevocation(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, db *gorm.DB, repo ShiftEntryRepository, clinicID, staffID uint64, date time.Time)
	}{
		{
			name: "revocation commits before write",
			run: func(t *testing.T, db *gorm.DB, repo ShiftEntryRepository, clinicID, staffID uint64, date time.Time) {
				require.NoError(t, mutateShiftStaffAssignment(db, clinicID, staffID))

				_, _, _, err := repo.SaveByStaffDate(context.Background(), clinicID, &model.ShiftEntry{
					ClinicID:  clinicID,
					StaffID:   staffID,
					Date:      date,
					ShiftType: model.ShiftTypeFull,
				}, nil)

				require.Error(t, err)
				var count int64
				require.NoError(t, db.Model(&model.ShiftEntry{}).
					Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicID, staffID, date).
					Count(&count).Error)
				assert.Zero(t, count)
			},
		},
		{
			name: "write holds ownership until transaction commits",
			run: func(t *testing.T, db *gorm.DB, repo ShiftEntryRepository, clinicID, staffID uint64, date time.Time) {
				holder := db.Begin()
				require.NoError(t, holder.Error)
				defer holder.Rollback()
				_, _, _, err := repo.SaveByStaffDate(
					persistence.WithTxValue(context.Background(), holder),
					clinicID,
					&model.ShiftEntry{
						ClinicID:  clinicID,
						StaffID:   staffID,
						Date:      date,
						ShiftType: model.ShiftTypeFull,
					},
					nil,
				)
				require.NoError(t, err)

				contender := db.Begin()
				require.NoError(t, contender.Error)
				defer contender.Rollback()
				require.NoError(t, contender.Exec("SET LOCAL lock_timeout = '100ms'").Error)
				var staff model.Staff
				contenderErr := contender.
					Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("id = ? AND deleted_at IS NULL", staffID).
					First(&staff).Error

				require.Error(t, contenderErr)
				assert.True(
					t,
					strings.Contains(contenderErr.Error(), "55P03") ||
						strings.Contains(contenderErr.Error(), "lock timeout"),
					"expected PostgreSQL lock timeout, got: %v",
					contenderErr,
				)
				require.NoError(t, contender.Rollback().Error)
				require.NoError(t, holder.Commit().Error)
				require.NoError(t, mutateShiftStaffAssignment(db, clinicID, staffID))

				var count int64
				require.NoError(t, db.Model(&model.ShiftEntry{}).
					Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicID, staffID, date).
					Count(&count).Error)
				assert.Equal(t, int64(1), count)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupShiftEntryTestDB(t)
			repo := NewShiftEntryRepository(db)
			const clinicID = uint64(1)
			staff := makeShiftEntryDoctor(t, db, clinicID, "revocation "+tt.name)
			tt.run(
				t,
				db,
				repo,
				clinicID,
				staff.ID,
				time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			)
		})
	}
}

func TestRepository_SaveDeleteByStaffDate_SerializeAbsentKey(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	const clinicID = uint64(1)
	staff := makeShiftEntryDoctor(t, db, clinicID, "absent key serialization")
	date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	holder := db.Begin()
	require.NoError(t, holder.Error)
	defer holder.Rollback()
	_, _, _, err := repo.SaveByStaffDate(
		persistence.WithTxValue(context.Background(), holder),
		clinicID,
		&model.ShiftEntry{
			ClinicID:  clinicID,
			StaffID:   staff.ID,
			Date:      date,
			ShiftType: model.ShiftTypeFull,
		},
		nil,
	)
	require.NoError(t, err)

	contender := db.Begin()
	require.NoError(t, contender.Error)
	defer contender.Rollback()
	require.NoError(t, contender.Exec("SET LOCAL lock_timeout = '100ms'").Error)
	deleteErr := repo.DeleteByStaffDate(
		persistence.WithTxValue(context.Background(), contender),
		clinicID,
		staff.ID,
		date,
	)

	require.Error(t, deleteErr)
	assert.True(
		t,
		strings.Contains(deleteErr.Error(), "55P03") ||
			strings.Contains(deleteErr.Error(), "lock timeout"),
		"delete must wait for the absent-key save ownership lock, got: %v",
		deleteErr,
	)
	require.NoError(t, contender.Rollback().Error)
	require.NoError(t, holder.Commit().Error)
	require.NoError(t, repo.DeleteByStaffDate(context.Background(), clinicID, staff.ID, date))
}

func TestRepository_DeleteByStaffDate_ReturnsNotFoundForAuthorizedAbsentKey(t *testing.T) {
	targetDate := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	otherDate := targetDate.AddDate(0, 0, 1)
	tests := []struct {
		name     string
		seedDate *time.Time
	}{
		{name: "no shift exists", seedDate: nil},
		{name: "only a different date exists", seedDate: &otherDate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupShiftEntryTestDB(t)
			repo := NewShiftEntryRepository(db)
			const clinicID = uint64(1)
			staff := makeShiftEntryDoctor(t, db, clinicID, "authorized absent delete")
			if tt.seedDate != nil {
				makeShiftEntryWithType(
					t,
					db,
					clinicID,
					staff.ID,
					*tt.seedDate,
					model.ShiftTypeFull,
				)
			}

			err := repo.DeleteByStaffDate(
				context.Background(),
				clinicID,
				staff.ID,
				targetDate,
			)

			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			if tt.seedDate != nil {
				var count int64
				require.NoError(t, db.Model(&model.ShiftEntry{}).
					Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicID, staff.ID, *tt.seedDate).
					Count(&count).Error)
				assert.Equal(t, int64(1), count, "a different date must remain untouched")
			}
		})
	}
}

func TestRepository_SaveByStaffDate_SerializesConcurrentAbsentUpserts(t *testing.T) {
	tests := []struct {
		name     string
		requests []struct {
			notes  string
			breaks []model.ShiftEntryBreak
		}
	}{
		{
			name: "both requests replace the full graph",
			requests: []struct {
				notes  string
				breaks []model.ShiftEntryBreak
			}{
				{
					notes: "first",
					breaks: []model.ShiftEntryBreak{{
						BreakStart: "11:00",
						BreakEnd:   "11:30",
					}},
				},
				{
					notes: "second",
					breaks: []model.ShiftEntryBreak{{
						BreakStart: "13:00",
						BreakEnd:   "14:00",
					}},
				},
			},
		},
		{
			name: "empty replacement remains exact",
			requests: []struct {
				notes  string
				breaks []model.ShiftEntryBreak
			}{
				{
					notes: "with break",
					breaks: []model.ShiftEntryBreak{{
						BreakStart: "12:00",
						BreakEnd:   "13:00",
					}},
				},
				{notes: "without break", breaks: []model.ShiftEntryBreak{}},
			},
		},
	}

	type outcome struct {
		saved   *model.ShiftEntry
		breaks  []model.ShiftEntryBreak
		created bool
		err     error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupShiftEntryTestDB(t)
			repo := NewShiftEntryRepository(db)
			const clinicID = uint64(1)
			staff := makeShiftEntryDoctor(t, db, clinicID, "concurrent absent upsert")
			date := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
			start := make(chan struct{})
			results := make(chan outcome, len(tt.requests))
			var wg sync.WaitGroup

			for _, request := range tt.requests {
				request := request
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					saved, savedBreaks, created, err := repo.SaveByStaffDate(
						ctx,
						clinicID,
						&model.ShiftEntry{
							ClinicID:  clinicID,
							StaffID:   staff.ID,
							Date:      date,
							ShiftType: model.ShiftTypeFull,
							Notes:     request.notes,
						},
						request.breaks,
					)
					results <- outcome{
						saved:   saved,
						breaks:  savedBreaks,
						created: created,
						err:     err,
					}
				}()
			}
			close(start)
			wg.Wait()
			close(results)

			createdCount := 0
			var updateResult outcome
			for result := range results {
				require.NoError(t, result.err)
				require.NotNil(t, result.saved)
				if result.created {
					createdCount++
				} else {
					updateResult = result
				}
			}
			require.Equal(t, 1, createdCount, "exactly one absent-key request must report creation")
			require.NotNil(t, updateResult.saved, "the serialized second request must report update")

			var persisted []model.ShiftEntry
			require.NoError(t, db.
				Preload("Breaks").
				Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicID, staff.ID, date).
				Find(&persisted).Error)
			require.Len(t, persisted, 1)
			assert.Equal(t, updateResult.saved.ID, persisted[0].ID)
			assert.Equal(t, updateResult.saved.Notes, persisted[0].Notes)
			require.Len(t, persisted[0].Breaks, len(updateResult.breaks))
			for i := range updateResult.breaks {
				assert.Equal(t, updateResult.breaks[i].ID, persisted[0].Breaks[i].ID)
				assert.Equal(t, updateResult.breaks[i].BreakStart, persisted[0].Breaks[i].BreakStart)
				assert.Equal(t, updateResult.breaks[i].BreakEnd, persisted[0].Breaks[i].BreakEnd)
			}
		})
	}
}
