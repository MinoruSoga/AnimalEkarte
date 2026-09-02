package reservation

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go — reservation-local nested DTOs. Pet summary includes
// danger_level and must not be merged with medicalrecord/billing summaries.

// staffSummaryResponse is nested staff JSON for reservation responses.
type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary returns nil when s is nil.
func toStaffSummary(s *model.Staff) *staffSummaryResponse {
	if s == nil {
		return nil
	}
	return &staffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// ownerSummaryResponse is nested owner JSON for reservation responses.
type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary returns nil when o is nil.
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}

// animalSpeciesSummaryResponse is nested species JSON for reservation responses.
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
