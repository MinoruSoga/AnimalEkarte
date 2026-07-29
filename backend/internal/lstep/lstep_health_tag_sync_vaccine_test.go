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

// Note: SyncVaccineDeadlineTag's opt-out / no-line-id / no-deadline / vacRepo-error branches are
// already covered by TestSyncVaccineDeadlineTag in lstep_health_tag_sync_test.go (which always
// exercises a nil lstep client). This file focuses on the branches only reachable with a real
// (mocked) lstep client: AddTag/RemoveTag success & failure.
//
// B-3: 旧公開ラッパー SyncVaccineDeadlineTag（thresholds fetch → syncVaccineDeadlineTagImpl 委譲）は
// 呼び出し元ゼロのため削除済み（batch は syncVaccineDeadlineTagImpl を thresholds 事前取得で直接呼ぶ）。
// ラッパー固有だった「shouldSkipSync 失敗」「thresholds 取得失敗」の分岐は syncVaccineDeadlineTagImpl の
// シグネチャ（thresholds を値で受け取る）には存在しないため、対応するテストは削除した。

func TestSyncVaccineDeadlineTagImpl_ClientInteraction(t *testing.T) {
	lineUID := "U_vaccine_deadline_test"
	near := time.Now().AddDate(0, 0, model.DefaultVaccineDeadlineDays-5)

	t.Run("returns wrapped error when checkOptOut fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) { return nil, errors.New("db error") },
			},
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("returns wrapped error when vacRepo.FindByOwner fails", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return nil, errors.New("db error")
				},
			},
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("deadline soon: adds the tag and upserts cache with the reason", func(t *testing.T) {
		var addedTag, upsertedTag, upsertedReason string
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, tagName string) error { addedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{NextDate: &near}}, nil
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, reason string) error {
					upsertedTag = tagName
					upsertedReason = reason
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}

		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.NoError(t, err)
		assert.Equal(t, PrevVaccineDeadlineTag, addedTag)
		assert.Equal(t, PrevVaccineDeadlineTag, upsertedTag)
		assert.NotEmpty(t, upsertedReason)
	})

	t.Run("deadline soon: AddTag failure is wrapped and returned", func(t *testing.T) {
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return []model.Vaccination{{NextDate: &near}}, nil
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("no deadline soon: removes the tag and deletes cache entry", func(t *testing.T) {
		var removedTag, deletedTag string
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, tagName string) error { removedTag = tagName; return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					return nil, nil // no vaccination records → deadlineSoon=false
				},
			},
			tagCacheRepo: &mockLstepTagCacheRepository{
				deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error {
					deletedTag = tagName
					return nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.NoError(t, err)
		assert.Equal(t, PrevVaccineDeadlineTag, removedTag)
		assert.Equal(t, PrevVaccineDeadlineTag, deletedTag)
	})

	t.Run("no deadline soon: RemoveTag failure propagates (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) { return nil, nil },
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		err := svc.syncVaccineDeadlineTagImpl(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults())
		assert.Error(t, err)
	})

	t.Run("preloaded vaccinations skip vacRepo.FindByOwner (G2F-02 bulk)", func(t *testing.T) {
		var findCalls int
		client := &mockLstepAPIClient{addTagFn: func(_ context.Context, _, _ string) error { return nil }}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
			ownerRepo: &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, LineUserID: &lineUID}, nil
				},
			},
			vacRepo: &mockVaccinationRepoForHealth{
				findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
					findCalls++
					return nil, errors.New("should not fetch")
				},
			},
			tagCacheRepo:  &mockLstepTagCacheRepository{},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
		}
		preloaded := []model.Vaccination{{NextDate: &near}}
		err := svc.syncVaccineDeadlineTagWithInputs(context.Background(), 1, 10, model.HealthPreventionThresholds{}.WithDefaults(), &preloaded)
		assert.NoError(t, err)
		assert.Equal(t, 0, findCalls)
	})
}
