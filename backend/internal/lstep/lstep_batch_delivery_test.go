package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockLstepDeliveryTriggerBatch は lstep_batch_delivery.go テスト専用の
// LstepDeliveryTriggerService モック。全13バッチトリガーの呼び出し回数を記録する。
type mockLstepDeliveryTriggerBatch struct {
	callCounts map[string]int
	triggerFn  func(name string, clinicID uint64, asOf time.Time) (int, []error)
}

func (m *mockLstepDeliveryTriggerBatch) call(name string, clinicID uint64, asOf time.Time) (int, []error) {
	if m.callCounts == nil {
		m.callCounts = map[string]int{}
	}
	m.callCounts[name]++
	if m.triggerFn != nil {
		return m.triggerFn(name, clinicID, asOf)
	}
	return 0, nil
}

func (m *mockLstepDeliveryTriggerBatch) TriggerFirstVisitFollowUp3D(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerFirstVisitFollowUp3D", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerFirstVisitFollowUp7D(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerFirstVisitFollowUp7D", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerNextVisitReminder(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerNextVisitReminder", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerVaccineDeadline60(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerVaccineDeadline60", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerVaccineDeadline30(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerVaccineDeadline30", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerBirthdayMessage(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerBirthdayMessage", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerDormantPrevention180(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerDormantPrevention180", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerDormantPrevention210(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerDormantPrevention210", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerDormantPrevention240(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerDormantPrevention240", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerDormantPrevention365(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerDormantPrevention365", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerFilariaAlert(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerFilariaAlert", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerFleaTickAlert(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerFleaTickAlert", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerFoodRefillReminder(_ context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return m.call("TriggerFoodRefillReminder", clinicID, asOf)
}
func (m *mockLstepDeliveryTriggerBatch) TriggerFirstVisitWelcome(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepDeliveryTriggerBatch) TriggerCheckupFollowUp(_ context.Context, _, _ uint64) error {
	return nil
}

// totalCalls は全トリガーメソッドの呼び出し合計回数を返す。
func (m *mockLstepDeliveryTriggerBatch) totalCalls() int {
	total := 0
	for _, c := range m.callCounts {
		total += c
	}
	return total
}

// jstTime は UTC 時刻を渡して JST 換算した time.Time を組み立てるテスト用ヘルパー。
// UTC 01:00 -> JST 10:00 (配信バッチ実行時刻)。
func jstFireHourTime() time.Time {
	return time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)
}

// jstOutsideFireHourTime は配信バッチ実行時刻(10:00 JST)の外側を表す。UTC 00:00 -> JST 09:00。
func jstOutsideFireHourTime() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

// ================================================================
// runDeliveryTriggersForClinic
// ================================================================

func TestRunDeliveryTriggersForClinic_NilTrigger(t *testing.T) {
	svc := &lstepBatchService{lstepDeliveryTrigger: nil}
	count, errs := svc.runDeliveryTriggersForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.Nil(t, errs)
}

func TestRunDeliveryTriggersForClinic_AggregatesCountsAndErrors(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(name string, _ uint64, _ time.Time) (int, []error) {
			if name == "TriggerFilariaAlert" {
				return 2, []error{errors.New("partial fail")}
			}
			return 1, nil
		},
	}
	svc := &lstepBatchService{lstepDeliveryTrigger: trigger}

	count, errs := svc.runDeliveryTriggersForClinic(context.Background(), 5)

	// 13 個のトリガーのうち 12 個が1件、TriggerFilariaAlert のみ2件 → 合計14件
	assert.Equal(t, 14, count)
	assert.Len(t, errs, 1)
	assert.Equal(t, 13, trigger.totalCalls(), "全13トリガーがちょうど1回ずつ呼ばれること")
	assert.Equal(t, 1, trigger.callCounts["TriggerFoodRefillReminder"], "末尾のトリガーも含め全て実行されること")
}

// ================================================================
// RunDeliveryTriggerBatchAllClinics
// ================================================================

func TestRunDeliveryTriggerBatchAllClinics_FindAllError(t *testing.T) {
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return nil, errors.New("db error")
		},
	}
	svc := &lstepBatchService{clinicRepo: clinicRepo, nowFn: jstFireHourTime}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.Error(t, err)
}

func TestRunDeliveryTriggerBatchAllClinics_SkipsOutsideFireHour(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		},
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             &batchMockAuditService{},
		nowFn:                jstOutsideFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.NoError(t, err)
	assert.Empty(t, trigger.callCounts, "実行時刻外はトリガーを一切実行しない")
}

func TestRunDeliveryTriggerBatchAllClinics_SettingsSvcNilFailsClosed(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		},
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             &batchMockAuditService{},
		settingsSvc:          nil,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.Error(t, err)
	assert.Empty(t, trigger.callCounts, "settingsSvc が nil ならトリガーを実行しない")
}

