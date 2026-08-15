package medicalrecord

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	reservationdomain "github.com/animal-ekarte/backend/internal/reservation"
)

type autoCreateRaceRepository struct {
	MedicalRecordRepository
	acquireLockCalls atomic.Int32
	acquireLockReady chan struct{}
}

func (r *autoCreateRaceRepository) AcquireAutoCreateLock(
	ctx context.Context,
	clinicID, petID uint64,
	date string,
) (bool, error) {
	// 両 caller が同じ logical key の lock 取得点へ到達した後で DB lock を競わせる。
	// sleep を使わず、実際の PostgreSQL advisory lock が直列化することを検証する。
	if r.acquireLockCalls.Add(1) == 2 {
		close(r.acquireLockReady)
	}
	if err := waitForAutoCreateBarrier(ctx, r.acquireLockReady); err != nil {
		return false, err
	}
	return r.MedicalRecordRepository.AcquireAutoCreateLock(ctx, clinicID, petID, date)
}

func (m *mockMedicalRecordRepository) AcquireAutoCreateLock(
	_ context.Context,
	_, _ uint64,
	_ string,
) (bool, error) {
	return true, nil
}

func (m *mockMedicalRecordRepository) CountByPetAndDate(
	_ context.Context,
	_, _ uint64,
	_ string,
) (int64, error) {
	return 0, nil
}

type autoCreateCountCaptureRepository struct {
	*mockMedicalRecordRepository
	countFn func(ctx context.Context, clinicID, petID uint64, date string) (int64, error)
}

func (r *autoCreateCountCaptureRepository) CountByPetAndDate(
	ctx context.Context,
	clinicID, petID uint64,
	date string,
) (int64, error) {
	return r.countFn(ctx, clinicID, petID, date)
}

type autoCreateLockCaptureRepository struct {
	*mockMedicalRecordRepository
	acquireFn func(ctx context.Context, clinicID, petID uint64, date string) (bool, error)
	countFn   func(ctx context.Context, clinicID, petID uint64, date string) (int64, error)
}

func (r *autoCreateLockCaptureRepository) AcquireAutoCreateLock(
	ctx context.Context,
	clinicID, petID uint64,
	date string,
) (bool, error) {
	return r.acquireFn(ctx, clinicID, petID, date)
}

func (r *autoCreateLockCaptureRepository) CountByPetAndDate(
	ctx context.Context,
	clinicID, petID uint64,
	date string,
) (int64, error) {
	if r.countFn != nil {
		return r.countFn(ctx, clinicID, petID, date)
	}
	return r.mockMedicalRecordRepository.CountByPetAndDate(ctx, clinicID, petID, date)
}

func waitForAutoCreateBarrier(ctx context.Context, ready <-chan struct{}) error {
	select {
	case <-ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("timed out waiting for auto-create barrier")
	}
}

type forceOuterRollbackTransactor struct {
	db       *gorm.DB
	sentinel error
}

func (t forceOuterRollbackTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	forceRollback := persistence.TxFromContext(ctx) == nil
	err := t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := fn(persistence.WithTxValue(ctx, tx)); err != nil {
			return err
		}
		if forceRollback {
			return t.sentinel
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type reservationCancelCleanupAuditCapture struct {
	*mockAuditService
	entries    []*AuditEntry
	err        error
	logEntryFn func(context.Context, *AuditEntry) error
}

func (c *reservationCancelCleanupAuditCapture) LogEntry(ctx context.Context, entry *AuditEntry) error {
	c.entries = append(c.entries, entry)
	if c.logEntryFn != nil {
		return c.logEntryFn(ctx, entry)
	}
	return c.err
}

func assertBoundedCleanupContext(ctx context.Context, t *testing.T) {
	t.Helper()

	assert.NoError(t, ctx.Err())
	assert.Nil(t, persistence.TxFromContext(ctx))
	deadline, ok := ctx.Deadline()
	if assert.True(t, ok, "cleanup context must have a deadline") {
		remaining := time.Until(deadline)
		assert.Greater(t, remaining, time.Duration(0))
		assert.LessOrEqual(t, remaining, 5*time.Second)
	}
}

func captureMedicalRecordCleanupLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buffer, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})
	return &buffer
}

// ================================================================
// DeleteDraftFromReservation
// ================================================================

