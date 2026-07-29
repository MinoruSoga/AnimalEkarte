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

var errDeliveryMonitorTest = errors.New("delivery monitor test error")

type mockDeliveryMonitorTriggerLogRepo struct {
	statusCounts         map[string]int64
	excludedReasons      map[string]int64
	suppressedByPriority int64
	statusCountsErr      error
	excludedReasonsErr   error
	suppressedErr        error
	findByDateRangeFn    func(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error)
}

func (m *mockDeliveryMonitorTriggerLogRepo) Create(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
	return nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CreateIfAbsentToday(_ context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if log != nil && log.ID == 0 {
		log.ID = 1
	}
	return true, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) ExistsTodayByOwnerAndType(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) UpdateStatus(_ context.Context, _, _ uint64, _ string, _ *time.Time, _ *string) error {
	return nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByClinicAndDate(_ context.Context, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	if m.statusCountsErr != nil {
		return nil, m.statusCountsErr
	}
	return m.statusCounts, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	if m.excludedReasonsErr != nil {
		return nil, m.excludedReasonsErr
	}
	return m.excludedReasons, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	if m.suppressedErr != nil {
		return 0, m.suppressedErr
	}
	return m.suppressedByPriority, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByDateRangeWithFilters(ctx context.Context, clinicID uint64, from, to time.Time, triggerType, status string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error) {
	if m.findByDateRangeFn != nil {
		return m.findByDateRangeFn(ctx, clinicID, from, to, triggerType, status, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByOwnerAndDate(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) UpdateSuppressed(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

func TestLstepDeliveryMonitorService_GetSummary_IncludesSuppressedByPriority(t *testing.T) {
	repo := &mockDeliveryMonitorTriggerLogRepo{
		statusCounts: map[string]int64{
			model.TriggerStatusScheduled: 10,
			model.TriggerStatusFired:     5,
			model.TriggerStatusExcluded:  3,
			model.TriggerStatusFailed:    2,
		},
		excludedReasons:      map[string]int64{"opt_out": 2},
		suppressedByPriority: 4,
	}
	svc := NewLstepDeliveryMonitorService(repo)

	got, err := svc.GetSummary(context.Background(), GetDeliveryMonitorSummaryInput{
		ClinicID: 1,
		From:     time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	assert.Equal(t, int64(10), got.Scheduled)
	assert.Equal(t, int64(5), got.Fired)
	assert.Equal(t, int64(3), got.Excluded)
	assert.Equal(t, int64(2), got.Failed)
	assert.Equal(t, int64(4), got.SuppressedByPriority)
	assert.Equal(t, map[string]int64{"opt_out": 2}, got.ExcludedReasonBreakdown)
}

func TestLstepDeliveryMonitorService_GetSummary_ErrorBranches(t *testing.T) {
	baseInput := GetDeliveryMonitorSummaryInput{
		ClinicID: 1,
		From:     time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
	}

	t.Run("CountByStatusAndDateRange error propagates wrapped", func(t *testing.T) {
		repo := &mockDeliveryMonitorTriggerLogRepo{statusCountsErr: errDeliveryMonitorTest}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetSummary(context.Background(), baseInput)
		assert.Error(t, err)
		assert.Equal(t, DeliveryTriggerSummary{}, got)
	})

	t.Run("CountExcludedReasonByDateRange error propagates wrapped", func(t *testing.T) {
		repo := &mockDeliveryMonitorTriggerLogRepo{excludedReasonsErr: errDeliveryMonitorTest}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetSummary(context.Background(), baseInput)
		assert.Error(t, err)
		assert.Equal(t, DeliveryTriggerSummary{}, got)
	})

	t.Run("CountSuppressedByPriorityDateRange error propagates wrapped", func(t *testing.T) {
		repo := &mockDeliveryMonitorTriggerLogRepo{suppressedErr: errDeliveryMonitorTest}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetSummary(context.Background(), baseInput)
		assert.Error(t, err)
		assert.Equal(t, DeliveryTriggerSummary{}, got)
	})
}

// ---- GetLogs tests ----

func TestLstepDeliveryMonitorService_GetLogs(t *testing.T) {
	t.Run("returns error when input is nil", func(t *testing.T) {
		svc := NewLstepDeliveryMonitorService(&mockDeliveryMonitorTriggerLogRepo{})
		got, err := svc.GetLogs(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, DeliveryTriggerLogsPage{}, got)
	})

	t.Run("applies default page/perPage and maps rows", func(t *testing.T) {
		firedAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
		reason := "opt_out"
		var capturedLimit, capturedOffset int
		repo := &mockDeliveryMonitorTriggerLogRepo{
			findByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time, _, _ string, limit, offset int) ([]DeliveryTriggerLogRow, int64, error) {
				capturedLimit = limit
				capturedOffset = offset
				return []DeliveryTriggerLogRow{
					{
						LstepDeliveryTriggerLog: model.LstepDeliveryTriggerLog{
							ID:             1,
							OwnerID:        10,
							TriggerType:    "birthday_message",
							ScheduledAt:    time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
							Status:         model.TriggerStatusFired,
							FiredAt:        &firedAt,
							ExcludedReason: &reason,
						},
						OwnerName: "テスト飼主",
					},
				}, 1, nil
			},
		}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetLogs(context.Background(), &GetDeliveryMonitorLogsInput{ClinicID: 1})
		require.NoError(t, err)
		assert.Equal(t, 20, capturedLimit)
		assert.Equal(t, 0, capturedOffset)
		assert.Equal(t, 1, got.Page)
		assert.Equal(t, 20, got.PerPage)
		assert.Equal(t, int64(1), got.Total)
		require.Len(t, got.Items, 1)
		assert.Equal(t, uint64(1), got.Items[0].ID)
		assert.Equal(t, uint64(10), got.Items[0].OwnerID)
		assert.Equal(t, "テスト飼主", got.Items[0].OwnerName)
		assert.Equal(t, &firedAt, got.Items[0].FiredAt)
		assert.Equal(t, &reason, got.Items[0].ExcludedReason)
	})

	t.Run("clamps perPage above 100 down to 100", func(t *testing.T) {
		var capturedLimit int
		repo := &mockDeliveryMonitorTriggerLogRepo{
			findByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time, _, _ string, limit, _ int) ([]DeliveryTriggerLogRow, int64, error) {
				capturedLimit = limit
				return nil, 0, nil
			},
		}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetLogs(context.Background(), &GetDeliveryMonitorLogsInput{ClinicID: 1, PerPage: 500})
		require.NoError(t, err)
		assert.Equal(t, 100, capturedLimit)
		assert.Equal(t, 100, got.PerPage)
	})

	t.Run("computes offset for page > 1", func(t *testing.T) {
		var capturedOffset int
		repo := &mockDeliveryMonitorTriggerLogRepo{
			findByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, offset int) ([]DeliveryTriggerLogRow, int64, error) {
				capturedOffset = offset
				return nil, 0, nil
			},
		}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetLogs(context.Background(), &GetDeliveryMonitorLogsInput{ClinicID: 1, Page: 3, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, 20, capturedOffset) // (3-1) * 10
		assert.Equal(t, 3, got.Page)
	})

	t.Run("negative page defaults to 1", func(t *testing.T) {
		repo := &mockDeliveryMonitorTriggerLogRepo{}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetLogs(context.Background(), &GetDeliveryMonitorLogsInput{ClinicID: 1, Page: -1})
		require.NoError(t, err)
		assert.Equal(t, 1, got.Page)
	})

	t.Run("propagates repo error wrapped", func(t *testing.T) {
		repo := &mockDeliveryMonitorTriggerLogRepo{
			findByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
				return nil, 0, errDeliveryMonitorTest
			},
		}
		svc := NewLstepDeliveryMonitorService(repo)
		got, err := svc.GetLogs(context.Background(), &GetDeliveryMonitorLogsInput{ClinicID: 1})
		assert.Error(t, err)
		assert.Equal(t, DeliveryTriggerLogsPage{}, got)
	})
}
