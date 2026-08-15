package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLstepTagSyncServiceDisabledSyncSkipsBeforeRepositories(t *testing.T) {
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, nil
		},
	}
	svc := NewLstepTagSyncService(
		settingsSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil, // errorCounterRepo — nil because sync is disabled, counter is never reached
		nil, // tagCodeRepo
		nil, // billingItemRepo
		nil, // tagConfigRepo
	)
	ctx := context.Background()
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "SyncVaccineTag", run: func() error { return svc.SyncVaccineTag(ctx, 1, 2, 3) }},
		{name: "SyncVisitCompletionTags", run: func() error { return svc.SyncVisitCompletionTags(ctx, 1, 2) }},
		{name: "SyncOwnerAnimalClassificationTags", run: func() error { return svc.SyncOwnerAnimalClassificationTags(ctx, 1, 2) }},
		{name: "SyncPetBasicInfoTags", run: func() error { return svc.SyncPetBasicInfoTags(ctx, 1, 2) }},
		{name: "SyncCPMStageTag", run: func() error { return svc.SyncCPMStageTag(ctx, 1, 2) }},
		{name: "SyncNextVisitTag", run: func() error { return svc.SyncNextVisitTag(ctx, 1, 2) }},
		{name: "SyncCheckupTag", run: func() error { return svc.SyncCheckupTag(ctx, 1, 2, 3, now, nil) }},
		{name: "SyncPrescriptionTag", run: func() error { return svc.SyncPrescriptionTag(ctx, 1, 2) }},
		{name: "SyncChronicConditionTags", run: func() error { return svc.SyncChronicConditionTags(ctx, 1, 2, []string{"kidney"}) }},
		{name: "SyncDormantTagsWithThresholds", run: func() error {
			return svc.SyncDormantTagsWithThresholds(ctx, 1, 2, 180, model.DormantThresholds{})
		}},
		{name: "ResyncOwnerVaccineTags", run: func() error { return svc.ResyncOwnerVaccineTags(ctx, 1, 2) }},
		{name: "ResyncOwnerCheckupTags", run: func() error { return svc.ResyncOwnerCheckupTags(ctx, 1, 2) }},
		{name: "SyncVisitDormantTags", run: func() error { return svc.SyncVisitDormantTags(ctx, 1, 2, 120) }},
		{name: "SyncExclusionTags", run: func() error { return svc.SyncExclusionTags(ctx, 1, 2) }},
		// B-3: SyncHealthcheckTags/SyncAnnual4CheckupTag/SyncVaccineDeadlineTag/SyncFilariaTag/
		// SyncFleaTickTag/SyncFoodPurchaseTag の公開ラッパーは呼び出し元ゼロのため削除済み。
		// これらの WithMappings 版・syncVaccineDeadlineTagImpl は LstepTagSyncService
		// interface に含まれないため本テーブル（interface 経由の一括検証）の対象外。
		// B-5: SyncCPMStageTagV2 は自ファイル内専用のため syncCPMStageTagV2（非公開）に
		// unexport 済みで、同様に interface から除外されたため本テーブルの対象外。
		// isSyncEnabled=false による早期 skip は各メソッドが共通で呼ぶ resolveSyncTarget の
		// 責務であり、本テーブルの残り約15メソッドで引き続き検証される。
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, tc.run())
		})
	}
}

// ---- isPetBasicInfoTag ----

func TestIsPetBasicInfoTag(t *testing.T) {
	c1Prefixes := []*model.LstepAutoManagedPrefix{
		{Prefix: "breed_", Category: "C1"},
		{Prefix: "sex_", Category: "C1"},
		{Prefix: "pet_birthday_", Category: "C1"},
		{Prefix: "birth_year_", Category: "C1"},
		{Prefix: "spay_neutered", Category: "C1"},
		{Prefix: "intact", Category: "C1"},
	}
	cases := []struct {
		tag  string
		want bool
	}{
		{"breed_shiba_inu", true},
		{"breed_mix_dog", true},
		{"sex_male", true},
		{"sex_female", true},
		{"sex_unknown", true},
		{"pet_birthday_04-20", true},
		{"birth_year_2020", true},
		{"spay_neutered", true},
		{"intact", true},
		{"ltv_amount_5", false},
		{"cpm_core", false},
		{"dormant_180d", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			assert.Equal(t, tc.want, isPetBasicInfoTagWithPrefixes(tc.tag, c1Prefixes))
		})
	}

	t.Run("prefixes outside category C1 are ignored even on an exact/prefix match", func(t *testing.T) {
		mixedPrefixes := []*model.LstepAutoManagedPrefix{
			{Prefix: "breed_", Category: "C2"}, // not C1 -> must not match despite the same prefix text
			{Prefix: "sex_", Category: "C1"},
		}
		assert.False(t, isPetBasicInfoTagWithPrefixes("breed_shiba_inu", mixedPrefixes), "C2-category prefix must not classify the tag as C1")
		assert.True(t, isPetBasicInfoTagWithPrefixes("sex_male", mixedPrefixes))
	})

	t.Run("empty dbPrefixes never matches", func(t *testing.T) {
		assert.False(t, isPetBasicInfoTagWithPrefixes("breed_shiba_inu", nil))
		assert.False(t, isPetBasicInfoTagWithPrefixes("", nil))
	})
}

