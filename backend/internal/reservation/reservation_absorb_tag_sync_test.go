package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// Previous owner is supplied by the canonical writer, after intent/row locking.
func TestReservationService_Create_AbsorbedOwnerTagSync(t *testing.T) {
	for _, status := range []model.ReservationStatus{model.ReservationStatusPending, model.ReservationStatusConfirmed, model.ReservationStatusCheckedIn} {
		for _, tt := range []struct {
			name               string
			previous           *uint64
			commitErr, syncErr error
			want               []uint64
		}{
			{name: "transferred pet", previous: uint64Ptr(6), want: []uint64{6, 7}},
			{name: "unchanged owner", previous: uint64Ptr(7), want: []uint64{7}},
			{name: "new booking", want: []uint64{7}},
			{name: "zero previous owner", previous: uint64Ptr(0), want: []uint64{7}},
			{name: "sync failure still attempts both", previous: uint64Ptr(6), syncErr: errors.New("sync unavailable"), want: []uint64{6, 7}},
			{name: "commit fails after write", previous: uint64Ptr(6), commitErr: errors.New("commit failed")},
		} {
			t.Run(string(status)+"/"+tt.name, func(t *testing.T) {
				committed := false
				repo := &mockReservationRepository{previousAbsorbedOwnerID: tt.previous, countOnDutyDoctorsFn: func(context.Context, uint64, time.Time) (int64, error) { return 1, nil }, countConflictsFn: func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil }, findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil }, absorbTentativeOnBookingFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { b.ID = 99; return b, nil }, assertOwnerInClinicFn: func(context.Context, uint64, uint64) error { return nil }, promoteTentativeOnCheckinFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { b.ID = 99; return b, nil }}
				tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					if err := fn(ctx); err != nil {
						return err
					}
					if tt.commitErr != nil {
						return tt.commitErr
					}
					committed = true
					return nil
				}}
				syncer := &commitAwareTagSyncer{t: t, committed: &committed, err: tt.syncErr}
				svc := newReservationService(repo, nil, tx, nil, nil, nil, &mockLineReservationSettingFinder{}, openDayHolidayFinder(), syncer)
				start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
				got, err := svc.Create(context.Background(), &CreateManualReservationInput{ClinicID: 3, StartTime: start, EndTime: start.Add(time.Hour), OwnerID: uint64Ptr(7), PetID: uint64Ptr(22), Status: status})
				if tt.commitErr != nil {
					require.ErrorIs(t, err, tt.commitErr)
					assert.Nil(t, got)
				} else {
					require.NoError(t, err)
					require.NotNil(t, got)
				}
				assert.Equal(t, tt.want, syncer.calls)
			})
		}
	}
}

type commitAwareTagSyncer struct {
	t         *testing.T
	committed *bool
	calls     []uint64
	err       error
}

func (s *commitAwareTagSyncer) SyncNextVisitTag(_ context.Context, clinicID, ownerID uint64) error {
	assert.True(s.t, *s.committed, "sync must run after outer commit")
	assert.Equal(s.t, uint64(3), clinicID)
	s.calls = append(s.calls, ownerID)
	return s.err
}

func TestReservationService_CreateBatch_AbsorbedOwnerTagSync(t *testing.T) {
	for _, failCommit := range []bool{false, true} {
		t.Run(map[bool]string{false: "dedupes old and new owners", true: "commit failure skips all"}[failCommit], func(t *testing.T) {
			committed := false
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				if failCommit {
					return errors.New("commit failed")
				}
				committed = true
				return nil
			}}
			repo := &mockReservationRepository{previousAbsorbedOwnerID: uint64Ptr(6), findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil }, promoteTentativeOnCheckinFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) {
				b.ID = *b.PetID
				return b, nil
			}}
			syncer := &commitAwareTagSyncer{t: t, committed: &committed}
			svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, syncer)
			start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
			got, err := svc.CreateBatch(context.Background(), &CreateManualReservationInput{ClinicID: 3, StartTime: start, EndTime: start.Add(time.Hour), Status: model.ReservationStatusCheckedIn}, []ReservationBatchPet{{OwnerID: 7, PetID: 22}, {OwnerID: 7, PetID: 23}})
			if failCommit {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Empty(t, syncer.calls)
			} else {
				require.NoError(t, err)
				assert.Len(t, got, 2)
				assert.Equal(t, []uint64{6, 7}, syncer.calls)
			}
		})
	}
}

