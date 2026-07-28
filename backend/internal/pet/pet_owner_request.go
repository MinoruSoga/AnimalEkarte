package pet

import "strings"

type replacePetOwnerRequest struct {
	OwnerID      uint64 `json:"owner_id" binding:"required"`
	Relationship string `json:"relationship"`
}

type replacePetOwnersRequest struct {
	Version   *int                      `json:"version" binding:"required"`
	SubOwners *[]replacePetOwnerRequest `json:"sub_owners" binding:"required,dive"`
}

func (r replacePetOwnersRequest) toServiceInput() *ReplacePetOwnersInput {
	links := make([]PetOwnerLinkInput, len(*r.SubOwners))
	for i, subOwner := range *r.SubOwners {
		links[i] = PetOwnerLinkInput{
			OwnerID:      subOwner.OwnerID,
			Relationship: strings.TrimSpace(subOwner.Relationship),
		}
	}
	return &ReplacePetOwnersInput{
		Version: r.Version,
		Links:   links,
	}
}
