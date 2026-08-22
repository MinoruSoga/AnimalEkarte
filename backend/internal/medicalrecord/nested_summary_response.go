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
// These are package-local, behavior-identical copies (exported for tygo wire codegen; TASK-444-S2) (same JSON field tags → byte-identical
// output) following the "文書化付き local copy" precedent established by
// sub-batch①'s validators.go. The originals stay in internal/handler (still used by ~10
// other, not-yet-migrated handlers); a later batch that migrates those handlers can collapse
// the duplication.

// StaffSummaryResponse mirrors internal/handler.StaffSummaryResponse (staff_response.go).
type StaffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary mirrors internal/handler.toStaffSummary. nil の場合は nil を返す。
func toStaffSummary(s *model.Staff) *StaffSummaryResponse {
	if s == nil {
		return nil
	}
	return &StaffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// OwnerSummaryResponse mirrors internal/handler.OwnerSummaryResponse (owner_response.go).
type OwnerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary mirrors internal/handler.toOwnerSummary. nil の場合は nil を返す。
func toOwnerSummary(o *model.Owner) *OwnerSummaryResponse {
	if o == nil {
		return nil
	}
	return &OwnerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}

// AnimalSpeciesSummaryResponse mirrors internal/handler.AnimalSpeciesSummaryResponse
// (pet_response.go).
type AnimalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// PetSummaryResponse mirrors internal/handler.PetSummaryResponse (pet_response.go). Only the
// fields the vaccination list response actually serializes are populated by toPetSummary,
// identical to the original.
type PetSummaryResponse struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	PetNumber string   `json:"pet_number"`
	Weight    *float64 `json:"weight,omitempty"`
	Status    string   `json:"status,omitempty"`
	Breed     string   `json:"breed,omitempty"`

	Gender        string                        `json:"gender,omitempty"`
	BirthDate     *string                       `json:"birth_date,omitempty"`
	NeuteredDate  *string                       `json:"neutered_date,omitempty"`
	AnimalSpecies *AnimalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *OwnerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary mirrors internal/handler.toPetSummary. nil の場合は nil を返す。
func toPetSummary(p *model.Pet) *PetSummaryResponse {
	if p == nil {
		return nil
	}
	resp := &PetSummaryResponse{
		ID:        p.ID,
		Name:      p.Name,
		PetNumber: p.PetNumber,
		Weight:    p.Weight,
		Status:    string(p.Status),
		Breed:     p.Breed,
		Gender:    string(p.Gender),
		Owner:     toOwnerSummary(p.Owner),
	}
	if p.BirthDate != nil {
		d := p.BirthDate.Format("2006-01-02")
		resp.BirthDate = &d
	}
	if p.NeuteredDate != nil {
		d := p.NeuteredDate.Format("2006-01-02")
		resp.NeuteredDate = &d
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &AnimalSpeciesSummaryResponse{
			ID:   p.AnimalSpecies.ID,
			Name: p.AnimalSpecies.Name,
		}
	}
	return resp
}
