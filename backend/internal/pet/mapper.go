package pet

import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

func normalizeNameKana(value string) string {
	return textsearch.NormalizeKana(value)
}

func buildPetModel(clinicID uint64, petNumber string, input *CreatePetInput) *model.Pet {
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       petNumber,
		Name:            input.Name,
		NameKana:        normalizeNameKana(input.PetNameKana),
		BirthDate:       input.BirthDate,
		Breed:           input.Breed,
		Color:           input.Color,
		Weight:          input.Weight,
		NeuteredDate:    input.NeuteredDate,
		DangerReason:    input.DangerReason,
		Food:            input.Food,
		Environment:     input.Environment,
		Phone:           input.Phone,
		InsuranceID:     input.InsuranceID,
		Remarks:         input.Remarks,
	}
	if input.Gender != "" {
		pet.Gender = model.PetGender(input.Gender)
	}
	if input.Status != "" {
		pet.Status = model.PetStatus(input.Status)
	}
	if input.AcquisitionType != "" {
		at := model.AcquisitionType(input.AcquisitionType)
		pet.AcquisitionType = &at
	}
	if input.DangerLevel != "" {
		pet.DangerLevel = model.DangerLevel(input.DangerLevel)
	}
	// 血液型 / マイクロチップ番号は nullable: 未入力（空文字）は NULL のまま残す（記録済みと誤表示しない）。
	if input.BloodType != "" {
		bt := input.BloodType
		pet.BloodType = &bt
	}
	if input.MicrochipNumber != "" {
		mc := input.MicrochipNumber
		pet.MicrochipNumber = &mc
	}
	return pet
}
