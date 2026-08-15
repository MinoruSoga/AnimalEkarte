package staff

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShiftEntryRepository_FindClinicIDsByStaffID_GeneratesScopedBatchQuery(t *testing.T) {
	var output bytes.Buffer
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info}),
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	repo := NewShiftEntryRepository(db)
	clinicIDs, err := repo.FindClinicIDsByStaffID(context.Background(), []uint64{7, 19}, 42)

	require.NoError(t, err)
	assert.Empty(t, clinicIDs)
	query := output.String()
	assert.Contains(t, query, `SELECT DISTINCT "clinic_id"`)
	assert.Contains(t, query, `FROM "shift_entries"`)
	assert.Contains(t, query, "clinic_id IN (7,19)")
	assert.Contains(t, query, "staff_id = 42")
}

func TestShiftEntryRepository_FindClinicIDsByStaffID_EmptyClinicScopeDoesNotQuery(t *testing.T) {
	var output bytes.Buffer
	db, err := gorm.Open(
		postgres.Open("host=localhost user=test dbname=test sslmode=disable"),
		&gorm.Config{
			DisableAutomaticPing:   true,
			DryRun:                 true,
			Logger:                 logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info}),
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	repo := NewShiftEntryRepository(db)
	clinicIDs, err := repo.FindClinicIDsByStaffID(context.Background(), nil, 42)

	require.NoError(t, err)
	assert.Empty(t, clinicIDs)
	assert.Empty(t, output.String())
}
