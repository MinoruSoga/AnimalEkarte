package lstep

// perf_n1_regression_test.go — N+1 クエリ回帰テスト (PERF-1 / PERF-2 / PERF-3)
//
// RED フェーズ: 現行実装に対してアサーションが失敗することを確認してから GREEN 化する。
// 各テストは「呼び出し回数」だけを spy し、ビジネスロジックには触れない。

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// ─────────────────────────────────────────────
// PERF-1: SyncHealthPreventionTagsForClinic が
//   GetHealthPreventionThresholds を N 回でなく 1 回だけ呼ぶこと
// ─────────────────────────────────────────────

func TestPERF1_HealthPreventionThresholdsFetchedOncePerClinic(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	// --- spy: GetHealthPreventionThresholds の呼び出し回数を数える ---
	var thresholdCallCount int64
	settingsSpy := &mockLstepSettingsService{
		getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
			atomic.AddInt64(&thresholdCallCount, 1)
			return model.HealthPreventionThresholds{}.WithDefaults(), nil
		},
	}

	// --- owner repository: ownerCount 件の飼い主を返す ---
	lineID := "line-u-test"
	owners := make([]model.Owner, ownerCount)
	for i := range owners {
		owners[i] = model.Owner{ID: uint64(i + 1), ClinicID: clinicID, LineUserID: &lineID}
	}
	ownerRepo := &mockOwnerRepository{
		findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, afterID uint64, _ int) ([]model.Owner, error) {
			if afterID != 0 {
				return nil, nil
			}
			return owners, nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			o := owners[0]
			return &o, nil
		},
	}

	// --- vaccination repo: 空結果（SyncVaccineDeadlineTag が失敗しないように） ---
	vacRepo := &mockVaccinationRepoForHealth{}

	// --- lstep.Client を nil で返す buildClientFn (API 呼び出しなし) ---
	buildClientFn := func(_ context.Context, _ uint64) (lstep.Client, error) {
		return nil, nil
	}

	// tagCodeRepo は nil 非許容（batch 関数が直接呼ぶ）→ 空結果を返す stub を渡す
	svc := &lstepTagSyncService{
		settingsSvc:   settingsSpy,
		ownerRepo:     ownerRepo,
		vacRepo:       vacRepo,
		tagCacheRepo:  &mockLstepTagCacheRepository{},
		tagCodeRepo:   &mockLstepTagCodeMappingRepository{}, // 空結果 → HLTH系タグはnoop
		buildClientFn: buildClientFn,
	}

	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), clinicID)
	_ = count
	_ = errs

	got := atomic.LoadInt64(&thresholdCallCount)
	// GREEN 条件: 1 回 (hoist 後)
	// RED  条件: ownerCount 回 (現行 SyncVaccineDeadlineTag がループ内で毎回 fetch)
	assert.Equal(t, int64(1), got,
		"GetHealthPreventionThresholds は clinic 単位で 1 回だけ呼ばれるべき (現在 %d 回)", got)
}

// ─────────────────────────────────────────────
// PERF-M1 / PERF-M2: SyncAnnual4CheckupTag / SyncFilariaTag / SyncFleaTickTag /
//   SyncFoodPurchaseTag が tag code mappings と thresholds を N 回でなく
//   clinic 単位で 1 回だけ fetch すること。
//
// PERF-1 のテストでは tagCodeRepo が空スタブのため、これら4関数は
// checkupCodes/testCodes/rxCodes が空で早期 return し、fetch 経路自体を
// 通過しなかった（"M-1 既知ギャップ"）。本テストでは有効な mapping を
// 用意して早期 return を無効化し、fetch 回数を実際に検証する。
// ─────────────────────────────────────────────

