package service

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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

func (m *batchMockReservationRepo) FindAll(_ context.Context, _ []uint64, _, _ int, _, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *batchMockReservationRepo) FindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *batchMockReservationRepo) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
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
func (m *batchMockReservationRepo) AcquireBookingLock(_ context.Context, _ uint64) error {
	return nil
}
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
func (m *batchMockReservationRepo) CountByTypeAndStartTime(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockReservationRepo) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *batchMockReservationRepo) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *batchMockReservationRepo) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
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

func (m *batchMockReservationRepo) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockReservationRepo) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *batchMockReservationRepo) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

// batchMockMedRecordRepo は batch テスト専用 MedicalRecordRepository モック
type batchMockMedRecordRepo struct {
	findDormantFn       func(ctx context.Context, clinicID uint64, minDays int) ([]repository.DormantOwnerEntry, error)
	findDormantCursorFn func(ctx context.Context, clinicID uint64, minDays int, afterOwnerID uint64, limit int) ([]repository.DormantOwnerEntry, error)
}

func (m *batchMockMedRecordRepo) FindAll(_ context.Context, _ []uint64, _ repository.MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (m *batchMockMedRecordRepo) FindByID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) Create(_ context.Context, _ *model.MedicalRecord) error { return nil }
func (m *batchMockMedRecordRepo) Update(_ context.Context, _, _ uint64, _ map[string]any, _ *int) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) Delete(_ context.Context, _, _ uint64) error { return nil }