func TestMedicalRecordService_DeleteDraftFromReservation(t *testing.T) {
	t.Run("draft は通常の安全な削除経路を通り監査される", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
			recordID      = uint64(91)
		)
		appointmentID := reservationID
		record := &model.MedicalRecord{
			ID:            recordID,
			ClinicID:      clinicID,
			AppointmentID: &appointmentID,
			Status:        model.MedicalRecordStatusDraft,
		}
		deleteCalled := false
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, gotClinicID, gotReservationID uint64) (*model.MedicalRecord, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, reservationID, gotReservationID)
				return record, nil
			},
			findByIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.MedicalRecord, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, recordID, gotRecordID)
				return record, nil
			},
			countEstimatesByMedicalRecordIDFn: func(_ context.Context, gotRecordID uint64) (int64, error) {
				assert.Equal(t, recordID, gotRecordID)
				return 0, nil
			},
			deleteFn: func(_ context.Context, gotClinicID, gotRecordID uint64) error {
				deleteCalled = true
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, recordID, gotRecordID)
				return nil
			},
		}
		auditSvc := &mockAuditService{}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.True(t, deleteCalled)
		assert.Contains(t, auditSvc.calls, "delete")
	})

	t.Run("検索エラーは best-effort で無視される（パニックしない）", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

		assert.NotPanics(t, func() {
			svc.DeleteDraftFromReservation(context.Background(), 3, 77)
		})
	})

	t.Run("検索DB障害は ERROR と internal_error 監査で可視化する", func(t *testing.T) {
		clinicID := uint64(3)
		reservationID := uint64(77)
		logs := captureMedicalRecordCleanupLogs(t)
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, errors.New("db unavailable")
			},
		}
		audit := &reservationCancelCleanupAuditCapture{mockAuditService: &mockAuditService{}}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.Contains(t, logs.String(), "level=ERROR")
		assert.Contains(t, logs.String(), "reservation cancel draft cleanup lookup failed")
		require.Len(t, audit.entries, 1)
		entry := audit.entries[0]
		assert.Equal(t, &clinicID, entry.ClinicID)
		assert.Nil(t, entry.ActorID)
		assert.Equal(t, model.AuditActorTypeSystem, entry.ActorType)
		assert.Equal(t, "reservation.draft_cleanup_failed", entry.Action)
		assert.Equal(t, model.AuditResourceReservation, entry.Resource)
		assert.Equal(t, &reservationID, entry.ResourceID)
		assert.Equal(t, map[string]any{
			"failure_category": "internal_error",
		}, entry.Metadata)
	})

	t.Run("見積依存 Conflict は ERROR と dependency_conflict 監査で障害から区別する", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
			recordID      = uint64(91)
		)
		logs := captureMedicalRecordCleanupLogs(t)
		record := &model.MedicalRecord{
			ID:       recordID,
			ClinicID: clinicID,
			Status:   model.MedicalRecordStatusDraft,
		}
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return record, nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return record, nil
			},
			countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
				return 1, nil
			},
		}
		audit := &reservationCancelCleanupAuditCapture{mockAuditService: &mockAuditService{}}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.Contains(t, logs.String(), "level=ERROR")
		assert.Contains(t, logs.String(), "reservation cancel draft cleanup blocked by dependency")
		require.Len(t, audit.entries, 1)
		assert.Equal(t, map[string]any{
			"failure_category":  "dependency_conflict",
			"medical_record_id": recordID,
		}, audit.entries[0].Metadata)
	})

	t.Run("lookup 後の状態変更 Conflict は ERROR と state_conflict 監査で見積依存から区別する", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
			recordID      = uint64(91)
		)
		logs := captureMedicalRecordCleanupLogs(t)
		lookupRecord := &model.MedicalRecord{
			ID:       recordID,
			ClinicID: clinicID,
			Status:   model.MedicalRecordStatusDraft,
		}
		lockedRecord := &model.MedicalRecord{
			ID:       recordID,
			ClinicID: clinicID,
			Status:   model.MedicalRecordStatusFinalized,
		}
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return lookupRecord, nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return lockedRecord, nil
			},
		}
		audit := &reservationCancelCleanupAuditCapture{mockAuditService: &mockAuditService{}}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.Contains(t, logs.String(), "level=ERROR")
		assert.Contains(t, logs.String(), "reservation cancel draft cleanup blocked by record state change")
		require.Len(t, audit.entries, 1)
		assert.Equal(t, map[string]any{
			"failure_category":  "state_conflict",
			"medical_record_id": recordID,
		}, audit.entries[0].Metadata)
	})

	t.Run("削除DB障害は ERROR と internal_error 監査で Conflict から区別する", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
			recordID      = uint64(91)
		)
		logs := captureMedicalRecordCleanupLogs(t)
		record := &model.MedicalRecord{
			ID:       recordID,
			ClinicID: clinicID,
			Status:   model.MedicalRecordStatusDraft,
		}
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return record, nil
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return record, nil
			},
			countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
				return 0, errors.New("db unavailable")
			},
		}
		audit := &reservationCancelCleanupAuditCapture{mockAuditService: &mockAuditService{}}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.Contains(t, logs.String(), "level=ERROR")
		assert.Contains(t, logs.String(), "reservation cancel draft cleanup failed")
		require.Len(t, audit.entries, 1)
		assert.Equal(t, map[string]any{
			"failure_category":  "internal_error",
			"medical_record_id": recordID,
		}, audit.entries[0].Metadata)
	})

	t.Run("failure 監査書込エラー自体も ERROR で可視化する", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
		)
		logs := captureMedicalRecordCleanupLogs(t)
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, errors.New("db unavailable")
			},
		}
		audit := &reservationCancelCleanupAuditCapture{
			mockAuditService: &mockAuditService{},
			err:              errors.New("audit unavailable"),
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})

		svc.DeleteDraftFromReservation(context.Background(), clinicID, reservationID)

		assert.Contains(t, logs.String(), "reservation cancel draft cleanup failure audit write failed")
	})

	t.Run("キャンセル済み request でも lookup・安全削除・failure 監査は独立した期限付き context を使う", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
			recordID      = uint64(91)
		)
		record := &model.MedicalRecord{
			ID:       recordID,
			ClinicID: clinicID,
			Status:   model.MedicalRecordStatusDraft,
		}
		lookupContextChecked := false
		deleteContextChecked := false
		auditContextChecked := false
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(ctx context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				lookupContextChecked = true
				assertBoundedCleanupContext(ctx, t)
				return record, nil
			},
			findByIDFn: func(ctx context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				deleteContextChecked = true
				assertBoundedCleanupContext(ctx, t)
				return record, nil
			},
			countEstimatesByMedicalRecordIDFn: func(ctx context.Context, _ uint64) (int64, error) {
				assertBoundedCleanupContext(ctx, t)
				return 1, nil
			},
		}
		audit := &reservationCancelCleanupAuditCapture{
			mockAuditService: &mockAuditService{},
			logEntryFn: func(ctx context.Context, _ *AuditEntry) error {
				auditContextChecked = true
				assertBoundedCleanupContext(ctx, t)
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{})
		parentWithTx := persistence.WithTxValue(context.Background(), &gorm.DB{})
		parentCtx, cancelParent := context.WithCancel(parentWithTx)
		cancelParent()

		svc.DeleteDraftFromReservation(parentCtx, clinicID, reservationID)

		assert.True(t, lookupContextChecked)
		assert.True(t, deleteContextChecked)
		assert.True(t, auditContextChecked)
	})

	t.Run("cleanup の期限を使い切っても failure 監査は独立した期限付き context を使う", func(t *testing.T) {
		const (
			clinicID      = uint64(3)
			reservationID = uint64(77)
		)
		auditContextChecked := false
		audit := &reservationCancelCleanupAuditCapture{
			mockAuditService: &mockAuditService{},
			logEntryFn: func(ctx context.Context, _ *AuditEntry) error {
				auditContextChecked = true
				assertBoundedCleanupContext(ctx, t)
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(
			&mockMedicalRecordRepository{}, nil, nil, nil, nil, nil, nil, nil, nil, audit, nil, &mockTransactor{}).(*medicalRecordService)
		expiredCleanupCtx, cancelCleanup := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
		defer cancelCleanup()

		svc.auditReservationDraftCleanupFailure(
			expiredCleanupCtx,
			clinicID,
			reservationID,
			nil,
			reservationDraftCleanupInternalError,
		)

		assert.True(t, auditContextChecked)
	})
}

// ================================================================
// fallbackFirstVisitCheck
// ================================================================

func TestMedicalRecordService_fallbackFirstVisitCheck(t *testing.T) {
	t.Run("count=0 -> 初診と判定して true を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) { return 0, nil },
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		impl, ok := svc.(*medicalRecordService)
		require.True(t, ok)

		assert.True(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("count>0 -> 再診と判定して false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) { return 3, nil },
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		impl := svc.(*medicalRecordService)

		assert.False(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("リポジトリエラー -> best-effort で false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		impl := svc.(*medicalRecordService)

		assert.False(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})
}

// ================================================================
// AutoCreateFromReservation: 追加分岐（BUG-386 テストで未カバーの経路）
// ================================================================

func TestAutoCreateFromReservation_AdditionalBranches(t *testing.T) {
	now := time.Now()
	ownerID := uint64(10)
	petID := uint64(20)
	doctorID := uint64(99)

	t.Run("reservation が nil の場合はパニックせず即座にスキップする", func(t *testing.T) {
		svc := NewMedicalRecordServiceWithTxAudit(&mockMedicalRecordRepository{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, nil)
		})
	})

	t.Run("同日同ペットの既存カルテがある場合はスキップする", func(t *testing.T) {
		created := false
		repo := &autoCreateCountCaptureRepository{
			mockMedicalRecordRepository: &mockMedicalRecordRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					created = true
					return nil
				},
			},
			countFn: func(_ context.Context, _ uint64, _ uint64, _ string) (int64, error) {
				return 1, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "同日重複カルテが存在する場合は作成しない")
	})

	t.Run("重複チェックでリポジトリエラーが発生した場合はスキップする", func(t *testing.T) {
		created := false
		repo := &autoCreateCountCaptureRepository{
			mockMedicalRecordRepository: &mockMedicalRecordRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					created = true
					return nil
				},
			},
			countFn: func(_ context.Context, _ uint64, _ uint64, _ string) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "重複チェック失敗時は安全側でスキップする")
	})

	t.Run("owner_id/pet_idが既に設定済みならLINE補完なしで作成される", func(t *testing.T) {
		var createdRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, record *model.MedicalRecord) error {
				createdRecord = record
				return nil
			},
		}
		// Create 成功後、AutoCreateFromReservation は必ず CreateSubRecords を呼び、
		// clinicalPlanRepo.FindByMedicalRecordID を無条件に参照する（medical_record_subrecords.go）ため
		// nil のままだとパニックする。既存 plan を返す最小モックを渡す。
		clinicalPlanRepo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1}, nil
			},
		}
		appt := &model.Reservation{
			ID: 1, ClinicID: 1, StartTime: now,
			OwnerID: &ownerID, PetID: &petID, DoctorID: &doctorID,
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return appt, nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, clinicalPlanRepo, nil, nil, nil, nil, reservationRepo, nil, nil, nil, &mockTransactor{})

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		require.NotNil(t, createdRecord)
		assert.Equal(t, &ownerID, createdRecord.OwnerID)
		assert.Equal(t, &petID, createdRecord.PetID)
		assert.Equal(t, &doctorID, createdRecord.DoctorID)
		if assert.NotNil(t, createdRecord.VisitType) {
			assert.Equal(t, model.VisitTypeRevisit, *createdRecord.VisitType)
		}
		if assert.NotNil(t, createdRecord.AppointmentID) {
			assert.Equal(t, uint64(1), *createdRecord.AppointmentID)
		}
	})

	t.Run("Createが失敗した場合はサブレコード作成を試みずパニックしない", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				return errors.New("db error")
			},
		}
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return appt, nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, &mockTransactor{})

		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, appt)
		})
	})
}

