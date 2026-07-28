package lstep

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestLstepCSVImportOwnerLookup_FindLineUserIDs_ClinicIsolation(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	clinicA := seedLstepCsvImportClinic(t, db)
	clinicB := seedLstepCsvImportClinic(t, db)
	lineA := "U-clinic-a"
	lineB := "U-clinic-b"
	lineDeleted := "U-deleted"
	owners := []*model.Owner{
		{ClinicID: clinicA, Name: "A", NameKana: "A", LineUserID: &lineA},
		{ClinicID: clinicB, Name: "B", NameKana: "B", LineUserID: &lineB},
		{ClinicID: clinicA, Name: "Deleted", NameKana: "Deleted", LineUserID: &lineDeleted},
		{ClinicID: clinicA, Name: "Unlinked", NameKana: "Unlinked"},
	}
	for _, owner := range owners {
		require.NoError(t, db.Create(owner).Error)
	}
	require.NoError(t, db.Delete(owners[2]).Error)
	lookup := NewLstepCSVImportOwnerLookup()

	got, err := lookup.FindExistingLineUserIDs(
		context.Background(),
		db,
		clinicA,
		[]string{lineA, lineB, lineDeleted, "U-missing"},
	)

	require.NoError(t, err)
	assert.Equal(t, map[string]struct{}{lineA: {}}, got)
}
