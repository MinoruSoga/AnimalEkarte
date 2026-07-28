package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- local minimal ReservationStaffRepository mock (scoped to this file) ----

type staffAvailMockReservationStaffRepository struct {
	findAllFn                                  func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                                 func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	supportsReservationTypeFn                  func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
	findAllReservationCapabilitiesByStaffIDsFn func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
}

func (m *staffAvailMockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) LockForMutation(
	_ context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}
func (m *staffAvailMockReservationStaffRepository) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *staffAvailMockReservationStaffRepository) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) FindAllExcludedReservationTypes(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) UpdateExcludedReservationTypes(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) FindAllReservationCapabilities(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if m.findAllReservationCapabilitiesByStaffIDsFn != nil {
		return m.findAllReservationCapabilitiesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return nil, nil
}
func (m *staffAvailMockReservationStaffRepository) UpdateReservationCapabilities(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *staffAvailMockReservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return false, nil
}

// ---- resolveTargetStaffs ----

func TestLiffService_ResolveTargetStaffs_WithStaffID(t *testing.T) {
	tests := []struct {
		name       string
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
		supportsFn func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
		wantErr    bool
		wantLen    int
	}{
		{
			name: "returns the staff when visible and supports the reservation type",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ReservationVisible: true}, nil
			},
			supportsFn: func(_ context.Context, _, _, _ uint64) (bool, error) { return true, nil },
			wantLen:    1,
		},
		{
			name: "returns empty when staff is not reservation-visible",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ReservationVisible: false}, nil
			},
			wantLen: 0,
		},
		{
			name: "returns empty when staff does not support the reservation type",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ReservationVisible: true}, nil
			},
			supportsFn: func(_ context.Context, _, _, _ uint64) (bool, error) { return false, nil },
			wantLen:    0,
		},
		{
			name: "returns wrapped error when FindByID fails",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Staff, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns wrapped error when SupportsReservationType fails",
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, ReservationVisible: true}, nil
			},
			supportsFn: func(_ context.Context, _, _, _ uint64) (bool, error) { return false, errors.New("db error") },
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &staffAvailMockReservationStaffRepository{
				findByIDFn:                tt.findByIDFn,
				supportsReservationTypeFn: tt.supportsFn,
			}
			svc := &liffService{staffRepo: repo}
			staffs, err := svc.resolveTargetStaffs(context.Background(), 1, 10, 5)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, staffs, tt.wantLen)
		})
	}
}

func TestLiffService_ResolveTargetStaffs_AllStaffs(t *testing.T) {
	tests := []struct {
		name         string
		findAllFn    func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
		capsFn       func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
		wantErr      bool
		wantStaffIDs []uint64
	}{
		{
			name: "returns error when FindAll fails",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "returns nil when no staff is visible",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: false}}, nil
			},
			// gotIDs は make([]uint64, 0, len(staffs)) で構築されるため、staffs が nil でも
			// 常に非nilの空スライスになる（nil ではなく []uint64{} と比較する）。
			wantStaffIDs: []uint64{},
		},
		{
			name: "returns error when capabilities lookup fails",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			capsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "filters to visible staff that support the reservation type",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{
					{ID: 1, ReservationVisible: true},
					{ID: 2, ReservationVisible: true},
					{ID: 3, ReservationVisible: false},
				}, nil
			},
			capsFn: func(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationCapability, error) {
				return []model.StaffReservationCapability{
					{StaffID: 1, ReservationTypeID: 10},
					{StaffID: 2, ReservationTypeID: 99}, // does not support typeID=10
				}, nil
			},
			wantStaffIDs: []uint64{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &staffAvailMockReservationStaffRepository{
				findAllFn: tt.findAllFn,
				findAllReservationCapabilitiesByStaffIDsFn: tt.capsFn,
			}
			svc := &liffService{staffRepo: repo}
			staffs, err := svc.resolveTargetStaffs(context.Background(), 1, 10, 0)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			gotIDs := make([]uint64, 0, len(staffs))
			for _, s := range staffs {
				gotIDs = append(gotIDs, s.ID)
			}
			assert.Equal(t, tt.wantStaffIDs, gotIDs)
		})
	}
}

// ---- isCapable ----

func TestIsCapable_AvailabilityStaff(t *testing.T) {
	caps := []model.StaffReservationCapability{{ReservationTypeID: 5}, {ReservationTypeID: 8}}
	assert.True(t, isCapable(caps, 5))
	assert.True(t, isCapable(caps, 8))
	assert.False(t, isCapable(caps, 99))
	assert.False(t, isCapable(nil, 5))
}
