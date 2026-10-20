package medicalrecord

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type checkupSQLRecorder struct{ statements []string }

func (r *checkupSQLRecorder) LogMode(logger.LogLevel) logger.Interface      { return r }
func (r *checkupSQLRecorder) Info(context.Context, string, ...interface{})  {}
func (r *checkupSQLRecorder) Warn(context.Context, string, ...interface{})  {}
func (r *checkupSQLRecorder) Error(context.Context, string, ...interface{}) {}
func (r *checkupSQLRecorder) Trace(_ context.Context, _ time.Time, sql func() (string, int64), _ error) {
	text, _ := sql()
	r.statements = append(r.statements, text)
}

// SQL construction only: DryRun never verifies a database write, lock or trigger.
func TestCheckupTypeRepository_UpdateSQLPreservesCASAndInsuranceOmission(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		name := "keeps legacy null by omission"
		if explicit {
			name = "writes explicit false"
		}
		t.Run(name, func(t *testing.T) {
			recorder := &checkupSQLRecorder{}
			db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused user=unused dbname=unused sslmode=disable"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: recorder})
			require.NoError(t, err)
			pool, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, pool.Close()) })
			// Deliberately unusable fallback verifies that every operation uses the
			// ambient transaction supplied through context, rather than r.db.
			repo := &checkupTypeRepository{}
			cmd := UpdateCheckupTypeInput{Name: checkupRuntimeString("renamed"), expectedVersion: int64Ptr(7)}
			if explicit {
				cmd.IsNonInsurance = boolPtr(false)
			}
			got, updateErr := repo.Update(persistence.WithTxValue(context.Background(), db), 1, 9, cmd)
			require.Nil(t, got)
			require.True(t, apperrors.IsConflict(updateErr), "DryRun has no affected row; unexpected error: %v", updateErr)
			var update string
			for _, statement := range recorder.statements {
				if strings.HasPrefix(statement, "UPDATE ") {
					update = statement
					break
				}
			}
			require.NotEmpty(t, update, "must inspect an actual repository UPDATE")
			require.Contains(t, update, "version = 7")
			require.Contains(t, update, "clinic_id = 1")
			beforeWhere, _, ok := strings.Cut(update, " WHERE ")
			require.True(t, ok)
			require.NotContains(t, beforeWhere, "version", "the database trigger owns version increments")
			require.Equal(t, explicit, strings.Contains(beforeWhere, "is_non_insurance"))
			if explicit {
				require.Contains(t, beforeWhere, `"is_non_insurance"=false`)
			}
		})
	}
}
