package lstep

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestPERF3_CreateCheckupSync_BatchFetchesOwners protects the batch lookup
// contract that prevents an owner-level N+1 query.
func TestPERF3_CreateCheckupSync_BatchFetchesOwners(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	var findByIDsCallCount int64
	lineID := "line-u-test"
	owners := make([]*model.Owner, ownerCount)
	for i := range owners {
		owners[i] = &model.Owner{ID: uint64(i + 1), ClinicID: clinicID, LineUserID: &lineID}
	}

	ownerRepo := &mockOwnerRepository{
		findByIDsFn: func(_ context.Context, gotClinicID uint64, _ []uint64) ([]*model.Owner, error) {
			assert.Equal(t, clinicID, gotClinicID)
			atomic.AddInt64(&findByIDsCallCount, 1)
			return owners, nil
		},
	}
	petRepo := &mockPetRepository{
		countLivingByOwnerIDsFn: func(_ context.Context, gotClinicID uint64, ids []uint64) (map[uint64]int64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			counts := make(map[uint64]int64, len(ids))
			for _, id := range ids {
				counts[id] = 1
			}
			return counts, nil
		},
	}

	svc := &checkupSyncService{
		ownerRepo:    ownerRepo,
		petRepo:      petRepo,
		tagCacheRepo: &mockLstepTagCacheRepository{},
		auditSvc:     &mockAuditService{},
		buildClientFn: func(_ context.Context, gotClinicID uint64) (lstepapi.Client, error) {
			assert.Equal(t, clinicID, gotClinicID)
			return &mockLstepAPIClient{}, nil
		},
	}

	ownerIDs := make([]uint64, ownerCount)
	for i := range ownerIDs {
		ownerIDs[i] = uint64(i + 1)
	}
	_, err := svc.CreateCheckupSync(context.Background(), clinicID, CreateCheckupSyncInput{
		CheckupType: "annual",
		OwnerIDs:    ownerIDs,
		TagName:     "checkup_perf_test",
	}, nil)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&findByIDsCallCount))
}
