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

// ---- applyTagAndLog ----

func TestApplyTagAndLog_AddTagSuccessUpdatesFired(t *testing.T) {
	var capturedStatus string
	client := &mockLstepClientForDelivery{
		addTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, _ *string) error {
				capturedStatus = status
				return nil
			},
		},
	}
	err := svc.applyTagAndLog(context.Background(), 1, client, "U1", "tag_x", 10)
	assert.NoError(t, err)
	assert.Equal(t, model.TriggerStatusFired, capturedStatus)
}

func TestApplyTagAndLog_AddTagSuccessButUpdateStatusFails(t *testing.T) {
	client := &mockLstepClientForDelivery{
		addTagFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, _ string, _ *time.Time, _ *string) error {
				return errors.New("update failed")
			},
		},
	}
	err := svc.applyTagAndLog(context.Background(), 1, client, "U1", "tag_x", 10)
	assert.Error(t, err)
}

func TestApplyTagAndLog_AddTagFailsRecordsFailedStatus(t *testing.T) {
	var capturedStatus string
	var capturedReason *string
	client := &mockLstepClientForDelivery{
		addTagFn: func(_ context.Context, _, _ string) error { return errors.New("lstep api error") },
	}
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, status string, _ *time.Time, reason *string) error {
				capturedStatus = status
				capturedReason = reason
				return nil
			},
		},
	}
	err := svc.applyTagAndLog(context.Background(), 1, client, "U1", "tag_x", 10)
	assert.Error(t, err)
	assert.Equal(t, model.TriggerStatusFailed, capturedStatus)
	if assert.NotNil(t, capturedReason) {
		assert.Contains(t, *capturedReason, "tag_x")
	}
}

func TestApplyTagAndLog_AddTagFailsAndUpdateStatusAlsoFailsIsFatal(t *testing.T) {
	// LSA-12 / DEC-35: failed status update failure must join the AddTag error (not silent).
	client := &mockLstepClientForDelivery{
		addTagFn: func(_ context.Context, _, _ string) error { return errors.New("lstep api error") },
	}
	svc := &lstepDeliveryTriggerService{
		triggerLogRepo: &mockDeliveryTriggerLogRepository{
			updateStatusFn: func(_ context.Context, _ uint64, _ string, _ *time.Time, _ *string) error {
				return errors.New("log update also failed")
			},
		},
	}
	err := svc.applyTagAndLog(context.Background(), 1, client, "U1", "tag_x", 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "add lstep tag")
	assert.Contains(t, err.Error(), "failed")
}

// ---- buildClient ----

func TestBuildClient_ClientBuilderFnTakesPrecedence(t *testing.T) {
	injected := &mockLstepClientForDelivery{}
	svc := &lstepDeliveryTriggerService{
		clientBuilderFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
			return injected, nil
		},
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, injected, client)
}

func TestBuildClient_IsSyncEnabledError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
				return false, errors.New("settings error")
			},
		},
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestBuildClient_SyncDisabledReturnsNilClient(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: disabledSettings(),
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.NoError(t, err)
	assert.Nil(t, client)
}

func TestBuildClient_GetRawCredentialsError(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "", "", "", errors.New("credentials error")
			},
		},
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestBuildClient_EmptyAPIKeyReturnsNilClient(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
				return "", "https://api.example.com", "", nil
			},
		},
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.NoError(t, err)
	assert.Nil(t, client)
}

func TestBuildClient_ValidCredentialsReturnsRealClient(t *testing.T) {
	svc := &lstepDeliveryTriggerService{
		settingsSvc: enabledSettings(),
	}
	client, err := svc.buildClient(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
