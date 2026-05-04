package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// batchMockReservationRepo は batch テスト専用 ReservationRepository モック
type batchMockReservationRepo struct {
	findNoShowCandidatesFn func(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	updateFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
}

func (m *batchMockReservationRepo) FindAll(_ context.Context, _ uint64, _, _ int, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *batchMockReservationRepo) FindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *batchMockReservationRepo) Create(_ context.Context, _ *model.Reservation) error { return nil }
func (m *batchMockReservationRepo) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{}, nil
}
func (m *batchMockReservationRepo) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *batchMockReservationRepo) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *batchMockReservationRepo) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}
func (m *batchMockReservationRepo) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	return 1, nil
}
func (m *batchMockReservationRepo) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockReservationRepo) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *batchMockReservationRepo) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *batchMockReservationRepo) CountMedicalRecordsByReservationID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockReservationRepo) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *batchMockReservationRepo) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}
func (m *batchMockReservationRepo) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *batchMockReservationRepo) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	if m.findNoShowCandidatesFn != nil {
		return m.findNoShowCandidatesFn(ctx, clinicID)
	}
	return nil, nil
}
func (m *batchMockReservationRepo) HasReservationByOwnerInRange(_ context.Context, _, _ uint64, _, _ time.Time) (bool, error) {
	return false, nil
}

// batchMockMedRecordRepo は batch テスト専用 MedicalRecordRepository モック
type batchMockMedRecordRepo struct {
	findDormantFn func(ctx context.Context, clinicID uint64, minDays int) ([]repository.DormantOwnerEntry, error)
}