// ---- isDormantTag ----

func TestIsDormantTag(t *testing.T) {
	cases := []struct {
		tag  string
		want bool
	}{
		{"dormant_180d", true},
		{"dormant_210d", true},
		{"dormant_240d", true},
		{"dormant_365d", true},
		{"dormant", false},
		{"dormant_90d", false},
		{"cpm_dormant", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			assert.Equal(t, tc.want, isDormantTag(tc.tag))
		})
	}
}

// ---- buildVisitTags ----

func TestBuildVisitTags_BothDates(t *testing.T) {
	first := time.Date(2022, 3, 10, 0, 0, 0, 0, time.UTC)
	last := time.Date(2024, 11, 5, 0, 0, 0, 0, time.UTC)
	summary := &medicalrecord.OwnerVisitSummary{
		FirstVisitAt: &first,
		LastVisitAt:  &last,
		AnnualCount:  7,
	}
	tags := buildVisitTags(summary, 60_000)
	assert.Contains(t, tags, "first_visit_2022-03-10")
	assert.Contains(t, tags, "last_visit_2024-11-05")
	assert.Contains(t, tags, "ltv_amount_5")
	assert.Contains(t, tags, "visit_count_annual_5")
	assert.Len(t, tags, 4)
}

func TestBuildVisitTags_NoDates(t *testing.T) {
	summary := &medicalrecord.OwnerVisitSummary{AnnualCount: 1}
	tags := buildVisitTags(summary, 0)
	assert.NotContains(t, tags, "first_visit_")
	assert.NotContains(t, tags, "last_visit_")
	assert.Contains(t, tags, "ltv_amount_0")
	assert.Contains(t, tags, "visit_count_annual_1")
	assert.Len(t, tags, 2)
}

// ---- buildPetBasicInfoTags ----

func TestBuildPetBasicInfoTags_SingleDogMale(t *testing.T) {
	dog := model.AnimalSpecies{Name: "犬種"}
	bd := time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)
	nd := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	pets := []model.Pet{
		{
			Breed:         "柴犬",
			Gender:        model.PetGenderMale,
			BirthDate:     &bd,
			NeuteredDate:  &nd,
			AnimalSpecies: &dog,
		},
	}
	tags := buildPetBasicInfoTags(pets)
	// breed resolved from breedCodeMap
	assert.Contains(t, tags, "breed_shiba_inu")
	assert.Contains(t, tags, "sex_male")
	assert.Contains(t, tags, "pet_birthday_04-20")
	assert.Contains(t, tags, "birth_year_2020")
	assert.Contains(t, tags, "spay_neutered")
	assert.NotContains(t, tags, "intact")
}

func TestBuildPetBasicInfoTags_UnknownBreedCatFallback(t *testing.T) {
	cat := model.AnimalSpecies{Name: "猫"}
	pets := []model.Pet{
		{
			Breed:         "ふわふわ猫",
			Gender:        model.PetGenderFemale,
			AnimalSpecies: &cat,
		},
	}
	tags := buildPetBasicInfoTags(pets)
	assert.Contains(t, tags, "breed_mix_cat")
	assert.Contains(t, tags, "sex_female")
	assert.Contains(t, tags, "intact")
	assert.NotContains(t, tags, "spay_neutered")
}

func TestBuildPetBasicInfoTags_EmptyBreedOtherFallback(t *testing.T) {
	pets := []model.Pet{
		{
			Breed:  "",
			Gender: model.PetGenderUnknown,
		},
	}
	tags := buildPetBasicInfoTags(pets)
	assert.Contains(t, tags, "breed_mix_other")
	assert.Contains(t, tags, "sex_unknown")
}

