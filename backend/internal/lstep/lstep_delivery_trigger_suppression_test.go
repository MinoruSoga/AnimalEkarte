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

// ---- local minimal mocks scoped to this file ----

type suppressionMockTriggerLogRepository struct {
	findByOwnerAndDateFn func(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error)
	updateSuppressedFn   func(ctx context.Context, clinicID, logID uint64, reason string) error
}

func (m *suppressionMockTriggerLogRepository) Create(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
	return nil
}
func (m *suppressionMockTriggerLogRepository) CreateIfAbsentToday(_ context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	if log != nil && log.ID == 0 {
		log.ID = 1
	}
	return true, nil
}
func (m *suppressionMockTriggerLogRepository) ExistsTodayByOwnerAndType(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (m *suppressionMockTriggerLogRepository) UpdateStatus(_ context.Context, _, _ uint64, _ string, _ *time.Time, _ *string) error {
	return nil
}
func (m *suppressionMockTriggerLogRepository) FindByClinicAndDate(_ context.Context, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) CountByStatusAndDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) CountExcludedReasonByDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (map[string]int64, error) {
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) CountSuppressedByPriorityDateRange(_ context.Context, _ uint64, _, _ time.Time, _ string) (int64, error) {
	return 0, nil
}
func (m *suppressionMockTriggerLogRepository) FindByDateRangeWithFilters(_ context.Context, _ uint64, _, _ time.Time, _, _ string, _, _ int) ([]DeliveryTriggerLogRow, int64, error) {
	return nil, 0, nil
}
func (m *suppressionMockTriggerLogRepository) CountByTypeAndStatus(_ context.Context, _ uint64, _, _ time.Time) ([]DeliveryStatsRow, error) {
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) CountVisitConversionsByType(_ context.Context, _ uint64, _, _ time.Time, _ int) ([]VisitConversionRow, error) {
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) FindByOwnerAndDate(ctx context.Context, clinicID, ownerID uint64, date time.Time) ([]model.LstepDeliveryTriggerLog, error) {
	if m.findByOwnerAndDateFn != nil {
		return m.findByOwnerAndDateFn(ctx, clinicID, ownerID, date)
	}
	return nil, nil
}
func (m *suppressionMockTriggerLogRepository) UpdateSuppressed(ctx context.Context, clinicID, logID uint64, reason string) error {
	if m.updateSuppressedFn != nil {
		return m.updateSuppressedFn(ctx, clinicID, logID, reason)
	}
	return nil
}

type suppressionMockPriorityService struct {
	getPriorityForFn func(ctx context.Context, clinicID uint64, triggerType string) (int, error)
}

func (m *suppressionMockPriorityService) GetByClinicID(_ context.Context, _ uint64) ([]model.LstepTriggerPriority, error) {
	return nil, nil
}
func (m *suppressionMockPriorityService) UpdatePriorities(_ context.Context, _ uint64, _ UpdateTriggerPrioritiesInput) error {
	return nil
}
func (m *suppressionMockPriorityService) GetPriorityFor(ctx context.Context, clinicID uint64, triggerType string) (int, error) {
	if m.getPriorityForFn != nil {
		return m.getPriorityForFn(ctx, clinicID, triggerType)
	}
	return 0, nil
}

func newSuppressionTestService(triggerLogRepo LstepDeliveryTriggerLogRepository, prioritySvc LstepTriggerPriorityService) *lstepDeliveryTriggerService {
	return &lstepDeliveryTriggerService{
		triggerLogRepo: triggerLogRepo,
		prioritySvc:    prioritySvc,
	}
}

