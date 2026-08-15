package lstep

// Tag-code mapping cardinality caps (SEC-CS-F11).
// Validated in the service BEFORE the replacement transaction so over-limit
// requests never call SoftDelete/Create.
const (
	// MaxTagCodeMappingEntries は 1 タグ置換リクエストの最大エントリ数。
	MaxTagCodeMappingEntries = 32
	// MaxTagCodeMappingCodesPerEntry は 1 エントリの codes 配列の最大長。
	MaxTagCodeMappingCodesPerEntry = 100
	// MaxTagCodeMappingTotalCodes は 1 リクエスト全体の codes 合計上限。
	MaxTagCodeMappingTotalCodes = 200
)

type putTagCodeMappingEntryRequest struct {
	CodeType     string   `json:"code_type"     binding:"required"`
	Codes        []string `json:"codes"         binding:"required"`
	SpeciesScope string   `json:"species_scope"`
	AgeMin       *int     `json:"age_min"`
}

type putTagCodeMappingsRequest struct {
	Entries []putTagCodeMappingEntryRequest `json:"entries" binding:"required"`
}