func (m *batchMockMedRecordRepo) FindAll(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (m *batchMockMedRecordRepo) FindByID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) Create(_ context.Context, _ *model.MedicalRecord) error { return nil }
func (m *batchMockMedRecordRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *batchMockMedRecordRepo) CountByPetID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockMedRecordRepo) CountEstimatesByMedicalRecordID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockMedRecordRepo) FindOwnerVisitSummary(_ context.Context, _, _ uint64) (*repository.OwnerVisitSummary, error) {
	return &repository.OwnerVisitSummary{}, nil
}
func (m *batchMockMedRecordRepo) FindLatestByOwner(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDays int) ([]repository.DormantOwnerEntry, error) {
	if m.findDormantFn != nil {
		return m.findDormantFn(ctx, clinicID, minDays)
	}
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindOwnersByFirstVisitDate(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindOwnersByLastVisitDays(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindOwnersByNextVisitRecommended(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func (m *batchMockMedRecordRepo) CountByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// batchMockTagSyncSvc は batch テスト専用 LstepTagSyncService モック
type batchMockTagSyncSvc struct {
	syncNoShowTagFn  func(ctx context.Context, clinicID, ownerID uint64, t time.Time) error
	syncDormantTagFn func(ctx context.Context, clinicID, ownerID uint64, daysSince int) error
}

func (m *batchMockTagSyncSvc) SyncVaccineTag(_ context.Context, _, _, _ uint64) error { return nil }
func (m *batchMockTagSyncSvc) SyncVisitCompletionTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncOwnerAnimalClassificationTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncPetBasicInfoTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncNextVisitTag(_ context.Context, _, _ uint64) error { return nil }
func (m *batchMockTagSyncSvc) SyncReservationTag(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncCancellationTag(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncCheckupTag(_ context.Context, _, _, _ uint64, _ time.Time, _ *time.Time) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncPrescriptionTag(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncChronicConditionTags(_ context.Context, _, _ uint64, _ []string) error {
	return nil
}
func (m *batchMockTagSyncSvc) SyncCPMStageTag(_ context.Context, _, _ uint64) error { return nil }
func (m *batchMockTagSyncSvc) SyncNoShowTag(ctx context.Context, clinicID, ownerID uint64, t time.Time) error {
	if m.syncNoShowTagFn != nil {
		return m.syncNoShowTagFn(ctx, clinicID, ownerID, t)
	}
	return nil
}
func (m *batchMockTagSyncSvc) SyncDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSince int) error {
	if m.syncDormantTagFn != nil {
		return m.syncDormantTagFn(ctx, clinicID, ownerID, daysSince)
	}
	return nil
}
func (m *batchMockTagSyncSvc) ResyncOwnerVaccineTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) ResyncOwnerCheckupTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) ResyncOwnerReservationTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncCPMStageTagV2(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncLTVTopPercent(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *batchMockTagSyncSvc) SyncVisitDormantTags(_ context.Context, _, _ uint64, _ int) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncPetSpeciesTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncSeniorTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncExclusionTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncHealthcheckTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncAnnual4CheckupTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncVaccineDeadlineTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncFilariaTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncFleaTickTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncFoodPurchaseTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncSpecialCheckupCandidateTag(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncHealthPreventionTagsForClinic(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

type batchMockAuditService struct {
	// ISSUE-010: 引数捕捉用の spy フィールド（既存テストでは未使用 — nil のまま）。
	capturedAction   string
	capturedMetadata any
	called           bool
}

func (m *batchMockAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *batchMockAuditService) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (m *batchMockAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}

func (m *batchMockAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, action, _ string, _ *uint64, metadata any) error {
	m.called = true
	m.capturedAction = action
	m.capturedMetadata = metadata
	return nil
}

func newBatchService(
	resRepo repository.ReservationRepository,
	tagSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRepo repository.MedicalRecordRepository,
) LstepBatchService {
	return NewLstepBatchService(resRepo, tagSvc, clinicRepo, medRepo, &batchMockAuditService{}, &mockLstepSettingsService{}, nil, nil)
}

// newBatchServiceWithAuditSpy は ISSUE-010 監査 metadata 検証用に audit spy を返す。
func newBatchServiceWithAuditSpy(
	resRepo repository.ReservationRepository,
	tagSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRepo repository.MedicalRecordRepository,
) (LstepBatchService, *batchMockAuditService) {
	spy := &batchMockAuditService{}
	return NewLstepBatchService(resRepo, tagSvc, clinicRepo, medRepo, spy, &mockLstepSettingsService{}, nil, nil), spy
}

func TestDetectNoShowReservations_Success(t *testing.T) {
	ownerID := uint64(42)
	now := time.Now()
	reservations := []model.Reservation{
		{ID: 1, OwnerID: &ownerID, StartTime: now},
		{ID: 2, OwnerID: nil, StartTime: now}, // no owner — tag skip
	}

	tagSyncCalled := 0
	svc := newBatchService(
		&batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return reservations, nil
			},
		},
		&batchMockTagSyncSvc{
			syncNoShowTagFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				tagSyncCalled++
				return nil
			},
		},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{},
	)

	count, errs := svc.DetectNoShowReservations(context.Background(), 1)

	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
	assert.Equal(t, 1, tagSyncCalled, "ownerIDなし予約はタグ同期をスキップする")
}

func TestDetectNoShowReservations_FindCandidatesError(t *testing.T) {
	svc := newBatchService(
		&batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{},
	)

	count, errs := svc.DetectNoShowReservations(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestDetectNoShowReservations_UpdateError(t *testing.T) {
	ownerID := uint64(10)
	svc := newBatchService(
		&batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 1, OwnerID: &ownerID}}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
				return nil, errors.New("update failed")
			},
		},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{},
	)

	count, errs := svc.DetectNoShowReservations(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestRunNoShowCheckAllClinics_Success(t *testing.T) {
	clinics := []model.Clinic{{ID: 1}, {ID: 2}}
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunNoShowCheckAllClinics(context.Background())
	assert.NoError(t, err)
}

func TestRunNoShowCheckAllClinics_FetchClinicsError(t *testing.T) {
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunNoShowCheckAllClinics(context.Background())
	assert.Error(t, err)
}

func TestDetectDormantOwners_Success(t *testing.T) {
	entries := []repository.DormantOwnerEntry{
		{OwnerID: 1, DaysSince: 200},
		{OwnerID: 2, DaysSince: 370},
	}
	synced := make([]uint64, 0)
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{
			syncDormantTagFn: func(_ context.Context, _, ownerID uint64, _ int) error {
				synced = append(synced, ownerID)
				return nil
			},
		},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{
			findDormantFn: func(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return entries, nil
			},
		},
	)

	count, errs := svc.DetectDormantOwners(context.Background(), 1)

	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
	assert.Equal(t, []uint64{1, 2}, synced)
}

func TestDetectDormantOwners_FindError(t *testing.T) {
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{
			findDormantFn: func(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return nil, errors.New("db error")
			},
		},
	)

	count, errs := svc.DetectDormantOwners(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestDetectDormantOwners_TagSyncError(t *testing.T) {
	entries := []repository.DormantOwnerEntry{{OwnerID: 5, DaysSince: 250}}
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{
			syncDormantTagFn: func(_ context.Context, _, _ uint64, _ int) error {
				return errors.New("lstep api error")
			},
		},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{
			findDormantFn: func(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return entries, nil
			},
		},
	)

	count, errs := svc.DetectDormantOwners(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestRunDormantDetectionAllClinics_Success(t *testing.T) {
	clinics := []model.Clinic{{ID: 1}}
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunDormantDetectionAllClinics(context.Background())
	assert.NoError(t, err)
}

func TestRunDormantDetectionAllClinics_FetchClinicsError(t *testing.T) {
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunDormantDetectionAllClinics(context.Background())
	assert.Error(t, err)
}

// ---- ISSUE-010: バッチ実行時の監査メタデータ永続化 ----

// TestRunNoShowCheckAllClinics_PersistsAuditMetadata は no-show バッチ完了時に
// processed_count / error_count を audit_logs.metadata に記録することを検証する。
func TestRunNoShowCheckAllClinics_PersistsAuditMetadata(t *testing.T) {
	ownerID := uint64(1)
	clinics := []model.Clinic{{ID: 1}}

	svc, spy := newBatchServiceWithAuditSpy(
		&batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 1, OwnerID: &ownerID, StartTime: time.Now()}}, nil
			},
		},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunNoShowCheckAllClinics(context.Background())
	assert.NoError(t, err)
	assert.True(t, spy.called, "監査ログ呼び出しが行われる")
	assert.Equal(t, "batch_no_show_detect", spy.capturedAction)

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "batch_no_show_detect", meta["operation"])
	assert.Equal(t, 1, meta["processed_count"])
	assert.Equal(t, 0, meta["error_count"])
}

// TestRunDormantDetectionAllClinics_PersistsAuditMetadata は休眠検知バッチで
// processed_count / error_count / min_days_since が metadata に記録されることを検証する。
func TestRunDormantDetectionAllClinics_PersistsAuditMetadata(t *testing.T) {
	clinics := []model.Clinic{{ID: 1}}
	entries := []repository.DormantOwnerEntry{
		{OwnerID: 1, DaysSince: 200},
		{OwnerID: 2, DaysSince: 250},
	}

	svc, spy := newBatchServiceWithAuditSpy(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{
			findDormantFn: func(_ context.Context, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return entries, nil
			},
		},
	)

	err := svc.RunDormantDetectionAllClinics(context.Background())
	assert.NoError(t, err)
	assert.True(t, spy.called)
	assert.Equal(t, "batch_dormant_detect", spy.capturedAction)

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "batch_dormant_detect", meta["operation"])
	assert.Equal(t, 2, meta["processed_count"])
	assert.Equal(t, 0, meta["error_count"])
	assert.Equal(t, 180, meta["min_days_since"], "判定閾値を後で再現できる")
}

