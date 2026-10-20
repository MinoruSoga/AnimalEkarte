package medicalrecord

import (
	"context"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Actual repository SQL and outcome classification with callback-controlled
// results. This does not exercise a PostgreSQL trigger or real concurrency.
func TestVaccineRepository_CASSQLAndOutcomes(t *testing.T) {
	for _, scenario := range []string{"updated", "stale", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			db, err := gorm.Open(postgres.Open("host=localhost user=unused dbname=unused sslmode=disable"), &gorm.Config{DisableAutomaticPing: true, DryRun: true, SkipDefaultTransaction: true})
			require.NoError(t, err)
			conn, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })
			var updateSQL string
			var updateVars []any
			reads := 0
			require.NoError(t, db.Callback().Update().After("gorm:update").Register("capture_cas", func(tx *gorm.DB) {
				updateSQL = tx.Statement.SQL.String()
				updateVars = append([]any(nil), tx.Statement.Vars...)
				if scenario == "updated" {
					tx.RowsAffected = 1
				} else {
					tx.RowsAffected = 0
				}
			}))
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("cas_result", func(tx *gorm.DB) {
				reads++
				require.Contains(t, tx.Statement.SQL.String(), "clinic_id")
				if scenario == "missing" {
					tx.AddError(gorm.ErrRecordNotFound)
					return
				}
				row := tx.Statement.Dest.(*model.Vaccine)
				row.ID = 17
				row.ClinicID = 2
				row.Version = 8
				tx.RowsAffected = 1
			}))
			repo := &vaccineRepository{db: nil}
			ctx := persistence.WithTxValue(t.Context(), db)
			name := "edit"
			got, err := repo.Update(ctx, 2, 17, UpdateVaccineInput{Version: 7, Name: &name})
			require.Contains(t, updateSQL, "id = $")
			require.Contains(t, updateSQL, "version = $")
			require.Contains(t, updateSQL, "clinic_id")
			require.Contains(t, updateSQL, "deleted_at")
			require.Contains(t, updateVars, int64(7))
			require.Contains(t, updateVars, uint64(2))
			setClause := strings.Split(updateSQL, " WHERE ")[0]
			require.NotContains(t, setClause, "\"version\"", "database trigger owns advancement")
			require.Equal(t, 1, reads)
			switch scenario {
			case "updated":
				require.NoError(t, err)
				require.Equal(t, int64(8), got.Version)
			case "stale":
				require.Error(t, err)
				require.True(t, apperrors.IsConflict(err))
				require.Nil(t, got)
			case "missing":
				require.Error(t, err)
				require.True(t, apperrors.IsNotFound(err))
				require.Nil(t, got)
			}
		})
	}
}

func TestVaccineService_CASConflictDoesNotAuditOrReplaceExpectedVersion(t *testing.T) {
	repo := &mockVaccineRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Vaccine, error) {
			return &model.Vaccine{ID: 17, ClinicID: 2, Version: 8}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, input UpdateVaccineInput) (*model.Vaccine, error) {
			require.Equal(t, int64(7), input.Version)
			return nil, apperrors.WrapConflict("stale")
		},
	}
	audited := false
	svc := NewVaccineServiceWithAudit(repo, MasterInsuranceWriteDependencies{
		Transactor: vaccineTestTransactorFn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }),
		Audit:      vaccineTestAuditFn(func(context.Context, *AuditEntry) error { audited = true; return nil }),
	})
	name := "edit"
	got, err := svc.Update(t.Context(), 2, 17, &UpdateVaccineInput{ActorID: vaccineTestActorID(), Version: 7, Name: &name})
	require.Error(t, err)
	require.True(t, apperrors.IsConflict(err))
	require.Nil(t, got)
	require.False(t, audited)
	require.Equal(t, int64(8), toVaccineResponse(&model.Vaccine{Version: 8}).Version)
}
