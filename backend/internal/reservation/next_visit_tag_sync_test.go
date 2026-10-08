package reservation

// EMR-308: post-commit next-visit タグ同期フック（next_visit_tag_sync.go と
// 各変更容易点の呼び出し）の検証。同期は best-effort で、失敗・未注入・
// owner 未確定は予約変更容易点の成否に影響しない。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// recordingNextVisitTagSyncer は NextVisitTagSyncer の呼び出し記録用モック。
type recordingNextVisitTagSyncer struct {
	calls []uint64
	err   error
}

func (r *recordingNextVisitTagSyncer) SyncNextVisitTag(_ context.Context, _, ownerID uint64) error {
	r.calls = append(r.calls, ownerID)
	return r.err
}

func uint64Ptr(v uint64) *uint64 { return &v }

func TestSyncNextVisitTagBestEffort(t *testing.T) {
	ctx := context.Background()

	t.Run("nil syncer is a no-op", func(t *testing.T) {
		syncNextVisitTagBestEffort(ctx, nil, 1, uint64Ptr(7))
	})

	t.Run("nil and zero owner are skipped", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		syncNextVisitTagBestEffort(ctx, syncer, 1, nil)
		syncNextVisitTagBestEffort(ctx, syncer, 1, uint64Ptr(0))
		assert.Empty(t, syncer.calls)
	})

	t.Run("syncer failure is swallowed (best-effort)", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{err: errors.New("lstep down")}
		syncNextVisitTagBestEffort(ctx, syncer, 1, uint64Ptr(7))
		assert.Equal(t, []uint64{7}, syncer.calls)
	})
}

func TestSyncNextVisitTagsBestEffort(t *testing.T) {
	ctx := context.Background()

	t.Run("deduplicates owner ids and skips nil", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		syncNextVisitTagsBestEffort(ctx, syncer, 1, uint64Ptr(7), nil, uint64Ptr(7), uint64Ptr(8))
		assert.ElementsMatch(t, []uint64{7, 8}, syncer.calls)
	})

	t.Run("nil syncer is a no-op", func(t *testing.T) {
		syncNextVisitTagsBestEffort(ctx, nil, 1, uint64Ptr(7), uint64Ptr(8))
	})
}

func TestReservationService_Delete_NextVisitTagSync(t *testing.T) {
	ctx := context.Background()

	newDeleteRepo := func() *mockReservationRepository {
		return &mockReservationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: 10, OwnerID: uint64Ptr(7)}, nil
			},
			countMedicalRecordsByReservationID: func(_ context.Context, _ uint64) (int64, error) {
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
		}
	}

	t.Run("syncs the deleted reservation owner after commit", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(newDeleteRepo(), nil, nil, nil, nil, nil, nil, nil, syncer)

		require.NoError(t, svc.Delete(ctx, 1, 10))
		assert.Equal(t, []uint64{7}, syncer.calls)
	})

	t.Run("syncer failure does not fail the delete", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{err: errors.New("lstep down")}
		svc := newReservationService(newDeleteRepo(), nil, nil, nil, nil, nil, nil, nil, syncer)

		require.NoError(t, svc.Delete(ctx, 1, 10))
		assert.Equal(t, []uint64{7}, syncer.calls)
	})

	t.Run("failed delete does not trigger the hook", func(t *testing.T) {
		repo := newDeleteRepo()
		repo.deleteFn = func(_ context.Context, _, _ uint64) error { return errors.New("db error") }
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(repo, nil, nil, nil, nil, nil, nil, nil, syncer)

		require.Error(t, svc.Delete(ctx, 1, 10))
		assert.Empty(t, syncer.calls)
	})
}

