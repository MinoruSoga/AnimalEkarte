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

var errReservationStaffGuardOutsideTx = errors.New("reservation staff guard called outside write transaction")

type recordingReservationStaffWriteGuard struct {
	inTx    func() bool
	calls   *[]string
	staff   model.Staff
	capable bool
}

func (g *recordingReservationStaffWriteGuard) record(call string) {
	*g.calls = append(*g.calls, call)
}

// FindByID represents the guard's two ordered identity checks: the staff row,
// followed by the active clinic assignment row.
func (g *recordingReservationStaffWriteGuard) FindByID(
	_ context.Context,
	_, _ uint64,
) (*model.Staff, error) {
	g.record("lock_staff")
	if !g.inTx() {
		return nil, errReservationStaffGuardOutsideTx
	}
	g.record("lock_assignment")
	staff := g.staff
	return &staff, nil
}

func (g *recordingReservationStaffWriteGuard) SupportsReservationType(
	_ context.Context,
	_, _, _ uint64,
) (bool, error) {
	g.record("lock_capability")
	if !g.inTx() {
		return false, errReservationStaffGuardOutsideTx
	}
	return g.capable, nil
}

func TestReservationStaffWriteGuard_OrderAndVisibilityPolicy(t *testing.T) {
	doctorID := uint64(7)
	zeroDoctorID := uint64(0)

	tests := []struct {
		name        string
		line        bool
		doctorID    *uint64
		staff       model.Staff
		capable     bool
		wantCalls   []string
		wantErr     bool
		wantInvalid bool
	}{
		{
			name:     "standard reservation permits inactive hidden staff after ordered locks",
			doctorID: &doctorID,
			staff: model.Staff{
				ID:                 doctorID,
				IsActive:           false,
				ReservationVisible: false,
			},
			capable:   true,
			wantCalls: []string{"lock_staff", "lock_assignment", "lock_capability"},
		},
		{
			name:     "LIFF reservation accepts active visible capable staff",
			line:     true,
			doctorID: &doctorID,
			staff: model.Staff{
				ID:                 doctorID,
				IsActive:           true,
				ReservationVisible: true,
			},
			capable:   true,
			wantCalls: []string{"lock_staff", "lock_assignment", "lock_capability"},
		},
		{
			name:     "LIFF reservation rejects inactive staff without weakening lock order",
			line:     true,
			doctorID: &doctorID,
			staff: model.Staff{
				ID:                 doctorID,
				IsActive:           false,
				ReservationVisible: true,
			},
			capable:     true,
			wantCalls:   []string{"lock_staff", "lock_assignment", "lock_capability"},
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:     "LIFF reservation rejects hidden staff without weakening lock order",
			line:     true,
			doctorID: &doctorID,
			staff: model.Staff{
				ID:                 doctorID,
				IsActive:           true,
				ReservationVisible: false,
			},
			capable:     true,
			wantCalls:   []string{"lock_staff", "lock_assignment", "lock_capability"},
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:     "missing capability is invalid after staff and assignment locks",
			doctorID: &doctorID,
			staff: model.Staff{
				ID:                 doctorID,
				IsActive:           true,
				ReservationVisible: true,
			},
			capable:     false,
			wantCalls:   []string{"lock_staff", "lock_assignment", "lock_capability"},
			wantErr:     true,
			wantInvalid: true,
		},
		{
			name:      "nil doctor is a no-op",
			doctorID:  nil,
			wantCalls: []string{},
		},
		{
			name:      "zero doctor is a no-op",
			doctorID:  &zeroDoctorID,
			wantCalls: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := make([]string, 0, 3)
			guard := &recordingReservationStaffWriteGuard{
				inTx:    func() bool { return true },
				calls:   &calls,
				staff:   tt.staff,
				capable: tt.capable,
			}

			var err error
			if tt.line {
				err = ValidateLineReservationStaffCapability(
					context.Background(),
					guard,
					1,
					tt.doctorID,
					11,
				)
			} else {
				err = ValidateReservationStaffCapability(
					context.Background(),
					guard,
					1,
					tt.doctorID,
					11,
				)
			}

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantInvalid {
				assert.True(t, apperrors.IsInvalidInput(err), "expected invalid input: %v", err)
			}
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestReservationWrites_RunStaffGuardInsideWriteTransaction(t *testing.T) {
	const (
		clinicID        = uint64(1)
		reservationID   = uint64(9)
		reservationType = uint64(11)
		doctorID        = uint64(7)
	)
	start := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)

	tests := []struct {
		name string
		run  func(
			t *testing.T,
			repo *mockReservationRepository,
			tx Transactor,
			guard *recordingReservationStaffWriteGuard,
		) error
		wantCalls []string
	}{
		{
			name: "standard create",
			run: func(
				_ *testing.T,
				repo *mockReservationRepository,
				tx Transactor,
				guard *recordingReservationStaffWriteGuard,
			) error {
				route := "reception"
				svc := NewReservationServiceWithAvailabilityAndType(
					repo,
					nil,
					tx,
					guard,
					nil,
				)
				_, err := svc.Create(context.Background(), &CreateManualReservationInput{
					ClinicID:          clinicID,
					StartTime:         start,
					EndTime:           end,
					VisitType:         model.VisitTypeRevisit,
					ReservationTypeID: reservationType,
					DoctorID:          uint64PtrForReservationStaffGuard(doctorID),
					Status:            model.ReservationStatusPending,
					Source:            model.ReservationSourceManual,
					ReservationRoute:  &route,
				})
				return err
			},
			wantCalls: []string{
				"tx_begin",
				"lock_staff",
				"lock_assignment",
				"lock_capability",
				"appointment_create",
				"tx_end",
			},
		},
		{
			name: "standard update",
			run: func(
				_ *testing.T,
				repo *mockReservationRepository,
				tx Transactor,
				guard *recordingReservationStaffWriteGuard,
			) error {
				svc := NewReservationServiceWithClinicHolidays(
					repo,
					nil,
					tx,
					guard,
					nil,
					nil,
					nil,
					openDayHolidayFinder(),
				)
				_, err := svc.Update(
					context.Background(),
					clinicID,
					reservationID,
					&UpdateReservationInput{
						DoctorID: uint64PtrForReservationStaffGuard(doctorID),
					},
				)
				return err
			},
			wantCalls: []string{
				"appointment_find",
				"tx_begin",
				"appointment_lock",
				"lock_staff",
				"lock_assignment",
				"lock_capability",
				"appointment_update",
				"tx_end",
			},
		},
		{
			name: "admin create",
			run: func(
				_ *testing.T,
				repo *mockReservationRepository,
				tx Transactor,
				guard *recordingReservationStaffWriteGuard,
			) error {
				svc := NewReservationAdminServiceWithClinicHolidays(
					&mockReservationAdminRepository{},
					repo,
					nil,
					tx,
					guard,
					nil,
					nil,
					openDayHolidayFinder(),
				)
				_, err := svc.Create(context.Background(), clinicID, &CreateReservationAdminInput{
					StartTime:         start,
					EndTime:           end,
					VisitType:         string(model.VisitTypeRevisit),
					ReservationTypeID: reservationType,
					DoctorID:          uint64PtrForReservationStaffGuard(doctorID),
				})
				return err
			},
			wantCalls: []string{
				"tx_begin",
				"lock_staff",
				"lock_assignment",
				"lock_capability",
				"appointment_create",
				"tx_end",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := make([]string, 0, 8)
			inTx := false
			current := &model.Reservation{
				ID:                reservationID,
				ClinicID:          clinicID,
				StartTime:         start,
				EndTime:           end,
				ReservationTypeID: reservationType,
				// Leave doctor unset so the Update path actually changes doctor_id
				// and still runs the staff write guard (BUG-006 skips unchanged schedule).
				Status: model.ReservationStatusPending,
			}
			repo := &mockReservationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					calls = append(calls, "appointment_find")
					copy := *current
					return &copy, nil
				},
				lockAndFindByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
					calls = append(calls, "appointment_lock")
					copy := *current
					return &copy, nil
				},
				createFn: func(_ context.Context, reservation *model.Reservation) error {
					calls = append(calls, "appointment_create")
					reservation.ID = reservationID
					return nil
				},
				updateFieldsFn: func(
					_ context.Context,
					_, _ uint64,
					_ map[string]any,
				) (*model.Reservation, error) {
					calls = append(calls, "appointment_update")
					copy := *current
					return &copy, nil
				},
			}
			tx := &mockTransactor{
				withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					calls = append(calls, "tx_begin")
					inTx = true
					err := fn(ctx)
					inTx = false
					calls = append(calls, "tx_end")
					return err
				},
			}
			guard := &recordingReservationStaffWriteGuard{
				inTx:  func() bool { return inTx },
				calls: &calls,
				staff: model.Staff{
					ID:                 doctorID,
					IsActive:           true,
					ReservationVisible: true,
				},
				capable: true,
			}

			err := tt.run(t, repo, tx, guard)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func uint64PtrForReservationStaffGuard(value uint64) *uint64 {
	return &value
}