func TestAutoCreateFromReservation_AuditLogBeforeCreateCoreExtraction(t *testing.T) {
	const clinicID = uint64(1)
	ownerID := uint64(10)
	petID := uint64(20)
	reservationID := uint64(30)
	start := time.Date(2026, time.July, 31, 15, 30, 0, 0, time.UTC)
	auditSvc := &mockAuditService{}
	repo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, record *model.MedicalRecord) error {
			record.ID = 990
			return nil
		},
	}
	reservation := &model.Reservation{
		ID:        reservationID,
		ClinicID:  clinicID,
		StartTime: start,
		OwnerID:   &ownerID,
		PetID:     &petID,
		VisitType: model.VisitTypeRevisit,
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findByIDFn: func(_ context.Context, gotClinicID, gotReservationID uint64) (*model.Reservation, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, reservationID, gotReservationID)
			return reservation, nil
		},
		assertOwnerFn: func(_ context.Context, gotClinicID, gotOwnerID uint64) error {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, ownerID, gotOwnerID)
			return nil
		},
		findPetOwnerFn: func(_ context.Context, gotClinicID, gotPetID uint64) (uint64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, petID, gotPetID)
			return ownerID, nil
		},
	}
	clinicalPlanRepo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1}, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(
		repo, nil, clinicalPlanRepo, nil, nil, nil, nil, reservationRepo, nil, auditSvc, nil, &mockTransactor{})

	svc.AutoCreateFromReservation(context.Background(), clinicID, reservation)

	assert.Contains(t, auditSvc.calls, "create")
}

