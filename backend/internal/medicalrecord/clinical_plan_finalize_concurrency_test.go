package medicalrecord

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const clinicalPlanFinalizeLockTimeout = 30 * time.Second

type blockingClinicalPlanAudit struct {
	reached chan<- struct{}
	release <-chan struct{}
}

type clinicalPlanFinalizePIDResult struct {
	pid int
	err error
}

func (a blockingClinicalPlanAudit) LogEntryTx(ctx context.Context, _ *AuditEntry) error {
	close(a.reached)
	select {
	case <-a.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func waitForClinicalPlanFinalizeLock(db *gorm.DB, finalizePID int, finalizeDone <-chan error) error {
	deadline := time.NewTimer(clinicalPlanFinalizeLockTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-finalizeDone:
			if err != nil {
				return fmt.Errorf("finalize completed before waiting for the parent row lock: %w", err)
			}
			return fmt.Errorf("finalize completed before waiting for the parent row lock")
		case <-deadline.C:
			return fmt.Errorf("finalize did not reach a PostgreSQL lock wait within %s", clinicalPlanFinalizeLockTimeout)
		case <-ticker.C:
			var lockWaiters int64
			if err := db.Raw(`
				SELECT count(*)
				FROM pg_stat_activity
				WHERE pid = ?
				  AND datname = current_database()
				  AND wait_event_type = 'Lock'
				  AND query LIKE 'UPDATE "medical_records"%'
			`, finalizePID).Scan(&lockWaiters).Error; err != nil {
				return fmt.Errorf("inspect finalize lock wait: %w", err)
			}
			if lockWaiters > 0 {
				return nil
			}
		}
	}
}

func waitForClinicalPlanAudit(t *testing.T, reached <-chan struct{}) {
	t.Helper()
	select {
	case <-reached:
	case <-time.After(clinicalPlanFinalizeLockTimeout):
		t.Fatal("clinical-plan mutation did not reach its transaction-scoped audit")
	}
}

func setupClinicalPlanFinalizeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupClinicalPlanTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.StaffClinicAssignment{}))
	return db
}

