package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockMRRepoDeleteDraftSpy は DeleteDraftByAppointmentID の呼び出し検証専用ラッパー。
// mockMedicalRecordRepository（medical_record_service_test.go 定義）の DeleteDraftByAppointmentID は
// 固定 nil を返す非可変実装のため、interface embedding で必要なメソッドのみ差し替える。
type mockMRRepoDeleteDraftSpy struct {
	repository.MedicalRecordRepository
	deleteDraftFn func(ctx context.Context, clinicID, reservationID uint64) error
}

func (m *mockMRRepoDeleteDraftSpy) DeleteDraftByAppointmentID(ctx context.Context, clinicID, reservationID uint64) error {
	if m.deleteDraftFn != nil {
		return m.deleteDraftFn(ctx, clinicID, reservationID)
	}
	return nil
}

// ================================================================
// DeleteDraftFromReservation
// ================================================================

func TestMedicalRecordService_DeleteDraftFromReservation(t *testing.T) {
	t.Run("成功時: repository.DeleteDraftByAppointmentID が正しい引数で呼ばれる", func(t *testing.T) {
		var capturedClinicID, capturedReservationID uint64
		called := false
		repo := &mockMRRepoDeleteDraftSpy{
			MedicalRecordRepository: &mockMedicalRecordRepository{},
			deleteDraftFn: func(_ context.Context, clinicID, reservationID uint64) error {
				called = true
				capturedClinicID = clinicID
				capturedReservationID = reservationID
				return nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)

		svc.DeleteDraftFromReservation(context.Background(), 3, 77)

		assert.True(t, called)
		assert.Equal(t, uint64(3), capturedClinicID)
		assert.Equal(t, uint64(77), capturedReservationID)
	})

	t.Run("リポジトリエラーは best-effort で無視される（パニックしない）", func(t *testing.T) {
		repo := &mockMRRepoDeleteDraftSpy{
			MedicalRecordRepository: &mockMedicalRecordRepository{},
			deleteDraftFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("db error")
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)

		assert.NotPanics(t, func() {
			svc.DeleteDraftFromReservation(context.Background(), 3, 77)
		})
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
		impl, ok := svc.(*medicalRecordService)
		require.True(t, ok)

		assert.True(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("count>0 -> 再診と判定して false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) { return 3, nil },
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
		impl := svc.(*medicalRecordService)

		assert.False(t, impl.fallbackFirstVisitCheck(context.Background(), 1, 10))
	})

	t.Run("リポジトリエラー -> best-effort で false を返す", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			countByOwnerIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
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
		svc := NewMedicalRecordService(&mockMedicalRecordRepository{}, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, nil)
		})
	})

	t.Run("同日同ペットの既存カルテがある場合はスキップする", func(t *testing.T) {
		created := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
				return []model.MedicalRecord{{ID: 1}}, 1, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "同日重複カルテが存在する場合は作成しない")
	})

	t.Run("重複チェックでリポジトリエラーが発生した場合はスキップする", func(t *testing.T) {
		created := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, errors.New("db error")
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "重複チェック失敗時は安全側でスキップする")
	})

	t.Run("owner_id/pet_idが既に設定済みならLINE補完なしで作成される", func(t *testing.T) {
		var createdRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, clinicalPlanRepo, nil, nil, nil, nil)
		appt := &model.Reservation{
			ID: 1, ClinicID: 1, StartTime: now,
			OwnerID: &ownerID, PetID: &petID, DoctorID: &doctorID,
		}

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
			findAllFn: func(_ context.Context, _ []uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				return errors.New("db error")
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil, nil, nil, nil)
		appt := &model.Reservation{ID: 1, ClinicID: 1, StartTime: now, OwnerID: &ownerID, PetID: &petID}

		assert.NotPanics(t, func() {
			svc.AutoCreateFromReservation(context.Background(), 1, appt)
		})
	})
}
