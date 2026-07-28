package sharedkernel

import "github.com/animal-ekarte/backend/internal/model"

// ItemCategoryResolverInput describes the backend-owned provenance used to
// derive a billing item category.
type ItemCategoryResolverInput struct {
	Source              model.ItemSource
	CarePlanType        model.CarePlanType
	TreatmentItemType   model.TreatmentItemType
	IsSurgery           bool
	HospitalizationType model.HospitalizationType
}

// ResolveItemCategory is the single backend authority for derived billing item
// categories. Unknown discriminator values intentionally fall back to other.
func ResolveItemCategory(input ItemCategoryResolverInput) model.ItemCategory {
	switch input.Source {
	case model.ItemSourceHospitalization:
		return resolveCarePlanItemCategory(input)
	case model.ItemSourceMedicalRecord:
		return resolveTreatmentItemCategory(input)
	case model.ItemSourceTrimming:
		return model.ItemCategoryTrimming
	default:
		return model.ItemCategoryOther
	}
}

func resolveCarePlanItemCategory(input ItemCategoryResolverInput) model.ItemCategory {
	switch input.CarePlanType {
	case model.CarePlanTypeMedicine:
		return model.ItemCategoryMedicine
	case model.CarePlanTypeFood:
		return model.ItemCategoryFood
	case model.CarePlanTypeTreatment:
		if input.IsSurgery {
			return model.ItemCategorySurgery
		}
		return model.ItemCategoryProcedure
	case model.CarePlanTypeItem:
		if input.HospitalizationType == model.HospitalizationTypeHotel {
			return model.ItemCategoryHotel
		}
		return model.ItemCategoryOther
	default:
		return model.ItemCategoryOther
	}
}

func resolveTreatmentItemCategory(input ItemCategoryResolverInput) model.ItemCategory {
	switch input.TreatmentItemType {
	case model.TreatmentItemTypeConsultation:
		return model.ItemCategoryExamination
	case model.TreatmentItemTypeProcedure:
		if input.IsSurgery {
			return model.ItemCategorySurgery
		}
		return model.ItemCategoryProcedure
	case model.TreatmentItemTypeMedicine:
		return model.ItemCategoryMedicine
	default:
		return model.ItemCategoryOther
	}
}
