package owner

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// buildOwnerModel は新規登録で作る owners 行（飼主ローカルの列だけ）を組み立てる。
// 氏名・住所・連絡先などの基本情報は owners の列ではなく共通 basic revision の項目で、ここでは扱わない。
func buildOwnerModel(clinicID uint64, input *CreateOwnerInput) *model.Owner {
	membershipType := input.MembershipType
	if membershipType == "" {
		membershipType = model.MembershipTypeNonMember
	}

	return &model.Owner{
		ClinicID:       clinicID,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: membershipType,
		DMPreference:   input.DMPreference,
	}
}

// buildOwnerPetModels は新規登録で作る pets 行（ペットローカルの列だけ）を組み立てる。
// 名前・動物種・性別などの基本情報は pets の列ではなく pet の basic revision の項目で、ここでは扱わない。
func buildOwnerPetModels(inputs []CreatePetForOwnerInput) []model.Pet {
	pets := make([]model.Pet, 0, len(inputs))
	for i := range inputs {
		pets = append(pets, model.Pet{InsuranceID: inputs[i].InsuranceID})
	}
	return pets
}
