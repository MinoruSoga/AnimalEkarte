package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
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

func validOwnerRepoForDelivery() *mockOwnerRepoForDelivery {
	return &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			return defaultOwnerWithLine(id), nil
		},
	}
}

func (m *mockDeliveryTriggerLogRepoForBatch) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if err := m.Create(ctx, log); err != nil {
		return false, err
	}
	return true, nil
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
func (m *mockDeliveryTriggerLogRepoForBatch) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
	return nil, 0, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *mockDeliveryTriggerLogRepoForBatch) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
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
		ownerRepo: validOwnerRepoForDelivery(),
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
		ownerRepo: validOwnerRepoForDelivery(),
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

func TestProcessSingleOwner_ValidatesOwnerScopeBeforeSuppressionLog(t *testing.T) {
	created := false
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), ownerID)
				return nil, errors.New("owner is outside clinic scope")
			},
		},
		triggerLogRepo: &mockDeliveryTriggerLogRepoForBatch{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "higher_priority_trigger"}}, nil
			},
			createFn: func(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
				created = true
				return nil
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
	assert.False(t, created)
}

func TestProcessSingleOwner_ValidatesOwnerScopeBeforeDuplicateLookup(t *testing.T) {
	duplicateLookupCalled := false
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("owner is outside clinic scope")
			},
		},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				duplicateLookupCalled = true
				return false, nil
			},
		},
	}

	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
	assert.False(t, duplicateLookupCalled)
}

func TestProcessSingleOwner_SuppressedCreatesLogAndSkips(t *testing.T) {
	var createdLog *model.LstepDeliveryTriggerLog
	svc := &lstepDeliveryTriggerService{
		ownerRepo: validOwnerRepoForDelivery(),
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
		ownerRepo: validOwnerRepoForDelivery(),
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

func TestProcessSingleOwner_ExcludedUpdateStatusFailureIsFatal(t *testing.T) {
	// LSA-12 / DEC-35: excluded status update failure must not return (false, nil).
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
				return errors.New("update failed")
			},
		},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
	assert.True(t, updateStatusCalled)
}

func TestProcessSingleOwner_ReusesScopedOwnerLookup(t *testing.T) {
	callCount := 0
	svc := &lstepDeliveryTriggerService{
		ownerRepo: &mockOwnerRepoForDelivery{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				callCount++
				return defaultOwnerWithLine(id), nil
			},
		},
		tagCacheRepo:   &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{},
	}
	fired, err := svc.processSingleOwner(context.Background(), &mockLstepClientForDelivery{}, 1, 10, "trigger_x", "tag_x", time.Now())
	assert.NoError(t, err)
	assert.True(t, fired)
	assert.Equal(t, 1, callCount)
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

// ---- N+1 bulk preload regression (BE-ACT-LSTEP-DELIVERY-BATCH-NPLUS1) ----

// bulkCountOwnerRepo counts FindByID vs FindByIDs for runBatch bulk-load proofs.
type bulkCountOwnerRepo struct {
	findByIDCalls  int
	findByIDsCalls int
	owners         map[uint64]*model.Owner
}

func (m *bulkCountOwnerRepo) FindByID(_ context.Context, _, id uint64) (*model.Owner, error) {
	m.findByIDCalls++
	if o, ok := m.owners[id]; ok {
		return o, nil
	}
	return nil, errors.New("owner not found")
}

func (m *bulkCountOwnerRepo) FindByIDs(_ context.Context, _ uint64, ids []uint64) ([]*model.Owner, error) {
	m.findByIDsCalls++
	out := make([]*model.Owner, 0, len(ids))
	for _, id := range ids {
		if o, ok := m.owners[id]; ok {
			out = append(out, o)
		}
	}
	return out, nil
}

// bulkCountTagRepo counts FindByOwner vs FindByOwners.
type bulkCountTagRepo struct {
	findByOwnerCalls  int
	findByOwnersCalls int
	tagsByOwner       map[uint64][]*model.LstepTagCache
}