func TestBuildPetBasicInfoTags_MultiPet_BothNeuteredStates(t *testing.T) {
	nd := time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC)
	pets := []model.Pet{
		{Breed: "", Gender: model.PetGenderMale, NeuteredDate: &nd},
		{Breed: "", Gender: model.PetGenderFemale},
	}
	tags := buildPetBasicInfoTags(pets)
	assert.Contains(t, tags, "spay_neutered")
	assert.Contains(t, tags, "intact")
	assert.Contains(t, tags, "sex_male")
	assert.Contains(t, tags, "sex_female")
}

func TestBuildPetBasicInfoTags_Empty(t *testing.T) {
	tags := buildPetBasicInfoTags(nil)
	assert.Empty(t, tags)
}

// ---- buildLatestVaccineTagSet (ISSUE-004) ----

func TestBuildLatestVaccineTagSet_EmptyReturnsEmpty(t *testing.T) {
	tagSet := buildLatestVaccineTagSet(nil)
	assert.Empty(t, tagSet)
}

func TestBuildLatestVaccineTagSet_SkipsNilVaccine(t *testing.T) {
	vaccinations := []model.Vaccination{
		{Date: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), Vaccine: nil},
	}
	tagSet := buildLatestVaccineTagSet(vaccinations)
	assert.Empty(t, tagSet)
}

func TestBuildLatestVaccineTagSet_SingleDog(t *testing.T) {
	dog := model.VaccineSpeciesDog
	vaccinations := []model.Vaccination{
		{
			Date:    time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog},
		},
	}
	tagSet := buildLatestVaccineTagSet(vaccinations)
	_, hasDog := tagSet["vaccine_dog_2024-05-01"]
	assert.True(t, hasDog)
	assert.Len(t, tagSet, 1)
}

func TestBuildLatestVaccineTagSet_KeepsOnlyLatestPerSpecies(t *testing.T) {
	// 同一種別に複数記録 → 最新日のみ保持される（仕様: 同一カテゴリ1タグ）
	dog := model.VaccineSpeciesDog
	vaccinations := []model.Vaccination{
		{
			Date:    time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog},
		},
		{
			Date:    time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
			Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog},
		},
		{
			Date:    time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC), // 最新
			Vaccine: &model.Vaccine{Name: "DHPP", Species: &dog},
		},
	}
	tagSet := buildLatestVaccineTagSet(vaccinations)
	_, hasLatest := tagSet["vaccine_dog_2025-01-10"]
	assert.True(t, hasLatest, "最新接種日のタグのみ採用される")
	assert.NotContains(t, tagSet, "vaccine_dog_2024-05-01")
	assert.NotContains(t, tagSet, "vaccine_dog_2023-06-15")
	assert.Len(t, tagSet, 1)
}

func TestBuildLatestVaccineTagSet_BothSpeciesAndRabies(t *testing.T) {
	both := model.VaccineSpeciesBoth
	dog := model.VaccineSpeciesDog
	vaccinations := []model.Vaccination{
		{
			Date:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			Vaccine: &model.Vaccine{Name: "総合ワクチン", Species: &both},
		},
		{
			Date:    time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			Vaccine: &model.Vaccine{Name: "狂犬病ワクチン", Species: &dog},
		},
	}
	tagSet := buildLatestVaccineTagSet(vaccinations)
	assert.Contains(t, tagSet, "vaccine_dog_2024-04-01") // both で 3/1, dog で 4/1 → 4/1 が勝つ
	assert.Contains(t, tagSet, "vaccine_cat_2024-03-01")
	assert.Contains(t, tagSet, "vaccine_rabies_2024-04-01")
}

// ---- buildLatestCheckupTagSet (ISSUE-004) ----

func TestBuildLatestCheckupTagSet_EmptyReturnsEmpty(t *testing.T) {
	tagSet := buildLatestCheckupTagSet(nil)
	assert.Empty(t, tagSet)
}

func TestBuildLatestCheckupTagSet_LatestPerType(t *testing.T) {
	checkups := []model.Checkup{
		{
			CheckupTypeID: 1,
			Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			CheckupTypeID: 1,
			Date:          time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC), // 最新
		},
		{
			CheckupTypeID: 2,
			Date:          time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
	}
	tagSet := buildLatestCheckupTagSet(checkups)
	assert.Contains(t, tagSet, "checkup_done_1_2024-08", "type=1 は最新の 2024-08")
	assert.NotContains(t, tagSet, "checkup_done_1_2024-01")
	assert.Contains(t, tagSet, "checkup_done_2_2023-12")
}