func TestPERFM1M2_MappingsAndThresholdsFetchedOncePerClinic(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	// --- spy: GetHealthPreventionThresholds の呼び出し回数 ---
	var thresholdCallCount int64
	settingsSpy := &mockLstepSettingsService{
		getHealthPreventionThresholdsFn: func(_ context.Context, _ uint64) (model.HealthPreventionThresholds, error) {
			atomic.AddInt64(&thresholdCallCount, 1)
			return model.HealthPreventionThresholds{}.WithDefaults(), nil
		},
	}

	// --- spy: FindByClinicIDAndTagName の呼び出し回数をタグ名ごとに数える ---
	mappingCallCounts := make(map[string]*int64)
	for _, tag := range []string{HlthHealthcheckDoneTag, PrevFilariaTag, PrevFleaTickTag, LtvFoodPurchaseTag} {
		var c int64
		mappingCallCounts[tag] = &c
	}
	tagCodeRepoSpy := &mockLstepTagCodeMappingRepository{
		findByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, tagName string) ([]*model.LstepTagCodeMapping, error) {
			if counter, ok := mappingCallCounts[tagName]; ok {
				atomic.AddInt64(counter, 1)
			}
			switch tagName {
			case HlthHealthcheckDoneTag:
				return []*model.LstepTagCodeMapping{{TagName: tagName, CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"健診A"}}}, nil
			case PrevFilariaTag:
				return []*model.LstepTagCodeMapping{{TagName: tagName, CodeType: model.CodeTypeCheckupType, Codes: pq.StringArray{"フィラリア検査"}}}, nil
			case PrevFleaTickTag:
				return []*model.LstepTagCodeMapping{{TagName: tagName, CodeType: model.CodeTypePrescription, Codes: pq.StringArray{"ノミダニ薬"}}}, nil
			case LtvFoodPurchaseTag:
				return []*model.LstepTagCodeMapping{{TagName: tagName, CodeType: model.CodeTypeMerchandiseItem, Codes: pq.StringArray{"フードA"}}}, nil
			}
			return nil, nil
		},
	}

	// --- owner repository: ownerCount 件の飼い主 ---
	lineID := "line-u-test"
	owners := make([]model.Owner, ownerCount)
	for i := range owners {
		owners[i] = model.Owner{ID: uint64(i + 1), ClinicID: clinicID, LineUserID: &lineID}
	}
	ownerRepo := &mockOwnerRepository{
		findAllWithLineUserIDCursorFn: func(_ context.Context, _ uint64, afterID uint64, _ int) ([]model.Owner, error) {
			if afterID != 0 {
				return nil, nil
			}
			return owners, nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
			o := model.Owner{ID: id, ClinicID: clinicID, LineUserID: &lineID}
			return &o, nil
		},
	}

	vacRepo := &mockVaccinationRepoForHealth{}
	petRepo := petRepoWithPets([]model.Pet{dogPet()}, nil)
	checkupRepo := checkupRepoWithResult(nil, nil)
	medRecordRepo := visitSummaryRepo(0)
	billingItemRepo := billingItemRepoReturning(false, false)

	buildClientFn := func(_ context.Context, _ uint64) (lstep.Client, error) {
		return nil, nil
	}

	svc := &lstepTagSyncService{
		settingsSvc:     settingsSpy,
		ownerRepo:       ownerRepo,
		vacRepo:         vacRepo,
		petRepo:         petRepo,
		checkupRepo:     checkupRepo,
		medRecordRepo:   medRecordRepo,
		billingItemRepo: billingItemRepo,
		tagCacheRepo:    &mockLstepTagCacheRepository{},
		tagCodeRepo:     tagCodeRepoSpy,
		buildClientFn:   buildClientFn,
	}

	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), clinicID)
	_ = count
	assert.Empty(t, errs)

	gotThresholds := atomic.LoadInt64(&thresholdCallCount)
	assert.Equal(t, int64(1), gotThresholds,
		"GetHealthPreventionThresholds は clinic 単位で 1 回だけ呼ばれるべき (現在 %d 回)", gotThresholds)

	for _, tag := range []string{HlthHealthcheckDoneTag, PrevFilariaTag, PrevFleaTickTag, LtvFoodPurchaseTag} {
		got := atomic.LoadInt64(mappingCallCounts[tag])
		assert.Equal(t, int64(1), got,
			"FindByClinicIDAndTagName(%s) は clinic 単位で 1 回だけ呼ばれるべき (現在 %d 回)", tag, got)
	}
}

// ─────────────────────────────────────────────
// H-1 回帰テスト: settingsSvc == nil でも DetectDormantOwners がパニックしないこと
// (PERF-2 fix で GetDormantThresholds を直接呼ぶようにしたため、nil ガードが必須)
// ─────────────────────────────────────────────

func TestPERFFOLLOWUP08_LTV_TagCacheFetchedOncePerClinic(t *testing.T) {
	const ownerCount = 3
	const clinicID = uint64(10)

	lineID := "line-u-ltv-test"
	owners := make([]model.Owner, ownerCount)
	for i := range owners {
		owners[i] = model.Owner{ID: uint64(i + 1), ClinicID: clinicID, LineUserID: &lineID}
	}

	accountRepo := newCPMAccountingRepository()
	accountRepo.findOwnersByAnnualRevenueFn = func(_ context.Context, _ uint64) ([]billing.OwnerAnnualRevenue, error) {
		return nil, nil // topN=0 -> 全owner が非top(タグ解除候補)
	}

	var findByOwnerCalls int64
	var findByOwnersCalls int64
	var lastBatchSize int
	tagCacheRepo := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			atomic.AddInt64(&findByOwnerCalls, 1)
			return nil, nil
		},
		findByOwnersFn: func(_ context.Context, _ uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
			atomic.AddInt64(&findByOwnersCalls, 1)
			lastBatchSize = len(ownerIDs)
			return map[uint64][]*model.LstepTagCache{}, nil
		},
	}

	client := &mockLstepAPIClient{}
	svc := &lstepTagSyncService{
		settingsSvc: &mockLstepSettingsService{},
		accountRepo: accountRepo,
		ownerRepo: &mockOwnerRepository{
			findAllWithLineUserIDCursorFn: func(_ context.Context, _, afterID uint64, _ int) ([]model.Owner, error) {
				if afterID != 0 {
					return nil, nil
				}
				return owners, nil
			},
		},
		tagCacheRepo:  tagCacheRepo,
		buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) { return client, nil },
	}

	count, errs := svc.SyncLTVTopPercent(context.Background(), clinicID)
	assert.Empty(t, errs)
	assert.Equal(t, ownerCount, count)

	gotFindByOwner := atomic.LoadInt64(&findByOwnerCalls)
	gotFindByOwners := atomic.LoadInt64(&findByOwnersCalls)
	// GREEN 条件: FindByOwner は 0 回、FindByOwners は 1 回 (hoist 後)
	// RED  条件: FindByOwner が ownerCount 回 (旧実装は owner ごとに再取得)
	assert.Equal(t, int64(0), gotFindByOwner,
		"FindByOwner は呼ばれるべきではない (バッチ化後, 現在 %d 回)", gotFindByOwner)
	assert.Equal(t, int64(1), gotFindByOwners,
		"FindByOwners は clinic 単位で 1 回だけ呼ばれるべき (現在 %d 回)", gotFindByOwners)
	assert.Equal(t, ownerCount, lastBatchSize, "バッチは非top owner全件をまとめて要求するべき")
}
