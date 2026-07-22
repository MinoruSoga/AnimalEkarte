package service

// Residual lifecycle tests use only the prefix-reading part of the tag-config
// repository. The full tag-service test double moved to internal/lstep in L③a.

import (
	"context"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagConfigRepository struct {
	findAllAutoManagedPrefixesFn func(context.Context) ([]*model.LstepAutoManagedPrefix, error)
}

func (m *mockLstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.findAllAutoManagedPrefixesFn != nil {
		return m.findAllAutoManagedPrefixesFn(ctx)
	}
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateAutoManagedPrefix(context.Context, *model.LstepAutoManagedPrefix) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteAutoManagedPrefix(context.Context, uint64) error {
	return nil
}

func (*mockLstepTagConfigRepository) FindAllConditionTagMappings(context.Context) ([]*model.LstepConditionTagMapping, error) {
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateConditionTagMapping(context.Context, *model.LstepConditionTagMapping) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteConditionTagMapping(context.Context, uint64) error {
	return nil
}

func (*mockLstepTagConfigRepository) FindAllSendPurposeTagPrefixes(context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	return nil, nil
}

func (*mockLstepTagConfigRepository) CreateSendPurposeTagPrefix(context.Context, *model.LstepSendPurposeTagPrefix) error {
	return nil
}

func (*mockLstepTagConfigRepository) DeleteSendPurposeTagPrefix(context.Context, uint64) error {
	return nil
}

func ltvBracketTag(ltv int64) string {
	switch {
	case ltv >= 80_000:
		return "ltv_amount_8"
	case ltv >= 50_000:
		return "ltv_amount_5"
	case ltv >= 20_000:
		return "ltv_amount_2"
	default:
		return "ltv_amount_0"
	}
}

func visitCountAnnualTag(count int64) string {
	switch {
	case count >= 10:
		return "visit_count_annual_10"
	case count >= 5:
		return "visit_count_annual_5"
	case count >= 3:
		return "visit_count_annual_3"
	case count >= 2:
		return "visit_count_annual_2"
	default:
		return "visit_count_annual_1"
	}
}

func vaccineTagNames(vac *model.Vaccination) []string {
	if vac.Vaccine == nil {
		return nil
	}
	date := vac.Date.Format(time.DateOnly)
	var tags []string
	if vac.Vaccine.Species != nil {
		switch *vac.Vaccine.Species {
		case model.VaccineSpeciesDog:
			tags = append(tags, "vaccine_dog_"+date)
		case model.VaccineSpeciesCat:
			tags = append(tags, "vaccine_cat_"+date)
		case model.VaccineSpeciesBoth:
			tags = append(tags, "vaccine_dog_"+date, "vaccine_cat_"+date)
		}
	}
	if isRabiesVaccine(vac.Vaccine.Name) {
		tags = append(tags, "vaccine_rabies_"+date)
	}
	return tags
}

func isRabiesVaccine(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "rabies") || strings.Contains(name, "狂犬病")
}
