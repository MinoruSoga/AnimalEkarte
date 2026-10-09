package reservation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestGuardedCheckInRequestRejectsWrongFields(t *testing.T) {
	for _, body := range []string{
		`{"expected_updated_at":"2026-10-08T00:00:00Z","status":"completed"}`,
		`{"expected_updated_at":"2026-10-08T00:00:00Z","status":"checked_in","owner_id":7}`,
		`{"expected_updated_at":"2026-10-08T00:00:00Z","status":"checked_in","pet_id":7}`,
		`{"expected_updated_at":"2026-10-08T00:00:00Z","status":"checked_in","notes":"edited"}`,
		`{"expected_updated_at":"0001-01-01T00:00:00Z","status":"checked_in"}`,
	} {
		t.Run(body, func(t *testing.T) {
			var req updateReservationRequest
			require.NoError(t, json.Unmarshal([]byte(body), &req))
			_, err := req.toServiceInput()
			require.Error(t, err)
		})
	}
}

type guardedCheckInFake struct {
	*mockReservationRepository
	checkInFn func(context.Context, uint64, uint64, time.Time, time.Time) (*model.Reservation, error)
}

func (f *guardedCheckInFake) checkInConfirmed(ctx context.Context, clinicID, id uint64, expected, now time.Time) (*model.Reservation, error) {
	return f.checkInFn(ctx, clinicID, id, expected, now)
}

type checkInTxKey struct{}

func TestGuardedCheckInRejectsChangedSnapshot(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name    string
		status  model.ReservationStatus
		version time.Time
		start   time.Time
	}{
		{"in consultation", model.ReservationStatusInConsultation, now, now},
		{"accounting", model.ReservationStatusAccounting, now, now},
		{"completed", model.ReservationStatusCompleted, now, now},
		{"already checked in", model.ReservationStatusCheckedIn, now, now},
		{"wrong version", model.ReservationStatusConfirmed, now.Add(time.Second), now},
		{"another day", model.ReservationStatusConfirmed, now, now.Add(48 * time.Hour)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{
				lockAndFindByIDFn: func(ctx context.Context, clinic, id uint64) (*model.Reservation, error) {
					require.Equal(t, true, ctx.Value(checkInTxKey{}))
					require.Equal(t, uint64(1), clinic)
					require.Equal(t, uint64(42), id)
					return &model.Reservation{ID: 42, ClinicID: 1, Status: tc.status, UpdatedAt: tc.version, StartTime: tc.start}, nil
				},
			}, checkInFn: func(context.Context, uint64, uint64, time.Time, time.Time) (*model.Reservation, error) {
				t.Fatal("stale reservation must not be written")
				return nil, nil
			}}
			svc := &reservationService{repo: repo, tx: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				return fn(context.WithValue(ctx, checkInTxKey{}, true))
			}}}
			status := model.ReservationStatusCheckedIn
			got, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &now})
			require.Error(t, err)
			require.True(t, apperrors.IsConflict(err))
			require.Nil(t, got)
		})
	}
}

func TestGuardedCheckInSuccessAndPetGuard(t *testing.T) {
	for _, deceased := range []bool{false, true} {
		t.Run(map[bool]string{false: "alive", true: "deceased"}[deceased], func(t *testing.T) {
			expected := time.Now()
			petID := uint64(7)
			ownerID := uint64(9)
			writes := 0
			repo := &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{
				lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: 42, ClinicID: 1, Status: model.ReservationStatusConfirmed, UpdatedAt: expected, StartTime: expected, PetID: &petID, OwnerID: &ownerID}, nil
				},
				findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return ownerID, nil },
				findPetByIDInClinicFn: func(ctx context.Context, clinic, id uint64) (*model.Pet, error) {
					require.Equal(t, true, ctx.Value(checkInTxKey{}))
					require.Equal(t, uint64(1), clinic)
					require.Equal(t, petID, id)
					status := model.PetStatusAlive
					if deceased {
						status = model.PetStatusDeceased
					}
					return &model.Pet{ID: petID, ClinicID: 1, OwnerID: ownerID, Status: status}, nil
				},
			}, checkInFn: func(ctx context.Context, clinic, id uint64, version, now time.Time) (*model.Reservation, error) {
				require.Equal(t, true, ctx.Value(checkInTxKey{}))
				require.True(t, version.Equal(expected))
				require.False(t, now.IsZero())
				writes++
				return &model.Reservation{ID: id, ClinicID: clinic, Status: model.ReservationStatusCheckedIn, CheckedInAt: &now, OwnerID: &ownerID}, nil
			}}
			svc := &reservationService{repo: repo, tx: &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				return fn(context.WithValue(ctx, checkInTxKey{}, true))
			}}}
			status := model.ReservationStatusCheckedIn
			got, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &expected})
			if deceased {
				require.Error(t, err)
				require.Nil(t, got)
				require.Zero(t, writes)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, writes)
				require.Equal(t, status, got.Status)
				require.NotNil(t, got.CheckedInAt)
			}
		})
	}
}