func TestAutoCreateFromReservation_ExistingAppointmentRepairsSubrecordsWithoutCreateSideEffects(t *testing.T) {
	const clinicID = uint64(1)
	ownerID := uint64(10)
	petID := uint64(20)
	reservationID := uint64(30)
	visitType := model.VisitTypeRevisit
	existing := &model.MedicalRecord{
		ID:        990,
		ClinicID:  clinicID,
		OwnerID:   &ownerID,
		PetID:     &petID,
		RecordNo:  "EXISTING-AUTO",
		Date:      time.Date(2026, time.August, 1, 0, 0, 0, 0, config.JST),
		Status:    model.MedicalRecordStatusDraft,
		VisitType: &visitType,
	}
	repo := &mockMedicalRecordRepository{
		findByAppointmentIDFn: func(_ context.Context, gotClinicID, gotReservationID uint64) (*model.MedicalRecord, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, reservationID, gotReservationID)
			return existing, nil
		},
		createFn: func(_ context.Context, _ *model.MedicalRecord) error {
			return errors.New("existing appointment must not insert another record")
		},
	}
	reservation := &model.Reservation{
		ID:        reservationID,
		ClinicID:  clinicID,
		StartTime: existing.Date,
		OwnerID:   &ownerID,
		PetID:     &petID,
		VisitType: model.VisitTypeRevisit,
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
			return reservation, nil
		},
		assertOwnerFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
		findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return ownerID, nil
		},
	}
	subrecordLookups := 0
	clinicalPlanRepo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.ClinicalPlan, error) {
			subrecordLookups++
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, existing.ID, gotRecordID)
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: existing.ID}, nil
		},
	}
	auditSvc := &mockAuditService{}
	nextVisitSyncs := 0
	tagSyncSvc := &mockLstepTagSyncService{
		syncNextVisitTagFn: func(_ context.Context, _, _ uint64) error {
			nextVisitSyncs++
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(
		repo, nil, clinicalPlanRepo, nil, nil, nil, nil, reservationRepo, nil, auditSvc, nil, &mockTransactor{}, tagSyncSvc)

	svc.AutoCreateFromReservation(context.Background(), clinicID, reservation)

	assert.Equal(t, 1, subrecordLookups, "existing appointment record must still receive best-effort subrecord repair")
	assert.Empty(t, auditSvc.calls, "existing appointment record must not emit another create audit")
	assert.Equal(t, 0, nextVisitSyncs, "existing appointment record must not repeat create tag side effects")
}

func TestAutoCreateFromReservation_RollsBackCreateWhenOuterTransactionFails(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "自動生成外側rollback飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "自動生成外側rollbackペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "自動生成外側rollback",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)
	start := time.Date(2026, time.July, 31, 15, 30, 0, 0, time.UTC)
	reservation := &model.Reservation{
		ClinicID: clinicID, StartTime: start, EndTime: start.Add(30 * time.Minute),
		OwnerID: &owner.ID, PetID: &pet.ID, VisitType: model.VisitTypeRevisit,
		ReservationTypeID: reservationType.ID, Status: model.ReservationStatusCheckedIn,
		Source: model.ReservationSourceManual, CustomerFields: []byte(`{}`),
	}
	require.NoError(t, db.Create(reservation).Error)
	clinicalPlanRepo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1}, nil
		},
	}
	sentinel := errors.New("force outer auto-create rollback")
	svc := NewMedicalRecordServiceWithTxAudit(
		NewMedicalRecordRepository(db),
		nil,
		clinicalPlanRepo,
		nil,
		nil,
		nil,
		nil,
		reservationdomain.NewReservationRepository(db),
		nil,
		nil,
		nil,
		forceOuterRollbackTransactor{db: db, sentinel: sentinel})

	svc.AutoCreateFromReservation(context.Background(), clinicID, reservation)

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("clinic_id = ? AND pet_id = ? AND date = ? AND deleted_at IS NULL", clinicID, pet.ID, "2026-08-01").
		Count(&count).Error)
	assert.Equal(t, int64(0), count, "outer auto-create rollback must roll back the medical_record insert")
}

