package reservation

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go contains reservation-local copies of the small nested-summary
// DTOs and converters established during BE9. Shared fields stay aligned where applicable;
// reservation-specific staff response fields are documented on the local DTO.

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

// petSummaryResponse is the staff-facing reservation pet summary. DangerLevel intentionally
// exists only in this domain copy to support the Reception clinical-safety badge.
type petSummaryResponse struct {
	ID           uint64   `json:"id"`
	Name         string   `json:"name"`
	PetNumber    string   `json:"pet_number"`
	Weight       *float64 `json:"weight,omitempty"`
	Status       string   `json:"status,omitempty"`
	DangerLevel  string   `json:"danger_level,omitempty"`
	DangerReason *string  `json:"danger_reason,omitempty"`
	Breed        string   `json:"breed,omitempty"`

	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *ownerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary maps a pet into the reservation-local summary. nil の場合は nil を返す。
func toPetSummary(p *model.Pet) *petSummaryResponse {
	if p == nil {
		return nil
	}
	resp := &petSummaryResponse{
		ID:           p.ID,
		Name:         p.Name,
		PetNumber:    p.PetNumber,
		Weight:       p.Weight,
		Status:       string(p.Status),
		DangerLevel:  string(p.DangerLevel),
		DangerReason: p.DangerReason,
		Breed:        p.Breed,
		Owner:        toOwnerSummary(p.Owner),
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   p.AnimalSpecies.ID,
			Name: p.AnimalSpecies.Name,
		}
	}
	return resp
}
