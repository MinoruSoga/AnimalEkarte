package model

import (
	"strings"
	"testing"
)

func TestCashRegisterCloseAdjustmentMigrationTenantBoundary(t *testing.T) {
	t.Parallel()

	sql := readMigrationFile(t, "../../migrations/001_init.sql")
	requiredSnippets := []string{
		"ADD CONSTRAINT uq_cash_register_closes_id_clinic\n        UNIQUE (id, clinic_id)",
		"ADD CONSTRAINT uq_staffs_id_clinic\n        UNIQUE (id, clinic_id)",
		"DROP CONSTRAINT IF EXISTS cash_register_close_adjustments_close_id_fkey",
		"CONSTRAINT fk_cash_register_close_adjustments_close_clinic\n        FOREIGN KEY (close_id, clinic_id)\n        REFERENCES cash_register_closes (id, clinic_id)\n        ON DELETE RESTRICT",
		"CONSTRAINT fk_cash_register_close_adjustments_billing_clinic\n        FOREIGN KEY (billing_id, clinic_id)\n        REFERENCES billings (id, clinic_id)\n        ON DELETE RESTRICT",
		"CONSTRAINT fk_cash_register_close_adjustments_actor_clinic\n        FOREIGN KEY (actor_id, clinic_id)\n        REFERENCES staffs (id, clinic_id)\n        ON DELETE RESTRICT",
		"SELECT app_private.apply_rls_policy(\n    'cash_register_close_adjustments'",
		"'tenant_cash_register_close_adjustments_isolation'",
		"'app_private.has_clinic_access(clinic_id)'",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("cash_register_close_adjustments migration missing tenant boundary:\n%s", snippet)
		}
	}

	forbiddenSnippets := []string{
		"billing_id           BIGINT       NOT NULL REFERENCES billings(id)",
		"actor_id             BIGINT       REFERENCES staffs(id)",
	}
	for _, snippet := range forbiddenSnippets {
		if strings.Contains(sql, snippet) {
			t.Fatalf("cash_register_close_adjustments retains a single-column tenant FK:\n%s", snippet)
		}
	}
}