func TestAutoCreateFromReservation_RollsBackWhenLockedAppointmentDateDriftsFromLockKey(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "自動生成日付drift飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "自動生成日付driftペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "自動生成日付drift",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)

	authoritativeStart := time.Date(2026, time.August, 1, 15, 30, 0, 0, time.UTC)
	stored := &model.Reservation{
		ClinicID: clinicID, StartTime: authoritativeStart, EndTime: authoritativeStart.Add(30 * time.Minute),
		OwnerID: &owner.ID, PetID: &pet.ID, VisitType: model.VisitTypeRevisit,
		ReservationTypeID: reservationType.ID, Status: model.ReservationStatusCheckedIn,
		Source: model.ReservationSourceManual, CustomerFields: []byte(`{}`),
	}
	require.NoError(t, db.Create(stored).Error)

	staleSnapshot := *stored
	staleSnapshot.StartTime = authoritativeStart.AddDate(0, 0, -1)
	staleSnapshot.EndTime = staleSnapshot.StartTime.Add(30 * time.Minute)
	clinicalPlanRepo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1}, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(
		NewMedicalRecordRepository(db),
		nil,
		clinicalPlanRepo,
		nil,
		nil,
		nil,
		nil,
		reservationdomain.NewReservationRepository(db),
		nil,
		nil,
		nil,
		testTransactor{db: db})

	svc.AutoCreateFromReservation(context.Background(), clinicID, &staleSnapshot)

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("clinic_id = ? AND pet_id = ? AND date = ? AND deleted_at IS NULL", clinicID, pet.ID, "2026-08-02").
		Count(&count).Error)
	assert.Equal(t, int64(0), count, "date drift after appointment row lock must roll back the insert made under the stale-day advisory key")
}

