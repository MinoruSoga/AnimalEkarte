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

func TestReservationUpdate_UsesLockedPreviousOwner(t *testing.T) {
	for _, cancel := range []bool{false, true} {
		for _, failCommit := range []bool{false, true} {
			t.Run(map[bool]string{false: "update", true: "cancel"}[cancel]+map[bool]string{false: "/committed", true: "/rollback"}[failCommit], func(t *testing.T) {
				committed := false
				start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)
				reservation := func(owner uint64) *model.Reservation {
					return &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(owner), Status: model.ReservationStatusPending, StartTime: start, EndTime: start.Add(time.Hour)}
				}
				repo := &mockReservationRepository{
					findByIDFn:        func(context.Context, uint64, uint64) (*model.Reservation, error) { return reservation(1), nil },
					lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) { return reservation(2), nil },
					updateFieldsFn: func(context.Context, uint64, uint64, map[string]any) (*model.Reservation, error) {
						return reservation(3), nil
					},
					deleteFn: func(context.Context, uint64, uint64) error { return nil },
				}
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
				syncer := &commitAwareTagSyncer{t: t, committed: &committed}
				svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, syncer)
				input := &UpdateReservationInput{OwnerID: uint64Ptr(3)}
				if cancel {
					status := model.ReservationStatusCancelled
					input.Status = &status
				}
				_, err := svc.Update(context.Background(), 3, 10, input)
				if failCommit {
					require.Error(t, err)
					assert.Empty(t, syncer.calls)
				} else {
					require.NoError(t, err)
					assert.Equal(t, []uint64{2, 3}, syncer.calls)
				}
			})
		}
	}
}

func TestReservationDelete_UsesLockedPreviousOwner(t *testing.T) {
	syncer := &recordingNextVisitTagSyncer{}
	repo := &mockReservationRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: 10, OwnerID: uint64Ptr(1)}, nil
		},
		lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: 10, OwnerID: uint64Ptr(2)}, nil
		},
		deleteFn: func(context.Context, uint64, uint64) error { return nil },
	}
	svc := newReservationService(repo, nil, nil, nil, nil, nil, nil, nil, syncer)
	require.NoError(t, svc.Delete(context.Background(), 3, 10))
	assert.Equal(t, []uint64{2}, syncer.calls)
}

func TestReservationAdminDelete_UsesLockedPreviousOwner(t *testing.T) {
	committed := false
	syncer := &commitAwareTagSyncer{t: t, committed: &committed}
	repo := &mockReservationRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: 10, OwnerID: uint64Ptr(1)}, nil
		},
		lockAndFindByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: 10, OwnerID: uint64Ptr(2)}, nil
		},
	}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			return err
		}
		committed = true
		return nil
	}}
	adminRepo := &mockReservationAdminRepository{softDeleteFn: func(context.Context, uint64, uint64) error { return nil }}
	svc := newReservationAdminService(adminRepo, repo, nil, tx, nil, nil, nil, nil, nil, syncer)
	require.NoError(t, svc.Delete(context.Background(), 3, 10))
	assert.Equal(t, []uint64{2}, syncer.calls)
}

func TestClearNextVisitIntent_SyncsStoredAndCurrentOwnersAfterCommit(t *testing.T) {
	committed, intentLocked := false, false
	repo := &mockReservationRepository{
		acquireNextVisitIntentLockFn:    func(context.Context, uint64, uint64) error { intentLocked = true; return nil },
		findPetOwnerInClinicFn:          func(context.Context, uint64, uint64) (uint64, error) { assert.True(t, intentLocked); return 7, nil },
		clearNextVisitIntentWithOwnerFn: func(context.Context, uint64, uint64) (bool, *uint64, error) { return true, uint64Ptr(6), nil },
	}
	tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			return err
		}
		committed = true
		return nil
	}}
	syncer := &commitAwareTagSyncer{t: t, committed: &committed}
	svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, syncer)
	require.NoError(t, svc.ClearNextVisitIntent(context.Background(), 3, 22))
	assert.Equal(t, []uint64{6, 7}, syncer.calls)
}

func TestTentativeTermination_SyncsLockedStoredAndTransferredPetOwners(t *testing.T) {
	for _, deleting := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "delete"}[deleting], func(t *testing.T) {
			committed := false
			locked := &model.Reservation{ID: 10, ClinicID: 3, OwnerID: uint64Ptr(6), PetID: uint64Ptr(22), Status: model.ReservationStatusTentative}
			repo := &mockReservationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					stale := *locked
					stale.OwnerID = uint64Ptr(1)
					return &stale, nil
				},
				findPetOwnerInClinicFn: func(context.Context, uint64, uint64) (uint64, error) { return 7, nil },
				updateTentativeStatusFn: func(ctx context.Context, _ uint64, _ uint64, _ map[string]any, validate tentativeTransitionValidator) (*model.Reservation, error) {
					require.NotNil(t, validate)
					if _, err := validate(ctx, locked); err != nil {
						return nil, err
					}
					updated := *locked
					updated.Status = model.ReservationStatusCancelled
					return &updated, nil
				},
				deleteFn: func(context.Context, uint64, uint64) error { return nil },
			}
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				if err := fn(ctx); err != nil {
					return err
				}
				committed = true
				return nil
			}}
			tags := &commitAwareTagSyncer{t: t, committed: &committed}
			svc := newReservationService(repo, nil, tx, nil, nil, nil, nil, nil, tags)
			if deleting {
				require.NoError(t, svc.Delete(context.Background(), 3, 10))
			} else {
				status := model.ReservationStatusCancelled
				_, err := svc.Update(context.Background(), 3, 10, &UpdateReservationInput{Status: &status})
				require.NoError(t, err)
			}
			assert.Equal(t, []uint64{6, 7}, tags.calls)
		})
	}
}
