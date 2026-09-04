package seedlogin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/seedbundle"
)

func TestCatalogMatchesLoginFormContract(t *testing.T) {
	t.Parallel()

	catalog := Catalog()
	require.Len(t, catalog, 40)

	seen := make(map[uint64]struct{}, len(catalog))
	emails := make(map[string]struct{}, len(catalog))
	for _, row := range catalog {
		_, dupID := seen[row.StaffID]
		assert.False(t, dupID, "duplicate staff id %d", row.StaffID)
		seen[row.StaffID] = struct{}{}
		_, dupEmail := emails[row.Email]
		assert.False(t, dupEmail, "duplicate email %s", row.Email)
		emails[row.Email] = struct{}{}
		assert.Equal(t, EmailForStaffID(row.StaffID), row.Email)
		assert.Contains(t, []uint64{1, 2, 3, 4}, row.ClinicID)
		assert.NotEmpty(t, row.Name)
		assert.Contains(t, []model.StaffType{model.StaffTypeDoctor, model.StaffTypeNurse}, row.StaffType)
		assert.NotEmpty(t, row.OccupationLabel)
		assert.NotEmpty(t, row.ClinicLabel)
	}

	assert.Equal(t, "stg-staff-10000021@example.test", catalog[0].Email)
	assert.Equal(t, uint64(1), catalog[0].ClinicID)
	assert.Equal(t, model.StaffTypeDoctor, catalog[0].StaffType)
	assert.Equal(t, "stg-staff-31000009@example.test", catalog[len(catalog)-1].Email)
	assert.Equal(t, uint64(4), catalog[len(catalog)-1].ClinicID)
	assert.Equal(t, model.StaffTypeNurse, catalog[len(catalog)-1].StaffType)
}

func TestMigrationKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, seedbundle.BundleMigrationKey(BundleDir), MigrationKey())
	assert.Equal(t, "seeds/003_login", MigrationKey())
}

func TestCatalogChecksumStable(t *testing.T) {
	t.Parallel()
	assert.Equal(t, CatalogChecksum(), CatalogChecksum())
	assert.Len(t, CatalogChecksum(), 64)
}
