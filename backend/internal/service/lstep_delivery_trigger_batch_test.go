package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// mockLstepTriggerPriorityServiceForBatch は LstepTriggerPriorityService の最小モック実装。
// processSingleOwner の抑制 (applySuppression) 分岐を駆動するためだけに使う。
type mockLstepTriggerPriorityServiceForBatch struct {
	getPriorityForFn func(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

func (m *mockLstepTriggerPriorityServiceForBatch) GetByClinicID(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
	return nil, nil
}
func (m *mockLstepTriggerPriorityServiceForBatch) UpdatePriorities(_ context.Context, _ uint64, _ UpdateTriggerPrioritiesInput) error {
	return nil
}
func (m *mockLstepTriggerPriorityServiceForBatch) GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	if m.getPriorityForFn != nil {
		return m.getPriorityForFn(ctx, clinicID, triggerType)
	}
	return 0, nil
}

// mockDeliveryTriggerLogRepoForBatch は LstepDeliveryTriggerLogRepository の完全実装モック。
// 既存の mockDeliveryTriggerLogRepository（lstep_delivery_trigger_service_test.go）は
// FindByOwnerAndDate を固定 nil,nil で返すため、processSingleOwner の抑制分岐を
// 駆動できない。このファイル専用に FindByOwnerAndDate を設定可能にした別モックを用意する。
type mockDeliveryTriggerLogRepoForBatch struct {
	createFn             func(ctx context.Context, log *model.LstepDeliveryTriggerLog) error
	existsTodayFn        func(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error)
	updateStatusFn       func(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error
	findByOwnerAndDateFn func(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error)
	updateSuppressedFn   func(ctx context.Context, clinicID, logID uint64, reason string) error
}

func (m *mockDeliveryTriggerLogRepoForBatch) Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	log.ID = 1
	return nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) ExistsTodayByOwnerAndType(ctx context.Context, clinicID, ownerID uint64, triggerType string, date time.Time) (bool, error) {
	if m.existsTodayFn != nil {
		return m.existsTodayFn(ctx, clinicID, ownerID, triggerType, date)
	}
	return false, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) UpdateStatus(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, clinicID, id, status, firedAt, excludedReason)
	}
	return nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) FindByClinicAndDate(_ context.Context, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]repository.DeliveryTriggerLogRow, int64, error) {
	return nil, 0, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) FindAllByOwnerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]repository.DeliveryStatsRow, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]repository.VisitConversionRow, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) FindByOwnerAndDate(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	if m.findByOwnerAndDateFn != nil {
		return m.findByOwnerAndDateFn(ctx, clinicID, ownerID, date)
	}
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) UpdateSuppressed(ctx context.Context, clinicID, logID uint64, reason string) error {
	if m.updateSuppressedFn != nil {
		return m.updateSuppressedFn(ctx, clinicID, logID, reason)
	}
	return nil
}

func TestRunBatch_BuildClientError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, errors.New("settings error")
			},
		},
	}
	count, errs := svc.runBatch(context.Background(), 1, []uint64{1, 2}, "trigger_x", "tag_x", time.Now())
	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}

func TestRunBatch_ClientNilIsNoop(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: disabledSettings(),
	}
	count, errs := svc.runBatch(context.Background(), 1, []uint64{1, 2}, "trigger_x", "tag_x", time.Now())
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

func TestProcessSingleOwner_AlreadyFiredTodayError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return false, errors.New("db error")
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

func TestProcessSingleOwner_ApplySuppressionError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepoForBatch{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return nil, errors.New("db error")
			},
		},
		prioritySvc: &mockLstepTriggerPriorityServiceForBatch{},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

func TestProcessSingleOwner_SuppressedCreatesLogAndSkips(t *testing.T) {
	var createdLog *model.LstepDeliveryTriggerLog
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepoForBatch{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "higher_priority_trigger"}}, nil
			},
			createFn: func(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
				createdLog = log
				log.ID = 5
				return nil
			},
		},
		prioritySvc: &mockLstepTriggerPriorityServiceForBatch{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "higher_priority_trigger" {
					return 1, nil // lower number = higher priority
				}
				return 10, nil // current trigger has lower priority
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.NoError(t, err)
	assert.False(t, fired)
	if assert.NotNil(t, createdLog) {
		assert.True(t, createdLog.SuppressedByPriority)
	}
}

func TestProcessSingleOwner_SuppressedLogCreateError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepoForBatch{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "higher_priority_trigger"}}, nil
			},
			createFn: func(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
				return errors.New("insert failed")
			},
		},
		prioritySvc: &mockLstepTriggerPriorityServiceForBatch{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "higher_priority_trigger" {
					return 1, nil
				}
				return 10, nil
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

func TestProcessSingleOwner_CheckExclusionError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

func TestProcessSingleOwner_RecordTriggerError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		},
		tagCacheRepo: &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			createFn: func(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
				return errors.New("insert failed")
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

func TestProcessSingleOwner_ExcludedUpdateStatusFailureIsNonFatal(t *testing.T) {
	updateStatusCalled := false
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerExcluded(id), nil
			},
		},
		tagCacheRepo: &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, _ string, _ *time.Time, _ *string) error {
				updateStatusCalled = true
				return errors.New("update failed (non-fatal)")
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.NoError(t, err)
	assert.False(t, fired)
	assert.True(t, updateStatusCalled)
}

func TestProcessSingleOwner_SecondFindByIDFailureAfterExclusionCheck(t *testing.T) {
	callCount := 0
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				callCount++
				if callCount == 1 {
					// checkExclusion 用の1回目の呼び出しは成功させる
					return defaultOwnerWithLine(id), nil
				}
				// processSingleOwner 直下の2回目の呼び出しで失敗させる
				return nil, errors.New("db error on second lookup")
			},
		},
		tagCacheRepo:   &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
	assert.Equal(t, 2, callCount)
}

// applyTagAndLog エラー伝播（AddTag 失敗）は runBatch/processSingleOwner の呼び出し元まで
// 伝わり、count に加算されないことを確認する回帰テスト。
func TestRunBatch_AddTagFailureDoesNotCountAsFired(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return defaultOwnerWithLine(id), nil
			},
		},
		tagCacheRepo:   &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{},
		settingsSvc:    enabledSettings(),
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForDelivery{
				addTagFn: func(_ context.Context, _, _ string) error {
					return errors.New("lstep api error")
				},
			}, nil
		},
	}
	count, errs := svc.runBatch(context.Background(), 1, []uint64{10}, "trigger_x", "tag_x", time.Now())
	assert.Equal(t, 0, count)
	assert.Len(t, errs, 1)
}