// EMR-495 P1: 旧「finalize commits before child mutation」ケース（親確定後の更新/削除が Conflict）は
// status 廃止に伴い削除した。ここでは子 mutation が親行ロックを保持する間、親 header 書込が待たされる
// 直列化のみを検証する。
func TestClinicalPlanParentWriteConcurrency(t *testing.T) {
	const clinicID = uint64(95101)

	operations := []struct {
		name            string
		mutate          func(ClinicalPlanService, context.Context, uint64, uint64, *uint64) error
		assertCommitted func(*testing.T, ClinicalPlanRepository, *gorm.DB, uint64, *model.ClinicalPlan)
		assertUnchanged func(*testing.T, ClinicalPlanRepository, *gorm.DB, uint64, *model.ClinicalPlan)
	}{
		{
			name: "update",
			mutate: func(svc ClinicalPlanService, ctx context.Context, cid, recordID uint64, actorID *uint64) error {
				physicalExam := "F1 concurrent update"
				_, err := svc.Update(ctx, cid, recordID, &UpdateClinicalPlanInput{
					PhysicalExam: &physicalExam,
					ActorID:      actorID,
				})
				return err
			},
			assertCommitted: func(t *testing.T, repo ClinicalPlanRepository, _ *gorm.DB, cid uint64, plan *model.ClinicalPlan) {
				t.Helper()
				persisted, err := repo.FindByMedicalRecordID(context.Background(), cid, plan.MedicalRecordID)
				require.NoError(t, err)
				assert.Equal(t, "F1 concurrent update", persisted.PhysicalExam)
			},
			assertUnchanged: func(t *testing.T, repo ClinicalPlanRepository, _ *gorm.DB, cid uint64, plan *model.ClinicalPlan) {
				t.Helper()
				persisted, err := repo.FindByMedicalRecordID(context.Background(), cid, plan.MedicalRecordID)
				require.NoError(t, err)
				assert.Equal(t, plan.PhysicalExam, persisted.PhysicalExam)
				assert.Equal(t, plan.Version, persisted.Version)
			},
		},
		{
			name: "delete",
			mutate: func(svc ClinicalPlanService, ctx context.Context, cid, recordID uint64, actorID *uint64) error {
				return svc.Delete(ctx, cid, recordID, actorID)
			},
			assertCommitted: func(t *testing.T, _ ClinicalPlanRepository, db *gorm.DB, _ uint64, plan *model.ClinicalPlan) {
				t.Helper()
				var persisted model.ClinicalPlan
				require.NoError(t, db.Unscoped().First(&persisted, plan.ID).Error)
				assert.True(t, persisted.DeletedAt.Valid)
			},
			assertUnchanged: func(t *testing.T, repo ClinicalPlanRepository, _ *gorm.DB, cid uint64, plan *model.ClinicalPlan) {
				t.Helper()
				persisted, err := repo.FindByMedicalRecordID(context.Background(), cid, plan.MedicalRecordID)
				require.NoError(t, err)
				assert.Equal(t, plan.ID, persisted.ID)
			},
		},
	}

	for _, operation := range operations {
		t.Run(operation.name+" child mutation commits before the parent header write", func(t *testing.T) {
			db := setupClinicalPlanFinalizeTestDB(t)
			clinicalPlanRepo := NewClinicalPlanRepository(db)
			medicalRecordRepo := NewMedicalRecordRepository(db)
			record := makeClinicalPlanMedicalRecord(t, db, clinicID, "F1-CHILD-FIRST-"+operation.name)
			plan := &model.ClinicalPlan{ClinicID: clinicID, MedicalRecordID: record.ID, PhysicalExam: "before parent write"}
			require.NoError(t, clinicalPlanRepo.Create(context.Background(), plan))

			auditReached := make(chan struct{})
			releaseAudit := make(chan struct{})
			actorID := uint64(95101)
			svc := NewClinicalPlanService(
				clinicalPlanRepo,
				medicalRecordRepo,
				nil,
				nil,
				persistence.NewTransactor(db),
				blockingClinicalPlanAudit{reached: auditReached, release: releaseAudit},
			)

			childDone := make(chan error, 1)
			go func() {
				childDone <- operation.mutate(svc, context.Background(), clinicID, record.ID, &actorID)
			}()
			waitForClinicalPlanAudit(t, auditReached)

			finalizeDone := make(chan error, 1)
			finalizePID := make(chan clinicalPlanFinalizePIDResult, 1)
			go func() {
				err := db.Connection(func(finalizeDB *gorm.DB) error {
					var pid int
					if err := finalizeDB.Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						finalizePID <- clinicalPlanFinalizePIDResult{err: fmt.Errorf("read finalizer backend pid: %w", err)}
						return err
					}
					finalizePID <- clinicalPlanFinalizePIDResult{pid: pid}
					_, err := NewMedicalRecordRepository(finalizeDB).Update(
						context.Background(),
						clinicID,
						record.ID,
						medicalRecordParentWrite(),
						nil,
					)
					return err
				})
				if err != nil {
					select {
					case finalizePID <- clinicalPlanFinalizePIDResult{err: err}:
					default:
					}
				}
				finalizeDone <- err
			}()

			pidResult := <-finalizePID
			lockWaitErr := pidResult.err
			if lockWaitErr == nil {
				lockWaitErr = waitForClinicalPlanFinalizeLock(db, pidResult.pid, finalizeDone)
			}
			close(releaseAudit)
			childErr := <-childDone
			var finalizeErr error
			if lockWaitErr == nil {
				finalizeErr = <-finalizeDone
			}

			require.NoError(t, lockWaitErr)
			require.NoError(t, childErr)
			require.NoError(t, finalizeErr)
			operation.assertCommitted(t, clinicalPlanRepo, db, clinicID, plan)

			persistedRecord, err := medicalRecordRepo.FindByID(context.Background(), clinicID, record.ID)
			require.NoError(t, err)
			require.NotNil(t, persistedRecord.VisitType)
			assert.Equal(t, model.VisitTypeRevisit, *persistedRecord.VisitType)
		})
	}
}
