package medicalrecord

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/testdb"
)

const examReferenceRangeTestSchema = "exam_reference_range_u1_test"

var errRollbackExamReferenceRangeTest = errors.New("rollback exam reference range migration test")

func readExamReferenceRangeMigration(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	initial := string(raw)

	const sourceMarker = "-- Source file: 005_exam_reference_ranges_and_clinic_fk.sql"
	const nextSourceMarker = "-- Source file: 006_payment_splits_payment_method_clinic_fk.sql"
	start := strings.Index(initial, sourceMarker)
	require.GreaterOrEqual(t, start, 0, "001_init.sql must contain the archived exam reference range migration")

	endOffset := strings.Index(initial[start:], "\n"+nextSourceMarker)
	require.Greater(t, endOffset, 0, "archived exam reference range migration must end at the 006 source marker")
	block := initial[start : start+endOffset]

	shaOffset := strings.Index(block, "-- Source SHA-256:")
	require.GreaterOrEqual(t, shaOffset, 0, "archived exam reference range migration must contain its SHA-256 header")
	bodyOffset := strings.Index(block[shaOffset:], "\n")
	require.GreaterOrEqual(t, bodyOffset, 0, "archived exam reference range migration metadata must end before its SQL body")

	return block[shaOffset+bodyOffset+1:]
}

func examReferenceRangeSchemaDDL(t *testing.T, migration string) string {
	t.Helper()
	const rlsMarker = "-- RLS policies"
	end := strings.Index(migration, rlsMarker)
	require.GreaterOrEqual(t, end, 0, "migration must contain the RLS marker")
	return migration[:end]
}

func withExamReferenceRangeMigrationSchema(t *testing.T, fn func(*gorm.DB)) {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)

	err := db.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, tx.Exec("CREATE SCHEMA "+examReferenceRangeTestSchema).Error)
		require.NoError(t, tx.Exec("SET LOCAL search_path TO "+examReferenceRangeTestSchema).Error)
		require.NoError(t, tx.Exec("CREATE TABLE clinics (id bigint PRIMARY KEY)").Error)
		require.NoError(t, tx.Exec("CREATE TABLE animal_species (id bigint PRIMARY KEY)").Error)

		initialRaw, readErr := os.ReadFile("../../migrations/001_init.sql")
		require.NoError(t, readErr)
		initial := string(initialRaw)
		require.NoError(t, tx.Exec(extractCreateTableDDL(t, initial, "exam_types")).Error)
		require.NoError(t, tx.Exec(extractCreateTableDDL(t, initial, "exam_type_fields")).Error)

		require.NoError(t, tx.Exec("INSERT INTO clinics (id) VALUES (1), (2)").Error)
		require.NoError(t, tx.Exec("INSERT INTO animal_species (id) VALUES (1), (2), (3)").Error)
		require.NoError(t, tx.Exec(`
			INSERT INTO exam_types (id, clinic_id, name)
			VALUES (101, 1, 'clinic A type'), (202, 2, 'clinic B type')
		`).Error)
		require.NoError(t, tx.Exec(`
			INSERT INTO exam_type_fields (id, exam_type_id, name)
			VALUES (1001, 101, 'clinic A field'), (2002, 202, 'clinic B field')
		`).Error)

		migration := readExamReferenceRangeMigration(t)
		require.NoError(t, tx.Exec(examReferenceRangeSchemaDDL(t, migration)).Error)
		fn(tx)
		return errRollbackExamReferenceRangeTest
	})
	require.ErrorIs(t, err, errRollbackExamReferenceRangeTest)
}

func TestExamReferenceRangesMigration_BackfillsClinicID(t *testing.T) {
	withExamReferenceRangeMigrationSchema(t, func(db *gorm.DB) {
		var counts struct {
			TotalCount    int64
			NullCount     int64
			MismatchCount int64
		}
		err := db.Raw(`
			SELECT
				COUNT(*) AS total_count,
				COUNT(*) FILTER (WHERE field.clinic_id IS NULL) AS null_count,
				COUNT(*) FILTER (
					WHERE field.clinic_id IS DISTINCT FROM exam_type.clinic_id
				) AS mismatch_count
			FROM exam_type_fields AS field
			JOIN exam_types AS exam_type ON exam_type.id = field.exam_type_id
		`).Scan(&counts).Error
		require.NoError(t, err)
		assert.Equal(t, int64(2), counts.TotalCount, "legacy rows must be present so the backfill check is not vacuous")
		assert.Zero(t, counts.NullCount)
		assert.Zero(t, counts.MismatchCount)
	})
}

