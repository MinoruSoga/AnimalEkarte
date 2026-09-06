package staff

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestShiftEntryService_Update(t *testing.T) {
	newShiftType := string(model.ShiftTypeAfternoon)
	newStartTime := "15:00"
	newEndTime := "19:00"
	newNote := "Updated note"
	// BUG-036: Full 勤務の既存行は start/end を持つ前提で検証する
	existingStart := "09:00:00"
	existingEnd := "18:00:00"

	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		input           *UpdateShiftEntryInput
		repoUpdateErr   error
		repoReturnEntry *model.ShiftEntry
		wantErr         bool
	}{
		{
			name:     "updates shift entry successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateShiftEntryInput{
				ShiftType: &newShiftType,
				StartTime: &newStartTime,
				EndTime:   &newEndTime,
				Notes:     &newNote,
			},
			repoUpdateErr: nil,
			repoReturnEntry: &model.ShiftEntry{
				ID:        1,
				ShiftType: model.ShiftType(newShiftType),
			},
			wantErr: false,
		},
		{
			name:            "returns error when no fields provided",
			clinicID:        1,
			id:              1,
			input:           &UpdateShiftEntryInput{},
			repoUpdateErr:   nil,
			repoReturnEntry: nil,
			wantErr:         true,
		},
		{
			name:     "returns error when update fails",
			clinicID: 1,
			id:       1,
			input: &UpdateShiftEntryInput{
				Notes: &newNote,
			},
			repoUpdateErr:   errors.New("db error"),
			repoReturnEntry: nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FindByID は2回呼ばれる: 1回目はバリデーション用（既存レコード取得）、2回目は更新後の取得
			// バリデーション用に常に有効なエントリを返す基底エントリを用意する
			baseEntry := &model.ShiftEntry{
				ID:        1,
				ClinicID:  1,
				ShiftType: model.ShiftTypeFull,
				StartTime: &existingStart,
				EndTime:   &existingEnd,
			}
			callCount := 0
			repo := &mockShiftEntryRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
					return tt.repoUpdateErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
					callCount++
					if callCount == 1 {
						// 1回目: バリデーション用の既存レコード
						return baseEntry, nil
					}
					// 2回目: 更新後のレコード取得
					return tt.repoReturnEntry, nil
				},
			}
			svc := newTestShiftEntryService(repo)

			entry, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
			}
		})
	}
}

func TestShiftEntryService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "deletes shift entry successfully",
			clinicID: 1,
			id:       1,
			repoErr:  nil,
			wantErr:  false,
		},
		{
			name:     "returns error when shift not found",
			clinicID: 1,
			id:       999,
			repoErr:  errors.New("not found"),
			wantErr:  true,
		},
		{
			name:     "returns error when delete fails",
			clinicID: 1,
			id:       1,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newTestShiftEntryService(repo)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestShiftEntryService_Delete_FindByIDError は P1: FindByID が失敗した場合、
// Delete 自体は呼ばれずに「find」文脈のエラーで返ることを検証する。
func TestShiftEntryService_Delete_FindByIDError(t *testing.T) {
	deleteCalled := false
	repo := &mockShiftEntryRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
			return nil, errors.New("not found")
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	svc := newTestShiftEntryService(repo)

	err := svc.Delete(context.Background(), 1, 999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find shift entry")
	assert.False(t, deleteCalled, "Delete should not be called when FindByID fails")
}

// ---- GetOnDutyStaffs ----

func TestShiftEntryService_GetOnDutyStaffs(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		repoStaffs []model.Staff
		repoErr    error
		wantLen    int
		wantErr    bool
	}{
		{
			name: "returns on-duty staffs",
			repoStaffs: []model.Staff{
				{ID: 1, Name: "Staff A"},
				{ID: 2, Name: "Staff B"},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:       "returns empty list when no staff on duty",
			repoStaffs: []model.Staff{},
			wantLen:    0,
			wantErr:    false,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				findOnDutyStaffsFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Staff, error) {
					return tt.repoStaffs, tt.repoErr
				},
			}
			svc := newTestShiftEntryService(repo)

			staffs, err := svc.GetOnDutyStaffs(context.Background(), 1, date)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, staffs, tt.wantLen)
			}
		})
	}
}

