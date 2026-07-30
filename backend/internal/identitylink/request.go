package identitylink

// CreateOwnerGroupRequest creates a new owner identity group from members.
type CreateOwnerGroupRequest struct {
	Members []OwnerMemberRef `json:"members" binding:"required,min=2,dive"`
}

// AddOwnerMembersRequest adds members to an existing owner identity group.
type AddOwnerMembersRequest struct {
	Members []OwnerMemberRef `json:"members" binding:"required,min=1,dive"`
}

// UnlinkOwnerMemberRequest unlinks one owner member.
type UnlinkOwnerMemberRequest struct {
	ClinicID uint64 `json:"clinic_id" binding:"required"`
	OwnerID  uint64 `json:"owner_id" binding:"required"`
}

// CreatePetGroupRequest creates a pet identity group under an owner identity group.
type CreatePetGroupRequest struct {
	OwnerGroupID uint64         `json:"owner_group_id" binding:"required"`
	Members      []PetMemberRef `json:"members" binding:"required,min=2,dive"`
}

// AddPetMembersRequest adds pets to an existing pet identity group.
type AddPetMembersRequest struct {
	Members []PetMemberRef `json:"members" binding:"required,min=1,dive"`
}

// UnlinkPetMemberRequest unlinks one pet member.
type UnlinkPetMemberRequest struct {
	ClinicID uint64 `json:"clinic_id" binding:"required"`
	PetID    uint64 `json:"pet_id" binding:"required"`
}