func TestRunDeliveryTriggerBatchAllClinics_SkipsWhenSyncCheckErrors(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, errors.New("db error")
		},
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             &batchMockAuditService{},
		settingsSvc:          settingsSvc,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.NoError(t, err)
	assert.Empty(t, trigger.callCounts, "sync確認エラー時はそのクリニックをスキップする")
}

func TestRunDeliveryTriggerBatchAllClinics_SkipsWhenSyncDisabled(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, nil
		},
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             &batchMockAuditService{},
		settingsSvc:          settingsSvc,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.NoError(t, err)
	assert.Empty(t, trigger.callCounts, "sync無効クリニックはスキップする")
}

func TestRunDeliveryTriggerBatchAllClinics_LogsAuditWhenCountPositive(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 9}}, nil
		},
	}
	audit := &batchMockAuditService{}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             audit,
		settingsSvc:          settingsSvc,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.NoError(t, err)
	assert.True(t, audit.called, "処理件数が正のとき監査ログが記録される")
	assert.Equal(t, "batch_delivery_trigger", audit.capturedAction)

	meta, ok := audit.capturedMetadata.(map[string]any)
	require.True(t, ok, "metadata は map[string]any で渡される")
	assert.Equal(t, "batch_delivery_trigger", meta["operation"])
	assert.Equal(t, 13, meta["processed_count"])
	assert.Equal(t, 0, meta["error_count"])
}

func TestRunDeliveryTriggerBatchAllClinics_NoAuditWhenCountZero(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 0, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 9}}, nil
		},
	}
	audit := &batchMockAuditService{}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             audit,
		settingsSvc:          settingsSvc,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	require.NoError(t, err)
	assert.False(t, audit.called, "処理件数0件のときは監査ログを記録しない")
}

// spyDeliveryAuditService は監査ログ失敗時の best-effort 挙動検証用の spy。
type spyDeliveryAuditService struct {
	batchMockAuditService
	logErr error
}

func (s *spyDeliveryAuditService) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error {
	_ = s.batchMockAuditService.LogLstepOperationWithMetadata(ctx, clinicID, actorID, action, resource, resourceID, metadata)
	return s.logErr
}

func TestRunDeliveryTriggerBatchAllClinics_AuditLogFailureIsBestEffort(t *testing.T) {
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, _ time.Time) (int, []error) { return 1, nil },
	}
	clinicRepo := &mockClinicRepository{
		findAllFn: func(_ context.Context) ([]model.Clinic, error) {
			return []model.Clinic{{ID: 1}, {ID: 2}}, nil
		},
	}
	audit := &spyDeliveryAuditService{logErr: errors.New("audit db down")}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
	}
	svc := &lstepBatchService{
		clinicRepo:           clinicRepo,
		lstepDeliveryTrigger: trigger,
		auditSvc:             audit,
		settingsSvc:          settingsSvc,
		nowFn:                jstFireHourTime,
	}

	err := svc.RunDeliveryTriggerBatchAllClinics(context.Background())
	assert.NoError(t, err, "監査ログ失敗はバッチ全体を失敗させない（best-effort）")
	assert.True(t, audit.called)
}