func TestLstepDeliveryTriggerService_ApplySuppression(t *testing.T) {
	asOf := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	t.Run("returns false,nil when prioritySvc is nil (suppression disabled)", func(t *testing.T) {
		svc := newSuppressionTestService(&suppressionMockTriggerLogRepository{}, nil)
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed)
	})

	t.Run("returns wrapped error when FindByOwnerAndDate fails", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newSuppressionTestService(repo, &suppressionMockPriorityService{})
		_, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.Error(t, err)
	})

	t.Run("returns false,nil when no active existing logs (all already suppressed or none)", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return nil, nil
			},
		}
		svc := newSuppressionTestService(repo, &suppressionMockPriorityService{})
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed)
	})

	t.Run("excludes already-suppressed logs from consideration", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{
					{ID: 1, TriggerType: "vaccine", SuppressedByPriority: true},
				}, nil
			},
		}
		svc := newSuppressionTestService(repo, &suppressionMockPriorityService{})
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed, "the only existing log is already suppressed, so it must not count as active")
	})

	t.Run("returns wrapped error when GetPriorityFor fails for the current trigger", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 0, errors.New("priority db error")
				}
				return 1, nil
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		_, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when GetPriorityFor fails for an existing trigger", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "vaccine" {
					return 0, errors.New("priority db error")
				}
				return 1, nil
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		_, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.Error(t, err)
	})

	t.Run("suppresses the current trigger when an existing trigger has higher priority (lower number)", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 5, nil // current: low priority (high number)
				}
				return 1, nil // existing (vaccine): high priority (low number)
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.True(t, suppressed)
	})

	t.Run("demotes existing lower-priority logs when the current trigger has higher priority", func(t *testing.T) {
		var suppressedLogID uint64
		var suppressedReason string
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
			updateSuppressedFn: func(_ context.Context, _, logID uint64, reason string) error {
				suppressedLogID = logID
				suppressedReason = reason
				return nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 1, nil // current: high priority
				}
				return 5, nil // existing (vaccine): low priority
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed, "current trigger fires normally; existing lower-priority logs are demoted instead")
		assert.Equal(t, uint64(1), suppressedLogID)
		assert.NotEmpty(t, suppressedReason)
	})

	t.Run("demote removes previously applied LSTEP tag before suppressing log (G2B-03)", func(t *testing.T) {
		lineID := "U-line-1"
		var removedTag string
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
			updateSuppressedFn: func(_ context.Context, _, _ uint64, _ string) error { return nil },
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 1, nil
				}
				return 5, nil
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		svc.ownerRepo = &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10, LineUserID: &lineID}, nil
			},
		}
		svc.clientBuilderFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepAPIClient{
				removeTagFn: func(_ context.Context, _, tag string) error {
					removedTag = tag
					return nil
				},
			}, nil
		}
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed)
		assert.Equal(t, "vaccine", removedTag)
	})

	t.Run("demote fails closed when RemoveTag fails (G2B-03)", func(t *testing.T) {
		lineID := "U-line-1"
		updateCalled := false
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
			updateSuppressedFn: func(_ context.Context, _, _ uint64, _ string) error {
				updateCalled = true
				return nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 1, nil
				}
				return 5, nil
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		svc.ownerRepo = &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: 10, LineUserID: &lineID}, nil
			},
		}
		svc.clientBuilderFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return &mockLstepAPIClient{
				removeTagFn: func(_ context.Context, _, _ string) error {
					return errors.New("lstep remove failed")
				},
			}, nil
		}
		_, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.Error(t, err)
		assert.False(t, updateCalled, "log demote must not proceed after RemoveTag failure")
	})

	t.Run("returns wrapped error when demoting an existing log fails", func(t *testing.T) {
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
			updateSuppressedFn: func(_ context.Context, _, _ uint64, _ string) error {
				return errors.New("db error")
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, triggerType string) (int, error) {
				if triggerType == "dormant" {
					return 1, nil
				}
				return 5, nil
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		_, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.Error(t, err)
	})

	t.Run("fires normally without demotion when priorities are equal", func(t *testing.T) {
		demoteCalled := false
		repo := &suppressionMockTriggerLogRepository{
			findByOwnerAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) ([]model.LstepDeliveryTriggerLog, error) {
				return []model.LstepDeliveryTriggerLog{{ID: 1, TriggerType: "vaccine"}}, nil
			},
			updateSuppressedFn: func(_ context.Context, _, _ uint64, _ string) error {
				demoteCalled = true
				return nil
			},
		}
		prioritySvc := &suppressionMockPriorityService{
			getPriorityForFn: func(_ context.Context, _ uint64, _ string) (int, error) {
				return 3, nil // equal priority for both current and existing
			},
		}
		svc := newSuppressionTestService(repo, prioritySvc)
		suppressed, err := svc.applySuppression(context.Background(), 1, 10, "dormant", asOf)
		assert.NoError(t, err)
		assert.False(t, suppressed)
		assert.False(t, demoteCalled, "equal priority must not trigger demotion")
	})
}
