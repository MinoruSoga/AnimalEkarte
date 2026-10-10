package medicalrecord

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

// SQL generation only: no connection, fixtures or database operations.
// The real-DB tied-row test covers the resulting page membership separately.
func TestHospitalizationRepository_FindAll_PaginationSQLContract(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		wantOffset string
	}{
		{"first page", 1, ""},
		{"second page", 2, "OFFSET 200"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{
				DisableAutomaticPing: true,
				DryRun:               true,
				Logger: logger.New(log.New(&output, "", 0), logger.Config{
					LogLevel: logger.Info,
				}),
			})
			require.NoError(t, err)
			// DryRun retains the COUNT SQL buffer; a real query clears it after
			// execution. Reset that buffer so the following Find builds its SQL.
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("reset_dry_run_count", func(tx *gorm.DB) {
				if _, count := tx.Statement.Dest.(*int64); count {
					tx.Statement.SQL.Reset()
					tx.Statement.Vars = nil
				}
			}))
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, sqlDB.Close()) })
			repo := NewHospitalizationRepository(db)

			_, _, err = repo.FindAll(context.Background(), 7, nil, nil, nil, nil, nil, tt.page, 200)

			require.NoError(t, err)
			sql := output.String()
			assert.Contains(t, sql, "ORDER BY start_date DESC, created_at DESC, id DESC")
			assert.Contains(t, sql, "LIMIT 200")
			assert.Contains(t, sql, "clinic_id = 7")
			if tt.wantOffset != "" {
				assert.Contains(t, sql, tt.wantOffset)
			} else {
				assert.NotContains(t, sql, "OFFSET")
			}
		})
	}
}
