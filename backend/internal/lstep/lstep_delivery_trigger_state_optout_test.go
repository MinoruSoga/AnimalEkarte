package lstep

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCheckExclusion_LstepOptOut(t *testing.T) {
	// LSB-01: opt-out must exclude even when line is linked and no EXCL cache tag remains.
	lineID := "U_optout"
	owner := &model.Owner{
		ID:               42,
		ClinicID:         1,
		LineUserID:       &lineID,
		LstepOptOut:      true,
		DeliveryExcluded: false,
	}
	tagRepo := &mockTagCacheRepoForDelivery{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return nil, nil
		},
	}
	svc := buildDeliverySvc(
		&mockOwnerRepoForDelivery{},
		&mockMedicalRecordRepository{},
		&mockVaccinationRepoForDelivery{},
		&mockPetRepoForDelivery{},
		tagRepo,
		&mockDeliveryTriggerLogRepository{},
		enabledSettings(),
	).(*lstepDeliveryTriggerService)

	excluded, reason, err := svc.checkExclusion(context.Background(), 1, 42, owner)
	require.NoError(t, err)
	assert.True(t, excluded)
	assert.Equal(t, "lstep_opt_out", reason)
}

// TestCheckExclusion_NilOwnerFailClosed pins the product defect at
// lstep_delivery_trigger_state.go:checkExclusion: nil owner must not panic and
// must not be treated as eligible for delivery (fail-closed).
func TestCheckExclusion_NilOwnerFailClosed(t *testing.T) {
	svc := buildDeliverySvc(
		&mockOwnerRepoForDelivery{},
		&mockMedicalRecordRepository{},
		&mockVaccinationRepoForDelivery{},
		&mockPetRepoForDelivery{},
		&mockTagCacheRepoForDelivery{},
		&mockDeliveryTriggerLogRepository{},
		enabledSettings(),
	).(*lstepDeliveryTriggerService)

	excluded, reason, err := svc.checkExclusion(context.Background(), 1, 99, nil)
	require.NoError(t, err)
	assert.True(t, excluded, "nil owner must be excluded from delivery (fail-closed)")
	assert.Equal(t, "owner_missing", reason)
	assert.NotEqual(t, "", reason)
}

// TestProcessSingleOwner_NilOwnerFromRepoFailClosed covers the real call path:
// ownerRepo.FindByID returning (nil, nil) must not panic and must not fire delivery.
func TestProcessSingleOwner_NilOwnerFromRepoFailClosed(t *testing.T) {
	ownerRepo := &mockOwnerRepoForDelivery{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, nil
		},
	}
	svc := &lstepDeliveryTriggerService{
		ownerRepo:      ownerRepo,
		tagCacheRepo:   &mockTagCacheRepoForDelivery{},
		triggerLogRepo: &mockDeliveryTriggerLogRepository{},
	}

	fired, err := svc.processSingleOwner(
		context.Background(),
		&mockLstepClientForDelivery{},
		1, 10,
		"trigger_x", "tag_x",
		time.Now(),
	)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "nil owner without error must map to not-found fail-closed")
	assert.False(t, fired, "nil owner must never count as fired delivery")
}
