package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// segTagSyncSvc は batchMockTagSyncSvc（lstep_batch_service_test.go）を埋め込み、
// lstep_batch_segmentation.go のテストに必要な3メソッドだけを設定可能にする。
// batchMockTagSyncSvc は LstepTagSyncService の全メソッドを実装済みのため、
// 埋め込みによって残りのメソッドは無変更のまま昇格させ、ここでは3メソッドだけを上書きする。
type segTagSyncSvc struct {
	*batchMockTagSyncSvc
	syncLTVTopPercentFn                 func(ctx context.Context, clinicID uint64) (int, []error)
	syncVisitDormantTagsFn              func(ctx context.Context, clinicID, ownerID uint64, daysSince int) error
	syncHealthPreventionTagsForClinicFn func(ctx context.Context, clinicID uint64) (int, []error)
}

func newSegTagSyncSvc() *segTagSyncSvc {
	return &segTagSyncSvc{batchMockTagSyncSvc: &batchMockTagSyncSvc{}}
}

func (m *segTagSyncSvc) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if m.syncLTVTopPercentFn != nil {
		return m.syncLTVTopPercentFn(ctx, clinicID)
	}
	return 0, nil
}

func (m *segTagSyncSvc) SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSince int) error {
	if m.syncVisitDormantTagsFn != nil {
		return m.syncVisitDormantTagsFn(ctx, clinicID, ownerID, daysSince)
	}
	return nil
}

func (m *segTagSyncSvc) SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if m.syncHealthPreventionTagsForClinicFn != nil {
		return m.syncHealthPreventionTagsForClinicFn(ctx, clinicID)
	}
	return 0, nil
}

// segAuditService は batchMockAuditService（lstep_batch_service_test.go）を埋め込み、
// LogLstepOperationWithMetadata のエラー注入だけを可能にする
// （batchMockAuditService はキャプチャ用 spy のみで常に nil を返すため）。
type segAuditService struct {
	*batchMockAuditService
	logMetadataErr error
}

func newSegAuditService() *segAuditService {
	return &segAuditService{batchMockAuditService: &batchMockAuditService{}}
}

func (m *segAuditService) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error {
	if m.logMetadataErr != nil {
		return m.logMetadataErr
	}
	return m.batchMockAuditService.LogLstepOperationWithMetadata(ctx, clinicID, actorID, action, resource, resourceID, metadata)
}

// newSegBatchService は lstep_batch_segmentation.go 用に settingsSvc / auditSvc を
// 個別に差し替え可能な LstepBatchService を構築する
// （lstep_batch_service_test.go の newBatchService はこれらを固定値で配線するため使えない）。
func newSegBatchService(
	clinicRepo lstepBatchClinicRepository,
	medRepo lstepBatchMedicalRecordRepository,
	tagSvc lstepBatchTagSyncer,
	auditSvc lstepBatchAuditService,
	settingsSvc lstepBatchSettingsService,
) LstepBatchService {
	return NewLstepBatchService(
		&batchMockReservationRepo{}, tagSvc, clinicRepo, medRepo,
		auditSvc, settingsSvc, nil,
		batchImmediateTransactor{}, &batchNoShowAuditTxLogger{},
	)
}

// ---- RunLTVTopPercentSyncAllClinics ----

func TestRunLTVTopPercentSyncAllClinics_FetchClinicsError(t *testing.T) {
	svc := newSegBatchService(
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockMedRecordRepo{},
		newSegTagSyncSvc(),
		newSegAuditService(),
		&mockLstepSettingsService{},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())
	assert.Error(t, err)
}

func TestRunLTVTopPercentSyncAllClinics_Success(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	var calledClinicID uint64
	tagSvc.syncLTVTopPercentFn = func(_ context.Context, clinicID uint64) (int, []error) {
		calledClinicID = clinicID
		return 3, nil
	}
	audit := newSegAuditService()
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 7}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, uint64(7), calledClinicID)
	assert.True(t, audit.called, "処理件数>0の場合は監査ログを記録する")
	assert.Equal(t, "batch_ltv_top_percent", audit.capturedAction)
}

