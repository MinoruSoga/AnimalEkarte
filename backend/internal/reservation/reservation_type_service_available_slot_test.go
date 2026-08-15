package reservation

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

// mockAvailableSlotRepository は ReservationTypeAvailableSlotRepository のテスト用モック実装
type mockAvailableSlotRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	createFn   func(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockAvailableSlotRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeAvailableSlot{}, nil
}

func (m *mockAvailableSlotRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeAvailableSlot{ID: id, ClinicID: clinicID}, nil
}

func (m *mockAvailableSlotRepository) Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error {
	if m.createFn != nil {
		return m.createFn(ctx, slot)
	}
	return nil
}

func (m *mockAvailableSlotRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func newTestReservationTypeAvailableSlotService(repo *mockReservationTypeRepository, slotRepo *mockAvailableSlotRepository) ReservationTypeService {
	return NewReservationTypeService(repo, &mockUnavailableTimeRepository{}, &mockReservationTypeOccupationRepository{}, &mockOccupationRepository{}, nil, slotRepo)
}

// ---- ListAvailableSlots ----

func TestReservationTypeService_ListAvailableSlots(t *testing.T) {
	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
		findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
		wantErr    bool
	}{
		{
			name:    "returns available slots successfully",
			wantErr: false,
		},
		{
			name: "returns error when reservation type not found",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
			wantErr: true,
		},
		{
			name: "propagates repository error listing slots",
			findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeAvailableSlot, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{findByIDFn: tt.findByIDFn}
			slotRepo := &mockAvailableSlotRepository{findAllFn: tt.findAllFn}
			svc := newTestReservationTypeAvailableSlotService(repo, slotRepo)

			got, err := svc.ListAvailableSlots(context.Background(), 1, 10)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// ---- CreateAvailableSlot ----

func TestReservationTypeService_CreateAvailableSlot(t *testing.T) {
	dayOfWeek := int8(1)
	invalidDay := int8(9)
	specificDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	activeFalse := false

	tests := []struct {
		name       string
		input      CreateAvailableSlotInput
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
		existing   []model.ReservationTypeAvailableSlot
		findAllErr error
		createErr  error
		wantErr    bool
		checkFn    func(t *testing.T, got *model.ReservationTypeAvailableSlot)
	}{
		{
			name: "creates weekly slot successfully with default IsActive true",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &dayOfWeek,
				StartTime:     "09:00",
			},
			wantErr: false,
			checkFn: func(t *testing.T, got *model.ReservationTypeAvailableSlot) {
				assert.True(t, got.IsActive)
			},
		},
		{
			name: "creates specific slot with explicit IsActive false",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeSpecific),
				SpecificDate:  &specificDate,
				StartTime:     "10:00",
				IsActive:      &activeFalse,
			},
			wantErr: false,
			checkFn: func(t *testing.T, got *model.ReservationTypeAvailableSlot) {
				assert.False(t, got.IsActive)
			},
		},
		{
			name: "returns error when reservation type not found",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &dayOfWeek,
				StartTime:     "09:00",
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
			wantErr: true,
		},
		{
			name: "returns invalid input for invalid day_of_week",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &invalidDay,
				StartTime:     "09:00",
			},
			wantErr: true,
		},
		{
			name: "propagates error checking existing slots",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &dayOfWeek,
				StartTime:     "09:00",
			},
			findAllErr: errors.New("db error"),
			wantErr:    true,
		},
		{
			name: "returns conflict when slot already exists",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &dayOfWeek,
				StartTime:     "09:00",
			},
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dayOfWeek, StartTime: "09:00"},
			},
			wantErr: true,
		},
		{
			name: "propagates repository create error",
			input: CreateAvailableSlotInput{
				AvailableType: string(model.AvailableSlotTypeWeekly),
				DayOfWeek:     &dayOfWeek,
				StartTime:     "09:00",
			},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{findByIDFn: tt.findByIDFn}
			slotRepo := &mockAvailableSlotRepository{
				findAllFn: func(_ context.Context, _, _ uint64) ([]model.ReservationTypeAvailableSlot, error) {
					return tt.existing, tt.findAllErr
				},
				createFn: func(_ context.Context, _ *model.ReservationTypeAvailableSlot) error {
					return tt.createErr
				},
			}
			svc := newTestReservationTypeAvailableSlotService(repo, slotRepo)

			got, err := svc.CreateAvailableSlot(context.Background(), 1, 10, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				if tt.checkFn != nil {
					tt.checkFn(t, got)
				}
			}
		})
	}
}

