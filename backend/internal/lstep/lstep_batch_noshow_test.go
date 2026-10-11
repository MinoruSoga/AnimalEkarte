package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// noShowMockReservationRepository implements only the consumer-side no-show intent.
type noShowMockReservationRepository struct {
	findNoShowCandidatesFn func(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	markNoShowFn           func(ctx context.Context, clinicID, id uint64) (reservation.NoShowTransition, error)
}

func (m *noShowMockReservationRepository) MarkNoShow(ctx context.Context, clinicID, id uint64) (reservation.NoShowTransition, error) {
	if m.markNoShowFn == nil {
		return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusPending}, nil
	}
	return m.markNoShowFn(ctx, clinicID, id)
}
func (m *noShowMockReservationRepository) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	if m.findNoShowCandidatesFn != nil {
		return m.findNoShowCandidatesFn(ctx, clinicID)
	}
	return nil, nil
}

// newNoShowBatchService は具象型を返す（B-5: detectNoShowReservations の unexport に伴い、
// テストが interface 外の非公開メソッドを直接呼ぶため）。
func newNoShowBatchService(reservationRepo *noShowMockReservationRepository, clinicRepo *dormantMockClinicRepository, auditSvc lstepBatchAuditService, settingsSvc lstepBatchSettingsService) *lstepBatchService {
	return &lstepBatchService{
		reservationRepo: reservationRepo,
		clinicRepo:      clinicRepo,
		auditSvc:        auditSvc,
		settingsSvc:     settingsSvc,
		transactor:      batchImmediateTransactor{},
		noShowAuditTx:   &batchNoShowAuditTxLogger{},
		nowFn:           time.Now,
	}
}

// ---- RunNoShowCheckAllClinics ----

func TestLstepBatchService_RunNoShowCheckAllClinics(t *testing.T) {
	t.Run("returns wrapped error when fetching clinics fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return nil, errors.New("db error") },
		}
		svc := newNoShowBatchService(&noShowMockReservationRepository{}, clinicRepo, &mockAuditService{}, &mockLstepSettingsService{})
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.Error(t, err)
	})

	t.Run("skips clinics where sync-enabled check fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				t.Fatal("DetectNoShowReservations must not run when sync-enabled check failed")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("skips clinics with sync disabled", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				t.Fatal("DetectNoShowReservations must not run for a clinic with sync disabled")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("processes enabled clinics and persists audit metadata when count>0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(action string) { auditCalled = true; assert.Equal(t, "batch_no_show_detect", action) }}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.True(t, auditCalled)
	})

	t.Run("does not persist audit metadata when count==0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(_ string) { auditCalled = true }}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.False(t, auditCalled)
	})

	t.Run("audit log failure is logged but does not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditSvc := &spyLstepBatchAuditService{err: errors.New("audit write failed")}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("per-reservation update errors are logged but do not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
			markNoShowFn: func(_ context.Context, _, _ uint64) (reservation.NoShowTransition, error) {
				return reservation.NoShowTransition{}, errors.New("update failed")
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("nil settingsSvc fails closed", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return nil, nil
			},
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, nil)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.Error(t, err)
	})
}

// ---- DetectNoShowReservations direct cases ----

func TestLstepBatchService_DetectNoShowReservations_FindError(t *testing.T) {
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectNoShowReservations(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestLstepBatchService_DetectNoShowReservations_Success(t *testing.T) {
	var capturedIDs []uint64
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5}, {ID: 6}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, id uint64) (reservation.NoShowTransition, error) {
			capturedIDs = append(capturedIDs, id)
			return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusConfirmed}, nil
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectNoShowReservations(context.Background(), 1)
	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
	assert.Equal(t, []uint64{5, 6}, capturedIDs)
}

func TestLstepBatchService_DetectNoShowReservations_StaleCandidateIsNotCounted(t *testing.T) {
	audit := &batchNoShowAuditTxLogger{}
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{}, nil
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.noShowAuditTx = audit
	count, errs := svc.detectNoShowReservations(context.Background(), 1)
	assert.Zero(t, count)
	assert.Empty(t, errs)
	assert.Empty(t, audit.entries, "a stale candidate must not create an audit record")
}

func TestLstepBatchService_DetectNoShowReservations_AuditsExactTransitionInTransaction(t *testing.T) {
	fixedNow := time.Date(2026, 7, 22, 12, 34, 56, 0, time.UTC)
	audit := &batchNoShowAuditTxLogger{}
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 17}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{
				Changed:        true,
				PreviousStatus: model.ReservationStatusConfirmed,
			}, nil
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.noShowAuditTx = audit
	svc.nowFn = func() time.Time { return fixedNow }

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	if assert.Len(t, audit.entries, 1) {
		entry := audit.entries[0]
		assert.Equal(t, uint64(7), entry.ClinicID)
		assert.Equal(t, uint64(17), entry.AppointmentID)
		assert.Equal(t, model.ReservationStatusConfirmed, entry.PreviousStatus)
		assert.Equal(t, fixedNow, entry.EvaluatedAt)
		assert.Equal(t, noShowRuleVersion, entry.RuleVersion)
		assert.NotEmpty(t, entry.BatchRunID)
	}
}

type noShowRollbackSpyTransactor struct {
	rolledBack bool
}

func (m *noShowRollbackSpyTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	err := fn(ctx)
	m.rolledBack = err != nil
	return err
}

// ---- next_visit_* post-commit resync (EMR-308) ----

