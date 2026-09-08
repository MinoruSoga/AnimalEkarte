package seedlogin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
)

// Apply upserts curated demo staffs + accounts using SharedPassword.
// The password value must not be written to logs.
func Apply(ctx context.Context, db *sql.DB) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(SharedPassword), config.BcryptCost)
	if err != nil {
		return 0, fmt.Errorf("hash demo login password: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin login seed tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	applied := 0
	for _, spec := range Catalog() {
		if err := upsertSpec(ctx, tx, spec, string(hash)); err != nil {
			return 0, err
		}
		applied++
	}
	if err := retireDuplicateHayashiLogins(ctx, tx); err != nil {
		return 0, err
	}
	if err := upsertOperatorFromEnv(ctx, tx); err != nil {
		return 0, err
	}
	if err := advanceLoginSequences(ctx, tx); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit login seed tx: %w", err)
	}
	return applied, nil
}

// CatalogChecksum is a non-secret fingerprint of the curated account set.
func CatalogChecksum() string {
	var b strings.Builder
	for _, spec := range Catalog() {
		b.WriteString(strconv.FormatUint(spec.StaffID, 10))
		b.WriteByte(',')
		b.WriteString(spec.Email)
		b.WriteByte(',')
		b.WriteString(spec.PermissionGroupName)
		b.WriteByte(',')
		if spec.AssignAllCatalogClinics {
			b.WriteString("all")
		} else {
			b.WriteString("home")
		}
		b.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func upsertSpec(ctx context.Context, tx *sql.Tx, spec AccountSpec, passwordHash string) error {
	var clinicID uint64
	err := tx.QueryRowContext(ctx, `SELECT id FROM clinics WHERE id = $1`, spec.ClinicID).
		Scan(&clinicID)
	if err != nil {
		return fmt.Errorf("login seed clinic %d missing (002_master required): %w", spec.ClinicID, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO staffs (
			id, clinic_id, name, is_active, license_number, staff_type,
			reservation_display_name, reservation_visible, reservation_comment, reservation_image_url
		) VALUES ($1, $2, $3, TRUE, '', $4, '', TRUE, '', '')
		ON CONFLICT (id) DO UPDATE
			SET is_active = TRUE, deleted_at = NULL
	`, spec.StaffID, spec.ClinicID, spec.Name, spec.StaffType)
	if err != nil {
		return fmt.Errorf("upsert staff %d: %w", spec.StaffID, err)
	}

	var accountID uint64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO accounts (email, password_hash, is_active, is_system_admin)
		VALUES ($1, $2, TRUE, FALSE)
		ON CONFLICT (email) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    is_active = TRUE,
			    deleted_at = NULL
		RETURNING id
	`, spec.Email, passwordHash).Scan(&accountID)
	if err != nil {
		return fmt.Errorf("upsert account for staff %d: %w", spec.StaffID, err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE staffs
		   SET account_id = $1, is_active = TRUE, deleted_at = NULL
		 WHERE id = $2
		   AND deleted_at IS NULL
		   AND (account_id IS NULL OR account_id = $1)
	`, accountID, spec.StaffID)
	if err != nil {
		return fmt.Errorf("link account for staff %d: %w", spec.StaffID, err)
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return fmt.Errorf("staff %d is linked to a different account", spec.StaffID)
	}

	if err := replaceClinicAssignments(ctx, tx, spec); err != nil {
		return err
	}
	if err := replacePermissionGroups(ctx, tx, spec); err != nil {
		return err
	}
	return nil
}

func replaceClinicAssignments(ctx context.Context, tx *sql.Tx, spec AccountSpec) error {
	keep := assignmentClinicIDs(spec)
	if len(keep) == 0 {
		return fmt.Errorf("login seed clinic assignments empty for staff %d", spec.StaffID)
	}
	for _, clinicID := range keep {
		var found uint64
		err := tx.QueryRowContext(ctx, `SELECT id FROM clinics WHERE id = $1`, clinicID).Scan(&found)
		if err != nil {
			return fmt.Errorf("login seed clinic %d missing (002_master required): %w", clinicID, err)
		}
	}

	placeholders := make([]string, len(keep))
	args := make([]any, 0, 1+len(keep))
	args = append(args, spec.StaffID)
	for i, clinicID := range keep {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, clinicID)
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE staff_clinic_assignments
		   SET deleted_at = NOW(), is_main = FALSE
		 WHERE staff_id = $1
		   AND deleted_at IS NULL
		   AND clinic_id NOT IN (%s)
	`, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return fmt.Errorf("retire extra clinic assignments for staff %d: %w", spec.StaffID, err)
	}

	for _, clinicID := range keep {
		isMain := clinicID == spec.ClinicID
		_, err := tx.ExecContext(ctx, `
			INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main)
			VALUES ($1, $2, $3)
			ON CONFLICT (staff_id, clinic_id) DO UPDATE
				SET deleted_at = NULL, is_main = EXCLUDED.is_main
		`, spec.StaffID, clinicID, isMain)
		if err != nil {
			return fmt.Errorf("assign clinic %d for staff %d: %w", clinicID, spec.StaffID, err)
		}
	}
	return nil
}

func replacePermissionGroups(ctx context.Context, tx *sql.Tx, spec AccountSpec) error {
	if spec.PermissionGroupName != PermissionGroupExecutive && spec.PermissionGroupName != PermissionGroupGeneral {
		return fmt.Errorf("login seed permission group %q is not allowed for staff %d", spec.PermissionGroupName, spec.StaffID)
	}
	if spec.AssignAllCatalogClinics && spec.PermissionGroupName != PermissionGroupExecutive {
		return fmt.Errorf("login seed all-clinic assignment requires %q for staff %d", PermissionGroupExecutive, spec.StaffID)
	}

	keep := assignmentClinicIDs(spec)
	groupIDs := make([]uint64, 0, len(keep))
	for _, clinicID := range keep {
		var groupID uint64
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM permission_groups
			 WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL
			 ORDER BY id
			 LIMIT 1
		`, clinicID, spec.PermissionGroupName).Scan(&groupID)
		if err != nil {
			return fmt.Errorf("permission group %q missing for clinic %d: %w", spec.PermissionGroupName, clinicID, err)
		}
		groupIDs = append(groupIDs, groupID)
	}

	_, err := tx.ExecContext(ctx, `
		DELETE FROM staff_permission_groups WHERE staff_id = $1
	`, spec.StaffID)
	if err != nil {
		return fmt.Errorf("replace permission groups for staff %d: %w", spec.StaffID, err)
	}
	for _, groupID := range groupIDs {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO staff_permission_groups (staff_id, group_id)
			VALUES ($1, $2)
			ON CONFLICT (staff_id, group_id) DO NOTHING
		`, spec.StaffID, groupID)
		if err != nil {
			return fmt.Errorf("assign permission group for staff %d: %w", spec.StaffID, err)
		}
	}
	return nil
}

func retireDuplicateHayashiLogins(ctx context.Context, tx *sql.Tx) error {
	for _, staffID := range retiredDuplicateHayashiStaffIDs() {
		email := EmailForStaffID(staffID)
		if _, err := tx.ExecContext(ctx, `
			UPDATE accounts
			   SET is_active = FALSE, deleted_at = NOW()
			 WHERE email = $1
			   AND deleted_at IS NULL
		`, email); err != nil {
			return fmt.Errorf("retire duplicate hayashi account %d: %w", staffID, err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE staffs
			   SET is_active = FALSE, deleted_at = NOW()
			 WHERE id = $1
			   AND deleted_at IS NULL
		`, staffID); err != nil {
			return fmt.Errorf("retire duplicate hayashi staff %d: %w", staffID, err)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE staff_clinic_assignments
			   SET deleted_at = NOW(), is_main = FALSE
			 WHERE staff_id = $1
			   AND deleted_at IS NULL
		`, staffID); err != nil {
			return fmt.Errorf("retire duplicate hayashi assignments %d: %w", staffID, err)
		}
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM staff_permission_groups WHERE staff_id = $1
		`, staffID); err != nil {
			return fmt.Errorf("retire duplicate hayashi groups %d: %w", staffID, err)
		}
	}
	return nil
}

func advanceLoginSequences(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`SELECT setval(pg_get_serial_sequence('staffs', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM staffs), 1), true)`,
		`SELECT setval(pg_get_serial_sequence('accounts', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM accounts), 1), true)`,
	}
	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("advance login seed sequence: %w", err)
		}
	}
	return nil
}

// LogApplied writes a non-secret summary. Do not pass the password.
func LogApplied(logger *slog.Logger, applied int) {
	if logger == nil {
		return
	}
	logger.Info("Login seed applied",
		slog.String("bundle", BundleDir),
		slog.Int("accounts", applied))
}
