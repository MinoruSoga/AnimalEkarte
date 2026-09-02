package model

import (
	"strings"
	"testing"
)

func TestTenantGraphFKMigrationsPinRequiredSnippets(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")
	required := []string{
		"ADD CONSTRAINT uq_reservation_types_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_cages_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_line_customers_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_consultations_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_procedures_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_inventory_items_id_clinic UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT fk_appointments_owner_clinic",
		"ADD CONSTRAINT fk_appointments_pet_clinic",
		"ADD CONSTRAINT fk_appointments_reservation_type_clinic",
		"ADD CONSTRAINT fk_appointments_doctor_clinic",
		"ADD CONSTRAINT fk_appointments_created_by_clinic",
		"ADD CONSTRAINT fk_appointments_line_customer_clinic",
		"ADD CONSTRAINT fk_hospitalizations_owner_clinic",
		"ADD CONSTRAINT fk_hospitalizations_pet_clinic",
		"ADD CONSTRAINT fk_hospitalizations_cage_clinic",
		"ADD CONSTRAINT fk_hospitalizations_doctor_clinic",
		"ADD CONSTRAINT fk_prescriptions_owner_clinic",
		"ADD CONSTRAINT fk_prescriptions_pet_clinic",
		"ADD CONSTRAINT fk_prescriptions_medical_record_clinic",
		"FOREIGN KEY (clinic_id, owner_id)\n    REFERENCES owners (clinic_id, id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (clinic_id, pet_id)\n    REFERENCES pets (clinic_id, id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (reservation_type_id, clinic_id)\n    REFERENCES reservation_types (id, clinic_id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (doctor_id, clinic_id)\n    REFERENCES staffs (id, clinic_id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (created_by, clinic_id)\n    REFERENCES staffs (id, clinic_id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (line_customer_id, clinic_id)\n    REFERENCES line_customers (id, clinic_id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (cage_id, clinic_id)\n    REFERENCES cages (id, clinic_id)\n    ON DELETE RESTRICT",
		"FOREIGN KEY (medical_record_id, clinic_id)\n    REFERENCES medical_records (id, clinic_id)\n    ON DELETE RESTRICT",
		"CREATE UNIQUE INDEX uq_medical_records_clinic_appointment_active",
		"ON medical_records (clinic_id, appointment_id)",
		"WHERE appointment_id IS NOT NULL AND deleted_at IS NULL",
		"COMMENT ON INDEX uq_medical_records_clinic_appointment_active",
		"CREATE EXTENSION IF NOT EXISTS btree_gist",
		"ADD CONSTRAINT excl_appointments_doctor_timerange",
		"EXCLUDE USING gist (",
		"clinic_id WITH =,",
		"doctor_id WITH =,",
		"tstzrange(start_time, end_time, '[)') WITH &&",
		"WHERE (deleted_at IS NULL AND status <> 'cancelled' AND doctor_id IS NOT NULL)",
		"COMMENT ON CONSTRAINT excl_appointments_doctor_timerange ON appointments",
		"SELECT app_private.apply_rls_policy(\n    'medicine_dose_params',\n    'tenant_medicine_dose_params_isolation',\n    'app_private.has_clinic_access(clinic_id)',\n    'app_private.has_clinic_access(clinic_id)'\n)",
		"ADD CONSTRAINT fk_medical_record_image_upload_quota_clinic",
		"FOREIGN KEY (clinic_id)\n    REFERENCES clinics (id)\n    ON DELETE RESTRICT",
		"ADD CONSTRAINT fk_medical_record_image_upload_quota_staff_clinic",
		"FOREIGN KEY (staff_id, clinic_id)\n    REFERENCES staffs (id, clinic_id)\n    ON DELETE RESTRICT",
		"SELECT app_private.apply_rls_policy(\n    'medical_record_image_upload_quota',\n    'tenant_medical_record_image_upload_quota_isolation',\n    'app_private.has_clinic_access(clinic_id)',\n    'app_private.has_clinic_access(clinic_id)'\n)",
		"ALTER TABLE treatments\n    ADD COLUMN clinic_id bigint NOT NULL",
		"ADD CONSTRAINT fk_treatments_clinic_id",
		"FOREIGN KEY (clinic_id)\n    REFERENCES clinics (id)\n    ON DELETE RESTRICT",
		"CREATE INDEX idx_treatments_clinic_id",
		"ADD CONSTRAINT fk_treatments_medicine_clinic",
		"ON DELETE SET NULL (medicine_id)",
		"ADD CONSTRAINT fk_treatments_consultation_clinic",
		"ON DELETE SET NULL (consultation_id)",
		"ADD CONSTRAINT fk_treatments_procedure_clinic",
		"ON DELETE SET NULL (procedure_id)",
		"ADD CONSTRAINT fk_treatments_inventory_clinic",
		"ON DELETE SET NULL (inventory_id)",
		"SELECT app_private.apply_rls_policy(\n    'treatments',\n    'tenant_treatments_isolation',\n    'app_private.has_clinic_access(clinic_id)',\n    'app_private.has_clinic_access(clinic_id)'\n)",
		"EXECUTE FUNCTION app_private.sync_treatments_clinic_id()",
		"ALTER TABLE billing_items\n    ALTER COLUMN clinic_id SET NOT NULL",
		"CREATE INDEX idx_billing_items_clinic_id",
		"ADD CONSTRAINT chk_billing_items_provenance_exclusive",
		"CHECK (num_nonnulls(vaccination_id, exam_id) <= 1)",
		"SELECT app_private.apply_rls_policy(\n    'billing_items',\n    'tenant_billing_items_isolation',\n    'app_private.has_clinic_access(clinic_id)',\n    'app_private.has_clinic_access(clinic_id)'\n)",
		"EXECUTE FUNCTION app_private.sync_billing_items_clinic_id()",
		"全会計明細の tenant scope。親 billings.clinic_id からトリガーが常に複製する",
	}

	for _, snippet := range required {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("001_init.sql missing required tenant-graph snippet:\n%s", snippet)
		}
	}
}

func TestTenantGraphFKMigrationsKeepExistingSingleColumnFKsAndRejectCascade(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")

	if strings.Contains(sql, "DROP INDEX") {
		for _, line := range strings.Split(sql, "\n") {
			if strings.Contains(line, "DROP INDEX") && strings.Contains(line, "uk_appointment_staff_time") {
				t.Fatal("001 must keep uk_appointment_staff_time; do not drop it")
			}
		}
	}
	if strings.Contains(sql, "fk_treatments_medical_record") {
		t.Fatal("001 must not add a treatments→medical_records composite FK")
	}
	if strings.Contains(strings.ToLower(sql), "execute procedure") {
		t.Fatal("001 must use EXECUTE FUNCTION (PG18), not EXECUTE PROCEDURE")
	}
	if strings.Contains(sql, "DROP CONSTRAINT fk_billing_items_billing_clinic") {
		t.Fatal("001 must keep fk_billing_items_billing_clinic")
	}
	if strings.Contains(sql, "ADD CONSTRAINT chk_billing_items_provenance_clinic_pair") {
		t.Fatal("001 must use chk_billing_items_provenance_exclusive, not provenance_clinic_pair")
	}

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if trimmed == "begin;" || trimmed == "commit;" {
			t.Fatal("001_init.sql must not contain a BEGIN;/COMMIT; statement")
		}
	}
}

func TestTenantGraphHardeningLivesInInitOnly(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")
	if strings.Contains(sql, "-- 16. 増分マイグレーション統合") {
		t.Fatal("do not archive hardening as a numbered incremental section; fold it into 001")
	}
}
