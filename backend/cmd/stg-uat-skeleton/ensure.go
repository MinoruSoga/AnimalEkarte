// Command stg-uat-skeleton applies the STG UAT clinic skeleton (clinics 1/2
// and F6 seed bindings) without writing any F6 cutover table, including staffs.
package main

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/model"
)

func applySkeleton(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if err := applySkeletonBindings(ctx, e, bootstrap); err != nil {
		return err
	}
	return verifyCutoverTablesEmpty(ctx, e)
}

func applySkeletonBindings(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if err := assertAllowlistDisjointFromCutover(); err != nil {
		return err
	}
	if err := requireCompany(ctx, e, companyID); err != nil {
		return err
	}
	if err := ensureClinics(ctx, e); err != nil {
		return err
	}
	if err := ensureExamTypes(ctx, e); err != nil {
		return err
	}
	if err := ensureTrimmingReservationTypes(ctx, e); err != nil {
		return err
	}
	if err := ensureClinicSettings(ctx, e); err != nil {
		return err
	}
	if err := ensurePermissionGroups(ctx, e); err != nil {
		return err
	}
	if err := ensurePermissionGroupRules(ctx, e); err != nil {
		return err
	}
	if err := ensureBootstrapAccount(ctx, e, bootstrap); err != nil {
		return err
	}
	if err := ensureDefaultPaymentMethods(ctx, e); err != nil {
		return err
	}
	return verifySkeletonBindings(ctx, e)
}

func ensureClinics(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var exists bool
		if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM clinics WHERE id = $1)`, clinic.id).Scan(&exists); err != nil {
			return fmt.Errorf("lookup clinic %d: %w", clinic.id, err)
		}
		if exists {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "clinics",
			sql:   `INSERT INTO clinics (id, company_id, name, is_active) VALUES ($1, $2, $3, true)`,
			args:  []any{clinic.id, companyID, clinic.name},
		}); err != nil {
			return err
		}
	}
	return guardedExec(ctx, e, writeOp{
		table: "clinics",
		sql:   `SELECT setval(pg_get_serial_sequence('clinics', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM clinics), 1))`,
	})
}

func ensureExamTypes(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM exam_types WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL`,
			clinic.id, examTypeName,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d exam type: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "exam_types",
			sql:   `INSERT INTO exam_types (clinic_id, name, is_active) VALUES ($1, $2, true)`,
			args:  []any{clinic.id, examTypeName},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensureTrimmingReservationTypes(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM reservation_types WHERE clinic_id = $1 AND category = $2 AND deleted_at IS NULL`,
			clinic.id, trimmingCategory,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d trimming reservation type: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "reservation_types",
			sql:   `INSERT INTO reservation_types (clinic_id, name, category, is_active, duration_minutes) VALUES ($1, $2, $3, true, 15)`,
			args:  []any{clinic.id, trimmingTypeName, trimmingCategory},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensureClinicSettings(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var n int64
		if err := e.QueryRow(ctx,
			`SELECT count(*) FROM clinic_settings WHERE clinic_id = $1`,
			clinic.id,
		).Scan(&n); err != nil {
			return fmt.Errorf("count clinic %d settings: %w", clinic.id, err)
		}
		if n >= 1 {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "clinic_settings",
			sql:   `INSERT INTO clinic_settings (clinic_id) VALUES ($1)`,
			args:  []any{clinic.id},
		}); err != nil {
			return err
		}
	}
	return nil
}

func ensurePermissionGroups(ctx context.Context, e execer) error {
	for _, group := range permissionGroupSeeds() {
		var exists bool
		if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM permission_groups WHERE id = $1)`, group.id).Scan(&exists); err != nil {
			return fmt.Errorf("lookup permission group %d: %w", group.id, err)
		}
		if exists {
			continue
		}
		if err := guardedExec(ctx, e, writeOp{
			table: "permission_groups",
			sql:   `INSERT INTO permission_groups (id, clinic_id, name, description, is_active, sort_order) VALUES ($1, $2, $3, $4, true, $5)`,
			args:  []any{group.id, group.clinicID, group.name, group.description, group.sortOrder},
		}); err != nil {
			return err
		}
	}
	return guardedExec(ctx, e, writeOp{
		table: "permission_groups",
		sql:   `SELECT setval(pg_get_serial_sequence('permission_groups', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM permission_groups), 1))`,
	})
}