// ---- normalizeTimeString ----

func TestNormalizeTimeString(t *testing.T) {
	hhmmss := "09:30:00"
	hhmm := "09:30"
	invalid := "not-a-time"
	empty := ""

	tests := []struct {
		name string
		in   *string
		want *string
	}{
		{name: "nil input returns nil", in: nil, want: nil},
		{name: "empty string returns nil", in: &empty, want: nil},
		{name: "HH:MM:SS format normalizes to itself", in: &hhmmss, want: &hhmmss},
		{name: "HH:MM format normalizes to HH:MM:SS", in: &hhmm, want: &hhmmss},
		{name: "invalid format returns nil", in: &invalid, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTimeString(tt.in)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

// ---- validateShiftType ----

func TestValidateShiftType(t *testing.T) {
	tests := []struct {
		name    string
		shift   model.ShiftType
		wantErr bool
	}{
		{name: "full is valid", shift: model.ShiftTypeFull, wantErr: false},
		{name: "morning is valid", shift: model.ShiftTypeMorning, wantErr: false},
		{name: "afternoon is valid", shift: model.ShiftTypeAfternoon, wantErr: false},
		{name: "off is valid", shift: model.ShiftTypeOff, wantErr: false},
		{name: "paid_leave is valid", shift: model.ShiftTypePaidLeave, wantErr: false},
		{name: "unknown value is invalid", shift: model.ShiftType("bogus"), wantErr: true},
		{name: "empty value is invalid", shift: model.ShiftType(""), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShiftType(tt.shift)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- validateShiftTimes ----

func TestValidateShiftTimes(t *testing.T) {
	early := "09:00:00"
	late := "18:00:00"
	same := "09:00:00"
	invalidFormat := "9am"

	tests := []struct {
		name      string
		shiftType model.ShiftType
		start     *string
		end       *string
		wantErr   bool
	}{
		{
			name:      "off type skips validation regardless of times",
			shiftType: model.ShiftTypeOff,
			start:     &late,
			end:       &early,
			wantErr:   false,
		},
		{
			name:      "paid_leave type skips validation regardless of times",
			shiftType: model.ShiftTypePaidLeave,
			start:     &late,
			end:       &early,
			wantErr:   false,
		},
		{
			// BUG-036: 勤務種別では start/end 必須（nil は拒否）
			name:      "nil start time is required for working shift",
			shiftType: model.ShiftTypeFull,
			start:     nil,
			end:       &late,
			wantErr:   true,
		},
		{
			// BUG-036: 勤務種別では start/end 必須（nil は拒否）
			name:      "nil end time is required for working shift",
			shiftType: model.ShiftTypeFull,
			start:     &early,
			end:       nil,
			wantErr:   true,
		},
		{
			// BUG-036: 両方 nil も拒否
			name:      "both nil times are required for working shift",
			shiftType: model.ShiftTypeFull,
			start:     nil,
			end:       nil,
			wantErr:   true,
		},
		{
			name:      "valid start before end",
			shiftType: model.ShiftTypeFull,
			start:     &early,
			end:       &late,
			wantErr:   false,
		},
		{
			name:      "end equal to start is invalid",
			shiftType: model.ShiftTypeFull,
			start:     &same,
			end:       &same,
			wantErr:   true,
		},
		{
			name:      "end before start is invalid",
			shiftType: model.ShiftTypeFull,
			start:     &late,
			end:       &early,
			wantErr:   true,
		},
		{
			name:      "invalid start_time format",
			shiftType: model.ShiftTypeFull,
			start:     &invalidFormat,
			end:       &late,
			wantErr:   true,
		},
		{
			name:      "invalid end_time format",
			shiftType: model.ShiftTypeFull,
			start:     &early,
			end:       &invalidFormat,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShiftTimes(tt.shiftType, tt.start, tt.end)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- Create: additional branch coverage ----

func TestShiftEntryService_Create_AdditionalBranches(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "09:00:00"
	endTime := "18:00:00"

	t.Run("returns invalid input for unknown shift type", func(t *testing.T) {
		repo := &mockShiftEntryRepository{}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Create(context.Background(), 1, &CreateShiftEntryInput{
			StaffID:   1,
			Date:      date,
			ShiftType: "bogus",
		})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("saves breaks when provided", func(t *testing.T) {
		var savedEntryID uint64
		var savedBreaks []model.ShiftEntryBreak
		repo := &mockShiftEntryRepository{
			createFn: func(_ context.Context, entry *model.ShiftEntry) error {
				entry.ID = 42
				return nil
			},
			replaceBreaksFn: func(_ context.Context, entryID uint64, breaks []model.ShiftEntryBreak) error {
				savedEntryID = entryID
				savedBreaks = breaks
				return nil
			},
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: id}, nil
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Create(context.Background(), 1, &CreateShiftEntryInput{
			StaffID:   1,
			Date:      date,
			ShiftType: string(model.ShiftTypeFull),
			StartTime: &startTime,
			EndTime:   &endTime,
			Breaks: []ShiftBreakInput{
				{BreakStart: "12:00:00", BreakEnd: "13:00:00"},
			},
		})

		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.Equal(t, uint64(42), savedEntryID)
		require.Len(t, savedBreaks, 1)
		assert.Equal(t, "12:00:00", savedBreaks[0].BreakStart)
	})

	t.Run("returns error when saving breaks fails", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			createFn: func(_ context.Context, entry *model.ShiftEntry) error {
				entry.ID = 1
				return nil
			},
			replaceBreaksFn: func(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
				return errors.New("db error")
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Create(context.Background(), 1, &CreateShiftEntryInput{
			StaffID:   1,
			Date:      date,
			ShiftType: string(model.ShiftTypeFull),
			StartTime: &startTime,
			EndTime:   &endTime,
			Breaks: []ShiftBreakInput{
				{BreakStart: "12:00:00", BreakEnd: "13:00:00"},
			},
		})

		require.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("returns error when fetching entry after create fails", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			createFn: func(_ context.Context, entry *model.ShiftEntry) error {
				entry.ID = 1
				return nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
				return nil, errors.New("not found")
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Create(context.Background(), 1, &CreateShiftEntryInput{
			StaffID:   1,
			Date:      date,
			ShiftType: string(model.ShiftTypeFull),
			StartTime: &startTime,
			EndTime:   &endTime,
		})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "failed to get shift entry after create")
	})
}

// ---- Update: additional branch coverage ----

func TestShiftEntryService_Update_AdditionalBranches(t *testing.T) {
	newStartTime := "20:00:00"
	newEndTime := "19:00:00" // end < start を意図的に指定
	newNote := "note"
	invalidShiftType := "bogus"
	// BUG-036: 勤務種別の既存行は有効な start/end を持つ
	existingStart := "09:00:00"
	existingEnd := "18:00:00"

	t.Run("returns error when FindByID fails before update", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
				return nil, errors.New("not found")
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{Notes: &newNote})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "failed to lock shift entry for update")
	})

	t.Run("returns invalid input for unknown shift type", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{
					ID: 1, ShiftType: model.ShiftTypeFull,
					StartTime: &existingStart, EndTime: &existingEnd,
				}, nil
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{ShiftType: &invalidShiftType})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("returns error when effective times are invalid", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{
					ID: 1, ShiftType: model.ShiftTypeFull,
					StartTime: &existingStart, EndTime: &existingEnd,
				}, nil
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{
			StartTime: &newStartTime,
			EndTime:   &newEndTime,
		})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "failed to validate shift times")
	})

	t.Run("saves breaks when provided", func(t *testing.T) {
		var savedBreaks []model.ShiftEntryBreak
		callCount := 0
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				callCount++
				return &model.ShiftEntry{
					ID: id, ShiftType: model.ShiftTypeFull,
					StartTime: &existingStart, EndTime: &existingEnd,
				}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
				return nil
			},
			replaceBreaksFn: func(_ context.Context, _ uint64, breaks []model.ShiftEntryBreak) error {
				savedBreaks = breaks
				return nil
			},
		}
		svc := newTestShiftEntryService(repo)

		breaks := []ShiftBreakInput{{BreakStart: "12:00:00", BreakEnd: "13:00:00"}}
		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{
			Notes:  &newNote,
			Breaks: &breaks,
		})

		require.NoError(t, err)
		require.NotNil(t, entry)
		require.Len(t, savedBreaks, 1)
		assert.Equal(t, 2, callCount)
	})

	t.Run("accepts breaks only update", func(t *testing.T) {
		breaks := []ShiftBreakInput{{BreakStart: "12:00:00", BreakEnd: "13:00:00"}}
		updateCalled := false
		replaceCalled := false
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: id, ShiftType: model.ShiftTypeOff}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
				updateCalled = true
				return nil
			},
			replaceBreaksFn: func(_ context.Context, _ uint64, got []model.ShiftEntryBreak) error {
				replaceCalled = true
				require.Len(t, got, 1)
				return nil
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{Breaks: &breaks})

		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.False(t, updateCalled)
		assert.True(t, replaceCalled)
	})

	t.Run("rolls back parent update when break replacement fails", func(t *testing.T) {
		state := &shiftUpdateTxState{notes: "before"}
		tx := &rollbackShiftUpdateTransactor{state: state}
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{ID: id, ShiftType: model.ShiftTypeOff, Notes: state.notes}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, cmd UpdateShiftEntryInput) error {
				if cmd.Notes != nil {
					state.notes = *cmd.Notes
				}
				return nil
			},
			replaceBreaksFn: func(_ context.Context, _ uint64, breaks []model.ShiftEntryBreak) error {
				state.breaks = append([]model.ShiftEntryBreak(nil), breaks...)
				return errors.New("replace breaks failed")
			},
		}
		svc := NewShiftEntryService(
			repo,
			&mockShiftEntryStaffLocker{},
			&mockShiftEntryStaffAssignments{},
			tx,
		)
		breaks := []ShiftBreakInput{{BreakStart: "12:00:00", BreakEnd: "13:00:00"}}

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{
			Notes:  &newNote,
			Breaks: &breaks,
		})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Equal(t, 1, tx.calls)
		assert.Equal(t, "before", state.notes)
		assert.Empty(t, state.breaks)
	})

	t.Run("returns error when saving breaks fails", func(t *testing.T) {
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				return &model.ShiftEntry{
					ID: id, ShiftType: model.ShiftTypeFull,
					StartTime: &existingStart, EndTime: &existingEnd,
				}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
				return nil
			},
			replaceBreaksFn: func(_ context.Context, _ uint64, _ []model.ShiftEntryBreak) error {
				return errors.New("db error")
			},
		}
		svc := newTestShiftEntryService(repo)

		breaks := []ShiftBreakInput{{BreakStart: "12:00:00", BreakEnd: "13:00:00"}}
		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{
			Notes:  &newNote,
			Breaks: &breaks,
		})

		require.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("returns error when fetching entry after update fails", func(t *testing.T) {
		callCount := 0
		repo := &mockShiftEntryRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.ShiftEntry, error) {
				callCount++
				if callCount == 1 {
					return &model.ShiftEntry{
						ID: id, ShiftType: model.ShiftTypeFull,
						StartTime: &existingStart, EndTime: &existingEnd,
					}, nil
				}
				return nil, errors.New("not found")
			},
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateShiftEntryInput) error {
				return nil
			},
		}
		svc := newTestShiftEntryService(repo)

		entry, err := svc.Update(context.Background(), 1, 1, &UpdateShiftEntryInput{Notes: &newNote})

		require.Error(t, err)
		assert.Nil(t, entry)
		assert.Contains(t, err.Error(), "failed to get shift entry after update")
	})
}
