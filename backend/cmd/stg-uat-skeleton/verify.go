// Command stg-uat-skeleton applies the STG UAT clinic skeleton (clinics 1/2
// and F6 seed bindings) without writing any F6 cutover table, including staffs.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/csvimport"
)

func verifySkeleton(ctx context.Context, e execer) error {
	if err := verifySkeletonBindings(ctx, e); err != nil {
		return err
	}
	return verifyCutoverTablesEmpty(ctx, e)
}

func verifySkeletonBindings(ctx context.Context, e execer) error {
	for _, clinic := range clinicSeeds() {
		var id, gotCompanyID int64
		var name string
		if err := e.QueryRow(ctx, `SELECT id, company_id, name FROM clinics WHERE id = $1`, clinic.id).
			Scan(&id, &gotCompanyID, &name); err != nil {
			return fmt.Errorf("clinic %d: %w", clinic.id, err)
		}
		if gotCompanyID != companyID || name != clinic.name {
			return fmt.Errorf("clinic %d identity mismatch", clinic.id)
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, "cash"},
			fmt.Sprintf("clinic %d cash payment method", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM payment_methods WHERE clinic_id = $1 AND system_key = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, "credit_card"},
			fmt.Sprintf("clinic %d credit_card payment method", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM exam_types WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, examTypeName},
			fmt.Sprintf("clinic %d exam type %s", clinic.id, examTypeName),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM reservation_types WHERE clinic_id = $1 AND category = $2 AND deleted_at IS NULL`,
			[]any{clinic.id, trimmingCategory},
			fmt.Sprintf("clinic %d trimming reservation type", clinic.id),
		); err != nil {
			return err
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM clinic_settings WHERE clinic_id = $1`,
			[]any{clinic.id},
			fmt.Sprintf("clinic %d clinic_settings", clinic.id),
		); err != nil {
			return err
		}
	}

	for _, group := range permissionGroupSeeds() {
		var gotClinic int64
		var name string
		if err := e.QueryRow(ctx,
			`SELECT clinic_id, name FROM permission_groups WHERE id = $1`,
			group.id,
		).Scan(&gotClinic, &name); err != nil {
			return fmt.Errorf("permission group %d: %w", group.id, err)
		}
		if gotClinic != group.clinicID || name != group.name {
			return fmt.Errorf("permission group %d clinic/name mismatch", group.id)
		}
		if err := requireCountAtLeast(ctx, e,
			`SELECT count(*) FROM permission_group_rules WHERE group_id = $1`,
			[]any{group.id},
			fmt.Sprintf("permission group %d rules", group.id),
		); err != nil {
			return err
		}
	}
	return nil
}

func verifyCutoverTablesEmpty(ctx context.Context, e execer) error {
	for _, name := range cutoverTableNames() {
		ident, err := quoteIdent(name)
		if err != nil {
			return err
		}
		var n int64
		if err := e.QueryRow(ctx, `SELECT count(*) FROM `+ident).Scan(&n); err != nil {
			return fmt.Errorf("count %s: %w", name, err)
		}
		if n != 0 {
			return fmt.Errorf("cutover table %s must be empty after skeleton apply, got %d rows", name, n)
		}
	}

	var staffsInBand int64
	if err := e.QueryRow(ctx,
		`SELECT count(*) FROM staffs WHERE id >= $1 AND id < $2`,
		hachiojiBandStart, hachiojiBandEnd,
	).Scan(&staffsInBand); err != nil {
		return fmt.Errorf("count staffs in hachioji band: %w", err)
	}
	if staffsInBand != 0 {
		return fmt.Errorf("staffs count in band [%d,%d) must be 0, got %d", hachiojiBandStart, hachiojiBandEnd, staffsInBand)
	}
	return nil
}

func requireCountAtLeast(ctx context.Context, e execer, sql string, args []any, label string) error {
	var n int64
	if err := e.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		return fmt.Errorf("count %s: %w", label, err)
	}
	if n < 1 {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}

func guardedExec(ctx context.Context, e execer, op writeOp) error {
	if err := rejectForbiddenWrite(op.table, op.sql); err != nil {
		return err
	}
	if !isAllowlisted(op.table) {
		return fmt.Errorf("stg-uat-skeleton refuses write to non-allowlisted table %s", op.table)
	}
	if err := e.Exec(ctx, op.sql, op.args...); err != nil {
		return fmt.Errorf("exec %s: %w", op.table, err)
	}
	return nil
}

func rejectForbiddenWrite(table, sql string) error {
	forbidden := cutoverTableSet()
	check := func(name string) error {
		name = strings.ToLower(strings.Trim(strings.TrimSpace(name), `"`))
		if name == "" {
			return nil
		}
		if _, ok := forbidden[name]; ok {
			return fmt.Errorf("stg-uat-skeleton refuses write to cutover table %s", name)
		}
		return nil
	}
	if err := check(table); err != nil {
		return err
	}
	if m := mutatingTableRe.FindStringSubmatch(sql); len(m) == 2 {
		if err := check(m[1]); err != nil {
			return err
		}
	}
	return nil
}

func assertAllowlistDisjointFromCutover() error {
	forbidden := cutoverTableSet()
	for _, name := range skeletonAllowlist {
		if _, ok := forbidden[name]; ok {
			return fmt.Errorf("skeleton allowlist includes cutover table %s", name)
		}
	}
	return nil
}

func isAllowlisted(table string) bool {
	table = strings.ToLower(strings.TrimSpace(table))
	for _, name := range skeletonAllowlist {
		if name == table {
			return true
		}
	}
	return false
}

func skeletonAllowlistTables() []string {
	return append([]string(nil), skeletonAllowlist...)
}

func cutoverTableNames() []string {
	specs := csvimport.CutoverTableSpecs()
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.Name)
	}
	return names
}

func cutoverTableSet() map[string]struct{} {
	names := cutoverTableNames()
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

func quoteIdent(name string) (string, error) {
	if !identRe.MatchString(name) {
		return "", fmt.Errorf("invalid identifier %q", name)
	}
	return `"` + name + `"`, nil
}

func (s *pgxSession) Begin(ctx context.Context) (dbTx, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTx{tx: tx}, nil
}

func (s *pgxSession) Close() {
	if s == nil || s.conn == nil {
		return
	}
	_ = s.conn.Close(context.Background())
}

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := t.tx.Exec(ctx, sql, args...)
	return err
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...any) rowScanner {
	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
