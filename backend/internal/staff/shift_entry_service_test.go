package staff

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ShiftEntry モック ----

type mockShiftEntryRepository struct {
	findAllFn          func(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error)
	findByIDFn         func(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error)
	createFn           func(ctx context.Context, entry *model.ShiftEntry) error
	updateFn           func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn           func(ctx context.Context, clinicID, id uint64) error
	replaceBreaksFn    func(ctx context.Context, entryID uint64, breaks []model.ShiftEntryBreak) error
	findOnDutyStaffsFn func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error)
}

func (m *mockShiftEntryRepository) FindAll(ctx context.Context, clinicID uint64, filter ShiftEntryFilter) ([]model.ShiftEntry, error) {
	return m.findAllFn(ctx, clinicID, filter)
}

func (m *mockShiftEntryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockShiftEntryRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.ShiftEntry, error) {
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockShiftEntryRepository) Create(ctx context.Context, entry *model.ShiftEntry) error {
	return m.createFn(ctx, entry)
}

func (m *mockShiftEntryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockShiftEntryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockShiftEntryRepository) ReplaceBreaks(ctx context.Context, entryID uint64, breaks []model.ShiftEntryBreak) error {
	if m.replaceBreaksFn != nil {
		return m.replaceBreaksFn(ctx, entryID, breaks)
	}
	return nil
}

func (m *mockShiftEntryRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockShiftEntryRepository) FindClinicIDsByStaffID(
	_ context.Context,
	_ []uint64,
	_ uint64,
) ([]uint64, error) {
	return nil, nil
}

func (m *mockShiftEntryRepository) FindOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	if m.findOnDutyStaffsFn != nil {
		return m.findOnDutyStaffsFn(ctx, clinicID, date)
	}
	return nil, nil
}

// 予約スケジュール用途 write（ADR-006 論点#1 案A）は shift entry service からは呼ばれない no-op スタブ。
func (m *mockShiftEntryRepository) SaveByStaffDate(
	_ context.Context,
	_ uint64,
	_ *model.ShiftEntry,
	_ []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return nil, nil, false, nil
}

