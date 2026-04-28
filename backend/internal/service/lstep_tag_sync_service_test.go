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