func (m *bulkCountTagRepo) FindByOwner(_ context.Context, _, ownerID uint64) ([]*model.LstepTagCache, error) {
	m.findByOwnerCalls++
	return m.tagsByOwner[ownerID], nil
}

func (m *bulkCountTagRepo) FindByOwners(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
	m.findByOwnersCalls++
	out := make(map[uint64][]*model.LstepTagCache, len(ownerIDs))
	for _, id := range ownerIDs {
		if tags, ok := m.tagsByOwner[id]; ok {
			out[id] = tags
		}
	}
	return out, nil
}

func (m *bulkCountTagRepo) FindOwnerIDsByTag(_ context.Context, _ uint64, _ string) ([]uint64, error) {
	return nil, nil
}

// bulkCountTriggerLogRepo counts per-owner daily/suppression reads vs clinic-day bulk.
type bulkCountTriggerLogRepo struct {
	existsTodayCalls           int
	findByOwnerAndDateCalls    int
	findByDateRangeFilterCalls int
	createCalls                int
	// dayLogs is the full clinic-day set returned by FindByDateRangeWithFilters.
	dayLogs []model.LstepDeliveryTriggerLog
}

func (m *bulkCountTriggerLogRepo) Create(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
	m.createCalls++
	log.ID = uint64(m.createCalls)
	return nil
}
func (m *bulkCountTriggerLogRepo) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if err := m.Create(ctx, log); err != nil {
		return false, err
	}
	return true, nil
}
func (m *bulkCountTriggerLogRepo) ExistsTodayByOwnerAndType(_ context.Context, _, ownerID uint64, triggerType string, _ time.Time) (bool, error) {
	m.existsTodayCalls++
	for i := range m.dayLogs {
		if m.dayLogs[i].OwnerID == ownerID && m.dayLogs[i].TriggerType == triggerType {
			return true, nil
		}
	}
	return false, nil
}
func (m *bulkCountTriggerLogRepo) UpdateStatus(_ context.Context, _, _ uint64, _ string, _ *time.Time, _ *string) error {
	return nil
}
func (m *bulkCountTriggerLogRepo) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *bulkCountTriggerLogRepo) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *bulkCountTriggerLogRepo) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}
func (m *bulkCountTriggerLogRepo) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error) {
	m.findByDateRangeFilterCalls++
	total := int64(len(m.dayLogs))
	if offset >= len(m.dayLogs) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(m.dayLogs) {
		end = len(m.dayLogs)
	}
	rows := make([]DeliveryTriggerLogRow, 0, end-offset)
	for _, log := range m.dayLogs[offset:end] {
		rows = append(rows, DeliveryTriggerLogRow{LstepDeliveryTriggerLog: log})
	}
	return rows, total, nil
}
func (m *bulkCountTriggerLogRepo) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *bulkCountTriggerLogRepo) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
	return nil, nil
}
func (m *bulkCountTriggerLogRepo) FindByOwnerAndDate(_ context.Context, _, ownerID uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	m.findByOwnerAndDateCalls++
	var out []model.LstepDeliveryTriggerLog
	for i := range m.dayLogs {
		if m.dayLogs[i].OwnerID == ownerID {
			out = append(out, m.dayLogs[i])
		}
	}
	return out, nil
}
func (m *bulkCountTriggerLogRepo) UpdateSuppressed(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

type deliveryBatchQueryCounts struct {
	ownerFindByID         int
	ownerFindByIDs        int
	tagFindByOwner        int
	tagFindByOwners       int
	logExistsToday        int
	logFindByOwnerAndDate int
	logFindByDateRange    int
	firedCount            int
}

func runDeliveryBatchWithCounts(t *testing.T, ownerCount int, dayLogs []model.LstepDeliveryTriggerLog, tagsByOwner map[uint64][]*model.LstepTagCache) deliveryBatchQueryCounts {
	t.Helper()
	owners := make(map[uint64]*model.Owner, ownerCount)
	ids := make([]uint64, ownerCount)
	for i := 0; i < ownerCount; i++ {
		id := uint64(i + 1)
		ids[i] = id
		owners[id] = defaultOwnerWithLine(id)
	}
	ownerRepo := &bulkCountOwnerRepo{owners: owners}
	tagRepo := &bulkCountTagRepo{tagsByOwner: tagsByOwner}
	if tagRepo.tagsByOwner == nil {
		tagRepo.tagsByOwner = map[uint64][]*model.LstepTagCache{}
	}
	logRepo := &bulkCountTriggerLogRepo{dayLogs: dayLogs}

	svc := &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   tagRepo,
		triggerLogRepo: logRepo,
		settingsSvc:    enabledSettings(),
		prioritySvc: &mockLstepTriggerPriorityServiceForBatch{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "higher_priority_trigger" {
					return 1, nil
				}
				return 10, nil
			},
		},
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForDelivery{}, nil
		},
	}
	asOf := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	fired, errs := svc.runBatch(context.Background(), 1, ids, "trigger_x", "tag_x", asOf)
	assert.Empty(t, errs)

	return deliveryBatchQueryCounts{
		ownerFindByID:         ownerRepo.findByIDCalls,
		ownerFindByIDs:        ownerRepo.findByIDsCalls,
		tagFindByOwner:        tagRepo.findByOwnerCalls,
		tagFindByOwners:       tagRepo.findByOwnersCalls,
		logExistsToday:        logRepo.existsTodayCalls,
		logFindByOwnerAndDate: logRepo.findByOwnerAndDateCalls,
		logFindByDateRange:    logRepo.findByDateRangeFilterCalls,
		firedCount:            fired,
	}
}

