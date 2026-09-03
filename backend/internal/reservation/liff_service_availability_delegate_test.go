package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
)

// ---- local minimal ReservationStaffRepository / ReservationAdminRepository mocks ----

type delegateMockReservationStaffRepository struct {
	findAllFn func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
}

func (m *delegateMockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *delegateMockReservationStaffRepository) FindByID(_ context.Context, _, _ uint64) (*model.Staff, error) {
	return nil, nil
}
func (m *delegateMockReservationStaffRepository) LockForMutation(
	_ context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}
func (m *delegateMockReservationStaffRepository) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) Update(_ context.Context, _, _ uint64, _ staffpkg.ReservationStaffUpdate) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *delegateMockReservationStaffRepository) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) FindAllExcludedReservationTypes(_ context.Context, _, _ uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *delegateMockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(_ context.Context, _ uint64, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}
func (m *delegateMockReservationStaffRepository) UpdateExcludedReservationTypes(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) FindAllReservationCapabilities(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}
func (m *delegateMockReservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(_ context.Context, _ uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	// By default, every resolved staff supports reservation type 10 (the typeID used across
	// TestLiffService_DelegateStaff) so delegateStaff's downstream availability logic is exercised
	// without every test case needing to wire up capability fixtures explicitly.
	caps := make([]model.StaffReservationCapability, 0, len(staffIDs))
	for _, id := range staffIDs {
		caps = append(caps, model.StaffReservationCapability{StaffID: id, ReservationTypeID: 10})
	}
	return caps, nil
}
func (m *delegateMockReservationStaffRepository) UpdateReservationCapabilities(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}
func (m *delegateMockReservationStaffRepository) SupportsReservationType(_ context.Context, _, _, _ uint64) (bool, error) {
	return false, nil
}

type delegateMockReservationAdminRepository struct {
	findAllByDayFn func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
}

func (m *delegateMockReservationAdminRepository) FindAllByMonth(_ context.Context, _ uint64, _ int, _ time.Month) ([]model.Reservation, error) {
	return nil, nil
}
func (m *delegateMockReservationAdminRepository) FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	if m.findAllByDayFn != nil {
		return m.findAllByDayFn(ctx, clinicID, date)
	}
	return nil, nil
}
func (m *delegateMockReservationAdminRepository) FindTimeRangesByDateRange(_ context.Context, _ uint64, _, _ time.Time) ([]model.Reservation, error) {
	return nil, nil
}
func (m *delegateMockReservationAdminRepository) Create(_ context.Context, _ *model.Reservation) error {
	return nil
}
func (m *delegateMockReservationAdminRepository) SoftDelete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *delegateMockReservationAdminRepository) FindAllByCustomerID(_ context.Context, _, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}
func (m *delegateMockReservationAdminRepository) CancelByID(_ context.Context, _, _, _ uint64) error {
	return nil
}
func (m *delegateMockReservationAdminRepository) FindByIDForNotify(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func TestLiffService_DelegateStaff(t *testing.T) {
	date := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	doctorID1 := uint64(1)

	tests := []struct {
		name           string
		findAllFn      func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
		findAllByDayFn func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
		mode           string
		startTime      string
		endTime        string
		wantErr        bool
		wantID         uint64
	}{
		{
			name: "returns wrapped error when resolveTargetStaffs fails",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return nil, errors.New("db error")
			},
			mode:    "top_priority",
			wantErr: true,
		},
		{
			name: "returns (0, nil) when no staff resolved (apperrors.Wrap(nil,...) is nil)",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return nil, nil
			},
			mode:    "top_priority",
			wantErr: false,
			wantID:  0,
		},
		{
			name: "top_priority mode returns the first staff by sort order",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			mode:   "top_priority",
			wantID: 1,
		},
		{
			name: "first_available falls back to no-staff (0,nil) when FindAllByDay fails",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			findAllByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
				return nil, errors.New("db error")
			},
			mode:   "first_available",
			wantID: 0,
		},
		{
			name: "first_available falls back to no-staff (0,nil) when startTime is malformed",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			mode:      "first_available",
			startTime: "not-a-time",
			endTime:   "1000",
			wantID:    0,
		},
		{
			name: "first_available falls back to no-staff (0,nil) when endTime is malformed",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			mode:      "first_available",
			startTime: "0900",
			endTime:   "not-a-time",
			wantID:    0,
		},
		{
			name: "first_available picks the first staff with no conflicting reservation",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{
					{ID: 1, ReservationVisible: true},
					{ID: 2, ReservationVisible: true},
				}, nil
			},
			findAllByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
				start1 := time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
				return []model.Reservation{
					{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed, StartTime: start1, EndTime: end1},
				}, nil
			},
			mode:      "first_available",
			startTime: "0900",
			endTime:   "1000",
			wantID:    2, // staff 1 is busy, staff 2 is free
		},
		{
			name: "first_available returns (0,nil) when every staff is busy",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			findAllByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
				start1 := time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
				return []model.Reservation{
					{DoctorID: &doctorID1, Status: model.ReservationStatusConfirmed, StartTime: start1, EndTime: end1},
				}, nil
			},
			mode:      "first_available",
			startTime: "0900",
			endTime:   "1000",
			wantID:    0,
		},
		{
			name: "first_available ignores cancelled reservations when checking availability",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			findAllByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
				start1 := time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC)
				end1 := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
				return []model.Reservation{
					{DoctorID: &doctorID1, Status: model.ReservationStatusCancelled, StartTime: start1, EndTime: end1},
				}, nil
			},
			mode:      "first_available",
			startTime: "0900",
			endTime:   "1000",
			wantID:    1,
		},
		{
			name: "default mode (empty string) behaves as first_available",
			findAllFn: func(_ context.Context, _ uint64) ([]model.Staff, error) {
				return []model.Staff{{ID: 1, ReservationVisible: true}}, nil
			},
			findAllByDayFn: func(_ context.Context, _ uint64, _ time.Time) ([]model.Reservation, error) {
				return nil, nil
			},
			mode:      "",
			startTime: "0900",
			endTime:   "1000",
			wantID:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staffRepo := &delegateMockReservationStaffRepository{findAllFn: tt.findAllFn}
			adminRepo := &delegateMockReservationAdminRepository{findAllByDayFn: tt.findAllByDayFn}
			svc := &liffService{staffRepo: staffRepo, adminRepo: adminRepo}

			gotID, err := svc.delegateStaff(context.Background(), 1, 10, tt.mode, date, tt.startTime, tt.endTime)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}