func TestGuardedCheckInMissingDependencies(t *testing.T) {
	expected := time.Now()
	status := model.ReservationStatusCheckedIn
	for _, svc := range []*reservationService{{repo: &mockReservationRepository{}, tx: &mockTransactor{}}, {repo: &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{}}}} {
		_, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &expected})
		require.Error(t, err)
	}
	repo := &reservationRepository{}
	_, err := repo.checkInConfirmed(context.Background(), 1, 42, expected, expected)
	require.Error(t, err)
}

func TestGuardedCheckInWrongClinic(t *testing.T) {
	expected := time.Now()
	status := model.ReservationStatusCheckedIn
	repo := &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{lockAndFindByIDFn: func(_ context.Context, clinic, id uint64) (*model.Reservation, error) {
		require.Equal(t, uint64(2), clinic)
		return nil, apperrors.WrapNotFound("reservation", "42")
	}}}
	svc := &reservationService{repo: repo, tx: &mockTransactor{}}
	got, err := svc.Update(context.Background(), 2, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &expected})
	require.Error(t, err)
	require.Nil(t, got)
}

func TestGuardedCheckInRequiresLinkedPatient(t *testing.T) {
	expected := time.Now()
	status := model.ReservationStatusCheckedIn
	positive := uint64(1)
	zero := uint64(0)
	for _, ids := range [][2]*uint64{{nil, &positive}, {&positive, nil}, {&zero, &positive}, {&positive, &zero}} {
		repo := &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: 42, ClinicID: 1, Status: model.ReservationStatusConfirmed, UpdatedAt: expected, StartTime: expected, OwnerID: ids[0], PetID: ids[1]}, nil
		}}}
		svc := &reservationService{repo: repo, tx: &mockTransactor{}}
		got, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &expected})
		require.Error(t, err)
		require.Nil(t, got)
	}
}

func TestGuardedCheckInRejectsUnverifiedPatient(t *testing.T) {
	expected := time.Now()
	status := model.ReservationStatusCheckedIn
	petID := uint64(7)
	ownerID := uint64(9)
	for _, pet := range []*model.Pet{nil, {ID: 7, ClinicID: 1, OwnerID: 9, Status: "unknown"}, {ID: 7, ClinicID: 1, OwnerID: 9, Status: ""}, {ID: 7, ClinicID: 2, OwnerID: 9, Status: model.PetStatusAlive}, {ID: 7, ClinicID: 1, OwnerID: 10, Status: model.PetStatusAlive}, {ID: 8, ClinicID: 1, OwnerID: 9, Status: model.PetStatusAlive}, {ID: 7, ClinicID: 1, OwnerID: 9, Status: model.PetStatusAlive, DeceasedAt: &expected}} {
		repo := &guardedCheckInFake{mockReservationRepository: &mockReservationRepository{
			lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: 42, ClinicID: 1, Status: model.ReservationStatusConfirmed, UpdatedAt: expected, StartTime: expected, OwnerID: &ownerID, PetID: &petID}, nil
			},
			findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return ownerID, nil },
			findPetByIDInClinicFn:  func(context.Context, uint64, uint64) (*model.Pet, error) { return pet, nil },
		}, checkInFn: func(context.Context, uint64, uint64, time.Time, time.Time) (*model.Reservation, error) {
			return &model.Reservation{ID: 42, OwnerID: &ownerID}, nil
		}}
		svc := &reservationService{repo: repo, tx: &mockTransactor{}}
		got, err := svc.Update(context.Background(), 1, 42, &UpdateReservationInput{Status: &status, ExpectedUpdatedAt: &expected})
		require.Error(t, err)
		require.Nil(t, got)
	}
}
