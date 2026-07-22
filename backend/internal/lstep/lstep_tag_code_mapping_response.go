package lstep

import (
	"github.com/animal-ekarte/backend/internal/model"
)

type tagCodeMappingResponse struct {
	ID           uint64   `json:"id"`
	ClinicID     uint64   `json:"clinic_id"`
	TagName      string   `json:"tag_name"`
	CodeType     string   `json:"code_type"`
	Codes        []string `json:"codes"`
	SpeciesScope *string  `json:"species_scope,omitempty"`
	AgeMin       *int     `json:"age_min,omitempty"`
}

func toTagCodeMappingResponse(m *model.LstepTagCodeMapping) tagCodeMappingResponse {
	codes := []string(m.Codes)
	if codes == nil {
		codes = []string{}
	}
	return tagCodeMappingResponse{
		ID:           m.ID,
		ClinicID:     m.ClinicID,
		TagName:      m.TagName,
		CodeType:     m.CodeType,
		Codes:        codes,
		SpeciesScope: m.SpeciesScope,
		AgeMin:       m.AgeMin,
	}
}

func toTagCodeMappingListResponse(ms []*model.LstepTagCodeMapping) []tagCodeMappingResponse {
	out := make([]tagCodeMappingResponse, len(ms))
	for i, m := range ms {
		out[i] = toTagCodeMappingResponse(m)
	}
	return out
}
