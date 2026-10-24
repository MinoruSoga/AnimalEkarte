package staff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestShiftEntryService_Update_WorkingToNonWorkingRequiresExplicitClearing は、勤務から
// off / paid_leave への変更が starts_at / ends_at の明示 null と breaks の明示空配列を
// 要求することを検証する（api.yaml UpdateShiftRequest）。breaks の省略は、保存済みの
// 休憩が既に空でも 400 で、何も保存しない。
func TestShiftEntryService_Update_WorkingToNonWorkingRequiresExplicitClearing(t *testing.T) {
	emptyBreaks := []ShiftBreakInput{}
	cases := []struct {
		name        string
		input       func(shiftType string) *UpdateShiftEntryInput
		wantInvalid bool
	}{
		{
			name: "breaks omitted is rejected even though stored breaks are empty",
			input: func(shiftType string) *UpdateShiftEntryInput {
				return &UpdateShiftEntryInput{
					ShiftType: &shiftType,
					StartsAt:  OptionalShiftTime{Set: true},
					EndsAt:    OptionalShiftTime{Set: true},
				}
			},
			wantInvalid: true,
		},
		{
			name: "both timestamps null and breaks empty array is accepted",
			input: func(shiftType string) *UpdateShiftEntryInput {
				return &UpdateShiftEntryInput{
					ShiftType: &shiftType,
					StartsAt:  OptionalShiftTime{Set: true},
					EndsAt:    OptionalShiftTime{Set: true},
					Breaks:    &emptyBreaks,
				}
			},
		},
		{
			name: "breaks empty array without clearing timestamps is rejected",
			input: func(shiftType string) *UpdateShiftEntryInput {
				return &UpdateShiftEntryInput{ShiftType: &shiftType, Breaks: &emptyBreaks}
			},
			wantInvalid: true,
		},
		{
			name: "breaks empty array with only starts_at cleared is rejected",
			input: func(shiftType string) *UpdateShiftEntryInput {
				return &UpdateShiftEntryInput{
					ShiftType: &shiftType,
					StartsAt:  OptionalShiftTime{Set: true},
					Breaks:    &emptyBreaks,
				}
			},
			wantInvalid: true,
		},
		{
			name: "breaks empty array with only ends_at cleared is rejected",
			input: func(shiftType string) *UpdateShiftEntryInput {
				return &UpdateShiftEntryInput{
					ShiftType: &shiftType,
					EndsAt:    OptionalShiftTime{Set: true},
					Breaks:    &emptyBreaks,
				}
			},
			wantInvalid: true,
		},
	}

	for _, target := range []model.ShiftType{model.ShiftTypeOff, model.ShiftTypePaidLeave} {
		for _, tc := range cases {
			t.Run(string(target)+"/"+tc.name, func(t *testing.T) {
				updates, replaces := 0, 0
				findCalls := 0
				repo := &mockShiftEntryRepository{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
						findCalls++
						if findCalls == 1 {
							return storedFullShift(t, 1), nil
						}
						return &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: target, Version: 2}, nil
					},
					updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
						updates++
						return nil
					},
					replaceBreaksFn: func(_ context.Context, _ uint64, breaks []model.ShiftEntryBreak) error {
						replaces++
						assert.Empty(t, breaks, "off / paid_leave の休憩は空配列で置換される")
						return nil
					},
				}
				svc := newTestShiftEntryService(repo)

				entry, err := svc.Update(context.Background(), 1, 1, tc.input(string(target)))

				if tc.wantInvalid {
					require.Error(t, err)
					assert.True(t, apperrors.IsInvalidInput(err), "unexpected error: %v", err)
					assert.Nil(t, entry)
					assert.Zero(t, updates+replaces, "400 では何も保存しない")
					return
				}
				require.NoError(t, err)
				require.NotNil(t, entry)
				assert.Equal(t, 1, updates)
				assert.Equal(t, 1, replaces)
			})
		}
	}
}

// TestShiftEntryService_Update_NonWorkingToNonWorkingNeedsNoExplicitClearing は、
// 既に終日非勤務の行の off <-> paid_leave 変更には明示 null / 空配列を要求しないことを検証する。
func TestShiftEntryService_Update_NonWorkingToNonWorkingNeedsNoExplicitClearing(t *testing.T) {
	findCalls := 0
	repo := &mockShiftEntryRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
			findCalls++
			if findCalls == 1 {
				return &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: model.ShiftTypeOff, Version: 1}, nil
			}
			return &model.ShiftEntry{ID: 1, Date: testShiftDate, ShiftType: model.ShiftTypePaidLeave, Version: 2}, nil
		},
	}
	svc := newTestShiftEntryService(repo)
	paidLeave := string(model.ShiftTypePaidLeave)

	entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{ShiftType: &paidLeave})

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, model.ShiftTypePaidLeave, entry.ShiftType)
}