// ---- DeleteAvailableSlot ----

func TestReservationTypeService_DeleteAvailableSlot(t *testing.T) {
	tests := []struct {
		name           string
		findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
		slotFindByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
		deleteErr      error
		wantErr        bool
	}{
		{
			name:    "deletes successfully",
			wantErr: false,
		},
		{
			name: "returns error when reservation type not found",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationType, error) {
				return nil, apperrors.WrapNotFound("reservation_type", "10")
			},
			wantErr: true,
		},
		{
			name: "returns error when available slot not found",
			slotFindByIDFn: func(_ context.Context, _, _ uint64) (*model.ReservationTypeAvailableSlot, error) {
				return nil, apperrors.WrapNotFound("reservation_type_available_slot", "5")
			},
			wantErr: true,
		},
		{
			name:      "propagates repository delete error",
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockReservationTypeRepository{findByIDFn: tt.findByIDFn}
			slotRepo := &mockAvailableSlotRepository{
				findByIDFn: tt.slotFindByIDFn,
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := newTestReservationTypeAvailableSlotService(repo, slotRepo)

			err := svc.DeleteAvailableSlot(context.Background(), 1, 10, 5)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- validateAvailableSlotInput ----

func TestValidateAvailableSlotInput(t *testing.T) {
	dayOfWeek := int8(3)
	invalidHigh := int8(7)
	invalidLow := int8(-1)
	specificDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   CreateAvailableSlotInput
		wantErr bool
	}{
		{
			name:    "valid weekly input",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &dayOfWeek, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name:    "weekly missing day_of_week",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "weekly day_of_week above range",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &invalidHigh, StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "weekly day_of_week below range",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &invalidLow, StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "valid specific input",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), SpecificDate: &specificDate, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name:    "specific missing specific_date",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "unknown available type",
			input:   CreateAvailableSlotInput{AvailableType: "bogus", StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "empty available type (zero value)",
			input:   CreateAvailableSlotInput{StartTime: "09:00"},
			wantErr: true,
		},
		{
			name:    "invalid start_time format",
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &dayOfWeek, StartTime: "25:99"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAvailableSlotInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- validateAvailableSlotNotDuplicated ----

func TestValidateAvailableSlotNotDuplicated(t *testing.T) {
	day1 := int8(1)
	day2 := int8(2)
	date1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		existing []model.ReservationTypeAvailableSlot
		input    CreateAvailableSlotInput
		wantErr  bool
	}{
		{
			name:     "no existing entries",
			existing: nil,
			input:    CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr:  false,
		},
		{
			name: "different available type is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &date1, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "different start_time is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &day1, StartTime: "10:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "weekly same day and time conflicts",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &day1, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr: true,
		},
		{
			name: "weekly different day does not conflict",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &day2, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "weekly nil day_of_week on input is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &day1, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: nil, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "weekly nil day_of_week on existing is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: nil, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeWeekly), DayOfWeek: &day1, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "specific same date and time conflicts",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &date1, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), SpecificDate: &date1, StartTime: "09:00"},
			wantErr: true,
		},
		{
			name: "specific different date does not conflict",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &date2, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), SpecificDate: &date1, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "specific nil specific_date on input is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &date1, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), SpecificDate: nil, StartTime: "09:00"},
			wantErr: false,
		},
		{
			name: "specific nil specific_date on existing is skipped",
			existing: []model.ReservationTypeAvailableSlot{
				{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: nil, StartTime: "09:00"},
			},
			input:   CreateAvailableSlotInput{AvailableType: string(model.AvailableSlotTypeSpecific), SpecificDate: &date1, StartTime: "09:00"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAvailableSlotNotDuplicated(tt.existing, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsConflict(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- HasActiveAvailableSlots ----

func TestHasActiveAvailableSlots(t *testing.T) {
	tests := []struct {
		name  string
		slots []model.ReservationTypeAvailableSlot
		want  bool
	}{
		{
			name:  "nil slice returns false",
			slots: nil,
			want:  false,
		},
		{
			name:  "empty slice returns false",
			slots: []model.ReservationTypeAvailableSlot{},
			want:  false,
		},
		{
			name:  "all inactive returns false",
			slots: []model.ReservationTypeAvailableSlot{{IsActive: false}, {IsActive: false}},
			want:  false,
		},
		{
			name:  "one active returns true",
			slots: []model.ReservationTypeAvailableSlot{{IsActive: false}, {IsActive: true}},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasActiveAvailableSlots(tt.slots))
		})
	}
}