// LockByIDForUpdate は X-11 finalize-lock テスト用に FindByID と同じ挙動へ委譲する。
func (m *batchMockMedRecordRepo) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}
func (m *batchMockMedRecordRepo) CountByPetID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *batchMockMedRecordRepo) FindFirstVisitDateByPetID(_ context.Context, _, _ uint64) (*time.Time, error) {
	return nil, nil
}
func (m *batchMockMedRecordRepo) CountEstimatesByMedicalRecordID(_ context.Context, _, _ uint64) (int64, error) {
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
func (m *batchMockMedRecordRepo) FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDays int, afterOwnerID uint64, limit int) ([]repository.DormantOwnerEntry, error) {
	if m.findDormantCursorFn != nil {
		return m.findDormantCursorFn(ctx, clinicID, minDays, afterOwnerID, limit)
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

func (m *batchMockMedRecordRepo) DeleteDraftByAppointmentID(_ context.Context, _, _ uint64) error {
	return nil
}

// batchMockTagSyncSvc は batch テスト専用 LstepTagSyncService モック
type batchMockTagSyncSvc struct {
	syncDormantTagsWithThresholdsFn func(ctx context.Context, clinicID, ownerID uint64, daysSince int, thresholds model.DormantThresholds) error
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
func (m *batchMockTagSyncSvc) ResyncOwnerVaccineTags(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *batchMockTagSyncSvc) ResyncOwnerCheckupTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncLTVTopPercent(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *batchMockTagSyncSvc) SyncVisitDormantTags(_ context.Context, _, _ uint64, _ int) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncExclusionTags(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *batchMockTagSyncSvc) SyncHealthPreventionTagsForClinic(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}

func (m *batchMockTagSyncSvc) SyncDormantTagsWithThresholds(ctx context.Context, clinicID, ownerID uint64, daysSince int, thresholds model.DormantThresholds) error {
	if m.syncDormantTagsWithThresholdsFn != nil {
		return m.syncDormantTagsWithThresholdsFn(ctx, clinicID, ownerID, daysSince, thresholds)
	}
	return nil
}

type batchMockAuditService struct {
	// ISSUE-010: 引数捕捉用の spy フィールド（既存テストでは未使用 — nil のまま）。
	capturedAction   string
	capturedMetadata any
	called           bool
}

func (m *batchMockAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *batchMockAuditService) LogEntry(_ context.Context, _ *AuditLogInput) error {
	return nil
}
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
func (m *batchMockAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *batchMockAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *batchMockAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (m *batchMockAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}

// newBatchService は具象型を返す（B-5: detectDormantOwners/detectNoShowReservations の
// unexport に伴い、テストが interface 外の非公開メソッドを直接呼ぶため）。
func newBatchService(
	resRepo repository.ReservationRepository,
	tagSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRepo repository.MedicalRecordRepository,
) *lstepBatchService {
	return NewLstepBatchService(resRepo, tagSvc, clinicRepo, medRepo, &batchMockAuditService{}, &mockLstepSettingsService{}, nil).(*lstepBatchService)
}

// newBatchServiceWithAuditSpy は ISSUE-010 監査 metadata 検証用に audit spy を返す。
func newBatchServiceWithAuditSpy(
	resRepo repository.ReservationRepository,
	tagSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	medRepo repository.MedicalRecordRepository,
) (*lstepBatchService, *batchMockAuditService) {
	spy := &batchMockAuditService{}
	return NewLstepBatchService(resRepo, tagSvc, clinicRepo, medRepo, spy, &mockLstepSettingsService{}, nil).(*lstepBatchService), spy
}

func TestDetectNoShowReservations_Success(t *testing.T) {
	ownerID := uint64(42)
	now := time.Now()
	reservations := []model.Reservation{
		{ID: 1, OwnerID: &ownerID, StartTime: now},
		{ID: 2, OwnerID: nil, StartTime: now},
	}

	svc := newBatchService(
		&batchMockReservationRepo{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return reservations, nil
			},
		},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{},
	)

	count, errs := svc.detectNoShowReservations(context.Background(), 1)

	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
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

	count, errs := svc.detectNoShowReservations(context.Background(), 1)

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

	count, errs := svc.detectNoShowReservations(context.Background(), 1)

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
			// PERF-2: DetectDormantOwners は SyncDormantTagsWithThresholds を呼ぶ。
			syncDormantTagsWithThresholdsFn: func(_ context.Context, _, ownerID uint64, _ int, _ model.DormantThresholds) error {
				synced = append(synced, ownerID)
				return nil
			},
		},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return entries, nil
			},
		},
	)

	count, errs := svc.detectDormantOwners(context.Background(), 1)

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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return nil, errors.New("db error")
			},
		},
	)

	count, errs := svc.detectDormantOwners(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestDetectDormantOwners_TagSyncError(t *testing.T) {
	entries := []repository.DormantOwnerEntry{{OwnerID: 5, DaysSince: 250}}
	svc := newBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{
			// PERF-2: DetectDormantOwners は SyncDormantTagsWithThresholds を呼ぶ。
			syncDormantTagsWithThresholdsFn: func(_ context.Context, _, _ uint64, _ int, _ model.DormantThresholds) error {
				return errors.New("lstep api error")
			},
		},
		&mockClinicRepository{},
		&batchMockMedRecordRepo{
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return entries, nil
			},
		},
	)

	count, errs := svc.detectDormantOwners(context.Background(), 1)

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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
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

// TestRunBatchAllClinics_全滅クリニックでも監査ログが記録されエラー内容がログに出る は
// perClinic が (0, errs) を返す全滅ケースでも audit が記録され、エラー本文がログに出ることを検証する（BE7-2）。
func TestRunBatchAllClinics_全滅クリニックでも監査ログが記録されエラー内容がログに出る(t *testing.T) {
	var logBuf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelError})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	clinics := []model.Clinic{{ID: 99}}
	svc, spy := newBatchServiceWithAuditSpy(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return clinics, nil
			},
		},
		&batchMockMedRecordRepo{},
	)

	err := svc.runBatchAllClinics(
		context.Background(),
		"test-batch",
		"test-batch",
		"synced",
		"batch_test_wipeout",
		nil,
		func(_ context.Context, _ uint64) (int, []error) {
			return 0, []error{errors.New("wipeout failure A"), errors.New("wipeout failure B")}
		},
	)
	assert.NoError(t, err)
	assert.True(t, spy.called, "全滅クリニックでも監査ログが記録される")
	assert.Equal(t, "batch_test_wipeout", spy.capturedAction)

	meta, ok := spy.capturedMetadata.(map[string]any)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, 0, meta["processed_count"])
	assert.Equal(t, 2, meta["error_count"])

	logOut := logBuf.String()
	assert.Contains(t, logOut, "wipeout failure A")
	assert.Contains(t, logOut, "wipeout failure B")
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
