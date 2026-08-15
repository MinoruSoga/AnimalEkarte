package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toStaffSummary(staff *model.Staff) *staffSummaryResponse {
	if staff == nil {
		return nil
	}
	return &staffSummaryResponse{ID: staff.ID, Name: staff.Name}
}

type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

func toOwnerSummary(owner *model.Owner) *ownerSummaryResponse {
	if owner == nil {
		return nil
	}
	return &ownerSummaryResponse{ID: owner.ID, OwnerName: owner.Name}
}

type animalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type petSummaryResponse struct {
	ID            uint64                        `json:"id"`
	Name          string                        `json:"name"`
	PetNumber     string                        `json:"pet_number"`
	Weight        *float64                      `json:"weight,omitempty"`
	Status        string                        `json:"status,omitempty"`
	Breed         string                        `json:"breed,omitempty"`
	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *ownerSummaryResponse         `json:"owner,omitempty"`
}

func toPetSummary(pet *model.Pet) *petSummaryResponse {
	if pet == nil {
		return nil
	}
	response := &petSummaryResponse{
		ID:        pet.ID,
		Name:      pet.Name,
		PetNumber: pet.PetNumber,
		Weight:    pet.Weight,
		Status:    string(pet.Status),
		Breed:     pet.Breed,
		Owner:     toOwnerSummary(pet.Owner),
	}
	if pet.AnimalSpecies != nil {
		response.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   pet.AnimalSpecies.ID,
			Name: pet.AnimalSpecies.Name,
		}
	}
	return response
}

func localTime(value time.Time) time.Time {
	return httpapi.LocalTime(value)
}