func ensurePermissionGroupRules(ctx context.Context, e execer) error {
	for _, group := range permissionGroupSeeds() {
		for _, resource := range model.AllResources {
			var n int64
			if err := e.QueryRow(ctx,
				`SELECT count(*) FROM permission_group_rules WHERE group_id = $1 AND resource = $2`,
				group.id, string(resource),
			).Scan(&n); err != nil {
				return fmt.Errorf("count permission group %d rule %s: %w", group.id, resource, err)
			}
			if n >= 1 {
				continue
			}
			canView, canCreate, canEdit, canDelete := permissionBits(resource, group.executive)
			if err := guardedExec(ctx, e, writeOp{
				table: "permission_group_rules",
				sql:   `INSERT INTO permission_group_rules (group_id, resource, can_view, can_create, can_edit, can_delete) VALUES ($1, $2, $3, $4, $5, $6)`,
				args:  []any{group.id, string(resource), canView, canCreate, canEdit, canDelete},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureBootstrapAccount(ctx context.Context, e execer, bootstrap bootstrapOpts) error {
	if (bootstrap.email == "") != (bootstrap.passwordHash == "") {
		return fmt.Errorf("bootstrap account requires both email and password hash")
	}
	if bootstrap.email == "" {
		return nil
	}
	var n int64
	if err := e.QueryRow(ctx,
		`SELECT count(*) FROM accounts WHERE email = $1`,
		bootstrap.email,
	).Scan(&n); err != nil {
		return fmt.Errorf("count bootstrap account: %w", err)
	}
	if n >= 1 {
		return nil
	}
	return guardedExec(ctx, e, writeOp{
		table: "accounts",
		sql:   `INSERT INTO accounts (email, password_hash, is_active, is_system_admin) VALUES ($1, $2, true, false)`,
		args:  []any{bootstrap.email, bootstrap.passwordHash},
	})
}

func clinicSeeds() []clinicSeed {
	return []clinicSeed{
		{id: clinicHachiojiID, name: clinicHachiojiName},
		{id: clinicJoutoID, name: clinicJoutoName},
	}
}

func permissionGroupSeeds() []permissionGroupSeed {
	return []permissionGroupSeed{
		{id: 1, clinicID: clinicHachiojiID, name: "執行", description: "執行権限", sortOrder: 1, executive: true},
		{id: 2, clinicID: clinicHachiojiID, name: "一般", description: "一般スタッフ権限", sortOrder: 2, executive: false},
		{id: 3, clinicID: clinicJoutoID, name: "執行", description: "執行権限", sortOrder: 1, executive: true},
		{id: 4, clinicID: clinicJoutoID, name: "一般", description: "一般スタッフ権限", sortOrder: 2, executive: false},
	}
}

func permissionBits(resource model.Resource, executive bool) (canView, canCreate, canEdit, canDelete bool) {
	for _, r := range skeletonPermissionRules {
		if r.resource != resource {
			continue
		}
		if executive {
			return r.execView, r.execCreate, r.execEdit, r.execDelete
		}
		return r.genView, r.genCreate, r.genEdit, r.genDelete
	}
	return false, false, false, false
}

func ensureDefaultPaymentMethods(ctx context.Context, e execer) error {
	// trg_create_default_payment_methods already inserts cash/credit_card on
	// clinic INSERT. Only fill F6-required keys when the trigger did not run.
	required := []struct {
		name       string
		systemKey  string
		displayOrd int
	}{
		{"現金", "cash", 1},
		{"クレジットカード", "credit_card", 2},
	}
	for _, clinic := range clinicSeeds() {
		for _, pm := range required {
			var n int64
			if err := e.QueryRow(ctx,
				`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
				clinic.id, pm.systemKey,
			).Scan(&n); err != nil {
				return fmt.Errorf("count clinic %d payment method %s: %w", clinic.id, pm.systemKey, err)
			}
			if n >= 1 {
				continue
			}
			if err := guardedExec(ctx, e, writeOp{
				table: "payment_methods",
				sql:   `INSERT INTO payment_methods (clinic_id, name, system_key, display_order, is_active) VALUES ($1, $2, $3, $4, true)`,
				args:  []any{clinic.id, pm.name, pm.systemKey, pm.displayOrd},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func requireCompany(ctx context.Context, e execer, id int64) error {
	var exists bool
	if err := e.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM companies WHERE id = $1)`, id).Scan(&exists); err != nil {
		return fmt.Errorf("lookup company %d: %w", id, err)
	}
	if !exists {
		return fmt.Errorf("company id=%d is required", id)
	}
	return nil
}