// TestDeliveryTriggerBatch_BulkQueryCountsConstantWithOwnerCount proves owner / daily /
// suppression / tag-cache read categories are CONSTANT w.r.t. candidate owner count
// (not linear). Writes (CreateIfAbsentToday) remain per-owner by design (LSA-15).
func TestDeliveryTriggerBatch_BulkQueryCountsConstantWithOwnerCount(t *testing.T) {
	// Pre-seed one higher-priority log for owner 1 so suppression path is exercised
	// without changing call-count shape across N.
	dayLogsSmall := []model.LstepDeliveryTriggerLog{
		{ID: 100, OwnerID: 1, ClinicID: 1, TriggerType: "higher_priority_trigger"},
	}
	dayLogsLarge := []model.LstepDeliveryTriggerLog{
		{ID: 100, OwnerID: 1, ClinicID: 1, TriggerType: "higher_priority_trigger"},
	}

	small := runDeliveryBatchWithCounts(t, 3, dayLogsSmall, nil)
	large := runDeliveryBatchWithCounts(t, 30, dayLogsLarge, nil)

	// Bulk categories: equal across owner counts.
	assert.Equal(t, small.ownerFindByIDs, large.ownerFindByIDs, "owner bulk FindByIDs must be constant w.r.t. owner count")
	assert.Equal(t, small.tagFindByOwners, large.tagFindByOwners, "tag-cache bulk FindByOwners must be constant w.r.t. owner count")
	assert.Equal(t, small.logFindByDateRange, large.logFindByDateRange, "day-log bulk FindByDateRangeWithFilters must be constant w.r.t. owner count")

	// Per-owner read categories must be zero when bulk preload is active.
	assert.Equal(t, 0, small.ownerFindByID, "per-owner FindByID must not run after bulk owner load")
	assert.Equal(t, 0, large.ownerFindByID, "per-owner FindByID must not run after bulk owner load")
	assert.Equal(t, 0, small.tagFindByOwner, "per-owner tag FindByOwner must not run after bulk tag load")
	assert.Equal(t, 0, large.tagFindByOwner, "per-owner tag FindByOwner must not run after bulk tag load")
	assert.Equal(t, 0, small.logExistsToday, "per-owner ExistsToday must not run after day-log bulk load")
	assert.Equal(t, 0, large.logExistsToday, "per-owner ExistsToday must not run after day-log bulk load")
	assert.Equal(t, 0, small.logFindByOwnerAndDate, "per-owner FindByOwnerAndDate must not run after day-log bulk load")
	assert.Equal(t, 0, large.logFindByOwnerAndDate, "per-owner FindByOwnerAndDate must not run after day-log bulk load")

	// Sanity: bulk calls themselves are non-zero (preload actually ran).
	assert.Equal(t, 1, small.ownerFindByIDs)
	assert.Equal(t, 1, small.tagFindByOwners)
	assert.Equal(t, 1, small.logFindByDateRange)

	// Owner 1 suppressed; owners 2..N fire. Linear growth in fired is expected.
	assert.Equal(t, 2, small.firedCount)
	assert.Equal(t, 29, large.firedCount)
}

