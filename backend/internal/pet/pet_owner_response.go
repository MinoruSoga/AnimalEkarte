package pet

import (
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ownerSharedPetAnimalSpeciesResponse struct {
	Name string `json:"name"`
}

type ownerSharedPetResponse struct {
	ID            uint64                              `json:"id"`
	PetNumber     string                              `json:"pet_number"`
	Name          string                              `json:"name"`
	Status        string                              `json:"status"`
	AnimalSpecies ownerSharedPetAnimalSpeciesResponse `json:"animal_species"`
	Gender        string                              `json:"gender"`
	BirthDate     *string                             `json:"birth_date"`
	Color         string                              `json:"color"`
	Weight        *float64                            `json:"weight"`
	Environment   string                              `json:"environment"`
	Remarks       string                              `json:"remarks"`
	Relationship  string                              `json:"relationship"`
}

type ownerSharedPetsResponse struct {
	SharedPets []ownerSharedPetResponse `json:"shared_pets"`
}

func newOwnerSharedPetsResponse(sharedPets []SharedPet) ownerSharedPetsResponse {
	items := make([]ownerSharedPetResponse, len(sharedPets))
	for i, sharedPet := range sharedPets {
		var birthDate *string
		if sharedPet.BirthDate != nil {
			value := sharedPet.BirthDate.In(time.Local).Format(time.DateOnly)
			birthDate = &value
		}
		items[i] = ownerSharedPetResponse{
			ID:        sharedPet.ID,
			PetNumber: sharedPet.PetNumber,
			Name:      sharedPet.Name,
			Status:    string(sharedPet.Status),
			AnimalSpecies: ownerSharedPetAnimalSpeciesResponse{
				Name: sharedPet.AnimalSpeciesName,
			},
			Gender:       string(sharedPet.Gender),
			BirthDate:    birthDate,
			Color:        sharedPet.Color,
			Weight:       sharedPet.Weight,
			Environment:  sharedPet.Environment,
			Remarks:      sharedPet.Remarks,
			Relationship: sharedPet.Relationship,
		}
	}
	return ownerSharedPetsResponse{SharedPets: items}
}

type petOwnerResponse struct {
	OwnerID      uint64 `json:"owner_id"`
	Name         string `json:"name"`
	NameKana     string `json:"name_kana"`
	Relationship string `json:"relationship"`
}

type petOwnersResponse struct {
	SubOwners []petOwnerResponse `json:"sub_owners"`
}

type petOwnerResponseLink struct {
	OwnerID      uint64
	Relationship string
}

func newPetOwnersResponse(
	links []model.PetOwner,
	owners []*model.Owner,
) (petOwnersResponse, error) {
	responseLinks := make([]petOwnerResponseLink, len(links))
	for i, link := range links {
		responseLinks[i] = petOwnerResponseLink{
			OwnerID:      link.OwnerID,
			Relationship: link.Relationship,
		}
	}
	return mapPetOwnersResponse(responseLinks, owners)
}

func mapPetOwnersResponse(
	links []petOwnerResponseLink,
	owners []*model.Owner,
) (petOwnersResponse, error) {
	ownersByID := make(map[uint64]*model.Owner, len(owners))
	for _, owner := range owners {
		if owner != nil {
			ownersByID[owner.ID] = owner
		}
	}

	subOwners := make([]petOwnerResponse, len(links))
	for i, link := range links {
		owner, ok := ownersByID[link.OwnerID]
		if !ok {
			return petOwnersResponse{}, apperrors.WrapNotFound(
				"owner",
				fmt.Sprintf("%d", link.OwnerID),
			)
		}
		subOwners[i] = petOwnerResponse{
			OwnerID:      link.OwnerID,
			Name:         owner.Name,
			NameKana:     owner.NameKana,
			Relationship: link.Relationship,
		}
	}
	return petOwnersResponse{SubOwners: subOwners}, nil
}
