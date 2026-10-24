package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// shiftConflictCode は 409 の固定 code（api.yaml ShiftContractError.code）を取り出す。
func shiftConflictCode(t *testing.T, err error) string {
	t.Helper()
	require.Error(t, err)
	require.True(t, apperrors.IsConflict(err), "unexpected error: %v", err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	return appErr.Code
}

// storedFullShift は保存済みの既知の全日勤務（09:00-18:00）を返す。
func storedFullShift(t *testing.T, version int64) *model.ShiftEntry {
	t.Helper()
	return &model.ShiftEntry{
		ID:        1,
		ClinicID:  1,
		Date:      testShiftDate,
		ShiftType: model.ShiftTypeFull,
		Version:   version,
		StartsAt:  shiftTS(t, testShiftDate, "09:00"),
		EndsAt:    shiftTS(t, testShiftDate, "18:00"),
	}
}

func TestShiftEntryService_Update_RejectsStaleExpectedVersion(t *testing.T) {
	// Arrange
	writes := 0
	repo := &mockShiftEntryRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
			return storedFullShift(t, 5), nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
			writes++
			return nil
		},
	}
	svc := newTestShiftEntryService(repo)
	staleVersion := int64(4)
	notes := "memo"

	// Act
	entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{Notes: &notes, ExpectedVersion: &staleVersion})

	// Assert
	assert.Nil(t, entry)
	assert.Equal(t, "version_conflict", shiftConflictCode(t, err))
	assert.Zero(t, writes, "版不一致では何も保存しない")
}

func TestShiftEntryService_Update_AcceptsTheCurrentExpectedVersion(t *testing.T) {
	// Arrange
	repo := &mockShiftEntryRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
			return storedFullShift(t, 5), nil
		},
	}
	svc := newTestShiftEntryService(repo)
	currentVersion := int64(5)
	notes := "memo"

	// Act
	entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{Notes: &notes, ExpectedVersion: &currentVersion})

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, entry)
}

func TestShiftEntryService_Update_StopsOnUnknownStoredInterval(t *testing.T) {
	stored := []struct {
		name  string
		entry *model.ShiftEntry
	}{
		{name: "working shift with NULL timestamps", entry: &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: model.ShiftTypeFull}},
		{name: "known shift with a break whose times are NULL", entry: &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: model.ShiftTypeFull, StartsAt: shiftTS(t, testShiftDate, "09:00"), EndsAt: shiftTS(t, testShiftDate, "18:00"), Breaks: []model.ShiftEntryBreak{{ID: 9}}}},
		{name: "off shift with leftover timestamps", entry: &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: model.ShiftTypeOff, StartsAt: shiftTS(t, testShiftDate, "09:00"), EndsAt: shiftTS(t, testShiftDate, "18:00")}},
	}
	notes := "memo"
	updates := []struct {
		name  string
		input *UpdateShiftEntryInput
	}{
		{name: "notes only", input: &UpdateShiftEntryInput{Notes: &notes}},
		{name: "explicit corrected times", input: &UpdateShiftEntryInput{
			StartsAt: OptionalShiftTime{Set: true, Value: shiftTS(t, testShiftDate, "09:00")},
			EndsAt:   OptionalShiftTime{Set: true, Value: shiftTS(t, testShiftDate, "18:00")},
		}},
	}

	for _, tt := range stored {
		for _, update := range updates {
			t.Run(tt.name+"/"+update.name, func(t *testing.T) {
				writes := 0
				repo := &mockShiftEntryRepository{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
						return tt.entry, nil
					},
					updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
						writes++
						return nil
					},
					replaceBreaksFn: func(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
						writes++
						return nil
					},
				}
				svc := newTestShiftEntryService(repo)

				entry, err := svc.Update(context.Background(), 1, 1, update.input)

				assert.Nil(t, entry)
				assert.Equal(t, "source_unresolved", shiftConflictCode(t, err))
				assert.Zero(t, writes, "unknown の勤務は訂正確認まで保存を停止する")
			})
		}
	}
}