func TestExamReferenceRangesMigration_EnforcesConstraints(t *testing.T) {
	tests := []struct {
		name     string
		wantCode string
		run      func(*gorm.DB) error
	}{
		{
			name:     "exam_type_fields INSERT cross-clinic parent",
			wantCode: "23503",
			run: func(db *gorm.DB) error {
				return db.Exec(`
					INSERT INTO exam_type_fields (id, exam_type_id, name, clinic_id)
					VALUES (1101, 101, 'cross-clinic insert', 2)
				`).Error
			},
		},
		{
			name:     "exam_type_fields UPDATE to cross-clinic parent",
			wantCode: "23503",
			run: func(db *gorm.DB) error {
				return db.Exec("UPDATE exam_type_fields SET clinic_id = 2 WHERE id = 1001").Error
			},
		},
		{
			name:     "exam_reference_ranges INSERT cross-clinic field",
			wantCode: "23503",
			run: func(db *gorm.DB) error {
				return db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id)
					VALUES (3001, 2, 1001, 1)
				`).Error
			},
		},
		{
			name:     "exam_reference_ranges UPDATE to cross-clinic field",
			wantCode: "23503",
			run: func(db *gorm.DB) error {
				if err := db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id)
					VALUES (3002, 1, 1001, 2)
				`).Error; err != nil {
					return err
				}
				return db.Exec("UPDATE exam_reference_ranges SET clinic_id = 2 WHERE id = 3002").Error
			},
		},
		{
			name:     "exam_reference_ranges duplicate clinic field species",
			wantCode: "23505",
			run: func(db *gorm.DB) error {
				if err := db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id)
					VALUES (3003, 1, 1001, 3)
				`).Error; err != nil {
					return err
				}
				return db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id)
					VALUES (3004, 1, 1001, 3)
				`).Error
			},
		},
		{
			name:     "exam_reference_ranges rejects unknown animal species",
			wantCode: "23503",
			run: func(db *gorm.DB) error {
				return db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id)
					VALUES (3005, 1, 1001, 999)
				`).Error
			},
		},
		{
			name:     "exam_reference_ranges rejects reversed numeric range",
			wantCode: "23514",
			run: func(db *gorm.DB) error {
				return db.Exec(`
					INSERT INTO exam_reference_ranges
						(id, clinic_id, exam_type_field_id, animal_species_id, ref_min, ref_max)
					VALUES (3006, 1, 1001, 1, 2, 1)
				`).Error
			},
		},
	}

	withExamReferenceRangeMigrationSchema(t, func(db *gorm.DB) {
		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				savepoint := fmt.Sprintf("exam_reference_range_case_%d", i)
				require.NoError(t, db.Exec("SAVEPOINT "+savepoint).Error)

				err := tt.run(db)
				require.Error(t, err)
				assertPostgresCode(t, err, tt.wantCode)

				require.NoError(t, db.Exec("ROLLBACK TO SAVEPOINT "+savepoint).Error)
			})
		}
	})
}

func TestExamReferenceRangesMigration_UsesDirectClinicRLS(t *testing.T) {
	migration := readExamReferenceRangeMigration(t)
	assert.Contains(t, migration, "'tenant_exam_type_fields_isolation'")
	assert.Contains(t, migration, "'tenant_exam_reference_ranges_isolation'")
	assert.Equal(t, 4, strings.Count(migration, "'app_private.has_clinic_access(clinic_id)'"))
	assert.NotContains(t, migration, "EXISTS (SELECT 1 FROM exam_types")
}

func assertPostgresCode(t *testing.T, err error, want string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, want, pgErr.Code)
}
