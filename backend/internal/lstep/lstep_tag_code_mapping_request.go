package lstep

type putTagCodeMappingEntryRequest struct {
	CodeType     string   `json:"code_type"     binding:"required"`
	Codes        []string `json:"codes"         binding:"required"`
	SpeciesScope string   `json:"species_scope"`
	AgeMin       *int     `json:"age_min"`
}

type putTagCodeMappingsRequest struct {
	Entries []putTagCodeMappingEntryRequest `json:"entries" binding:"required"`
}