func (m *mockShiftEntryRepository) DeleteByStaffDate(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

type mockShiftEntryStaffAssignments struct {
	countFn func(ctx context.Context, staffID, clinicID uint64) (int64, error)
	lockFn  func(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error)
}

func (m *mockShiftEntryStaffAssignments) CountByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (int64, error) {
	if m.countFn != nil {
		return m.countFn(ctx, staffID, clinicID)
	}
	return 1, nil
}

func (m *mockShiftEntryStaffAssignments) LockActiveByStaffAndClinic(
	ctx context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	if m.lockFn != nil {
		return m.lockFn(ctx, staffID, clinicID)
	}
	return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
}

type mockShiftEntryStaffLocker struct {
	lockFn func(ctx context.Context, staffID uint64) (*model.Staff, error)
}

func (m *mockShiftEntryStaffLocker) LockActiveByIDForShare(
	ctx context.Context,
	staffID uint64,
) (*model.Staff, error) {
	if m.lockFn != nil {
		return m.lockFn(ctx, staffID)
	}
	return &model.Staff{ID: staffID}, nil
}

type shiftCreateTxState struct {
	entryCreated bool
	breaks       []model.ShiftEntryBreak
}

type rollbackShiftCreateTransactor struct {
	state *shiftCreateTxState
}

func (t *rollbackShiftCreateTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	entryCreatedBefore := t.state.entryCreated
	breaksBefore := append([]model.ShiftEntryBreak(nil), t.state.breaks...)
	if err := fn(ctx); err != nil {
		t.state.entryCreated = entryCreatedBefore
		t.state.breaks = breaksBefore
		return err
	}
	return nil
}

type passthroughShiftEntryTransactor struct{}

func (passthroughShiftEntryTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type shiftUpdateTxState struct {
	notes  string
	breaks []model.ShiftEntryBreak
}

type rollbackShiftUpdateTransactor struct {
	state *shiftUpdateTxState
	calls int
}

func (t *rollbackShiftUpdateTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	t.calls++
	notesBefore := t.state.notes
	breaksBefore := append([]model.ShiftEntryBreak(nil), t.state.breaks...)
	if err := fn(ctx); err != nil {
		t.state.notes = notesBefore
		t.state.breaks = breaksBefore
		return err
	}
	return nil
}

func newTestShiftEntryService(repo ShiftEntryRepository) ShiftEntryService {
	return NewShiftEntryService(
		repo,
		&mockShiftEntryStaffLocker{},
		&mockShiftEntryStaffAssignments{},
		passthroughShiftEntryTransactor{},
	)
}

// ---- Tests ----

func TestShiftEntryService_List(t *testing.T) {
	tests := []struct {
		name        string
		clinicID    uint64
		yearMonth   string
		staffID     *uint64
		repoEntries []model.ShiftEntry
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name:      "returns shifts for clinic without filter",
			clinicID:  1,
			yearMonth: "",
			staffID:   nil,
			repoEntries: []model.ShiftEntry{
				{ID: 1, ClinicID: 1, StaffID: 1, ShiftType: model.ShiftTypeMorning},
				{ID: 2, ClinicID: 1, StaffID: 2, ShiftType: model.ShiftTypeAfternoon},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "returns empty list when no shifts exist",
			clinicID:    1,
			yearMonth:   "2024-03",
			staffID:     nil,
			repoEntries: []model.ShiftEntry{},
			repoErr:     nil,
			wantLen:     0,
			wantErr:     false,
		},
		{
			name:        "returns error on invalid yearMonth format",
			clinicID:    1,
			yearMonth:   "2024/03",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     nil,
			wantErr:     true,
		},
		{
			name:        "returns error on invalid yearMonth value",
			clinicID:    1,
			yearMonth:   "2024-13",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     nil,
			wantErr:     true,
		},
		{
			name:        "propagates repository error",
			clinicID:    1,
			yearMonth:   "2024-03",
			staffID:     nil,
			repoEntries: nil,
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				findAllFn: func(_ context.Context, _ uint64, _ ShiftEntryFilter) ([]model.ShiftEntry, error) {
					return tt.repoEntries, tt.repoErr
				},
			}
			svc := newTestShiftEntryService(repo)

			entries, err := svc.List(context.Background(), tt.clinicID, tt.yearMonth, tt.staffID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, entries, tt.wantLen)
			}
		})
	}
}

func TestShiftEntryService_Create(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	startTime := "09:00:00"
	endTime := "18:00:00"
	// BUG-028: end_time <= start_time のテスト用
	sameTime := "09:00:00"
	earlierTime := "08:00:00"

	tests := []struct {
		name             string
		clinicID         uint64
		input            *CreateShiftEntryInput
		repoErr          error
		wantErr          bool
		wantInvalidInput bool
	}{
		{
			name:     "creates shift entry successfully",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: &startTime,
				EndTime:   &endTime,
				Notes:     "Regular shift",
			},
			repoErr:          nil,
			wantErr:          false,
			wantInvalidInput: false,
		},
		{
			name:     "creates shift without times",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeOff),
				Notes:     "Day off",
			},
			repoErr:          nil,
			wantErr:          false,
			wantInvalidInput: false,
		},
		{
			// BUG-028: paid_leave は時刻不要なので end_time <= start_time でもエラーなし
			name:     "creates paid_leave shift without time validation",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypePaidLeave),
				StartTime: &sameTime,
				EndTime:   &sameTime,
			},
			repoErr:          nil,
			wantErr:          false,
			wantInvalidInput: false,
		},
		{
			// BUG-036: full で時刻未指定は InvalidInput（DB に NULL を残さない）
			name:     "returns invalid input when full shift has no times",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeFull),
			},
			repoErr:          nil,
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			// BUG-028: end_time == start_time は InvalidInput
			name:     "returns invalid input when end_time equals start_time",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeFull),
				StartTime: &sameTime,
				EndTime:   &sameTime,
			},
			repoErr:          nil,
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			// BUG-028: end_time < start_time は InvalidInput
			name:     "returns invalid input when end_time is before start_time",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: &startTime,
				EndTime:   &earlierTime,
			},
			repoErr:          nil,
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name:     "returns error when repository fails",
			clinicID: 1,
			input: &CreateShiftEntryInput{
				StaffID:   1,
				Date:      date,
				ShiftType: string(model.ShiftTypeMorning),
				StartTime: &startTime,
				EndTime:   &endTime,
			},
			repoErr:          errors.New("db error"),
			wantErr:          true,
			wantInvalidInput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				createFn: func(_ context.Context, _ *model.ShiftEntry) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
					return &model.ShiftEntry{ID: 1, ClinicID: tt.clinicID}, nil
				},
			}
			svc := newTestShiftEntryService(repo)

			entry, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, entry)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err), "unexpected error: %v", err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
			}
		})
	}
}

