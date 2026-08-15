package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// buildPreventionTagSvc は client!=nil の実 API 呼び出し経路を検証するために
// buildClientFn を直接注入した lstepTagSyncService を組み立てる。
func buildPreventionTagSvc(
	tagCodeRepo *mockLstepTagCodeMappingRepository,
	ownerRepo *mockOwnerRepository,
	checkupRepo *mockCheckupRepository,
	petRepo *mockPetRepository,
	billingItemRepo *mockBillingItemRepository,
	tagCache *mockLstepTagCacheRepository,
	client *mockLstepAPIClient,
) *lstepTagSyncService {
	return &lstepTagSyncService{
		tagCodeRepo:     tagCodeRepo,
		ownerRepo:       ownerRepo,
		checkupRepo:     checkupRepo,
		petRepo:         petRepo,
		billingItemRepo: billingItemRepo,
		settingsSvc:     &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		tagCacheRepo:    tagCache,
		buildClientFn:   func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
	}
}

func TestSyncFilariaTag_ClientCalls(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1
	tagRepo := mappingWith(PrevFilariaTag, model.CodeTypeCheckupType, []string{"フィラリア検査"})

	t.Run("incomplete=true: AddTag成功でタグ更新", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tag string) error { addedTag = tag; return nil },
		}
		var upsertedTag string
		tagCache := &mockLstepTagCacheRepository{
			upsertTagFn: func(_ context.Context, _, _ uint64, tag, _, _ string) error { upsertedTag = tag; return nil },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, tagCache, client)

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, PrevFilariaTag, addedTag)
		assert.Equal(t, PrevFilariaTag, upsertedTag)
	})

	t.Run("incomplete=true: AddTag失敗はエラー伝播", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("complete=true(testDone): RemoveTag成功", func(t *testing.T) {
		var removedTag string
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
		}
		checkups := []model.Checkup{recentCheckup("フィラリア検査")}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, PrevFilariaTag, removedTag)
	})

	t.Run("complete=true: RemoveTag失敗はエラーを返す (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		checkups := []model.Checkup{recentCheckup("フィラリア検査")}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(checkups, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("buildClientエラーは伝播する", func(t *testing.T) {
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return nil, errors.New("client build failed")
		}

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("thresholds取得エラーはラップされて返る", func(t *testing.T) {
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), petRepoWithPets([]model.Pet{dogPet()}, nil), nil, &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.settingsSvc = &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				return model.HealthPreventionThresholds{}, errors.New("db error")
			},
		}

		err := svc.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})
}

func TestSyncFleaTickTag_ClientCalls(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1
	tagRepo := mappingWith(PrevFleaTickTag, model.CodeTypePrescription, []string{"ノミダニ薬"})

	t.Run("処方なし: AddTag成功でタグ更新", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tag string) error { addedTag = tag; return nil },
		}
		var upsertedTag string
		tagCache := &mockLstepTagCacheRepository{
			upsertTagFn: func(_ context.Context, _, _ uint64, tag, _, _ string) error { upsertedTag = tag; return nil },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(false, false), tagCache, client)

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, PrevFleaTickTag, addedTag)
		assert.Equal(t, PrevFleaTickTag, upsertedTag)
	})

	t.Run("処方なし: AddTag失敗はエラー伝播", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(false, false), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("処方あり: RemoveTag成功", func(t *testing.T) {
		var removedTag string
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(true, false), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, PrevFleaTickTag, removedTag)
	})

	t.Run("処方あり: RemoveTag失敗はエラーを返す (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(true, false), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("buildClientエラーは伝播する", func(t *testing.T) {
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(false, false), &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return nil, errors.New("client build failed")
		}

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("thresholds取得エラーはラップされて返る", func(t *testing.T) {
		svc := buildPreventionTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			nil, nil, billingItemRepoReturning(false, false), &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.settingsSvc = &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				return model.HealthPreventionThresholds{}, errors.New("db error")
			},
		}

		err := svc.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})
}
