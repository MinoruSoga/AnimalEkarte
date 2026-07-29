package lstep

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
