package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type mockDeliveryMonitorTriggerLogRepo struct {
	statusCounts         map[string]int64
	excludedReasons      map[string]int64
	suppressedByPriority int64
}

func (m *mockDeliveryMonitorTriggerLogRepo) Create(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
	return nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) ExistsTodayByOwnerAndType(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) UpdateStatus(_ context.Context, _ uint64, _ string, _ *time.Time, _ *string) error {
	return nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByClinicAndDate(_ context.Context, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return m.statusCounts, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return m.excludedReasons, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return m.suppressedByPriority, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]repository.DeliveryTriggerLogRow, int64, error) {
	return nil, 0, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) ListByOwnerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]repository.DeliveryStatsRow, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) FindByOwnerAndDate(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *mockDeliveryMonitorTriggerLogRepo) UpdateSuppressed(_ context.Context, _ uint64, _ string) error {
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
