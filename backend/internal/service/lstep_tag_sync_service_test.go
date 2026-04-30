package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- isPetBasicInfoTag ----

func TestIsPetBasicInfoTag(t *testing.T) {
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
			assert.Equal(t, tc.want, isPetBasicInfoTag(tc.tag))
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

// ---- isReservationRelatedTag ----

func TestIsReservationRelatedTag(t *testing.T) {
	cases := []struct {
		tag  string
		want bool
	}{
		{"reserved_2024-05-01", true},
		{"reserved_", true},
		{"canceled_visit", true},
		{"no_show_3", true},
		{"no_show_", true},
		{"canceled", false},
		{"cpm_core", false},
		{"dormant_180d", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			assert.Equal(t, tc.want, isReservationRelatedTag(tc.tag))
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
		got, ok := conditionTagMap[code]
		assert.True(t, ok, "conditionTagMap missing key: %s", code)
		assert.Equal(t, wantTag, got)
	}
	// unknown code must not be present
	_, hasUnknown := conditionTagMap["unknown"]
	assert.False(t, hasUnknown)
}
