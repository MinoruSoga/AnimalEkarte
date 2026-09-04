package owner

import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

func normalizeNameKana(value string) string {
	return textsearch.NormalizeKana(value)
}

func buildOwnerModel(clinicID uint64, input *CreateOwnerInput) *model.Owner {
	membershipType := input.MembershipType
	if membershipType == "" {
		membershipType = model.MembershipTypeNonMember
	}

	return &model.Owner{
		ClinicID:       clinicID,
		Name:           input.OwnerName,
		NameKana:       normalizeNameKana(input.OwnerNameKana),
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: membershipType,
		DMPreference:   input.DMPreference,
	}
}

func buildOwnerPetModels(inputs []CreatePetForOwnerInput) []model.Pet {
	pets := make([]model.Pet, 0, len(inputs))
	for i := range inputs {
		p := &inputs[i]
		pet := model.Pet{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			NameKana:        normalizeNameKana(p.PetNameKana),
			Breed:           p.Breed,
			Color:           p.Color,
			BirthDate:       p.BirthDate,
			Weight:          p.Weight,
			NeuteredDate:    p.NeuteredDate,
			DangerReason:    p.DangerReason,
			Food:            p.Food,
			Environment:     p.Environment,
			InsuranceID:     p.InsuranceID,
			Remarks:         p.Remarks,
		}
		// 血液型 / マイクロチップ番号は nullable。空文字は未設定として NULL のままにする。
		if p.BloodType != "" {
			bt := p.BloodType
			pet.BloodType = &bt
		}
		if p.MicrochipNumber != "" {
			mc := p.MicrochipNumber
			pet.MicrochipNumber = &mc
		}
		if p.Gender != "" {
			pet.Gender = model.PetGender(p.Gender)
		}
		if p.Status != "" {
			pet.Status = model.PetStatus(p.Status)
		}
		if p.AcquisitionType != "" {
			at := model.AcquisitionType(p.AcquisitionType)
			pet.AcquisitionType = &at
		}
		if p.DangerLevel != "" {
			pet.DangerLevel = model.DangerLevel(p.DangerLevel)
		}
		pets = append(pets, pet)
	}
	return pets
}
