package lstep

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

type mockClinicRepository = dormantMockClinicRepository

// batchMockReservationRepo は batch テスト専用 ReservationRepository モック
type batchMockReservationRepo struct {
	findNoShowCandidatesFn func(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	updateFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
}

func (m *batchMockReservationRepo) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{}, nil
}
func (m *batchMockReservationRepo) MarkNoShow(ctx context.Context, clinicID, id uint64) (reservation.NoShowTransition, error) {
	if m.updateFn == nil {
		return reservation.NoShowTransition{Changed: true, PreviousStatus: model.ReservationStatusPending}, nil
	}
	_, err := m.updateFn(ctx, clinicID, id, map[string]any{"status": model.ReservationStatusNoShow})
	return reservation.NoShowTransition{Changed: err == nil, PreviousStatus: model.ReservationStatusPending}, err
}
func (m *batchMockReservationRepo) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	if m.findNoShowCandidatesFn != nil {
		return m.findNoShowCandidatesFn(ctx, clinicID)
	}
	return nil, nil
}

// batchMockMedRecordRepo は batch テスト専用 MedicalRecordRepository モック
type batchMockMedRecordRepo struct {
	findDormantFn       func(ctx context.Context, clinicID uint64, minDays int) ([]medicalrecord.DormantOwnerEntry, error)
	findDormantCursorFn func(ctx context.Context, clinicID uint64, minDays int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error)
}

func (m *batchMockMedRecordRepo) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDays int) ([]medicalrecord.DormantOwnerEntry, error) {
	if m.findDormantFn != nil {
		return m.findDormantFn(ctx, clinicID, minDays)
	}
	return nil, nil
}
func (m *batchMockMedRecordRepo) FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDays int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error) {
	if m.findDormantCursorFn != nil {
		return m.findDormantCursorFn(ctx, clinicID, minDays, afterOwnerID, limit)
	}
	// G2F-04: production uses Cursor; tests that only stub findDormantFn still get one page.
	if m.findDormantFn != nil {
		if afterOwnerID != 0 {
			return nil, nil
		}
		return m.findDormantFn(ctx, clinicID, minDays)
	}
	return nil, nil
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

type batchImmediateTransactor struct{}

func (batchImmediateTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type batchNoShowAuditTxLogger struct {
	entries []*NoShowAuditEntry
	err     error
}

func (m *batchNoShowAuditTxLogger) LogNoShowTransitionTx(_ context.Context, entry *NoShowAuditEntry) error {
	m.entries = append(m.entries, entry)
	return m.err
}

func (m *batchMockAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, action, _ string, _ *uint64, metadata any) error {
	m.called = true
	m.capturedAction = action
	m.capturedMetadata = metadata
	return nil
}

// newBatchService は具象型を返す（B-5: detectDormantOwners/detectNoShowReservations の
// unexport に伴い、テストが interface 外の非公開メソッドを直接呼ぶため）。
func newBatchService(
	resRepo lstepBatchReservationRepository,
	tagSvc lstepBatchTagSyncer,
	clinicRepo lstepBatchClinicRepository,
	medRepo lstepBatchMedicalRecordRepository,
) *lstepBatchService {
	return NewLstepBatchService(
		resRepo, tagSvc, clinicRepo, medRepo,
		&batchMockAuditService{}, &mockLstepSettingsService{}, nil,
		batchImmediateTransactor{}, &batchNoShowAuditTxLogger{},
	).(*lstepBatchService)
}

// newBatchServiceWithAuditSpy は ISSUE-010 監査 metadata 検証用に audit spy を返す。
func newBatchServiceWithAuditSpy(
	resRepo lstepBatchReservationRepository,
	tagSvc lstepBatchTagSyncer,
	clinicRepo lstepBatchClinicRepository,
	medRepo lstepBatchMedicalRecordRepository,
) (*lstepBatchService, *batchMockAuditService) {
	spy := &batchMockAuditService{}
	return NewLstepBatchService(
		resRepo, tagSvc, clinicRepo, medRepo,
		spy, &mockLstepSettingsService{}, nil,
		batchImmediateTransactor{}, &batchNoShowAuditTxLogger{},
	).(*lstepBatchService), spy
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
	entries := []medicalrecord.DormantOwnerEntry{
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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return nil, errors.New("db error")
			},
		},
	)

	count, errs := svc.detectDormantOwners(context.Background(), 1)

	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestDetectDormantOwners_TagSyncError(t *testing.T) {
	entries := []medicalrecord.DormantOwnerEntry{{OwnerID: 5, DaysSince: 250}}
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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
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
	entries := []medicalrecord.DormantOwnerEntry{
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
			findDormantCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
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

// TestRunBatchAllClinics_全滅クリニックでも監査ログが記録されエラー本文は秘匿される は
// perClinic が (0, errs) を返す全滅ケースでも audit が記録される一方、外部API由来の
// エラー本文（LINE user ID 等を含み得る）はログへ出さないことを検証する。
func TestRunBatchAllClinics_全滅クリニックでも監査ログが記録されエラー本文は秘匿される(t *testing.T) {
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
			return 0, []error{errors.New("sensitive-line-user-id-U123"), errors.New("sensitive-owner-name")}
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
	assert.Contains(t, logOut, `"error_count":2`)
	assert.NotContains(t, logOut, "sensitive-line-user-id-U123")
	assert.NotContains(t, logOut, "sensitive-owner-name")
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
