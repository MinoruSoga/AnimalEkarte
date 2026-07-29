package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// buildCheckupTagSvc は client!=nil の実 API 呼び出し経路を検証するために
// buildClientFn を直接注入した lstepTagSyncService を組み立てる。
func buildCheckupTagSvc(
	tagCodeRepo *mockLstepTagCodeMappingRepository,
	ownerRepo *mockOwnerRepository,
	checkupRepo *mockCheckupRepository,
	medRecordRepo *mockMedicalRecordRepository,
	tagCache *mockLstepTagCacheRepository,
	client *mockLstepAPIClient,
) *lstepTagSyncService {
	return &lstepTagSyncService{
		tagCodeRepo:   tagCodeRepo,
		ownerRepo:     ownerRepo,
		checkupRepo:   checkupRepo,
		medRecordRepo: medRecordRepo,
		settingsSvc:   &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		tagCacheRepo:  tagCache,
		buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
	}
}

func TestSyncHealthcheckTagsWithMappings_ClientCalls(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1
	tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})

	t.Run("hasHealthcheck=true: AddTag/RemoveTag成功でタグ更新", func(t *testing.T) {
		var addedTag, removedTag string
		client := &mockLstepAPIClient{
			addTagFn:    func(_ context.Context, _, tag string) error { addedTag = tag; return nil },
			removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
		}
		var upsertedTag, deletedTag string
		tagCache := &mockLstepTagCacheRepository{
			upsertTagFn: func(_ context.Context, _, _ uint64, tag, _, _ string) error { upsertedTag = tag; return nil },
			deleteTagFn: func(_ context.Context, _, _ uint64, tag string) error { deletedTag = tag; return nil },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, tagCache, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, HlthHealthcheckDoneTag, addedTag)
		assert.Equal(t, HlthHealthcheckNeverTag, removedTag)
		assert.Equal(t, HlthHealthcheckDoneTag, upsertedTag)
		assert.Equal(t, HlthHealthcheckNeverTag, deletedTag)
	})

	t.Run("hasHealthcheck=true: AddTag失敗はエラー伝播", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("hasHealthcheck=true: RemoveTag失敗はエラーを返す (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("hasHealthcheck=false: AddTag(never)/RemoveTag(done)成功", func(t *testing.T) {
		var addedTag, removedTag string
		client := &mockLstepAPIClient{
			addTagFn:    func(_ context.Context, _, tag string) error { addedTag = tag; return nil },
			removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, HlthHealthcheckNeverTag, addedTag)
		assert.Equal(t, HlthHealthcheckDoneTag, removedTag)
	})

	t.Run("hasHealthcheck=false: AddTag(never)失敗はエラー伝播", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("hasHealthcheck=false: RemoveTag(done)失敗はエラーを返す (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("cachedMappingsが渡されればtagCodeRepoは呼ばれない", func(t *testing.T) {
		called := false
		panicRepo := &mockLstepTagCodeMappingRepository{
			findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) ([]*model.LstepTagCodeMapping, error) {
				called = true
				return nil, nil
			},
		}
		client := &mockLstepAPIClient{}
		cached := []*model.LstepTagCodeMapping{{CodeType: model.CodeTypeCheckupType, Codes: []string{"健診A"}}}
		svc := buildCheckupTagSvc(panicRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, client)

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, cached, nil)
		assert.NoError(t, err)
		assert.False(t, called, "cachedMappings が渡された場合 tagCodeRepo.FindByClinicIDAndTagName は呼ばれない")
	})

	t.Run("cachedThresholdsが渡されればsettingsSvc.GetHealthPreventionThresholdsは呼ばれない", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		cachedThresholds := &model.HealthPreventionThresholds{LookbackDays: 365}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, client)
		// settingsSvc の GetHealthPreventionThresholds が呼ばれたらpanicするモックに差し替える
		svc.settingsSvc = &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				t.Fatal("GetHealthPreventionThresholds should not be called when cachedThresholds is provided")
				return model.HealthPreventionThresholds{}, nil
			},
		}

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, cachedThresholds)
		assert.NoError(t, err)
	})

	t.Run("thresholds取得エラーはラップされて返る", func(t *testing.T) {
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.settingsSvc = &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
			getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
				return model.HealthPreventionThresholds{}, errors.New("db error")
			},
		}

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("buildClientエラーは伝播する", func(t *testing.T) {
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), nil, &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return nil, errors.New("client build failed")
		}

		err := svc.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})
}

func TestSyncAnnual4CheckupTag_ClientCalls(t *testing.T) {
	ctx := context.Background()
	const clinicID, ownerID uint64 = 10, 1
	tagRepo := mappingWith(HlthHealthcheckDoneTag, model.CodeTypeCheckupType, []string{"健診A"})

	t.Run("qualified=true: AddTag成功", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tag string) error { addedTag = tag; return nil },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), visitSummaryRepo(2), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, HlthAnnual4CheckupTag, addedTag)
	})

	t.Run("qualified=true: AddTag失敗はエラー伝播", func(t *testing.T) {
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), visitSummaryRepo(2), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("qualified=false: RemoveTag成功", func(t *testing.T) {
		var removedTag string
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tag string) error { removedTag = tag; return nil },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), visitSummaryRepo(0), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, HlthAnnual4CheckupTag, removedTag)
	})

	t.Run("qualified=false: RemoveTag失敗はエラーを返す (G2B-01)", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error { return errors.New("api error") },
		}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult(nil, nil), visitSummaryRepo(0), &mockLstepTagCacheRepository{}, client)

		err := svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("buildClientエラーは伝播する", func(t *testing.T) {
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			checkupRepoWithResult([]model.Checkup{recentCheckup("健診A")}, nil), visitSummaryRepo(2), &mockLstepTagCacheRepository{}, &mockLstepAPIClient{})
		svc.buildClientFn = func(_ context.Context, _ uint64) (lstep.Client, error) {
			return nil, errors.New("client build failed")
		}

		err := svc.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil)
		assert.Error(t, err)
	})

	t.Run("preloaded checkups/visit summary skip repo fetches (G2F-02)", func(t *testing.T) {
		var findCheckupCalls, findVisitCalls int
		panicCheckup := &mockCheckupRepository{
			findByOwnerIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
				findCheckupCalls++
				return nil, errors.New("should not fetch")
			},
		}
		panicVisit := &mockMedicalRecordRepository{
			findOwnerVisitSummaryFn: func(_ context.Context, _, _ uint64) (*medicalrecord.OwnerVisitSummary, error) {
				findVisitCalls++
				return nil, errors.New("should not fetch")
			},
		}
		client := &mockLstepAPIClient{}
		svc := buildCheckupTagSvc(tagRepo, ownerRepoReturning(defaultOwnerWithLineID()),
			panicCheckup, panicVisit, &mockLstepTagCacheRepository{}, client)

		checkups := []model.Checkup{recentCheckup("健診A")}
		summary := &medicalrecord.OwnerVisitSummary{AnnualCount: 2}
		err := svc.syncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, nil, nil, &checkups, &summary)
		assert.NoError(t, err)
		assert.Equal(t, 0, findCheckupCalls)
		assert.Equal(t, 0, findVisitCalls)
	})
}
