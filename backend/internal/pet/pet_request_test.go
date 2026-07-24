package pet

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestNewListPetQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   listPetQuery
	}{
		{
			name:   "parses owner_id and search",
			values: url.Values{"owner_id": {"10"}, "search": {"momo"}},
			want:   listPetQuery{OwnerID: "10", Search: "momo"},
		},
		{
			name:   "parses species and include_deceased",
			values: url.Values{"species": {"3"}, "include_deceased": {"true"}},
			want:   listPetQuery{Species: "3", IncludeDeceased: "true"},
		},
		{
			name:   "empty values yield zero query",
			values: url.Values{},
			want:   listPetQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newListPetQuery(tt.values))
		})
	}
}

func TestListPetQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listPetQuery{
		OwnerID: "10",
		Search:  "momo",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	require.NotNil(t, filters.OwnerID)
	assert.Equal(t, uint64(10), *filters.OwnerID)
	assert.Equal(t, "momo", filters.Search)
	assert.Nil(t, filters.AnimalSpeciesID)
	assert.False(t, filters.IncludeDeceased, "include_deceased 未指定は既定 false（生存のみ）")
}

func TestListPetQuery_ToServiceFilters_SpeciesAndIncludeDeceased(t *testing.T) {
	filters, err := (&listPetQuery{
		Species:         "3",
		IncludeDeceased: "true",
	}).toServiceFilters()
	require.NoError(t, err)

	require.NotNil(t, filters.AnimalSpeciesID)
	assert.Equal(t, uint64(3), *filters.AnimalSpeciesID)
	assert.True(t, filters.IncludeDeceased)
}

func TestListPetQuery_ToServiceFilters_InvalidOwnerID(t *testing.T) {
	filters, err := (&listPetQuery{OwnerID: "abc"}).toServiceFilters()
	require.Error(t, err)
	assert.Equal(t, listPetFilters{}, filters)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestListPetQuery_ToServiceFilters_InvalidSpecies(t *testing.T) {
	filters, err := (&listPetQuery{Species: "abc"}).toServiceFilters()
	require.Error(t, err)
	assert.Equal(t, listPetFilters{}, filters)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestListPetQuery_ToServiceFilters_InvalidIncludeDeceased(t *testing.T) {
	filters, err := (&listPetQuery{IncludeDeceased: "not-a-bool"}).toServiceFilters()
	require.Error(t, err)
	assert.Equal(t, listPetFilters{}, filters)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestCreatePetRequest_ToServiceInput(t *testing.T) {
	birthDate := &jsonDate{Time: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)}
	neuteredDate := &jsonDate{Time: time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC)}
	weight := 4.2
	insuranceID := uint64(9)

	input := (&createPetRequest{
		OwnerID:         5,
		AnimalSpeciesID: 1,
		Name:            "ポチ",
		NameKana:        "ポチ",
		Gender:          "male",
		Status:          "alive",
		BirthDate:       birthDate,
		Breed:           "柴犬",
		Color:           "茶",
		Weight:          &weight,
		NeuteredDate:    neuteredDate,
		AcquisitionType: "purchase",
		DangerLevel:     "low",
		Food:            "ドライ",
		Environment:     "室内",
		Phone:           "090-1234-5678",
		InsuranceID:     &insuranceID,
		Remarks:         "備考",
	}).toServiceInput()

	assert.Equal(t, uint64(5), input.OwnerID)
	assert.Equal(t, uint64(1), input.AnimalSpeciesID)
	assert.Equal(t, "ポチ", input.Name)
	assert.Equal(t, "ポチ", input.PetNameKana)
	assert.Equal(t, "male", input.Gender)
	assert.Equal(t, "alive", input.Status)
	require.NotNil(t, input.BirthDate)
	assert.Equal(t, birthDate.Time, *input.BirthDate)
	require.NotNil(t, input.NeuteredDate)
	assert.Equal(t, neuteredDate.Time, *input.NeuteredDate)
	assert.Same(t, &weight, input.Weight)
	assert.Same(t, &insuranceID, input.InsuranceID)
	assert.Equal(t, "備考", input.Remarks)
}

func TestUpdatePetRequest_ToServiceInput(t *testing.T) {
	lastVisit := &jsonDate{Time: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC)}
	ownerID := uint64(6)
	animalSpeciesID := uint64(2)
	name := "タマ"
	insuranceID := uint64(12)
	insuranceIDField := &insuranceID

	input := (&updatePetRequest{
		OwnerID:         &ownerID,
		AnimalSpeciesID: &animalSpeciesID,
		Name:            &name,
		LastVisit:       lastVisit,
		InsuranceID:     &insuranceIDField,
	}).toServiceInput()

	assert.Same(t, &ownerID, input.OwnerID)
	assert.Same(t, &animalSpeciesID, input.AnimalSpeciesID)
	assert.Same(t, &name, input.Name)
	require.NotNil(t, input.LastVisit)
	assert.Equal(t, lastVisit.Time, *input.LastVisit)
	require.NotNil(t, input.InsuranceID)
	require.NotNil(t, *input.InsuranceID)
	assert.Equal(t, insuranceID, **input.InsuranceID)
}

func TestUpdatePetRequest_ToServiceInput_InsuranceIDClear(t *testing.T) {
	var insuranceID *uint64

	input := (&updatePetRequest{
		InsuranceID: &insuranceID,
	}).toServiceInput()

	require.NotNil(t, input.InsuranceID)
	assert.Nil(t, *input.InsuranceID)
}