func TestReservationService_Update_NextVisitTagSync(t *testing.T) {
	ctx := context.Background()
	notes := "follow-up"
	start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)

	newUpdateRepo := func(updatedOwnerID uint64) *mockReservationRepository {
		return &mockReservationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{
					ID:        10,
					OwnerID:   uint64Ptr(1),
					Status:    model.ReservationStatusPending,
					StartTime: start,
					EndTime:   start.Add(30 * time.Minute),
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
				return &model.Reservation{ID: 10, OwnerID: uint64Ptr(updatedOwnerID)}, nil
			},
		}
	}

	t.Run("syncs both old and new owners when ownership changes", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(newUpdateRepo(2), nil, nil, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Update(ctx, 1, 10, &UpdateReservationInput{Notes: &notes})
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{1, 2}, syncer.calls)
	})

	t.Run("same owner is synced once", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(newUpdateRepo(1), nil, nil, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Update(ctx, 1, 10, &UpdateReservationInput{Notes: &notes})
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, syncer.calls)
	})

	t.Run("syncer failure does not fail the update", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{err: errors.New("lstep down")}
		svc := newReservationService(newUpdateRepo(1), nil, nil, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Update(ctx, 1, 10, &UpdateReservationInput{Notes: &notes})
		require.NoError(t, err)
		assert.Equal(t, []uint64{1}, syncer.calls)
	})

	t.Run("failed update does not trigger the hook", func(t *testing.T) {
		repo := newUpdateRepo(1)
		repo.updateFieldsFn = func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
			return nil, errors.New("db error")
		}
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(repo, nil, nil, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Update(ctx, 1, 10, &UpdateReservationInput{Notes: &notes})
		require.Error(t, err)
		assert.Empty(t, syncer.calls)
	})
}

func TestReservationService_Create_NextVisitTagSync(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2030, 1, 10, 10, 0, 0, 0, time.UTC)

	newCreateRepo := func() *mockReservationRepository {
		return &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
			// EMR-308 absorb: tentative がなければ通常 INSERT（booking をそのまま返す）。
			absorbTentativeOnBookingFn: func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
				booking.ID = 99
				return booking, nil
			},
		}
	}

	newCreateInput := func() *CreateManualReservationInput {
		return &CreateManualReservationInput{
			ClinicID:  1,
			StartTime: start,
			EndTime:   start.Add(30 * time.Minute),
			OwnerID:   uint64Ptr(7),
			VisitType: model.VisitTypeRevisit,
			Status:    model.ReservationStatusCheckedIn, // 当日受付: booking constraints 不発・Promote 経路
			Source:    model.ReservationSourceManual,
		}
	}

	t.Run("syncs the created reservation owner after commit", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{}
		repo := newCreateRepo()
		repo.promoteTentativeOnCheckinFn = func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			booking.ID = 99
			return booking, nil
		}
		svc := newReservationService(repo, nil, &mockTransactor{}, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Create(ctx, newCreateInput())
		require.NoError(t, err)
		assert.Equal(t, []uint64{7}, syncer.calls)
	})

	t.Run("syncer failure does not fail the create", func(t *testing.T) {
		syncer := &recordingNextVisitTagSyncer{err: errors.New("lstep down")}
		repo := newCreateRepo()
		repo.promoteTentativeOnCheckinFn = func(_ context.Context, booking *model.Reservation) (*model.Reservation, error) {
			booking.ID = 99
			return booking, nil
		}
		svc := newReservationService(repo, nil, &mockTransactor{}, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Create(ctx, newCreateInput())
		require.NoError(t, err)
		assert.Equal(t, []uint64{7}, syncer.calls)
	})

	t.Run("failed create does not trigger the hook", func(t *testing.T) {
		repo := newCreateRepo()
		repo.promoteTentativeOnCheckinFn = func(_ context.Context, _ *model.Reservation) (*model.Reservation, error) {
			return nil, errors.New("db error")
		}
		syncer := &recordingNextVisitTagSyncer{}
		svc := newReservationService(repo, nil, &mockTransactor{}, nil, nil, nil, nil, nil, syncer)

		_, err := svc.Create(ctx, newCreateInput())
		require.Error(t, err)
		assert.Empty(t, syncer.calls)
	})
}