func TestAutoCreateFromReservation_JSTDateBoundaryPreventsDuplicate(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	const clinicID = uint64(1)
	owner := makeTestOwner(t, db, clinicID, "JST境界テスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "JST境界テストペット")
	reservationStart := time.Date(2026, time.July, 27, 23, 0, 0, 0, time.UTC)
	existingRecord := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "JST-BOUNDARY-EXISTING",
		Date:     time.Date(2026, time.July, 28, 0, 0, 0, 0, config.JST),
		OwnerID:  &owner.ID,
		PetID:    &pet.ID,
		Status:   model.MedicalRecordStatusDraft,
	})

	realRepo := NewMedicalRecordRepository(db)
	createAttempted := false
	repo := &autoCreateCountCaptureRepository{
		mockMedicalRecordRepository: &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createAttempted = true
				return errors.New("unexpected duplicate create attempt")
			},
		},
		countFn: func(ctx context.Context, gotClinicID, gotPetID uint64, date string) (int64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, pet.ID, gotPetID)
			assert.Equal(t, "2026-07-28", date)
			total, err := realRepo.CountByPetAndDate(ctx, gotClinicID, gotPetID, date)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			assert.Equal(t, existingRecord.ID, existingRecord.ID)
			return total, nil
		},
	}
	reservation := &model.Reservation{
		ID:        7001,
		ClinicID:  clinicID,
		StartTime: reservationStart,
		OwnerID:   &owner.ID,
		PetID:     &pet.ID,
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findByIDFn: func(_ context.Context, gotClinicID, gotReservationID uint64) (*model.Reservation, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, reservation.ID, gotReservationID)
			return reservation, nil
		},
		assertOwnerFn: func(_ context.Context, gotClinicID, gotOwnerID uint64) error {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, owner.ID, gotOwnerID)
			return nil
		},
		findPetOwnerFn: func(_ context.Context, gotClinicID, gotPetID uint64) (uint64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, pet.ID, gotPetID)
			return owner.ID, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, &mockTransactor{})

	svc.AutoCreateFromReservation(context.Background(), clinicID, reservation)

	assert.False(t, createAttempted, "JST同日の既存カルテがある場合は重複作成を試みない")
}

