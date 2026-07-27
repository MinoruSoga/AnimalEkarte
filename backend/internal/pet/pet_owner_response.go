package pet

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

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
