package csvimport

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Opt-in disposable integration: MIGRATE_SQL_INTEGRATION=1 + DATABASE_URL.
// Uses a disposable migrated DB only — never shared/F6 volumes.
func TestPreflightAcceptsPost002EnteredBySingleColumnFK(t *testing.T) {
	pool := openCutoverIntegrationPool(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	seeds := cutoverIntegrationSeeds()
	ensureCutoverIntegrationSeedBaseline(t, ctx, pool, seeds)

	manifest := cutoverManifestForTargetTests()
	manifest.IDBand = CutoverIDBand{Base: 10_000_000, EndExclusive: 20_000_000}
	if err := PreflightCutoverTarget(ctx, pool, manifest, seeds); err != nil {
		t.Fatalf("PreflightCutoverTarget after 002/003 = %v, want success", err)
	}
}

func TestHistoricalRecorderSemanticsOnPost002Schema(t *testing.T) {
	pool := openCutoverIntegrationPool(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	seeds := cutoverIntegrationSeeds()
	ensureCutoverIntegrationSeedBaseline(t, ctx, pool, seeds)

	mustExec(t, ctx, pool, `INSERT INTO companies (id, name) VALUES (1, 'other co') ON CONFLICT (id) DO NOTHING`)
	mustExec(t, ctx, pool, `INSERT INTO clinics (id, company_id, name) VALUES (1, 1, 'other clinic') ON CONFLICT (id) DO NOTHING`)
	mustExec(t, ctx, pool, `
INSERT INTO staffs (id, clinic_id, name, is_active, reservation_visible)
VALUES (900001, 1, 'home-other-clinic', true, false)
ON CONFLICT (id) DO UPDATE SET clinic_id = EXCLUDED.clinic_id, is_active = true, deleted_at = NULL`)

	ownerID := seeds.ClinicID*10_000_000 + 1
	petID := seeds.ClinicID*10_000_000 + 1
	mustExec(t, ctx, pool, `
INSERT INTO owners (id, clinic_id, name) VALUES ($1, $2, 'o')
ON CONFLICT (id) DO NOTHING`, ownerID, seeds.ClinicID)
	mustExec(t, ctx, pool, `
INSERT INTO pets (id, clinic_id, owner_id, name, animal_species_id)
VALUES ($1, $2, $3, 'p', $4)
ON CONFLICT (id) DO NOTHING`, petID, seeds.ClinicID, ownerID, seeds.AnimalSpeciesID)

	doctorID := seeds.ClinicID
	mustExec(t, ctx, pool, `
INSERT INTO staffs (id, clinic_id, name, is_active, reservation_visible)
VALUES ($1, $2, 'doctor-home', true, true)
ON CONFLICT (id) DO UPDATE SET clinic_id = EXCLUDED.clinic_id, is_active = true, deleted_at = NULL`,
		doctorID, seeds.ClinicID)

	t.Run("R1_home_clinic_mismatch_entered_by_allowed", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
INSERT INTO medical_records (
  id, clinic_id, record_no, date, owner_id, pet_id, status, visit_type, doctor_id, entered_by
) VALUES (
  $1, $2, 'R1', CURRENT_DATE, $3, $4, 'draft', 'revisit', $5, 900001
)`, seeds.ClinicID*10_000_000+101, seeds.ClinicID, ownerID, petID, doctorID)
		if err != nil {
			t.Fatalf("home-clinic mismatch entered_by insert = %v, want allow", err)
		}
	})

	t.Run("R4_inactive_entered_by_row_allowed_when_present", func(t *testing.T) {
		mustExec(t, ctx, pool, `
INSERT INTO staffs (id, clinic_id, name, is_active, reservation_visible)
VALUES (900004, $1, 'inactive-recorder', false, false)
ON CONFLICT (id) DO UPDATE SET is_active = false, deleted_at = NULL`, seeds.ClinicID)
		_, err := pool.Exec(ctx, `
INSERT INTO medical_records (
  id, clinic_id, record_no, date, owner_id, pet_id, status, visit_type, doctor_id, entered_by
) VALUES (
  $1, $2, 'R4', CURRENT_DATE, $3, $4, 'draft', 'revisit', $5, 900004
)`, seeds.ClinicID*10_000_000+104, seeds.ClinicID, ownerID, petID, doctorID)
		if err != nil {
			t.Fatalf("inactive entered_by insert = %v, want allow (cutover does not re-check is_active)", err)
		}
	})

	t.Run("R3_missing_entered_by_rejected", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
INSERT INTO medical_records (
  id, clinic_id, record_no, date, owner_id, pet_id, status, visit_type, doctor_id, entered_by
) VALUES (
  $1, $2, 'R3', CURRENT_DATE, $3, $4, 'draft', 'revisit', $5, 999999001
)`, seeds.ClinicID*10_000_000+103, seeds.ClinicID, ownerID, petID, doctorID)
		if err == nil {
			t.Fatal("missing entered_by insert succeeded, want FK reject")
		}
	})

	t.Run("R5_doctor_clinic_mismatch_rejected", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
INSERT INTO medical_records (
  id, clinic_id, record_no, date, owner_id, pet_id, status, visit_type, doctor_id, entered_by
) VALUES (
  $1, $2, 'R5', CURRENT_DATE, $3, $4, 'draft', 'revisit', 900001, $5
)`, seeds.ClinicID*10_000_000+105, seeds.ClinicID, ownerID, petID, doctorID)
		if err == nil {
			t.Fatal("doctor home-other-clinic insert succeeded, want composite FK reject")
		}
	})

	t.Run("R2_no_assignment_not_enforced_by_cutover_path", func(t *testing.T) {
		// Cutover COPY/INSERT does not call AssertEnteredByActor and does not
		// import staff_clinic_assignments. HTTP create still rejects via AssertEnteredByActor.
		mustExec(t, ctx, pool, `
INSERT INTO staffs (id, clinic_id, name, is_active, reservation_visible)
VALUES (900002, 1, 'no-assignment-to-target', true, false)
ON CONFLICT (id) DO UPDATE SET clinic_id = 1, is_active = true, deleted_at = NULL`)
		_, err := pool.Exec(ctx, `
INSERT INTO medical_records (
  id, clinic_id, record_no, date, owner_id, pet_id, status, visit_type, doctor_id, entered_by
) VALUES (
  $1, $2, 'R2', CURRENT_DATE, $3, $4, 'draft', 'revisit', $5, 900002
)`, seeds.ClinicID*10_000_000+102, seeds.ClinicID, ownerID, petID, doctorID)
		if err != nil {
			t.Fatalf("cutover path insert without assignment = %v; want allow (HTTP authz is separate)", err)
		}
	})
}

func TestPreflightRejectsMissingEnteredByOrCreatedBySingleColumnFK(t *testing.T) {
	pool := openCutoverIntegrationPool(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	seeds := cutoverIntegrationSeeds()
	ensureCutoverIntegrationSeedBaseline(t, ctx, pool, seeds)
	manifest := cutoverManifestForTargetTests()
	manifest.IDBand = CutoverIDBand{Base: 10_000_000, EndExclusive: 20_000_000}

	t.Run("R6a_drop_entered_by_single_fk", func(t *testing.T) {
		mustExec(t, ctx, pool, `ALTER TABLE medical_records DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by`)
		t.Cleanup(func() {
			mustExec(t, context.Background(), pool, `
ALTER TABLE medical_records
  ADD CONSTRAINT fk_medical_records_entered_by
  FOREIGN KEY (entered_by) REFERENCES staffs (id) ON DELETE RESTRICT`)
		})
		err := PreflightCutoverTarget(ctx, pool, manifest, seeds)
		if err == nil || !strings.Contains(err.Error(), "medical_records.entered_by") {
			t.Fatalf("error = %v, want medical_records.entered_by single FK rejection", err)
		}
	})

	t.Run("R6b_drop_created_by_single_fk", func(t *testing.T) {
		mustExec(t, ctx, pool, `ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_created_by`)
		t.Cleanup(func() {
			mustExec(t, context.Background(), pool, `
ALTER TABLE appointments
  ADD CONSTRAINT fk_appointments_created_by
  FOREIGN KEY (created_by) REFERENCES staffs (id) ON DELETE RESTRICT`)
		})
		err := PreflightCutoverTarget(ctx, pool, manifest, seeds)
		if err == nil || !strings.Contains(err.Error(), "appointments.created_by") {
			t.Fatalf("error = %v, want appointments.created_by single FK rejection", err)
		}
	})

	t.Run("R6_wrong_parent_for_entered_by", func(t *testing.T) {
		mustExec(t, ctx, pool, `DELETE FROM medical_records`)
		mustExec(t, ctx, pool, `ALTER TABLE medical_records DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by`)
		mustExec(t, ctx, pool, `
ALTER TABLE medical_records
  ADD CONSTRAINT fk_medical_records_entered_by_wrong
  FOREIGN KEY (entered_by) REFERENCES clinics (id) ON DELETE RESTRICT`)
		t.Cleanup(func() {
			mustExec(t, context.Background(), pool, `ALTER TABLE medical_records DROP CONSTRAINT IF EXISTS fk_medical_records_entered_by_wrong`)
			mustExec(t, context.Background(), pool, `
ALTER TABLE medical_records
  ADD CONSTRAINT fk_medical_records_entered_by
  FOREIGN KEY (entered_by) REFERENCES staffs (id) ON DELETE RESTRICT`)
		})
		err := PreflightCutoverTarget(ctx, pool, manifest, seeds)
		if err == nil || !strings.Contains(err.Error(), "medical_records.entered_by") {
			t.Fatalf("error = %v, want entered_by parent staffs.id rejection", err)
		}
	})

	t.Run("R6_wrong_parent_for_created_by", func(t *testing.T) {
		mustExec(t, ctx, pool, `DELETE FROM appointments`)
		mustExec(t, ctx, pool, `ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_created_by`)
		mustExec(t, ctx, pool, `
ALTER TABLE appointments
  ADD CONSTRAINT fk_appointments_created_by_wrong
  FOREIGN KEY (created_by) REFERENCES clinics (id) ON DELETE RESTRICT`)
		t.Cleanup(func() {
			mustExec(t, context.Background(), pool, `ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_created_by_wrong`)
			mustExec(t, context.Background(), pool, `
ALTER TABLE appointments
  ADD CONSTRAINT fk_appointments_created_by
  FOREIGN KEY (created_by) REFERENCES staffs (id) ON DELETE RESTRICT`)
		})
		err := PreflightCutoverTarget(ctx, pool, manifest, seeds)
		if err == nil || !strings.Contains(err.Error(), "appointments.created_by") {
			t.Fatalf("error = %v, want created_by parent staffs.id rejection", err)
		}
	})
}

func cutoverIntegrationSeeds() CutoverSeedIDs {
	// Clinic insert triggers payment_methods ids 1..4; align seed IDs to that contract.
	return CutoverSeedIDs{
		ClinicID: 2, AnimalSpeciesID: 1, ExamTypeID: 11008, TrimmingReservationTypeID: 9,
		CashPaymentMethodID: 1, CreditCardPaymentMethodID: 2,
	}
}

func openCutoverIntegrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("MIGRATE_SQL_INTEGRATION") != "1" {
		t.Skip("set MIGRATE_SQL_INTEGRATION=1 for disposable cutover FK contract tests")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for cutover FK contract integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	return pool
}

func ensureCutoverIntegrationSeedBaseline(t *testing.T, ctx context.Context, pool *pgxpool.Pool, seeds CutoverSeedIDs) {
	t.Helper()
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&n); err != nil {
		t.Fatalf("schema_migrations: %v", err)
	}
	if n < 3 {
		t.Fatalf("schema_migrations count = %d, want >= 3 (001-003 applied)", n)
	}
	mustExec(t, ctx, pool, `INSERT INTO companies (id, name) VALUES ($1, 'cutover-co') ON CONFLICT (id) DO NOTHING`, seeds.ClinicID)
	mustExec(t, ctx, pool, `INSERT INTO clinics (id, company_id, name) VALUES ($1, $1, 'cutover-clinic') ON CONFLICT (id) DO NOTHING`, seeds.ClinicID)
	// Trigger may have created payment_methods 1..4; ensure cash/credit ids match seeds.
	mustExec(t, ctx, pool, `
INSERT INTO animal_species (id, name) VALUES
 (1,'犬'),(2,'猫'),(3,'鳥'),(4,'うさぎ'),(5,'ハムスター'),(6,'その他')
ON CONFLICT (id) DO NOTHING`)
	mustExec(t, ctx, pool, `
INSERT INTO exam_types (id, clinic_id, name) VALUES ($1, $2, '検査')
ON CONFLICT (id) DO NOTHING`, seeds.ExamTypeID, seeds.ClinicID)
	mustExec(t, ctx, pool, `
INSERT INTO reservation_types (id, clinic_id, name, category)
VALUES ($1, $2, 'trimming', 'trimming')
ON CONFLICT (id) DO NOTHING`, seeds.TrimmingReservationTypeID, seeds.ClinicID)
}

func mustExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		t.Fatalf("exec %s: %v", sql, err)
	}
}