func TestAutoCreateFromReservation_UsesReservationDateInJST(t *testing.T) {
	const (
		clinicID = uint64(7)
		ownerID  = uint64(70)
		petID    = uint64(700)
	)
	tests := []struct {
		name             string
		reservationStart time.Time
		wantDate         string
	}{
		{
			name:             "過去日の予約は予約日時のJST日付で検索する",
			reservationStart: time.Date(2020, time.January, 1, 23, 0, 0, 0, time.UTC),
			wantDate:         "2020-01-02",
		},
		{
			name:             "未来日の予約は予約日時のJST日付で検索する",
			reservationStart: time.Date(2099, time.December, 31, 23, 0, 0, 0, time.UTC),
			wantDate:         "2100-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			countCalled := false
			reservationOwnerID := ownerID
			reservationPetID := petID
			repo := &autoCreateCountCaptureRepository{
				mockMedicalRecordRepository: &mockMedicalRecordRepository{},
				countFn: func(_ context.Context, gotClinicID, gotPetID uint64, date string) (int64, error) {
					countCalled = true
					assert.Equal(t, clinicID, gotClinicID)
					assert.Equal(t, petID, gotPetID)
					assert.Equal(t, tt.wantDate, date)
					return 1, nil
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
			reservation := &model.Reservation{
				ID:        7100,
				ClinicID:  clinicID,
				StartTime: tt.reservationStart,
				OwnerID:   &reservationOwnerID,
				PetID:     &reservationPetID,
			}

			svc.AutoCreateFromReservation(context.Background(), clinicID, reservation)

			assert.True(t, countCalled)
		})
	}
}

func TestAutoCreateFromReservation_ConcurrentSameClinicPetJSTDayCreatesOneRecord(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	require.NoError(t, db.Exec("TRUNCATE TABLE medical_records, appointments, reservation_types CASCADE").Error)
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "自動生成競合テスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "自動生成競合テストペット")
	reservationType := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "自動生成競合テスト",
		Category: model.ReservationTypeCategoryGeneral,
		IsActive: true,
	}
	require.NoError(t, db.Create(reservationType).Error)

	start := time.Date(2026, time.July, 31, 15, 30, 0, 0, time.UTC)
	reservations := []*model.Reservation{
		{
			ClinicID: clinicID, StartTime: start, EndTime: start.Add(30 * time.Minute),
			OwnerID: &owner.ID, PetID: &pet.ID, VisitType: model.VisitTypeRevisit,
			ReservationTypeID: reservationType.ID, Status: model.ReservationStatusCheckedIn,
			Source: model.ReservationSourceManual, CustomerFields: []byte(`{}`),
		},
		{
			ClinicID: clinicID, StartTime: start, EndTime: start.Add(60 * time.Minute),
			OwnerID: &owner.ID, PetID: &pet.ID, VisitType: model.VisitTypeRevisit,
			ReservationTypeID: reservationType.ID, Status: model.ReservationStatusCheckedIn,
			Source: model.ReservationSourceManual, CustomerFields: []byte(`{}`),
		},
	}
	require.NoError(t, db.Create(&reservations[0]).Error)
	require.NoError(t, db.Create(&reservations[1]).Error)
	require.NotEqual(t, reservations[0].ID, reservations[1].ID)

	realRepo := NewMedicalRecordRepository(db)
	raceRepo := &autoCreateRaceRepository{
		MedicalRecordRepository: realRepo,
		acquireLockReady:        make(chan struct{}),
	}
	clinicalPlanRepo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1}, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(
		raceRepo,
		nil,
		clinicalPlanRepo,
		nil,
		nil,
		nil,
		nil,
		reservationdomain.NewReservationRepository(db),
		nil,
		nil,
		nil,
		testTransactor{db: db})

	ready := make(chan struct{})
	done := make(chan struct{}, len(reservations))
	runCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, reservation := range reservations {
		go func(candidate *model.Reservation) {
			<-ready
			svc.AutoCreateFromReservation(runCtx, clinicID, candidate)
			done <- struct{}{}
		}(reservation)
	}
	close(ready)
	for range reservations {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for concurrent auto-create worker")
		}
	}

	var count int64
	require.NoError(t, db.Model(&model.MedicalRecord{}).
		Where("clinic_id = ? AND pet_id = ? AND date = ? AND deleted_at IS NULL", clinicID, pet.ID, "2026-08-01").
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "same clinic/pet/JST day must have one auto-created record; actual=%d", count)
}

