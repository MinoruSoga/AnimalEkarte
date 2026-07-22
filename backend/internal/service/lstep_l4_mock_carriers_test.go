package service

// L④ moved the owning lstep tests into internal/lstep. These small carriers
// keep residual L⑤/service tests on consumer-side contracts until their batch.

import (
	"context"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepSettingsService struct {
	lstep.LstepSettingsService
	isSyncEnabledFn        func(context.Context, uint64) (bool, error)
	getDormantThresholdsFn func(context.Context, uint64) (model.DormantThresholds, error)
}

func (m *mockLstepSettingsService) IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error) {
	if m.isSyncEnabledFn != nil {
		return m.isSyncEnabledFn(ctx, clinicID)
	}
	return true, nil
}

func (m *mockLstepSettingsService) GetDormantThresholds(ctx context.Context, clinicID uint64) (model.DormantThresholds, error) {
	if m.getDormantThresholdsFn != nil {
		return m.getDormantThresholdsFn(ctx, clinicID)
	}
	return model.DormantThresholds{}.WithDefaults(), nil
}

func (*mockLstepSettingsService) GetCPMV1Thresholds(context.Context, uint64) (model.CPMV1Thresholds, error) {
	return model.CPMV1Thresholds{}.WithDefaults(), nil
}

type mockLstepTagCacheRepository struct{ lstep.LstepTagCacheRepository }

type mockLstepTagSyncService struct {
	lstep.LstepTagSyncService
	syncOwnerAnimalClassificationTagFn func(context.Context, uint64, uint64) error
	syncPetBasicInfoTagsFn             func(context.Context, uint64, uint64) error
	syncChronicConditionTagsFn         func(context.Context, uint64, uint64, []string) error
}

func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncOwnerAnimalClassificationTagFn != nil {
		return m.syncOwnerAnimalClassificationTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
	if m.syncPetBasicInfoTagsFn != nil {
		return m.syncPetBasicInfoTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, codes []string) error {
	if m.syncChronicConditionTagsFn != nil {
		return m.syncChronicConditionTagsFn(ctx, clinicID, ownerID, codes)
	}
	return nil
}

func (*mockLstepTagSyncService) SyncExclusionTags(context.Context, uint64, uint64) error { return nil }
func (*mockLstepTagSyncService) SyncDormantTagsWithThresholds(context.Context, uint64, uint64, int, model.DormantThresholds) error {
	return nil
}
func (*mockLstepTagSyncService) SyncLTVTopPercent(context.Context, uint64) (int, []error) {
	return 0, nil
}
func (*mockLstepTagSyncService) SyncVisitDormantTags(context.Context, uint64, uint64, int) error {
	return nil
}
func (*mockLstepTagSyncService) SyncHealthPreventionTagsForClinic(context.Context, uint64) (int, []error) {
	return 0, nil
}
