package csvimport

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/seedlogin"
)

// TestSeedloginCatalogDoesNotOccupyCutoverNonOwnerBands keeps migrate-phase
// synthetic logins out of the csv-import ID band. Occupying staffs.ids here
// makes make reset fail with CUTOVER_REF_BAND_OCCUPIED before old_db handoff.
func TestSeedloginCatalogDoesNotOccupyCutoverNonOwnerBands(t *testing.T) {
	t.Parallel()

	catalog := seedlogin.Catalog()
	if len(catalog) == 0 {
		t.Fatal("seedlogin catalog is empty")
	}

	const maxOrdinal int64 = 50
	for _, spec := range catalog {
		id := int64(spec.StaffID)
		if id <= 0 {
			t.Fatalf("staff id must be positive, got %d", spec.StaffID)
		}
		if spec.ClinicID == 0 || spec.ClinicID > uint64(maxOrdinal) {
			t.Fatalf("seedlogin clinic %d is outside cutover ordinal range 1..%d", spec.ClinicID, maxOrdinal)
		}
		base := int64(spec.ClinicID) * clinicBandSize
		if id < base || id >= base+nonOwnerBandOffset {
			t.Fatalf(
				"seedlogin staff %d (clinic %d) must sit in EndExclusive gap [%d,%d)",
				spec.StaffID, spec.ClinicID, base, base+nonOwnerBandOffset,
			)
		}
		for ordinal := int64(1); ordinal <= maxOrdinal; ordinal++ {
			floor := (ordinal-1)*clinicBandSize + nonOwnerBandOffset
			end := ordinal * clinicBandSize
			if id >= floor && id < end {
				t.Fatalf(
					"seedlogin staff %d (clinic %d) occupies cutover non-owner band for ordinal %d [%d,%d); demo IDs must use clinic EndExclusive",
					spec.StaffID, spec.ClinicID, ordinal, floor, end,
				)
			}
		}
	}
}