func TestBuildLatestCheckupTagSet_NextCheckupLatest(t *testing.T) {
	near := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	far := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	checkups := []model.Checkup{
		{
			CheckupTypeID: 1,
			Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			NextDate:      &near,
		},
		{
			CheckupTypeID: 2,
			Date:          time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			NextDate:      &far,
		},
	}
	tagSet := buildLatestCheckupTagSet(checkups)
	assert.Contains(t, tagSet, "next_checkup_2026-03-01", "next_checkup は最遠の next_date を採用")
	assert.NotContains(t, tagSet, "next_checkup_2025-06-01")
}

func TestBuildLatestCheckupTagSet_SkipsNilNextDate(t *testing.T) {
	checkups := []model.Checkup{
		{
			CheckupTypeID: 1,
			Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			NextDate:      nil,
		},
	}
	tagSet := buildLatestCheckupTagSet(checkups)
	assert.Contains(t, tagSet, "checkup_done_1_2024-01")
	for k := range tagSet {
		assert.NotContains(t, k, "next_checkup_", "next_date=nil なら next_checkup タグは生成しない")
	}
}

// ---- conditionTagMap ----

func TestConditionTagMap(t *testing.T) {
	mappings := []*model.LstepConditionTagMapping{
		{ConditionCode: "ckd", TagName: "chronic_ckd"},
		{ConditionCode: "heart", TagName: "chronic_heart"},
		{ConditionCode: "skin", TagName: "chronic_skin"},
		{ConditionCode: "diabetes", TagName: "chronic_diabetes"},
		{ConditionCode: "liver", TagName: "chronic_liver"},
		{ConditionCode: "thyroid", TagName: "chronic_thyroid"},
		{ConditionCode: "other", TagName: "chronic_other"},
	}
	m := conditionTagMapFromMappings(mappings)
	cases := map[string]string{
		"ckd":      "chronic_ckd",
		"heart":    "chronic_heart",
		"skin":     "chronic_skin",
		"diabetes": "chronic_diabetes",
		"liver":    "chronic_liver",
		"thyroid":  "chronic_thyroid",
		"other":    "chronic_other",
	}
	for code, wantTag := range cases {
		got, ok := m[code]
		assert.True(t, ok, "conditionTagMap missing key: %s", code)
		assert.Equal(t, wantTag, got)
	}
	_, hasUnknown := m["unknown"]
	assert.False(t, hasUnknown)
}

// ---- FEAT-375: EXCL_カルテ連携エラー 自動タグ ----

// mockLstepAPIClient は lstep.Client の最小モック実装。
type mockLstepAPIClient struct {
	addTagFn    func(ctx context.Context, lineUserID, tagName string) error
	removeTagFn func(ctx context.Context, lineUserID, tagName string) error
}

func (m *mockLstepAPIClient) AddTag(ctx context.Context, lineUserID, tagName string) error {
	if m.addTagFn != nil {
		return m.addTagFn(ctx, lineUserID, tagName)
	}
	return nil
}
func (m *mockLstepAPIClient) RemoveTag(ctx context.Context, lineUserID, tagName string) error {
	if m.removeTagFn != nil {
		return m.removeTagFn(ctx, lineUserID, tagName)
	}
	return nil
}
func (m *mockLstepAPIClient) GetUserTags(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockLstepAPIClient) AddTagBulk(_ context.Context, _ []string, _ string) error { return nil }
func (m *mockLstepAPIClient) GetUser(_ context.Context, _ string) (*lstep.UserInfo, error) {
	return nil, nil
}
func (m *mockLstepAPIClient) SetProperty(_ context.Context, _, _, _ string) error { return nil }

// mockErrorCounterRepo は LstepSyncErrorCounterRepository の最小モック実装。
type mockErrorCounterRepo struct {
	incrementFn func(ctx context.Context, clinicID, ownerID uint64) (int, error)
	resetFn     func(ctx context.Context, clinicID, ownerID uint64) error
	findFn      func(ctx context.Context, clinicID, ownerID uint64) (*model.LstepSyncErrorCounter, error)
}