func TestRunLTVTopPercentSyncAllClinics_PartialErrorsStillSucceeds(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	tagSvc.syncLTVTopPercentFn = func(_ context.Context, _ uint64) (int, []error) {
		return 2, []error{errors.New("partial fail")}
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())
	assert.NoError(t, err, "個別エラーはバッチ全体を失敗させない")
}

func TestRunLTVTopPercentSyncAllClinics_SkipsWhenSyncDisabled(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	called := false
	tagSvc.syncLTVTopPercentFn = func(_ context.Context, _ uint64) (int, []error) {
		called = true
		return 0, nil
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())
	assert.NoError(t, err)
	assert.False(t, called, "sync 無効なクリニックは SyncLTVTopPercent を呼ばない")
}

func TestRunLTVTopPercentSyncAllClinics_SkipsOnSyncCheckError(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	called := false
	tagSvc.syncLTVTopPercentFn = func(_ context.Context, _ uint64) (int, []error) {
		called = true
		return 0, nil
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("check failed") },
		},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())
	assert.NoError(t, err, "sync 有効判定の失敗はクリニックをスキップしバッチ全体は継続する")
	assert.False(t, called)
}

func TestRunLTVTopPercentSyncAllClinics_AuditLogFailureDoesNotFailBatch(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	tagSvc.syncLTVTopPercentFn = func(_ context.Context, _ uint64) (int, []error) {
		return 1, nil
	}
	audit := newSegAuditService()
	audit.logMetadataErr = errors.New("audit db down")
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	err := svc.RunLTVTopPercentSyncAllClinics(context.Background())
	assert.NoError(t, err, "監査ログ失敗はバッチ全体を失敗させない")
}

// ---- RunVisitDormantSyncAllClinics ----

func TestRunVisitDormantSyncAllClinics_FetchClinicsError(t *testing.T) {
	svc := newSegBatchService(
		&mockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) {
				return nil, errors.New("db error")
			},
		},
		&batchMockMedRecordRepo{},
		newSegTagSyncSvc(),
		newSegAuditService(),
		&mockLstepSettingsService{},
	)

	err := svc.RunVisitDormantSyncAllClinics(context.Background())
	assert.Error(t, err)
}

func TestRunVisitDormantSyncAllClinics_Success(t *testing.T) {
	medRepo := &batchMockMedRecordRepo{
		findDormantFn: func(_ context.Context, _ uint64, minDays int) ([]medicalrecord.DormantOwnerEntry, error) {
			assert.Equal(t, 120, minDays, "VISIT_* バッチは120日閾値を使う")
			return []medicalrecord.DormantOwnerEntry{
				{OwnerID: 1, DaysSince: 200},
				{OwnerID: 2, DaysSince: 250},
			}, nil
		},
	}
	tagSvc := newSegTagSyncSvc()
	synced := make([]uint64, 0)
	tagSvc.syncVisitDormantTagsFn = func(_ context.Context, _, ownerID uint64, _ int) error {
		synced = append(synced, ownerID)
		return nil
	}
	audit := newSegAuditService()
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		medRepo,
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	err := svc.RunVisitDormantSyncAllClinics(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, synced)
	assert.True(t, audit.called)
	assert.Equal(t, "batch_visit_dormant", audit.capturedAction)
}

func TestRunVisitDormantSyncAllClinics_FindDormantEntriesErrorCountsFailed(t *testing.T) {
	medRepo := &batchMockMedRecordRepo{
		findDormantFn: func(_ context.Context, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			return nil, errors.New("db error")
		},
	}
	tagSvc := newSegTagSyncSvc()
	called := false
	tagSvc.syncVisitDormantTagsFn = func(_ context.Context, _, _ uint64, _ int) error {
		called = true
		return nil
	}
	audit := newSegAuditService()
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		medRepo,
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	// LSA-03: FindDormant failure must surface as Failed/audit, not silent (0,nil).
	err := svc.RunVisitDormantSyncAllClinics(context.Background())
	assert.NoError(t, err, "per-clinic find errors do not abort the whole multi-clinic run")
	assert.False(t, called)
	assert.Equal(t, "batch_visit_dormant", audit.capturedAction)
	meta, ok := audit.capturedMetadata.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, 1, meta["error_count"])
	assert.Equal(t, 0, meta["processed_count"])
}