// ---- RunHealthPreventionTagSyncAllClinics (FEAT-379) ----

func TestRunHealthPreventionTagSyncAllClinics_Success(t *testing.T) {
	clinics := []model.Clinic{{ID: 1}}
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.NoError(t, err)
}

func TestRunHealthPreventionTagSyncAllClinics_FetchClinicsError(t *testing.T) {
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.Error(t, err)
}

// ---- FEAT-383-supplement Scope 3: クリニック別配信実行時刻フィルタ ----

// batchMockDeliveryTrigger は RunDeliveryTriggerBatchAllClinics の配信時刻フィルタ検証用モック。
// TriggerFirstVisitFollowUp3D のみ clinicID を記録し (1, nil) を返す。他は (0, nil)。
type batchMockDeliveryTrigger struct {
	processedClinics []uint64
}

func (m *batchMockDeliveryTrigger) TriggerFirstVisitFollowUp3D(_ context.Context, clinicID uint64, _ time.Time) (int, []error) {
	m.processedClinics = append(m.processedClinics, clinicID)
	return 1, nil
}
func (m *batchMockDeliveryTrigger) TriggerFirstVisitFollowUp7D(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerNextVisitReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerVaccineDeadline60(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerVaccineDeadline30(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerBirthdayMessage(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerDormantPrevention120(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerDormantPrevention180(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerDormantPrevention220(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerFilariaAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerFleaTickAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerFoodRefillReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerSuppRefillReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *batchMockDeliveryTrigger) TriggerFirstVisitWelcome(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockDeliveryTrigger) TriggerCheckupFollowUp(_ context.Context, _, _ uint64) error {
	return nil
}

// TestFireHourFor_UsesDefault は fireHours マップにエントリがない場合 defaultFireHourJST (10) を返すことを検証する。
func TestFireHourFor_UsesDefault(t *testing.T) {
	svc := &lstepBatchService{fireHours: map[uint64]int{}}
	assert.Equal(t, defaultFireHourJST, svc.fireHourFor(99))
}

// TestFireHourFor_UsesConfigured は fireHours マップのエントリを返すことを検証する。
func TestFireHourFor_UsesConfigured(t *testing.T) {
	svc := &lstepBatchService{fireHours: map[uint64]int{1: 9, 2: 22}}
	assert.Equal(t, 9, svc.fireHourFor(1))
	assert.Equal(t, 22, svc.fireHourFor(2))
	assert.Equal(t, defaultFireHourJST, svc.fireHourFor(3), "未設定クリニックはデフォルト")
}

// TestRunDeliveryTriggerBatchAllClinics_FireHourFilter は配信時刻が一致するクリニックのみ処理されることを検証する。
// clinic1: fire at 9 JST, clinic2: fire at 10 JST (default)。nowFn = 9:00 JST → clinic1 のみ処理。
func TestRunDeliveryTriggerBatchAllClinics_FireHourFilter(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	trigger := &batchMockDeliveryTrigger{}
	svc := &lstepBatchService{
		fireHours: map[uint64]int{1: 9}, // clinic2 は未設定 → デフォルト 10
		clinicRepo: &mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 1}, {ID: 2}}, nil
			},
		},
		auditSvc:             &batchMockAuditService{},
		lstepDeliveryTrigger: trigger,
		nowFn: func() time.Time {
			return time.Date(2026, 5, 5, 9, 0, 0, 0, jst)
		},
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []uint64{1}, trigger.processedClinics, "9時設定の clinic1 のみ処理される")
}
