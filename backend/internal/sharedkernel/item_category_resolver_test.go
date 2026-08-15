package sharedkernel

import (
	"fmt"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestResolveItemCategory_AllDiscriminatorCombinations(t *testing.T) {
	carePlanTypes := []model.CarePlanType{
		model.CarePlanTypeFood,
		model.CarePlanTypeMedicine,
		model.CarePlanTypeTreatment,
		model.CarePlanTypeInstruction,
		model.CarePlanTypeItem,
	}
	treatmentItemTypes := []model.TreatmentItemType{
		model.TreatmentItemTypeConsultation,
		model.TreatmentItemTypeProcedure,
		model.TreatmentItemTypeMedicine,
		model.TreatmentItemTypeOther,
	}
	isSurgeryValues := []bool{false, true}
	hospitalizationTypes := []model.HospitalizationType{
		model.HospitalizationTypeInpatient,
		model.HospitalizationTypeHotel,
	}

	for _, carePlanType := range carePlanTypes {
		for _, treatmentItemType := range treatmentItemTypes {
			for _, isSurgery := range isSurgeryValues {
				for _, hospitalizationType := range hospitalizationTypes {
					input := ItemCategoryResolverInput{
						CarePlanType:        carePlanType,
						TreatmentItemType:   treatmentItemType,
						IsSurgery:           isSurgery,
						HospitalizationType: hospitalizationType,
					}
					name := fmt.Sprintf(
						"care_plan=%s/treatment=%s/surgery=%t/hospitalization=%s",
						carePlanType,
						treatmentItemType,
						isSurgery,
						hospitalizationType,
					)

					t.Run("hospitalization/"+name, func(t *testing.T) {
						input.Source = model.ItemSourceHospitalization
						assertItemCategory(
							t,
							ResolveItemCategory(input),
							expectedCarePlanItemCategory(carePlanType, isSurgery, hospitalizationType),
						)
					})
					t.Run("medical_record/"+name, func(t *testing.T) {
						input.Source = model.ItemSourceMedicalRecord
						assertItemCategory(
							t,
							ResolveItemCategory(input),
							expectedTreatmentItemCategory(treatmentItemType, isSurgery),
						)
					})
					t.Run("trimming/"+name, func(t *testing.T) {
						input.Source = model.ItemSourceTrimming
						assertItemCategory(t, ResolveItemCategory(input), model.ItemCategoryTrimming)
					})
				}
			}
		}
	}
}

func TestResolveItemCategory_UnknownDiscriminatorsReturnOther(t *testing.T) {
	tests := []struct {
		name  string
		input ItemCategoryResolverInput
	}{
		{
			name: "unknown source",
			input: ItemCategoryResolverInput{
				Source: model.ItemSource("unknown"),
			},
		},
		{
			name: "unknown care plan type",
			input: ItemCategoryResolverInput{
				Source:              model.ItemSourceHospitalization,
				CarePlanType:        model.CarePlanType("unknown"),
				HospitalizationType: model.HospitalizationTypeInpatient,
			},
		},
		{
			name: "unknown treatment item type",
			input: ItemCategoryResolverInput{
				Source:            model.ItemSourceMedicalRecord,
				TreatmentItemType: model.TreatmentItemType("unknown"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertItemCategory(t, ResolveItemCategory(tt.input), model.ItemCategoryOther)
		})
	}
}

func expectedCarePlanItemCategory(
	carePlanType model.CarePlanType,
	isSurgery bool,
	hospitalizationType model.HospitalizationType,
) model.ItemCategory {
	switch carePlanType {
	case model.CarePlanTypeMedicine:
		return model.ItemCategoryMedicine
	case model.CarePlanTypeFood:
		return model.ItemCategoryFood
	case model.CarePlanTypeTreatment:
		if isSurgery {
			return model.ItemCategorySurgery
		}
		return model.ItemCategoryProcedure
	case model.CarePlanTypeItem:
		if hospitalizationType == model.HospitalizationTypeHotel {
			return model.ItemCategoryHotel
		}
		return model.ItemCategoryOther
	default:
		return model.ItemCategoryOther
	}
}

func expectedTreatmentItemCategory(
	treatmentItemType model.TreatmentItemType,
	isSurgery bool,
) model.ItemCategory {
	switch treatmentItemType {
	case model.TreatmentItemTypeConsultation:
		return model.ItemCategoryExamination
	case model.TreatmentItemTypeProcedure:
		if isSurgery {
			return model.ItemCategorySurgery
		}
		return model.ItemCategoryProcedure
	case model.TreatmentItemTypeMedicine:
		return model.ItemCategoryMedicine
	default:
		return model.ItemCategoryOther
	}
}

func assertItemCategory(t *testing.T, got, want model.ItemCategory) {
	t.Helper()
	if got != want {
		t.Fatalf("ResolveItemCategory() = %q, want %q", got, want)
	}
}