func TestDeliveryTriggerBatch_BulkOptOutExcludesWithoutPerOwnerTagRead(t *testing.T) {
	owners := map[uint64]*model.Owner{
		1: {ID: 1, ClinicID: 1, LineUserID: strPtr("U1"), LstepOptOut: true},
		2: defaultOwnerWithLine(2),
	}
	ownerRepo := &bulkCountOwnerRepo{owners: owners}
	tagRepo := &bulkCountTagRepo{tagsByOwner: map[uint64][]*model.LstepTagCache{
		2: {{TagName: exclTagDeliveryStop}},
	}}
	logRepo := &bulkCountTriggerLogRepo{}

	svc := &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   tagRepo,
		triggerLogRepo: logRepo,
		settingsSvc:    enabledSettings(),
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForDelivery{}, nil
		},
	}
	fired, errs := svc.runBatch(context.Background(), 1, []uint64{1, 2}, "trigger_x", "tag_x", time.Now())
	assert.Empty(t, errs)
	assert.Equal(t, 0, fired, "opt-out and excl tag must both exclude")
	assert.Equal(t, 0, ownerRepo.findByIDCalls)
	assert.Equal(t, 0, tagRepo.findByOwnerCalls)
	assert.Equal(t, 1, tagRepo.findByOwnersCalls)
}

func TestDeliveryTriggerBatch_DailyClaimSkipsAlreadyFired(t *testing.T) {
	owners := map[uint64]*model.Owner{
		1: defaultOwnerWithLine(1),
		2: defaultOwnerWithLine(2),
	}
	ownerRepo := &bulkCountOwnerRepo{owners: owners}
	tagRepo := &bulkCountTagRepo{tagsByOwner: map[uint64][]*model.LstepTagCache{}}
	logRepo := &bulkCountTriggerLogRepo{
		dayLogs: []model.LstepDeliveryTriggerLog{
			{ID: 1, OwnerID: 1, ClinicID: 1, TriggerType: "trigger_x"},
		},
	}
	svc := &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   tagRepo,
		triggerLogRepo: logRepo,
		settingsSvc:    enabledSettings(),
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForDelivery{}, nil
		},
	}
	fired, errs := svc.runBatch(context.Background(), 1, []uint64{1, 2}, "trigger_x", "tag_x", time.Now())
	assert.Empty(t, errs)
	assert.Equal(t, 1, fired, "only owner 2 should fire; owner 1 already claimed today")
	assert.Equal(t, 0, logRepo.existsTodayCalls, "daily check must use bulk day logs")
	assert.Equal(t, 1, logRepo.createCalls, "only the non-claimed owner creates a log")
}