func TestReservationAdminService_Create_AbsorbedOwnerTagSync(t *testing.T) {
	for _, failCommit := range []bool{false, true} {
		t.Run(map[bool]string{false: "syncs both after commit", true: "commit failure skips both"}[failCommit], func(t *testing.T) {
			committed := false
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				if failCommit {
					return errors.New("commit failed")
				}
				committed = true
				return nil
			}}
			repo := &mockReservationRepository{previousAbsorbedOwnerID: uint64Ptr(6), countOnDutyDoctorsFn: func(context.Context, uint64, time.Time) (int64, error) { return 1, nil }, countConflictsFn: func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil }, findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil }, absorbTentativeOnBookingFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { b.ID = 99; return b, nil }}
			syncer := &commitAwareTagSyncer{t: t, committed: &committed}
			svc := newReservationAdminService(&mockReservationAdminRepository{}, repo, nil, tx, nil, nil, nil, nil, openDayHolidayFinder(), syncer)
			start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
			got, err := svc.Create(context.Background(), 3, &CreateReservationAdminInput{StartTime: start, EndTime: start.Add(time.Hour), OwnerID: uint64Ptr(7), PetID: uint64Ptr(22)})
			if failCommit {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Empty(t, syncer.calls)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, []uint64{6, 7}, syncer.calls)
			}
		})
	}
}

func TestLiffService_CreateReservation_AbsorbedOwnerTagSync(t *testing.T) {
	for _, tt := range []struct {
		name     string
		previous *uint64
		want     []uint64
		err      error
	}{
		{name: "transferred pet", previous: uint64Ptr(6), want: []uint64{6, 7}},
		{name: "unchanged owner", previous: uint64Ptr(7), want: []uint64{7}},
		{name: "outer transaction failed", previous: uint64Ptr(6), err: errors.New("commit failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			committed := false
			syncer := &commitAwareTagSyncer{t: t, committed: &committed}
			svc := newReservationTestSvc(func(s *liffService) {
				s.nextVisitTagSync = syncer
				s.validators = &mockLiffValidators{previousAbsorbedOwnerID: tt.previous, validateAndCreateFn: func(context.Context, *CreateReservationInput) (*model.Reservation, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					committed = true
					return &model.Reservation{ID: 99, ClinicID: 3, OwnerID: uint64Ptr(7)}, nil
				}}
			})
			got, err := svc.CreateReservation(context.Background(), 3, 1, reservationBaseInput())
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			assert.Equal(t, tt.want, syncer.calls)
		})
	}
}

func TestReservationService_Create_GuardedReceptionAbsorbedOwnerTagSync(t *testing.T) {
	committed := false
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			return err
		}
		committed = true
		return nil
	}}
	repo := &receptionCreateFake{mockReservationRepository: &mockReservationRepository{previousAbsorbedOwnerID: uint64Ptr(6), findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil }}, guardedFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { return b, nil }}
	syncer := &commitAwareTagSyncer{t: t, committed: &committed}
	svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, syncer)
	input := receptionCreateInput()
	input.ClinicID = 3
	input.OwnerID = uint64Ptr(7)
	_, err := svc.Create(context.Background(), input)
	require.NoError(t, err)
	assert.Equal(t, []uint64{6, 7}, syncer.calls)
}

func TestReservationValidators_ValidateAndCreate_AbsorbedOwnerResult(t *testing.T) {
	for _, failCommit := range []bool{false, true} {
		t.Run(map[bool]string{false: "propagates locked previous owner after commit", true: "failed commit returns no metadata"}[failCommit], func(t *testing.T) {
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				if failCommit {
					return errors.New("commit failed")
				}
				return nil
			}}
			repo := &mockReservationRepository{previousAbsorbedOwnerID: uint64Ptr(6), countOnDutyDoctorsFn: func(context.Context, uint64, time.Time) (int64, error) { return 1, nil }, countConflictsFn: func(context.Context, uint64, time.Time, time.Time, *uint64) (int64, error) { return 0, nil }, absorbTentativeOnBookingFn: func(_ context.Context, b *model.Reservation) (*model.Reservation, error) { b.ID = 99; return b, nil }}
			typeRepo := mockReservationTypeFinder{findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ReservationType, error) {
				return &model.ReservationType{ID: id, ClinicID: clinicID, IsActive: true, ReservationVisible: true, Category: model.ReservationTypeCategoryGeneral}, nil
			}}
			validators := NewReservationValidators(tx, repo, typeRepo, nil, nil, nil, nil, openDayHolidayFinder())
			got, previous, err := validators.ValidateAndCreate(context.Background(), &CreateReservationInput{ClinicID: 3, CustomerID: 1, ReservationTypeID: 9, OwnerID: uint64Ptr(7), PetID: uint64Ptr(22), Date: dateInDays(3), StartTime: "1000", EndTime: "1030", Settings: newSettingForValidation()})
			if failCommit {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Nil(t, previous)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.NotNil(t, previous)
				assert.Equal(t, uint64(6), *previous)
			}
		})
	}
}
