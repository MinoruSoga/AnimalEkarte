package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
		{name: "SyncDormantTags", run: func() error { return svc.SyncDormantTags(ctx, 1, 2, 180) }},
		{name: "ResyncOwnerVaccineTags", run: func() error { return svc.ResyncOwnerVaccineTags(ctx, 1, 2) }},
		{name: "ResyncOwnerCheckupTags", run: func() error { return svc.ResyncOwnerCheckupTags(ctx, 1, 2) }},
		{name: "SyncCPMStageTagV2", run: func() error { return svc.SyncCPMStageTagV2(ctx, 1, 2) }},
		{name: "SyncVisitDormantTags", run: func() error { return svc.SyncVisitDormantTags(ctx, 1, 2, 120) }},
		{name: "SyncPetSpeciesTags", run: func() error { return svc.SyncPetSpeciesTags(ctx, 1, 2) }},
		{name: "SyncSeniorTag", run: func() error { return svc.SyncSeniorTag(ctx, 1, 2) }},
		{name: "SyncExclusionTags", run: func() error { return svc.SyncExclusionTags(ctx, 1, 2) }},
		{name: "SyncHealthcheckTags", run: func() error { return svc.SyncHealthcheckTags(ctx, 1, 2) }},
		{name: "SyncAnnual4CheckupTag", run: func() error { return svc.SyncAnnual4CheckupTag(ctx, 1, 2) }},
		{name: "SyncVaccineDeadlineTag", run: func() error { return svc.SyncVaccineDeadlineTag(ctx, 1, 2) }},
		{name: "SyncFilariaTag", run: func() error { return svc.SyncFilariaTag(ctx, 1, 2) }},
		{name: "SyncFleaTickTag", run: func() error { return svc.SyncFleaTickTag(ctx, 1, 2) }},
		{name: "SyncFoodPurchaseTag", run: func() error { return svc.SyncFoodPurchaseTag(ctx, 1, 2) }},
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
	summary := &repository.OwnerVisitSummary{
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
	summary := &repository.OwnerVisitSummary{AnnualCount: 1}
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

	apiFailed := svc.removeStaleTagsByPrefixes(
		context.Background(),
		client,
		1,
		2,
		"u1",
		[]string{"vaccine_dog_"},
		map[string]struct{}{},
	)

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

// ---- petSpeciesFlags (pure function) ----

func TestPetSpeciesFlags(t *testing.T) {
	dog := &model.AnimalSpecies{Name: "犬"}
	cat := &model.AnimalSpecies{Name: "猫"}
	bird := &model.AnimalSpecies{Name: "鳥"}

	hasDog, hasCat := petSpeciesFlags(nil)
	assert.False(t, hasDog)
	assert.False(t, hasCat)

	hasDog, hasCat = petSpeciesFlags([]model.Pet{})
	assert.False(t, hasDog)
	assert.False(t, hasCat)

	hasDog, hasCat = petSpeciesFlags([]model.Pet{{AnimalSpecies: dog}})
	assert.True(t, hasDog)
	assert.False(t, hasCat)

	hasDog, hasCat = petSpeciesFlags([]model.Pet{{AnimalSpecies: cat}})
	assert.False(t, hasDog)
	assert.True(t, hasCat)

	hasDog, hasCat = petSpeciesFlags([]model.Pet{{AnimalSpecies: dog}, {AnimalSpecies: cat}})
	assert.True(t, hasDog)
	assert.True(t, hasCat)

	hasDog, hasCat = petSpeciesFlags([]model.Pet{{AnimalSpecies: bird}})
	assert.False(t, hasDog)
	assert.False(t, hasCat)

	// nil AnimalSpecies is skipped
	hasDog, hasCat = petSpeciesFlags([]model.Pet{{AnimalSpecies: nil}})
	assert.False(t, hasDog)
	assert.False(t, hasCat)
}

// ---- hasSeniorPet (pure function) ----

func TestHasSeniorPet(t *testing.T) {
	dog := &model.AnimalSpecies{Name: "犬"}
	cat := &model.AnimalSpecies{Name: "猫"}
	bird := &model.AnimalSpecies{Name: "鳥"}
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	bd7 := now.AddDate(-7, 0, -1) // 7歳超: senior
	bd7e := now.AddDate(-7, 0, 0) // 7歳ちょうど: senior
	bd6 := now.AddDate(-6, 0, -1) // 6歳: not senior

	assert.False(t, hasSeniorPet(nil, now))
	assert.False(t, hasSeniorPet([]model.Pet{}, now))

	// 7歳超 犬 → senior
	assert.True(t, hasSeniorPet([]model.Pet{{AnimalSpecies: dog, BirthDate: &bd7}}, now))

	// 7歳ちょうど 犬 → senior
	assert.True(t, hasSeniorPet([]model.Pet{{AnimalSpecies: dog, BirthDate: &bd7e}}, now))

	// 6歳 犬 → not senior
	assert.False(t, hasSeniorPet([]model.Pet{{AnimalSpecies: dog, BirthDate: &bd6}}, now))

	// 7歳超 猫 → senior
	assert.True(t, hasSeniorPet([]model.Pet{{AnimalSpecies: cat, BirthDate: &bd7}}, now))

	// 7歳超 鳥 → not senior (犬猫以外は対象外)
	assert.False(t, hasSeniorPet([]model.Pet{{AnimalSpecies: bird, BirthDate: &bd7}}, now))

	// BirthDate nil → not senior
	assert.False(t, hasSeniorPet([]model.Pet{{AnimalSpecies: dog, BirthDate: nil}}, now))

	// AnimalSpecies nil → not senior
	assert.False(t, hasSeniorPet([]model.Pet{{AnimalSpecies: nil, BirthDate: &bd7}}, now))
}

// ---- hasVaccineDeadlineSoon (pure function) ----

func TestHasVaccineDeadlineSoon(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	days := 60

	// nil / empty slice → false
	assert.False(t, hasVaccineDeadlineSoon(nil, now, days))
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{}, now, days))

	// NextDate nil → skipped
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: nil}}, now, days))

	// NextDate が now ちょうど → true (境界値)
	exact := now
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &exact}}, now, days))

	// NextDate が deadline ちょうど (now + 60 days) → true
	deadline := now.AddDate(0, 0, days)
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &deadline}}, now, days))

	// NextDate が deadline より 1 日先 → false
	beyond := now.AddDate(0, 0, days+1)
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &beyond}}, now, days))

	// NextDate が now より前 (過去) → false
	past := now.AddDate(0, 0, -1)
	assert.False(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &past}}, now, days))

	// 複数レコードのうち 1 件が範囲内 → true
	inner := now.AddDate(0, 0, 30)
	assert.True(t, hasVaccineDeadlineSoon([]model.Vaccination{{NextDate: &beyond}, {NextDate: &inner}}, now, days))
}

// ---- tagCodeRepo == nil noop テスト ----
// 判定コードは lstep_tag_code_mappings テーブル（DB）から取得する。
// tagCodeRepo が nil の場合、各 Sync 関数は noop（nil 返却）になる。

func TestSyncHealthcheckTagsNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, svc.SyncHealthcheckTags(context.Background(), 1, 2))
}

func TestSyncAnnual4CheckupTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, svc.SyncAnnual4CheckupTag(context.Background(), 1, 2))
}

func TestSyncFilariaTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, svc.SyncFilariaTag(context.Background(), 1, 2))
}

func TestSyncFleaTickTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, svc.SyncFleaTickTag(context.Background(), 1, 2))
}

func TestSyncFoodPurchaseTagNoopWhenTagCodeRepoNil(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, svc.SyncFoodPurchaseTag(context.Background(), 1, 2))
}

func TestSyncHealthPreventionTagsForClinicDisabledSync(t *testing.T) {
	svc := NewLstepTagSyncService(
		&mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil }},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
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
		assert.Contains(t, addedTags, "EXCL_配信注意")
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
		assert.Contains(t, removedTags, "EXCL_配信注意")
	})
}
