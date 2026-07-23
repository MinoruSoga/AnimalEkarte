package medicalrecord

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

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
	assert.Nil(t, repohelpers.TxFromContext(ctx))
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})

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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{})
		parentWithTx := repohelpers.WithTxValue(context.Background(), &gorm.DB{})
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
		svc := NewMedicalRecordService(
			&mockMedicalRecordRepository{}, nil, nil, nil, nil, nil, nil, nil, nil, audit, &mockTransactor{},
		).(*medicalRecordService)
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		impl, ok := svc.(*medicalRecordService)
		require.True(t, ok)

		assert.True(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("count>0 -> 再診と判定して false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) { return 3, nil },
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		impl := svc.(*medicalRecordService)

		assert.False(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("リポジトリエラー -> best-effort で false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
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
		svc := NewMedicalRecordService(&mockMedicalRecordRepository{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, nil)
		})
	})

	t.Run("同日同ペットの既存カルテがある場合はスキップする", func(t *testing.T) {
		created := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return []model.MedicalRecord{{ID: 1}}, 1, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "同日重複カルテが存在する場合は作成しない")
	})

	t.Run("重複チェックでリポジトリエラーが発生した場合はスキップする", func(t *testing.T) {
		created := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, errors.New("db error")
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "重複チェック失敗時は安全側でスキップする")
	})

	t.Run("owner_id/pet_idが既に設定済みならLINE補完なしで作成される", func(t *testing.T) {
		var createdRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
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
		svc := NewMedicalRecordService(repo, nil, clinicalPlanRepo, nil, nil, nil, nil, reservationRepo, nil, nil, &mockTransactor{})

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
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, &mockTransactor{})

		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, appt)
		})
	})
}
