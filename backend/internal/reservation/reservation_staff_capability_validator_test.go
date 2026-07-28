package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockReservationStaffRepositoryForCapability is a minimal ReservationStaffRepository
// implementation used only for ValidateReservationStaffCapability tests.
type mockReservationStaffRepositoryForCapability struct {
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	supportsReservationTypeFn func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationStaffRepositoryForCapability) FindAll(_ context.Context, _ uint64) ([]model.Staff, error) {
	return nil, nil
}
func (m *mockReservationStaffRepositoryForCapability) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Staff{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true}, nil
}
func (m *mockReservationStaffRepositoryForCapability) LockForMutation(
	_ context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}
func (m *mockReservationStaffRepositoryForCapability) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockReservationStaffRepositoryForCapability) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) FindAllExcludedReservationTypes(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockReservationStaffRepositoryForCapability) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *mockReservationStaffRepositoryForCapability) UpdateExcludedReservationTypes(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) FindAllReservationCapabilities(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *mockReservationStaffRepositoryForCapability) FindAllReservationCapabilitiesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *mockReservationStaffRepositoryForCapability) UpdateReservationCapabilities(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *mockReservationStaffRepositoryForCapability) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}

func TestValidateReservationStaffCapability(t *testing.T) {
	doctorID := uint64(3)
	zero := uint64(0)

	tests := []struct {
		name              string
		repo              ReservationStaffRepository
		doctorID          *uint64
		reservationTypeID uint64
		wantErr           bool
		wantInvalidInput  bool
		wantInternal      bool
	}{
		{
			name:              "nil repo: fail closed",
			repo:              nil,
			doctorID:          &doctorID,
			reservationTypeID: 1,
			wantErr:           true,
			wantInternal:      true,
		},
		{
			name:              "nil doctorID: noop",
			repo:              &mockReservationStaffRepositoryForCapability{},
			doctorID:          nil,
			reservationTypeID: 1,
			wantErr:           false,
		},
		{
			name:              "zero doctorID: noop",
			repo:              &mockReservationStaffRepositoryForCapability{},
			doctorID:          &zero,
			reservationTypeID: 1,
			wantErr:           false,
		},
		{
			name:              "zero reservationTypeID: fail closed",
			repo:              &mockReservationStaffRepositoryForCapability{},
			doctorID:          &doctorID,
			reservationTypeID: 0,
			wantErr:           true,
			wantInternal:      true,
		},
		{
			name: "staff not found: wrapped error",
			repo: &mockReservationStaffRepositoryForCapability{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
					return nil, errors.New("not found")
				},
			},
			doctorID:          &doctorID,
			reservationTypeID: 1,
			wantErr:           true,
		},
		{
			name: "SupportsReservationType lookup fails: wrapped error",
			repo: &mockReservationStaffRepositoryForCapability{
				supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
					return false, errors.New("db error")
				},
			},
			doctorID:          &doctorID,
			reservationTypeID: 1,
			wantErr:           true,
		},
		{
			name: "staff does not support the reservation type: invalid input",
			repo: &mockReservationStaffRepositoryForCapability{
				supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
					return false, nil
				},
			},
			doctorID:          &doctorID,
			reservationTypeID: 1,
			wantErr:           true,
			wantInvalidInput:  true,
		},
		{
			name: "staff supports the reservation type: success",
			repo: &mockReservationStaffRepositoryForCapability{
				supportsReservationTypeFn: func(_ context.Context, _, _, _ uint64) (bool, error) {
					return true, nil
				},
			},
			doctorID:          &doctorID,
			reservationTypeID: 1,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReservationStaffCapability(context.Background(), tt.repo, 1, tt.doctorID, tt.reservationTypeID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				if tt.wantInternal {
					var appErr *apperrors.AppError
					require.ErrorAs(t, err, &appErr)
					assert.Equal(t, "INTERNAL", appErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