func (m *mockErrorCounterRepo) IncrementFailure(ctx context.Context, clinicID, ownerID uint64) (int, error) {
	if m.incrementFn != nil {
		return m.incrementFn(ctx, clinicID, ownerID)
	}
	return 0, nil
}
func (m *mockErrorCounterRepo) ResetFailure(ctx context.Context, clinicID, ownerID uint64) error {
	if m.resetFn != nil {
		return m.resetFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockErrorCounterRepo) FindByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.LstepSyncErrorCounter, error) {
	if m.findFn != nil {
		return m.findFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func TestNotifyAPIFailure(t *testing.T) {
	t.Run("below threshold: no EXCL tag added", func(t *testing.T) {
		addTagCalled := false
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalled = true
				return nil
			},
		}
		repo := &mockErrorCounterRepo{
			incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
				return lstepSyncErrorThreshold - 1, nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
		svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
		assert.False(t, addTagCalled)
	})

	t.Run("at threshold: EXCL tag added and cached", func(t *testing.T) {
		var addedTag string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTag = tagName
				return nil
			},
		}
		var upsertedTag string
		tagCache := &mockLstepTagCacheRepository{
			upsertTagFn: func(_ context.Context, _, _ uint64, tagName, _, _ string) error {
				upsertedTag = tagName
				return nil
			},
		}
		repo := &mockErrorCounterRepo{
			incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
				return lstepSyncErrorThreshold, nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
		svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
		assert.Equal(t, lstepErrorTag, addedTag)
		assert.Equal(t, lstepErrorTag, upsertedTag)
	})

	t.Run("nil errorCounterRepo is noop (sync disabled path)", func(t *testing.T) {
		addTagCalled := false
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error {
				addTagCalled = true
				return nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: nil, tagCacheRepo: &mockLstepTagCacheRepository{}}
		svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
		assert.False(t, addTagCalled)
	})

	t.Run("increment error is logged and does not propagate", func(t *testing.T) {
		client := &mockLstepAPIClient{}
		repo := &mockErrorCounterRepo{
			incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
				return 0, errors.New("db error")
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
		// must not panic; best-effort — no return value to check
		svc.notifyAPIFailure(context.Background(), client, 1, 2, "u1")
	})
}

func TestRemoveStaleTagsByPrefixesRecordsRemoveFailure(t *testing.T) {
	client := &mockLstepAPIClient{
		removeTagFn: func(_ context.Context, _, _ string) error {
			return errors.New("lstep remove failed")
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "vaccine_dog_2026-05-01"}}, nil
		},
	}
	incrementCalled := false
	repo := &mockErrorCounterRepo{
		incrementFn: func(_ context.Context, _, _ uint64) (int, error) {
			incrementCalled = true
			return lstepSyncErrorThreshold - 1, nil
		},
	}
	svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}

	apiFailed, err := svc.removeStaleTagsByPrefixes(
		context.Background(),
		client,
		1,
		2,
		"u1",
		[]string{"vaccine_dog_"},
		map[string]struct{}{},
	)

	assert.NoError(t, err)
	assert.True(t, apiFailed)
	assert.True(t, incrementCalled)
}

func TestNotifyAPISuccess(t *testing.T) {
	t.Run("counter zero: noop, no RemoveTag call", func(t *testing.T) {
		removeTagCalled := false
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error {
				removeTagCalled = true
				return nil
			},
		}
		repo := &mockErrorCounterRepo{
			findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				return &model.LstepSyncErrorCounter{FailureCount: 0}, nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
		svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
		assert.False(t, removeTagCalled)
	})

	t.Run("positive counter: reset and remove EXCL tag", func(t *testing.T) {
		var removedTag string
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTag = tagName
				return nil
			},
		}
		resetCalled := false
		var deletedTag string
		tagCache := &mockLstepTagCacheRepository{
			deleteTagFn: func(_ context.Context, _, _ uint64, tagName string) error {
				deletedTag = tagName
				return nil
			},
		}
		repo := &mockErrorCounterRepo{
			findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				return &model.LstepSyncErrorCounter{FailureCount: lstepSyncErrorThreshold}, nil
			},
			resetFn: func(_ context.Context, _, _ uint64) error {
				resetCalled = true
				return nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: tagCache}
		svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
		assert.True(t, resetCalled)
		assert.Equal(t, lstepErrorTag, removedTag)
		assert.Equal(t, lstepErrorTag, deletedTag)
	})

	t.Run("remove EXCL failure keeps counter for retry", func(t *testing.T) {
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error {
				return errors.New("remove failed")
			},
		}
		resetCalled := false
		repo := &mockErrorCounterRepo{
			findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				return &model.LstepSyncErrorCounter{FailureCount: lstepSyncErrorThreshold}, nil
			},
			resetFn: func(_ context.Context, _, _ uint64) error {
				resetCalled = true
				return nil
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
		svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
		assert.False(t, resetCalled)
	})

	t.Run("not found counter: noop (owner never failed)", func(t *testing.T) {
		removeTagCalled := false
		client := &mockLstepAPIClient{
			removeTagFn: func(_ context.Context, _, _ string) error {
				removeTagCalled = true
				return nil
			},
		}
		repo := &mockErrorCounterRepo{
			findFn: func(_ context.Context, _, _ uint64) (*model.LstepSyncErrorCounter, error) {
				return nil, apperrors.WrapNotFound("lstep_sync_error_counter", "owner=2")
			},
		}
		svc := &lstepTagSyncService{errorCounterRepo: repo, tagCacheRepo: &mockLstepTagCacheRepository{}}
		svc.notifyAPISuccess(context.Background(), client, 1, 2, "u1")
		assert.False(t, removeTagCalled)
	})
}

