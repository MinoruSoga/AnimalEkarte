package handler

// owner_lstep_mock_test.go — BE9-2D ⑦ carrier: owner_handler_test が使う全メソッド no-op の
// service.LstepTagSyncService mock（移動した medical_record_handler_test 由来の空 struct 版を
// interface 全メソッドへ機械展開。解消 = owner domain 移行時）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagSyncService struct{}

func (m *mockLstepTagSyncService) SyncVaccineTag(_ context.Context, _ uint64, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncVisitCompletionTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCPMStageTag(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncNextVisitTag(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncCheckupTag(_ context.Context, _ uint64, _ uint64, _ uint64, _ time.Time, _ *time.Time) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncPrescriptionTag(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncChronicConditionTags(_ context.Context, _ uint64, _ uint64, _ []string) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncDormantTagsWithThresholds(_ context.Context, _ uint64, _ uint64, _ int, _ model.DormantThresholds) error {
	return nil
}
func (m *mockLstepTagSyncService) ResyncOwnerVaccineTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) ResyncOwnerCheckupTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncLTVTopPercent(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}
func (m *mockLstepTagSyncService) SyncVisitDormantTags(_ context.Context, _ uint64, _ uint64, _ int) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncExclusionTags(_ context.Context, _ uint64, _ uint64) error {
	return nil
}
func (m *mockLstepTagSyncService) SyncHealthPreventionTagsForClinic(_ context.Context, _ uint64) (int, []error) {
	return 0, nil
}
