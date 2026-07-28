package lstep

import "strings"

// lstepBatchPageSize is the cursor page size for tag-sync batch processing.
const lstepBatchPageSize = 500

// These fuzzy matches are intentionally limited to marketing-tag classification.
// Medication dose matching has a separate exact-match, fail-closed contract.
func isDogSpeciesName(name string) bool {
	return strings.Contains(name, "犬")
}

func isCatSpeciesName(name string) bool {
	return strings.Contains(name, "猫")
}