// ---- isStaffAvailable ----

func TestIsStaffAvailable_Delegate(t *testing.T) {
	doctorID := uint64(1)
	otherDoctorID := uint64(2)
	start := time.Date(2026, 4, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		staffID  uint64
		startMin int
		endMin   int
		dayResv  []model.Reservation
		want     bool
	}{
		{
			name:     "available when no reservations exist",
			staffID:  1,
			startMin: 9 * 60,
			endMin:   10 * 60,
			want:     true,
		},
		{
			name:     "unavailable when overlapping reservation exists for the same doctor",
			staffID:  1,
			startMin: 9 * 60,
			endMin:   10 * 60,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
			},
			want: false,
		},
		{
			name:     "available when reservation belongs to a different doctor",
			staffID:  1,
			startMin: 9 * 60,
			endMin:   10 * 60,
			dayResv: []model.Reservation{
				{DoctorID: &otherDoctorID, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
			},
			want: true,
		},
		{
			name:     "available when reservation is cancelled",
			staffID:  1,
			startMin: 9 * 60,
			endMin:   10 * 60,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, Status: model.ReservationStatusCancelled, StartTime: start, EndTime: end},
			},
			want: true,
		},
		{
			name:     "available when reservation has nil DoctorID",
			staffID:  1,
			startMin: 9 * 60,
			endMin:   10 * 60,
			dayResv: []model.Reservation{
				{DoctorID: nil, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
			},
			want: true,
		},
		{
			name:     "available when the new slot does not overlap (before)",
			staffID:  1,
			startMin: 7 * 60,
			endMin:   8*60 + 30,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
			},
			want: true,
		},
		{
			name:     "available when the new slot does not overlap (after)",
			staffID:  1,
			startMin: 10 * 60,
			endMin:   11 * 60,
			dayResv: []model.Reservation{
				{DoctorID: &doctorID, Status: model.ReservationStatusConfirmed, StartTime: start, EndTime: end},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaffAvailable(tt.staffID, tt.startMin, tt.endMin, tt.dayResv)
			assert.Equal(t, tt.want, got)
		})
	}
}