// ---- CalculateCPMStageV2 (Q19 確定 2026-05-08) ----

func TestCalculateCPMStageV2(t *testing.T) {
	cases := []struct {
		name string
		in   CPMStageV2Input
		want CPMStageV2
	}{
		// 出会い: 累計 0 回
		{
			name: "Encounter: 0 visits",
			in:   CPMStageV2Input{TotalVisitCount: 0},
			want: CPMStageV2Encounter,
		},
		// 出会い: 累計 1 回
		{
			name: "Encounter: 1 visit",
			in:   CPMStageV2Input{TotalVisitCount: 1},
			want: CPMStageV2Encounter,
		},
		// これから: 累計 2 回（下限境界）
		{
			name: "Coming: 2 visits (lower boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 2},
			want: CPMStageV2Coming,
		},
		// これから: 累計 3 回（上限境界）
		{
			name: "Coming: 3 visits (upper boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 3},
			want: CPMStageV2Coming,
		},
		// いいかんじ: 累計 4 回（下限境界）
		{
			name: "Good: 4 visits (lower boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 4},
			want: CPMStageV2Good,
		},
		// いいかんじ: 累計 7 回（上限境界）
		{
			name: "Good: 7 visits (upper boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 7},
			want: CPMStageV2Good,
		},
		// ファミリー: 累計 8 回（下限境界）
		{
			name: "Family: 8 visits (lower boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 8},
			want: CPMStageV2Family,
		},
		// ファミリー: 累計 12 回（上限境界）
		{
			name: "Family: 12 visits (upper boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 12},
			want: CPMStageV2Family,
		},
		// ノア: 累計 13 回（下限境界）
		{
			name: "Noah: 13 visits (lower boundary)",
			in:   CPMStageV2Input{TotalVisitCount: 13},
			want: CPMStageV2Noah,
		},
		// ノア: 累計大量来院
		{
			name: "Noah: 100 visits",
			in:   CPMStageV2Input{TotalVisitCount: 100},
			want: CPMStageV2Noah,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateCPMStageV2(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestCalculateCPMStageV2_CustomThresholds: クリニック単位閾値を反映する
func TestCalculateCPMStageV2_CustomThresholds(t *testing.T) {
	custom := model.CPMV2Thresholds{Coming: 5, Good: 10, Family: 20, Noah: 30}
	cases := []struct {
		name  string
		count int64
		want  CPMStageV2
	}{
		{"Encounter: 4 visits (below custom Coming=5)", 4, CPMStageV2Encounter},
		{"Coming: 5 visits (custom Coming=5 lower boundary)", 5, CPMStageV2Coming},
		{"Coming: 9 visits (below custom Good=10)", 9, CPMStageV2Coming},
		{"Good: 10 visits (custom Good=10 lower boundary)", 10, CPMStageV2Good},
		{"Good: 19 visits (below custom Family=20)", 19, CPMStageV2Good},
		{"Family: 20 visits (custom Family=20 lower boundary)", 20, CPMStageV2Family},
		{"Family: 29 visits (below custom Noah=30)", 29, CPMStageV2Family},
		{"Noah: 30 visits (custom Noah=30 lower boundary)", 30, CPMStageV2Noah},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateCPMStageV2(CPMStageV2Input{
				TotalVisitCount: tc.count,
				CPMV2Thresholds: custom,
			})
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---- isVisitDormantTag ----

func TestIsVisitDormantTag(t *testing.T) {
	assert.True(t, isVisitDormantTag("VISIT_120日超"))
	assert.True(t, isVisitDormantTag("VISIT_180日超"))
	assert.True(t, isVisitDormantTag("VISIT_220日超"))
	assert.True(t, isVisitDormantTag("VISIT_240日超"))
	assert.False(t, isVisitDormantTag("dormant_180d"))
	assert.False(t, isVisitDormantTag("VISIT_120"))
	assert.False(t, isVisitDormantTag(""))
}

// ---- visitDormantTagsForDays (pure function) ----

func TestVisitDormantTagsForDays(t *testing.T) {
	assert.Equal(t, []string(nil), visitDormantTagsForDays(119))
	assert.Equal(t, []string(nil), visitDormantTagsForDays(120))
	assert.Equal(t, []string{"VISIT_120日超"}, visitDormantTagsForDays(121))
	assert.Equal(t, []string{"VISIT_120日超"}, visitDormantTagsForDays(180))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超"}, visitDormantTagsForDays(181))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超"}, visitDormantTagsForDays(220))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超", "VISIT_220日超"}, visitDormantTagsForDays(221))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超", "VISIT_220日超"}, visitDormantTagsForDays(240))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超", "VISIT_220日超", "VISIT_240日超"}, visitDormantTagsForDays(241))
	assert.Equal(t, []string{"VISIT_120日超", "VISIT_180日超", "VISIT_220日超", "VISIT_240日超"}, visitDormantTagsForDays(300))
}

// ---- hasVaccineDeadlineSoon (pure function) ----

func TestHasVaccineDeadlineSoon(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	days := model.DefaultVaccineDeadlineDays

	// nil / empty slice → false
	assert.False(t, hasVaccineDeadlineSoon(nil, now))
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{}, now))

	// NextDate nil → skipped
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: nil}}, now))

	// NextDate が now ちょうど → true (境界値)
	exact := now
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &exact}}, now))

	// NextDate が deadline ちょうど (now + DefaultVaccineDeadlineDays) → true
	deadline := now.AddDate(0, 0, days)
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &deadline}}, now))

	// NextDate が deadline より 1 日先 → false
	beyond := now.AddDate(0, 0, days+1)
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &beyond}}, now))

	// NextDate が now より前 (過去) → false
	past := now.AddDate(0, 0, -1)
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &past}}, now))

	// 複数レコードのうち 1 件が範囲内 → true
	inner := now.AddDate(0, 0, 30)
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &beyond}, {NextDate: &inner}}, now))
}