func TestShiftEntryService_Create_RequiresActiveStaffClinicAssignmentBeforeWrite(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		assignment   *model.StaffClinicAssignment
		lockErr      error
		wantNotFound bool
	}{
		{
			name:         "rejects staff without an active assignment in the target clinic",
			lockErr:      apperrors.WrapNotFound("staff_clinic_assignment", "staff=7,clinic=2"),
			wantNotFound: true,
		},
		{
			name:    "fails closed when assignment lock fails",
			lockErr: errors.New("assignment lock failed"),
		},
		{
			name:       "fails closed when lock returns a mismatched assignment",
			assignment: &model.StaffClinicAssignment{StaffID: 7, ClinicID: 99},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockShiftEntryRepository{
				createFn: func(_ context.Context, _ *model.ShiftEntry) error {
					t.Fatal("shift entry must not be written before staff assignment is verified")
					return nil
				},
			}
			assignments := &mockShiftEntryStaffAssignments{
				countFn: func(_ context.Context, _, _ uint64) (int64, error) {
					t.Fatal("aggregate count must not be used for write authorization")
					return 0, nil
				},
				lockFn: func(_ context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error) {
					assert.Equal(t, uint64(7), staffID)
					assert.Equal(t, uint64(2), clinicID)
					return tt.assignment, tt.lockErr
				},
			}
			svc := NewShiftEntryService(repo, &mockShiftEntryStaffLocker{}, assignments, &rollbackShiftCreateTransactor{state: &shiftCreateTxState{}})

			entry, err := svc.Create(context.Background(), 2, &CreateShiftEntryInput{
				StaffID:   7,
				Date:      date,
				ShiftType: string(model.ShiftTypeOff),
			})

			require.Error(t, err)
			assert.Nil(t, entry)
			if tt.wantNotFound {
				assert.True(t, apperrors.IsNotFound(err))
			}
		})
	}
}

func TestShiftEntryService_Create_LocksAssignmentBeforeWrite(t *testing.T) {
	events := make([]string, 0, 2)
	repo := &mockShiftEntryRepository{
		createFn: func(_ context.Context, entry *model.ShiftEntry) error {
			events = append(events, "create")
			entry.ID = 42
			return nil
		},
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
			return &model.ShiftEntry{ID: id, ClinicID: clinicID, StaffID: 7}, nil
		},
	}
	assignments := &mockShiftEntryStaffAssignments{
		countFn: func(_ context.Context, _, _ uint64) (int64, error) {
			t.Fatal("aggregate count must not be used for write authorization")
			return 0, nil
		},
		lockFn: func(_ context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error) {
			events = append(events, "lock-assignment")
			return &model.StaffClinicAssignment{ID: 9, StaffID: staffID, ClinicID: clinicID}, nil
		},
	}
	svc := NewShiftEntryService(repo, &mockShiftEntryStaffLocker{}, assignments, passthroughShiftEntryTransactor{})

	entry, err := svc.Create(context.Background(), 2, &CreateShiftEntryInput{
		StaffID:   7,
		Date:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		ShiftType: string(model.ShiftTypeOff),
	})

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, []string{"lock-assignment", "create"}, events)
}

