package service

// This tag-sync carrier has real owner/pet/chronic-condition test consumers. Delete it with
// those production consumers in BE9-2E (BE9-2F compatibility backstop).

import (
	"context"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

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
