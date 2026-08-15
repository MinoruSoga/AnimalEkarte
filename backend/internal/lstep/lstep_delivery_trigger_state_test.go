package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- alreadyFiredToday ----

func TestAlreadyFiredToday_True(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return true, nil
			},
		},
	}
	fired, err := svc.alreadyFiredToday(context.Background(), 1, 10, "trigger_x", time.Now())
	assert.NoError(t, err)
	assert.True(t, fired)
}

func TestAlreadyFiredToday_False(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return false, nil
			},
		},
	}
	fired, err := svc.alreadyFiredToday(context.Background(), 1, 10, "trigger_x", time.Now())
	assert.NoError(t, err)
	assert.False(t, fired)
}

func TestAlreadyFiredToday_RepoErrorIsWrapped(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			existsTodayFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return false, errors.New("db error")
			},
		},
	}
	fired, err := svc.alreadyFiredToday(context.Background(), 1, 10, "trigger_x", time.Now())
	assert.Error(t, err)
	assert.False(t, fired)
}

// ---- recordTrigger ----

func TestRecordTrigger_Success(t *testing.T) {
	var created *model.LstepDeliveryTriggerLog
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			createFn: func(_ context.Context, log *model.LstepDeliveryTriggerLog) error {
				log.ID = 42
				created = log
				return nil
			},
		},
	}
	asOf := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	id, claimed, err := svc.recordTrigger(context.Background(), 1, 10, "trigger_x", asOf)
	assert.NoError(t, err)
	assert.True(t, claimed)
	assert.Equal(t, uint64(42), id)
	if assert.NotNil(t, created) {
		assert.Equal(t, uint64(10), created.OwnerID)
		assert.Equal(t, uint64(1), created.ClinicID)
		assert.Equal(t, "trigger_x", created.TriggerType)
		assert.Equal(t, model.TriggerStatusScheduled, created.Status)
		assert.True(t, asOf.Equal(created.ScheduledAt))
	}
}

func TestRecordTrigger_RepoErrorIsWrapped(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			createFn: func(_ context.Context, _ *model.LstepDeliveryTriggerLog) error {
				return errors.New("insert failed")
			},
		},
	}
	id, claimed, err := svc.recordTrigger(context.Background(), 1, 10, "trigger_x", time.Now())
	assert.Error(t, err)
	assert.False(t, claimed)
	assert.Equal(t, uint64(0), id)
}
