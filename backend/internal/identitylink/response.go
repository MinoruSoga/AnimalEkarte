package identitylink

import "github.com/animal-ekarte/backend/internal/model"

// OwnerSearchItem is a non-PHI-minimized search hit (name/phone needed for manual link UI).
type OwnerSearchItem struct {
	ClinicID uint64 `json:"clinic_id"`
	OwnerID  uint64 `json:"owner_id"`
	Name     string `json:"name"`
	NameKana string `json:"name_kana"`
	Phone    string `json:"phone"`
}

// PetSearchItem is a search hit for manual pet linking.
type PetSearchItem struct {
	ClinicID  uint64 `json:"clinic_id"`
	PetID     uint64 `json:"pet_id"`
	OwnerID   uint64 `json:"owner_id"`
	Name      string `json:"name"`
	NameKana  string `json:"name_kana"`
	PetNumber string `json:"pet_number"`
}

// OwnerGroupMemberResponse is a visible owner identity member.
type OwnerGroupMemberResponse struct {
	ClinicID uint64 `json:"clinic_id"`
	OwnerID  uint64 `json:"owner_id"`
}

// OwnerGroupResponse is the owner identity group wire DTO.
type OwnerGroupResponse struct {
	ID              uint64                     `json:"id"`
	CreatedClinicID uint64                     `json:"created_clinic_id"`
	Version         int64                      `json:"version"`
	Members         []OwnerGroupMemberResponse `json:"members"`
}

// PetGroupMemberResponse is a visible pet identity member.
type PetGroupMemberResponse struct {
	ClinicID uint64 `json:"clinic_id"`
	PetID    uint64 `json:"pet_id"`
}

// PetGroupResponse is the pet identity group wire DTO.
type PetGroupResponse struct {
	ID                        uint64                   `json:"id"`
	CreatedClinicID           uint64                   `json:"created_clinic_id"`
	OwnerGroupCreatedClinicID uint64                   `json:"owner_group_created_clinic_id"`
	OwnerGroupID              uint64                   `json:"owner_group_id"`
	Version                   int64                    `json:"version"`
	Members                   []PetGroupMemberResponse `json:"members"`
}

// LinkedTreatmentHistoryItem is a minimal treatment history row for the Phase 1 slice.
// It lives with the response DTOs so the response-only tygo source set is self-contained.
type LinkedTreatmentHistoryItem struct {
	ClinicID        uint64  `json:"clinic_id"`
	PetID           uint64  `json:"pet_id"`
	MedicalRecordID uint64  `json:"medical_record_id"`
	RecordNo        string  `json:"record_no"`
	RecordDate      string  `json:"record_date"`
	TreatmentID     uint64  `json:"treatment_id"`
	ItemType        string  `json:"item_type"`
	Content         string  `json:"content"`
	UnitPrice       int64   `json:"unit_price"`
	Quantity        float64 `json:"quantity"`
}

// LinkedTreatmentHistoryResponse wraps paginated linked history.
type LinkedTreatmentHistoryResponse struct {
	Items []LinkedTreatmentHistoryItem `json:"items"`
	Total int64                        `json:"total"`
	Page  int                          `json:"page"`
	Limit int                          `json:"limit"`
}

func toOwnerSearchItems(owners []model.Owner) []OwnerSearchItem {
	out := make([]OwnerSearchItem, 0, len(owners))
	for _, o := range owners {
		out = append(out, OwnerSearchItem{
			ClinicID: o.ClinicID,
			OwnerID:  o.ID,
			Name:     o.Name,
			NameKana: o.NameKana,
			Phone:    o.Phone,
		})
	}
	return out
}

func toPetSearchItems(pets []model.Pet) []PetSearchItem {
	out := make([]PetSearchItem, 0, len(pets))
	for _, p := range pets {
		out = append(out, PetSearchItem{
			ClinicID:  p.ClinicID,
			PetID:     p.ID,
			OwnerID:   p.OwnerID,
			Name:      p.Name,
			NameKana:  p.NameKana,
			PetNumber: p.PetNumber,
		})
	}
	return out
}

func toOwnerGroupResponse(group *model.OwnerIdentityGroup, members []model.OwnerIdentityGroupMember) OwnerGroupResponse {
	ms := make([]OwnerGroupMemberResponse, 0, len(members))
	for _, m := range members {
		ms = append(ms, OwnerGroupMemberResponse{ClinicID: m.ClinicID, OwnerID: m.OwnerID})
	}
	version := int64(1)
	createdClinicID := uint64(0)
	id := uint64(0)
	if group != nil {
		id = group.ID
		createdClinicID = group.CreatedClinicID
		if group.Version > 0 {
			version = group.Version
		}
	}
	return OwnerGroupResponse{
		ID:              id,
		CreatedClinicID: createdClinicID,
		Version:         version,
		Members:         ms,
	}
}

func toPetGroupResponse(group *model.PetIdentityGroup, members []model.PetIdentityGroupMember) PetGroupResponse {
	ms := make([]PetGroupMemberResponse, 0, len(members))
	for _, m := range members {
		ms = append(ms, PetGroupMemberResponse{ClinicID: m.ClinicID, PetID: m.PetID})
	}
	resp := PetGroupResponse{Members: ms, Version: 1}
	if group != nil {
		resp.ID = group.ID
		resp.CreatedClinicID = group.CreatedClinicID
		resp.OwnerGroupCreatedClinicID = group.OwnerGroupCreatedClinicID
		resp.OwnerGroupID = group.OwnerGroupID
		if group.Version > 0 {
			resp.Version = group.Version
		}
	}
	return resp
}
