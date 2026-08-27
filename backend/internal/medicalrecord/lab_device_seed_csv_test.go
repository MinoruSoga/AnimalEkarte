package medicalrecord

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedDir returns the 003_demo seed directory relative to this source file.
func seedDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	// Local:     backend/internal/medicalrecord → ../.. → backend → migrations/seeds/003_demo
	// Container: /app/internal/medicalrecord   → ../.. → /app   → migrations/seeds/003_demo
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations", "seeds", "003_demo")
}

func readSeedCSV(t *testing.T, dir, csvFile string) []map[string]string {
	t.Helper()
	path := filepath.Join(dir, csvFile)
	f, err := os.Open(path)
	require.NoError(t, err, "open %s", path)
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	require.NoError(t, err, "read %s", path)
	require.Greater(t, len(records), 1, "CSV %s must have header + at least one row", csvFile)
	headers := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, rec := range records[1:] {
		row := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(rec) {
				row[h] = rec[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func TestLabDeviceSeedCSV_ItemMastersMappedAll(t *testing.T) {
	dir := seedDir(t)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("003_demo lab device CSV retired; catalog lives in clinic master UI")
	}
	rows := readSeedCSV(t, dir, "lab_device_item_masters.csv")

	jouto := make([]map[string]string, 0, LabDeviceItemCatalogCount)
	for _, r := range rows {
		if r["clinic_id"] == "2" {
			jouto = append(jouto, r)
		}
	}

	assert.Len(t, jouto, LabDeviceItemCatalogCount,
		"lab_device_item_masters.csv must have %d rows for clinic_id=2", LabDeviceItemCatalogCount)

	for _, r := range jouto {
		assert.NotEmpty(t, r["exam_type_field_id"],
			"clinic_id=2 source=%s code=%s must have exam_type_field_id set",
			r["source_type"], r["device_item_code"])
		assert.NotEmpty(t, r["source_type"], "source_type must not be empty")
		assert.NotEmpty(t, r["device_item_code"], "device_item_code must not be empty")
	}
}

func TestLabDeviceSeedCSV_DevicesExamTypeMapping(t *testing.T) {
	dir := seedDir(t)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("003_demo lab device CSV retired; catalog lives in clinic master UI")
	}
	rows := readSeedCSV(t, dir, "lab_devices.csv")

	wantExamType := map[string]string{
		"fuji_nx600":    "7",  // 血液化学検査
		"fuji_au10v":    "7",  // 血液化学検査 (SAA)
		"arkray_pu4010": "13", // 尿検査
		"idexx_vetlab":  "6",  // 血液検査(CBC)
	}

	jouto := make([]map[string]string, 0, 4)
	for _, r := range rows {
		if r["clinic_id"] == "2" {
			jouto = append(jouto, r)
		}
	}
	assert.Len(t, jouto, len(wantExamType),
		"lab_devices.csv must have %d rows for clinic_id=2", len(wantExamType))

	for _, r := range jouto {
		st := r["source_type"]
		want, ok := wantExamType[st]
		assert.True(t, ok, "unexpected source_type %q in seed CSV", st)
		assert.Equal(t, want, r["exam_type_id"],
			"source_type %s exam_type_id want %s got %s", st, want, r["exam_type_id"])
	}
}

func TestLabDeviceSeedCSV_ItemMastersCoverCatalog(t *testing.T) {
	dir := seedDir(t)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("003_demo lab device CSV retired; catalog lives in clinic master UI")
	}
	rows := readSeedCSV(t, dir, "lab_device_item_masters.csv")

	// Build set of (source_type, device_item_code) from seed CSV for clinic_id=2.
	inSeed := make(map[string]bool, LabDeviceItemCatalogCount)
	for _, r := range rows {
		if r["clinic_id"] == "2" {
			key := r["source_type"] + "\x00" + r["device_item_code"]
			inSeed[key] = true
		}
	}

	// Every catalog item must appear in the seed.
	catalog := labDeviceItemCatalog()
	for _, item := range catalog {
		key := string(item.SourceType) + "\x00" + item.DeviceItemCode
		assert.True(t, inSeed[key],
			"catalog item (%s, %s) not found in seed CSV", item.SourceType, item.DeviceItemCode)
	}
}