func TestLstepBatchService_DetectNoShowReservations_ResyncsNextVisitTagPostCommit(t *testing.T) {
	ownerID := uint64(42)
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5, OwnerID: &ownerID}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusConfirmed, OwnerID: ownerID}, nil
		},
	}
	type resyncCall struct {
		clinicID, ownerID uint64
	}
	var calls []resyncCall
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.tagSyncSvc = &batchMockTagSyncSvc{
		syncNextVisitTagFn: func(_ context.Context, clinicID, oid uint64) error {
			calls = append(calls, resyncCall{clinicID: clinicID, ownerID: oid})
			return nil
		},
	}

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	assert.Equal(t, []resyncCall{{clinicID: 7, ownerID: ownerID}}, calls,
		"a changed no-show transition must resync the appointment's owner tag exactly once")
}

func TestLstepBatchService_DetectNoShowReservations_SyncsTransitionedRowOwnerAfterTransfer(t *testing.T) {
	// Codex #25: candidate 抽出〜CAS 行ロック確定の間に飼い主譲渡が割り込み得る。
	// 同期対象は stale な candidate.OwnerID ではなく MarkNoShowAt がロックした行の
	// owner_id（transition.OwnerID）。
	staleOwnerID := uint64(42)
	currentOwnerID := uint64(77)
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5, OwnerID: &staleOwnerID}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{
				Changed:        true,
				PreviousStatus: model.ReservationStatusConfirmed,
				OwnerID:        currentOwnerID,
			}, nil
		},
	}
	var synced []uint64
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.tagSyncSvc = &batchMockTagSyncSvc{
		syncNextVisitTagFn: func(_ context.Context, _, oid uint64) error {
			synced = append(synced, oid)
			return nil
		},
	}

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	assert.Equal(t, []uint64{currentOwnerID}, synced,
		"Codex #25: CAS で確定した行の owner を同期する（candidate の stale owner ではない）")
}

func TestLstepBatchService_DetectNoShowReservations_NoResyncForNilOwnerOrUnchanged(t *testing.T) {
	ownerID := uint64(42)
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{
				{ID: 5},                    // owner 未紐付け
				{ID: 6, OwnerID: &ownerID}, // owner あり
			}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, id uint64) (reservation.NoShowTransition, error) {
			if id == 6 {
				// 未変化（stale candidate）は遷移しないため resync しない。
				return reservation.NoShowTransition{}, nil
			}
			return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusPending}, nil
		},
	}
	var calls int
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.tagSyncSvc = &batchMockTagSyncSvc{
		syncNextVisitTagFn: func(_ context.Context, _, _ uint64) error {
			calls++
			return nil
		},
	}

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Equal(t, 1, count)
	assert.Empty(t, errs)
	assert.Zero(t, calls, "nil-owner and unchanged transitions must not trigger a resync")
}

func TestLstepBatchService_DetectNoShowReservations_ResyncFailureDoesNotFailItem(t *testing.T) {
	ownerID := uint64(42)
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5, OwnerID: &ownerID}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusPending, OwnerID: ownerID}, nil
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	synced := 0
	svc.tagSyncSvc = &batchMockTagSyncSvc{
		syncNextVisitTagFn: func(_ context.Context, _, _ uint64) error {
			synced++
			return errors.New("tag sync down")
		},
	}

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Equal(t, 1, count, "tag resync failure must not fail the no-show batch item")
	assert.Equal(t, 1, synced)
	assert.Empty(t, errs)
}

func TestLstepBatchService_DetectNoShowReservations_AuditFailureFailsClosed(t *testing.T) {
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 19}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, _ uint64) (reservation.NoShowTransition, error) {
			return reservation.NoShowTransition{
				Changed:        true,
				PreviousStatus: model.ReservationStatusPending,
			}, nil
		},
	}
	tx := &noShowRollbackSpyTransactor{}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.transactor = tx
	svc.noShowAuditTx = &batchNoShowAuditTxLogger{err: errors.New("audit unavailable")}

	count, errs := svc.detectNoShowReservations(context.Background(), 7)

	assert.Zero(t, count)
	assert.Len(t, errs, 1)
	assert.True(t, tx.rolledBack, "audit failure must escape the transaction callback")
}

type noShowCancelOnCommitTransactor struct {
	cancel context.CancelFunc
	calls  int
}

func (m *noShowCancelOnCommitTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	m.calls++
	err := fn(ctx)
	if err == nil && m.calls == 1 {
		m.cancel()
	}
	return err
}

func TestLstepBatchService_DetectNoShowReservations_StopsAfterCommitCancellation(t *testing.T) {
	ownerID := uint64(42)
	var marked []uint64
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5, OwnerID: &ownerID}, {ID: 6, OwnerID: &ownerID}}, nil
		},
		markNoShowFn: func(_ context.Context, _ uint64, id uint64) (reservation.NoShowTransition, error) {
			marked = append(marked, id)
			return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusConfirmed, OwnerID: ownerID}, nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	svc.transactor = &noShowCancelOnCommitTransactor{cancel: cancel}
	var syncCtx context.Context
	svc.tagSyncSvc = &batchMockTagSyncSvc{
		syncNextVisitTagFn: func(c context.Context, _, _ uint64) error {
			syncCtx = c
			assert.NoError(t, c.Err())
			assert.Nil(t, persistence.TxFromContext(c))
			deadline, ok := c.Deadline()
			require.True(t, ok)
			assert.InDelta(t, noShowTagSyncTimeout.Seconds(), time.Until(deadline).Seconds(), 5)
			return nil
		},
	}

	count, errs := svc.detectNoShowReservations(ctx, 7)

	assert.Equal(t, 1, count)
	assert.Equal(t, []uint64{5}, marked, "the second candidate must not be marked after cancellation")
	require.NotEmpty(t, errs)
	assert.ErrorIs(t, errs[len(errs)-1], context.Canceled)
	require.NotNil(t, syncCtx, "the committed item must still resync on the detached context")
	require.ErrorIs(t, syncCtx.Err(), context.Canceled)
}
