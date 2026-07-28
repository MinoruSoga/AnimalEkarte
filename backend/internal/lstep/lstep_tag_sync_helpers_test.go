package lstep

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestLtvBracketTag(t *testing.T) {
	assert.Equal(t, "ltv_amount_8", ltvBracketTag(80_000))
	assert.Equal(t, "ltv_amount_8", ltvBracketTag(200_000))
	assert.Equal(t, "ltv_amount_5", ltvBracketTag(50_000))
	assert.Equal(t, "ltv_amount_5", ltvBracketTag(79_999))
	assert.Equal(t, "ltv_amount_2", ltvBracketTag(20_000))
	assert.Equal(t, "ltv_amount_2", ltvBracketTag(49_999))
	assert.Equal(t, "ltv_amount_0", ltvBracketTag(0))
	assert.Equal(t, "ltv_amount_0", ltvBracketTag(19_999))
}

func TestVisitCountAnnualTag(t *testing.T) {
	assert.Equal(t, "visit_count_annual_10", visitCountAnnualTag(10))
	assert.Equal(t, "visit_count_annual_10", visitCountAnnualTag(15))
	assert.Equal(t, "visit_count_annual_5", visitCountAnnualTag(5))
	assert.Equal(t, "visit_count_annual_5", visitCountAnnualTag(9))
	assert.Equal(t, "visit_count_annual_3", visitCountAnnualTag(3))
	assert.Equal(t, "visit_count_annual_3", visitCountAnnualTag(4))
	assert.Equal(t, "visit_count_annual_2", visitCountAnnualTag(2))
	assert.Equal(t, "visit_count_annual_1", visitCountAnnualTag(0))
	assert.Equal(t, "visit_count_annual_1", visitCountAnnualTag(1))
}

func TestVaccineTagNames(t *testing.T) {
	dogSpecies := model.VaccineSpeciesDog
	catSpecies := model.VaccineSpeciesCat
	bothSpecies := model.VaccineSpeciesBoth
	baseDate := time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		vaccine *model.Vaccination
		want    []string
	}{
		{
			name:    "dog species",
			vaccine: &model.Vaccination{Date: baseDate, Vaccine: &model.Vaccine{Name: "混合ワクチン", Species: &dogSpecies}},
			want:    []string{"vaccine_dog_2026-04-25"},
		},
		{
			name:    "cat species",
			vaccine: &model.Vaccination{Date: baseDate, Vaccine: &model.Vaccine{Name: "猫ワクチン", Species: &catSpecies}},
			want:    []string{"vaccine_cat_2026-04-25"},
		},
		{
			name:    "both species",
			vaccine: &model.Vaccination{Date: baseDate, Vaccine: &model.Vaccine{Name: "共通ワクチン", Species: &bothSpecies}},
			want:    []string{"vaccine_dog_2026-04-25", "vaccine_cat_2026-04-25"},
		},
		{
			name:    "rabies",
			vaccine: &model.Vaccination{Date: baseDate, Vaccine: &model.Vaccine{Name: "狂犬病ワクチン", Species: &dogSpecies}},
			want:    []string{"vaccine_dog_2026-04-25", "vaccine_rabies_2026-04-25"},
		},
		{name: "nil vaccine", vaccine: &model.Vaccination{Date: baseDate}, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, vaccineTagNames(tt.vaccine))
		})
	}
}

func TestIsRabiesVaccine(t *testing.T) {
	assert.True(t, isRabiesVaccine("rabies"))
	assert.True(t, isRabiesVaccine("RABIES vaccine"))
	assert.True(t, isRabiesVaccine("狂犬病予防ワクチン"))
	assert.False(t, isRabiesVaccine("混合ワクチン"))
	assert.False(t, isRabiesVaccine("フィラリア"))
}
