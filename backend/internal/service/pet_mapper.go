package service

import "github.com/animal-ekarte/backend/internal/model"

func buildPetModel(clinicID uint64, petNumber string, input *CreatePetInput) *model.Pet {
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         input.OwnerID,
		AnimalSpeciesID: input.AnimalSpeciesID,
		PetNumber:       petNumber,
		Name:            input.Name,
		NameKana:        input.PetNameKana,
		BirthDate:       input.BirthDate,
		Breed:           input.Breed,
		Color:           input.Color,
		Weight:          input.Weight,
		NeuteredDate:    input.NeuteredDate,
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
	return pet
}