func TestDeliveryTriggerBatch_SuppressionUsesDayLogBulk(t *testing.T) {
	owners := map[uint64]*model.Owner{1: defaultOwnerWithLine(1)}
	ownerRepo := &bulkCountOwnerRepo{owners: owners}
	tagRepo := &bulkCountTagRepo{tagsByOwner: map[uint64][]*model.LstepTagCache{}}
	logRepo := &bulkCountTriggerLogRepo{
		dayLogs: []model.LstepDeliveryTriggerLog{
			{ID: 9, OwnerID: 1, ClinicID: 1, TriggerType: "higher_priority_trigger"},
		},
	}
	var created *model.LstepDeliveryTriggerLog
	// Wrap create to capture suppressed log: use custom create via embedding override.
	// bulkCountTriggerLogRepo.Create assigns IDs; inspect after run via createCalls + re-run capture.
	svc := &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   tagRepo,
		triggerLogRepo: logRepo,
		settingsSvc:    enabledSettings(),
		prioritySvc: &mockLstepTriggerPriorityServiceForBatch{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "higher_priority_trigger" {
					return 1, nil
				}
				return 10, nil
			},
		},
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepClientForDelivery{}, nil
		},
	}
	// Capturing create path for suppressed-log assertion; day logs served via bulk adapter.
	capturing := &mockDeliveryTriggerLogRepoForBatch{
		findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
			t.Fatal("FindByOwnerAndDate must not be called when day-log bulk is installed")
			return nil, nil
		},
		createFn: func(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
			created = log
			log.ID = 1
			return nil
		},
	}
	adapter := &suppressionBulkLogAdapter{
		dayLogs:   logRepo.dayLogs,
		capturing: capturing,
	}
	svc.triggerLogRepo = adapter

	fired, errs := svc.runBatch(context.Background(), 1, []uint64{1}, "trigger_x", "tag_x", time.Now())
	assert.Empty(t, errs)
	assert.Equal(t, 0, fired)
	if assert.NotNil(t, created) {
		assert.True(t, created.SuppressedByPriority)
	}
	assert.Equal(t, 0, adapter.findByOwnerAndDateCalls)
	assert.Equal(t, 0, adapter.existsTodayCalls)
	assert.GreaterOrEqual(t, adapter.findByDateRangeCalls, 1)
}

// suppressionBulkLogAdapter is a test double that counts bulk vs per-owner log reads
// while delegating Create to a capturing mock.
type suppressionBulkLogAdapter struct {
	dayLogs                 []model.LstepDeliveryTriggerLog
	capturing               *mockDeliveryTriggerLogRepoForBatch
	existsTodayCalls        int
	findByOwnerAndDateCalls int
	findByDateRangeCalls    int
}

func (m *suppressionBulkLogAdapter) Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error {
	return m.capturing.Create(ctx, log)
}
func (m *suppressionBulkLogAdapter) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	return m.capturing.CreateIfAbsentToday(ctx, log)
}
func (m *suppressionBulkLogAdapter) ExistsTodayByOwnerAndType(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
	m.existsTodayCalls++
	return false, nil
}
func (m *suppressionBulkLogAdapter) UpdateStatus(ctx context.Context, clinicID, id uint64, status string, firedAt *time.Time, excludedReason *string) error {
	return m.capturing.UpdateStatus(ctx, clinicID, id, status, firedAt, excludedReason)
}
func (m *suppressionBulkLogAdapter) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *suppressionBulkLogAdapter) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *suppressionBulkLogAdapter) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}
func (m *suppressionBulkLogAdapter) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
	m.findByDateRangeCalls++
	rows := make([]DeliveryTriggerLogRow, len(m.dayLogs))
	for i, log := range m.dayLogs {
		rows[i] = DeliveryTriggerLogRow{LstepDeliveryTriggerLog: log}
	}
	return rows, int64(len(rows)), nil
}
func (m *suppressionBulkLogAdapter) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *suppressionBulkLogAdapter) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
	return nil, nil
}
func (m *suppressionBulkLogAdapter) FindByOwnerAndDate(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	m.findByOwnerAndDateCalls++
	return nil, nil
}
func (m *suppressionBulkLogAdapter) UpdateSuppressed(_ context.Context, _, _ uint64, _ string) error {
	return nil
}