func TestAutoCreateFromReservation_LockScopeAndFailure(t *testing.T) {
	const petID = uint64(700)
	start := time.Date(2026, time.July, 31, 15, 30, 0, 0, time.UTC)

	t.Run("clinicとpetとJST日付をlock keyへ渡しclinic間で分離する", func(t *testing.T) {
		tests := []struct {
			name     string
			clinicID uint64
		}{
			{name: "clinic A", clinicID: 7},
			{name: "clinic B", clinicID: 8},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var gotClinicID, gotPetID uint64
				var gotDate string
				baseRepo := &mockMedicalRecordRepository{
					findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
						return []model.MedicalRecord{{ID: 1}}, 1, nil
					},
				}
				repo := &autoCreateLockCaptureRepository{
					mockMedicalRecordRepository: baseRepo,
					acquireFn: func(_ context.Context, clinicID, petID uint64, date string) (bool, error) {
						gotClinicID, gotPetID, gotDate = clinicID, petID, date
						return true, nil
					},
					countFn: func(_ context.Context, _, _ uint64, _ string) (int64, error) {
						return 1, nil
					},
				}
				ownerID := uint64(70)
				reservationPetID := petID
				reservation := &model.Reservation{
					ID:        7100,
					ClinicID:  tt.clinicID,
					StartTime: start,
					OwnerID:   &ownerID,
					PetID:     &reservationPetID,
				}
				svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

				svc.AutoCreateFromReservation(context.Background(), tt.clinicID, reservation)

				assert.Equal(t, tt.clinicID, gotClinicID)
				assert.Equal(t, petID, gotPetID)
				assert.Equal(t, "2026-08-01", gotDate)
			})
		}
	})

	t.Run("lock取得エラー時はlookupとcreateを行わずbest-effortでreturnする", func(t *testing.T) {
		logs := captureMedicalRecordCleanupLogs(t)
		countCalled := false
		createCalled := false
		baseRepo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		repo := &autoCreateLockCaptureRepository{
			mockMedicalRecordRepository: baseRepo,
			acquireFn: func(_ context.Context, _, _ uint64, _ string) (bool, error) {
				return false, errors.New("lock unavailable")
			},
			countFn: func(_ context.Context, _, _ uint64, _ string) (int64, error) {
				countCalled = true
				return 0, nil
			},
		}
		ownerID := uint64(70)
		reservationPetID := petID
		reservation := &model.Reservation{
			ID:        7101,
			ClinicID:  7,
			StartTime: start,
			OwnerID:   &ownerID,
			PetID:     &reservationPetID,
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 7, reservation)
		})

		assert.False(t, countCalled)
		assert.False(t, createCalled)
		assert.Contains(t, logs.String(), "failed to acquire auto-create lock")
		assert.Contains(t, logs.String(), "auto-create transaction failed")
	})

	t.Run("lock未取得時はlookupとcreateを行わずbest-effortでreturnする", func(t *testing.T) {
		logs := captureMedicalRecordCleanupLogs(t)
		countCalled := false
		createCalled := false
		baseRepo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		repo := &autoCreateLockCaptureRepository{
			mockMedicalRecordRepository: baseRepo,
			acquireFn: func(_ context.Context, _, _ uint64, _ string) (bool, error) {
				return false, nil
			},
			countFn: func(_ context.Context, _, _ uint64, _ string) (int64, error) {
				countCalled = true
				return 0, nil
			},
		}
		ownerID := uint64(70)
		reservationPetID := petID
		reservation := &model.Reservation{
			ID:        7102,
			ClinicID:  7,
			StartTime: start,
			OwnerID:   &ownerID,
			PetID:     &reservationPetID,
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 7, reservation)
		})

		assert.False(t, countCalled)
		assert.False(t, createCalled)
		assert.Contains(t, logs.String(), "auto-create lock is busy")
	})
}

func TestMedicalRecordRepository_AcquireAutoCreateLock_RequiresAmbientTransaction(t *testing.T) {
	repo := NewMedicalRecordRepository(&gorm.DB{})

	acquired, err := repo.AcquireAutoCreateLock(context.Background(), 7, 700, "2026-08-01")

	assert.False(t, acquired)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires an ambient transaction")
}

func TestMedicalRecordRepository_AcquireAutoCreateLock_IsNonBlockingWhenContended(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	holder := db.Begin()
	require.NoError(t, holder.Error)
	defer holder.Rollback()
	holderCtx := persistence.WithTxValue(ctx, holder)
	acquired, err := repo.AcquireAutoCreateLock(holderCtx, 7, 700, "2026-08-01")
	require.NoError(t, err)
	require.True(t, acquired)

	competitor := db.Begin()
	require.NoError(t, competitor.Error)
	defer competitor.Rollback()
	competitorCtx, cancel := context.WithTimeout(
		persistence.WithTxValue(ctx, competitor),
		2*time.Second,
	)
	defer cancel()

	started := time.Now()
	acquired, err = repo.AcquireAutoCreateLock(competitorCtx, 7, 700, "2026-08-01")
	require.NoError(t, err)
	assert.False(t, acquired)
	assert.Less(t, time.Since(started), time.Second, "try-lock must not wait for the holder transaction")
}