// ---- tagCodeRepo == nil noop テスト ----
// 判定コードは lstep_tag_code_mappings テーブル（DB）から取得する。
// tagCodeRepo が nil の場合、各 Sync 関数は noop（nil 返却）になる。

func TestSyncHealthcheckTagsNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	).(*lstepTagSyncService)
	assert.NoError(t, svc.SyncHealthcheckTagsWithMappings(context.Background(), 1, 2, nil, nil))
}

func TestSyncAnnual4CheckupTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	).(*lstepTagSyncService)
	assert.NoError(t, svc.SyncAnnual4CheckupTagWithMappings(context.Background(), 1, 2, nil, nil))
}

func TestSyncFilariaTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	).(*lstepTagSyncService)
	assert.NoError(t, svc.SyncFilariaTagWithMappings(context.Background(), 1, 2, nil, nil))
}

func TestSyncFleaTickTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	).(*lstepTagSyncService)
	assert.NoError(t, svc.SyncFleaTickTagWithMappings(context.Background(), 1, 2, nil, nil))
}

func TestSyncFoodPurchaseTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	).(*lstepTagSyncService)
	assert.NoError(t, svc.SyncFoodPurchaseTagWithMappings(context.Background(), 1, 2, nil, nil))
}

func TestSyncHealthPreventionTagsForClinicDisabledSync(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	count, errs := svc.SyncHealthPreventionTagsForClinic(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.Empty(t, errs)
}

func TestSyncExclusionTagsCaution(t *testing.T) {
	lineUID := "U_test"

	buildOwnerRepo := func(caution bool) *mockOwnerRepository {
		return &mockOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{
					ID:              10,
					ClinicID:        1,
					LineUserID:      &lineUID,
					DeliveryCaution: caution,
				}, nil
			},
		}
	}
	buildPetRepo := func() *mockPetRepository {
		return &mockPetRepository{
			countByOwnerFn:       func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
			countLivingByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) { return 1, nil },
		}
	}

	t.Run("delivery_caution=true adds EXCL_配信注意 tag", func(t *testing.T) {
		var addedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, tagName string) error {
				addedTags = append(addedTags, tagName)
				return nil
			},
			removeTagFn: func(_ context.Context, _, _ string) error { return nil },
		}
		svc := &lstepTagSyncService{
			settingsSvc:      &mockLstepSettingsService{},
			ownerRepo:        buildOwnerRepo(true),
			petRepo:          buildPetRepo(),
			tagCacheRepo:     &mockLstepTagCacheRepository{},
			errorCounterRepo: nil, // nil → notifyAPISuccess は noop
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Contains(t, addedTags, exclTagDeliveryCaution)
	})

	t.Run("delivery_caution=false removes EXCL_配信注意 tag", func(t *testing.T) {
		var removedTags []string
		client := &mockLstepAPIClient{
			addTagFn: func(_ context.Context, _, _ string) error { return nil },
			removeTagFn: func(_ context.Context, _, tagName string) error {
				removedTags = append(removedTags, tagName)
				return nil
			},
		}
		svc := &lstepTagSyncService{
			settingsSvc:      &mockLstepSettingsService{},
			ownerRepo:        buildOwnerRepo(false),
			petRepo:          buildPetRepo(),
			tagCacheRepo:     &mockLstepTagCacheRepository{},
			errorCounterRepo: nil, // nil → notifyAPISuccess は noop
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				return client, nil
			},
		}
		err := svc.SyncExclusionTags(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Contains(t, removedTags, exclTagDeliveryCaution)
	})
}