func TestRunVisitDormantSyncAllClinics_PartialSyncErrorsStillSucceeds(t *testing.T) {
	medRepo := &batchMockMedRecordRepo{
		findDormantFn: func(_ context.Context, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			return []medicalrecord.DormantOwnerEntry{{OwnerID: 1, DaysSince: 200}, {OwnerID: 2, DaysSince: 250}}, nil
		},
	}
	tagSvc := newSegTagSyncSvc()
	tagSvc.syncVisitDormantTagsFn = func(_ context.Context, _, ownerID uint64, _ int) error {
		if ownerID == 2 {
			return errors.New("lstep api error")
		}
		return nil
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		medRepo,
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{},
	)

	err := svc.RunVisitDormantSyncAllClinics(context.Background())
	assert.NoError(t, err, "個別オーナー同期の失敗は集約されバッチ全体は継続する")
}

func TestRunVisitDormantSyncAllClinics_SkipsWhenSyncDisabled(t *testing.T) {
	medRepo := &batchMockMedRecordRepo{
		findDormantFn: func(_ context.Context, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			t.Fatal("sync 無効なクリニックでは日次記録取得を呼ばない")
			return nil, nil
		},
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		medRepo,
		newSegTagSyncSvc(),
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		},
	)

	err := svc.RunVisitDormantSyncAllClinics(context.Background())
	assert.NoError(t, err)
}

func TestRunVisitDormantSyncAllClinics_SkipsOnSyncCheckError(t *testing.T) {
	medRepo := &batchMockMedRecordRepo{
		findDormantFn: func(_ context.Context, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			t.Fatal("sync 有効判定に失敗したクリニックでは日次記録取得を呼ばない")
			return nil, nil
		},
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		medRepo,
		newSegTagSyncSvc(),
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("check failed") },
		},
	)

	err := svc.RunVisitDormantSyncAllClinics(context.Background())
	assert.NoError(t, err)
}

// ---- RunHealthPreventionTagSyncAllClinics ----
// TestRunHealthPreventionTagSyncAllClinics_Success / _FetchClinicsError は
// lstep_batch_service_test.go に既存のため、ここでは残りの未カバー分岐のみ追加する。

func TestRunHealthPreventionTagSyncAllClinics_SkipsWhenSyncDisabled(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	called := false
	tagSvc.syncHealthPreventionTagsForClinicFn = func(_ context.Context, _ uint64) (int, []error) {
		called = true
		return 0, nil
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestRunHealthPreventionTagSyncAllClinics_SkipsOnSyncCheckError(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	called := false
	tagSvc.syncHealthPreventionTagsForClinicFn = func(_ context.Context, _ uint64) (int, []error) {
		called = true
		return 0, nil
	}
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		newSegAuditService(),
		&mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("check failed") },
		},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestRunHealthPreventionTagSyncAllClinics_PartialErrorsStillSucceeds(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	tagSvc.syncHealthPreventionTagsForClinicFn = func(_ context.Context, _ uint64) (int, []error) {
		return 4, []error{errors.New("partial")}
	}
	audit := newSegAuditService()
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 9}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.NoError(t, err)
	assert.True(t, audit.called)
	assert.Equal(t, "batch_health_prevention", audit.capturedAction)
}

func TestRunHealthPreventionTagSyncAllClinics_AuditLogFailureDoesNotFailBatch(t *testing.T) {
	tagSvc := newSegTagSyncSvc()
	tagSvc.syncHealthPreventionTagsForClinicFn = func(_ context.Context, _ uint64) (int, []error) {
		return 1, nil
	}
	audit := newSegAuditService()
	audit.logMetadataErr = errors.New("audit failure")
	svc := newSegBatchService(
		&mockClinicRepository{findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		}},
		&batchMockMedRecordRepo{},
		tagSvc,
		audit,
		&mockLstepSettingsService{},
	)

	err := svc.RunHealthPreventionTagSyncAllClinics(context.Background())
	assert.NoError(t, err)
}
