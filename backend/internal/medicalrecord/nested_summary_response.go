package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go — documented local copies (BE9-2D Batch C) of the small,
// shared nested-summary response DTOs and their converters that live in
// internal/handler ({staff,pet,owner}_response.go). The checkup / vaccination response
// bodies moved into this package embed these summaries (Doctor, Pet, Owner, AnimalSpecies),
// but internal/handler already imports internal/medicalrecord — so medicalrecord cannot
// import internal/handler back without creating an import cycle (ADR-006 aggregator 非経由).
// These are unexported, behavior-identical copies (same JSON field tags → byte-identical
// output) following the "文書化付き unexported local copy" precedent established by
// sub-batch①'s validators.go. The originals stay in internal/handler (still used by ~10
// other, not-yet-migrated handlers); a later batch that migrates those handlers can collapse
// the duplication.

// staffSummaryResponse mirrors internal/handler.staffSummaryResponse (staff_response.go).
type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary mirrors internal/handler.toStaffSummary. nil の場合は nil を返す。
func toStaffSummary(s *model.Staff) *staffSummaryResponse {
	if s == nil {
		return nil
	}
	return &staffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// ownerSummaryResponse mirrors internal/handler.ownerSummaryResponse (owner_response.go).
type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary mirrors internal/handler.toOwnerSummary. nil の場合は nil を返す。
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}

// animalSpeciesSummaryResponse mirrors internal/handler.animalSpeciesSummaryResponse
// (pet_response.go).
type animalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// petSummaryResponse mirrors internal/handler.petSummaryResponse (pet_response.go). Only the
// fields the vaccination list response actually serializes are populated by toPetSummary,
// identical to the original.
type petSummaryResponse struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	PetNumber string   `json:"pet_number"`
	Weight    *float64 `json:"weight,omitempty"`
	Status    string   `json:"status,omitempty"`
	Breed     string   `json:"breed,omitempty"`

	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *ownerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary mirrors internal/handler.toPetSummary. nil の場合は nil を返す。
func toPetSummary(p *model.Pet) *petSummaryResponse {
	if p == nil {
		return nil
	}
	resp := &petSummaryResponse{
		ID:        p.ID,
		Name:      p.Name,
		PetNumber: p.PetNumber,
		Weight:    p.Weight,
		Status:    string(p.Status),
		Breed:     p.Breed,
		Owner:     toOwnerSummary(p.Owner),
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   p.AnimalSpecies.ID,
			Name: p.AnimalSpecies.Name,
		}
	}
	return resp
}
