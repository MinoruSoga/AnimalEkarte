package medicalrecord

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestLabDeviceItemCatalog_JoutoObservedRows(t *testing.T) {
	rows := labDeviceItemCatalog()
	require.Len(t, rows, LabDeviceItemCatalogCount)

	seen := make(map[string]struct{}, len(rows))
	nx, au, urine := 0, 0, 0
	for _, row := range rows {
		key := string(row.SourceType) + "\x00" + row.DeviceItemCode
		_, dup := seen[key]
		assert.False(t, dup, "duplicate catalog key %s %s", row.SourceType, row.DeviceItemCode)
		seen[key] = struct{}{}
		assert.NotEmpty(t, row.DeviceItemCode)
		assert.NotEmpty(t, row.DisplayName)
		assert.True(t, isLabDeviceSourceType(string(row.SourceType)))
		assert.True(t, isLabDeviceValueShape(row.ValueShape))
		assert.NotEqual(t, "T.26", row.DeviceItemCode)
		assert.NotEqual(t, "COM.", row.DeviceItemCode)
		switch row.SourceType {
		case model.LabImportSourceTypeFujiNX600:
			nx++
		case model.LabImportSourceTypeFujiAU10V:
			au++
			assert.Equal(t, "vf-SAA", row.DeviceItemCode)
		case model.LabImportSourceTypeArkrayPU4010:
			urine++
		}
	}
	assert.Equal(t, 16, nx)
	assert.Equal(t, 1, au)
	assert.Equal(t, 8, urine)
}

func TestLabDeviceItemCatalogRow_NoLegacyColumn(t *testing.T) {
	assertNoLegacyField(t, reflect.TypeOf(labDeviceItemCatalogRow{}))
	assertNoLegacyField(t, reflect.TypeOf(model.LabDeviceItemMaster{}))
}

func assertNoLegacyField(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		assert.NotContains(t, name, "legacy", typ.Name()+" must not store legacy_name_candidate")
	}
}