func TestShiftEntryService_Create_RejectsDeletedStaffBeforeAssignmentLock(t *testing.T) {
	repo := &mockShiftEntryRepository{
		createFn: func(_ context.Context, _ *model.ShiftEntry) error {
			t.Fatal("deleted staff must not receive a shift")
			return nil
		},
	}
	staffLocker := &mockShiftEntryStaffLocker{
		lockFn: func(_ context.Context, staffID uint64) (*model.Staff, error) {
			return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
		},
	}
	assignments := &mockShiftEntryStaffAssignments{
		lockFn: func(_ context.Context, _, _ uint64) (*model.StaffClinicAssignment, error) {
			t.Fatal("staff identity must be locked before its assignment")
			return nil, nil
		},
	}
	svc := NewShiftEntryService(repo, staffLocker, assignments, passthroughShiftEntryTransactor{})

	entry, err := svc.Create(context.Background(), 2, &CreateShiftEntryInput{
		StaffID:   7,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: string(model.ShiftTypeOff),
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
	assert.Nil(t, entry)
}

func TestShiftEntryService_Create_UsesStaffThenAssignmentLockAndTransactionContext(t *testing.T) {
	events := make([]string, 0, 5)
	repo := &mockShiftEntryRepository{
		createFn: func(ctx context.Context, entry *model.ShiftEntry) error {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "create")
			entry.ID = 42
			return nil
		},
		replaceBreaksFn: func(ctx context.Context, entryID uint64, _ []model.ShiftEntryBreak) error {
			requireStaffSecurityTxContext(t, ctx)
			assert.Equal(t, uint64(42), entryID)
			events = append(events, "replace-breaks")
			return nil
		},
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.ShiftEntry, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "find")
			return &model.ShiftEntry{ID: id, ClinicID: clinicID, StaffID: 7}, nil
		},
	}
	staffLocker := &mockShiftEntryStaffLocker{
		lockFn: func(ctx context.Context, staffID uint64) (*model.Staff, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-staff")
			return &model.Staff{ID: staffID}, nil
		},
	}
	assignments := &mockShiftEntryStaffAssignments{
		lockFn: func(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error) {
			requireStaffSecurityTxContext(t, ctx)
			events = append(events, "lock-assignment")
			return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
		},
	}
	svc := NewShiftEntryService(repo, staffLocker, assignments, markedStaffSecurityTransactor{})

	entry, err := svc.Create(context.Background(), 2, &CreateShiftEntryInput{
		StaffID:   7,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: string(model.ShiftTypeOff),
		Breaks: []ShiftBreakInput{
			{BreakStart: "12:00:00", BreakEnd: "13:00:00"},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, []string{
		"lock-staff",
		"lock-assignment",
		"create",
		"replace-breaks",
		"find",
	}, events)
}

func TestShiftEntryService_Create_RollsBackEntryWhenBreakReplacementFails(t *testing.T) {
	state := &shiftCreateTxState{}
	events := make([]string, 0, 3)
	repo := &mockShiftEntryRepository{
		createFn: func(_ context.Context, entry *model.ShiftEntry) error {
			events = append(events, "create")
			state.entryCreated = true
			entry.ID = 42
			return nil
		},
		replaceBreaksFn: func(_ context.Context, entryID uint64, breaks []model.ShiftEntryBreak) error {
			events = append(events, "replace-breaks")
			assert.Equal(t, uint64(42), entryID)
			state.breaks = append([]model.ShiftEntryBreak(nil), breaks...)
			return errors.New("replace breaks failed")
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.ShiftEntry, error) {
			t.Fatal("failed transaction must not fetch a result")
			return nil, nil
		},
	}
	assignments := &mockShiftEntryStaffAssignments{
		lockFn: func(_ context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error) {
			events = append(events, "lock-assignment")
			assert.Equal(t, uint64(7), staffID)
			assert.Equal(t, uint64(2), clinicID)
			return &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}, nil
		},
	}
	svc := NewShiftEntryService(repo, &mockShiftEntryStaffLocker{}, assignments, &rollbackShiftCreateTransactor{state: state})

	entry, err := svc.Create(context.Background(), 2, &CreateShiftEntryInput{
		StaffID:   7,
		Date:      time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		ShiftType: string(model.ShiftTypeOff),
		Breaks: []ShiftBreakInput{
			{BreakStart: "12:00:00", BreakEnd: "13:00:00"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, entry)
	assert.Equal(t, []string{"lock-assignment", "create", "replace-breaks"}, events)
	assert.False(t, state.entryCreated)
	assert.Empty(t, state.breaks)
}