// ---- buildClient ----

func TestLstepTagSyncService_BuildClient(t *testing.T) {
	tests := []struct {
		name          string
		isSyncEnabled bool
		syncErr       error
		apiKey        string
		credErr       error
		wantErr       bool
		wantNilClient bool
	}{
		{
			name:          "sync disabled -> nil client, nil error",
			isSyncEnabled: false,
			wantNilClient: true,
		},
		{
			name:          "IsSyncEnabled error -> wrapped error",
			isSyncEnabled: false,
			syncErr:       errors.New("db error"),
			wantErr:       true,
			wantNilClient: true,
		},
		{
			name:          "GetRawCredentials error -> wrapped error",
			isSyncEnabled: true,
			credErr:       errors.New("db error"),
			wantErr:       true,
			wantNilClient: true,
		},
		{
			name:          "empty api key -> nil client, nil error",
			isSyncEnabled: true,
			apiKey:        "",
			wantNilClient: true,
		},
		{
			name:          "valid credentials -> non-nil client",
			isSyncEnabled: true,
			apiKey:        "secret-key",
			wantNilClient: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingsSvc := &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
					return tt.isSyncEnabled, tt.syncErr
				},
				getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
					return tt.apiKey, "https://example.com", "", tt.credErr
				},
			}
			svc := &lstepTagSyncService{settingsSvc: settingsSvc}

			client, err := svc.buildClient(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantNilClient {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}

	t.Run("buildClientFn override takes precedence over real settingsSvc lookups", func(t *testing.T) {
		called := false
		fakeClient := &mockLstepAPIClient{}
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{
				isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
					t.Fatal("real settingsSvc must not be consulted when buildClientFn is set")
					return false, nil
				},
			},
			buildClientFn: func(_ context.Context, _ uint64) (lstep.Client, error) {
				called = true
				return fakeClient, nil
			},
		}
		client, err := svc.buildClient(context.Background(), 1)
		assert.NoError(t, err)
		assert.Same(t, fakeClient, client)
		assert.True(t, called)
	})
}

// ---- shouldSkipSync ----

func TestLstepTagSyncService_ShouldSkipSync(t *testing.T) {
	t.Run("nil settingsSvc always skips", func(t *testing.T) {
		svc := &lstepTagSyncService{}
		skip, err := svc.shouldSkipSync(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, skip)
	})

	t.Run("IsSyncEnabled error is wrapped and skip=false", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") }},
		}
		skip, err := svc.shouldSkipSync(context.Background(), 1)
		assert.Error(t, err)
		assert.False(t, skip)
	})

	t.Run("sync disabled -> skip=true, nil error", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		}
		skip, err := svc.shouldSkipSync(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, skip)
	})

	t.Run("sync enabled -> skip=false, nil error", func(t *testing.T) {
		svc := &lstepTagSyncService{
			settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		}
		skip, err := svc.shouldSkipSync(context.Background(), 1)
		assert.NoError(t, err)
		assert.False(t, skip)
	})
}
