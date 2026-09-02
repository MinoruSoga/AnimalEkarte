package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go — medicalrecord-local nested DTOs for checkup/vaccination
// JSON (exported for tygo). Field sets are domain-specific and must not be merged with
// billing/reservation summaries.

// StaffSummaryResponse is the nested staff JSON for medicalrecord responses.
type StaffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary returns nil when s is nil.
func toStaffSummary(s *model.Staff) *StaffSummaryResponse {
	if s == nil {
		return nil
	}
	return &StaffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// OwnerSummaryResponse is the nested owner JSON for medicalrecord responses.
type OwnerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary returns nil when o is nil.
func toOwnerSummary(o *model.Owner) *OwnerSummaryResponse {
	if o == nil {
		return nil
	}
	return &OwnerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}

// AnimalSpeciesSummaryResponse is the nested species JSON for medicalrecord responses.
type AnimalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// PetSummaryResponse is the nested pet JSON for vaccination/checkup lists.
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

// toPetSummary returns nil when p is nil.
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
