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

	_, err = tx.ExecContext(ctx, `
		INSERT INTO staff_clinic_assignments (staff_id, clinic_id, is_main)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (staff_id, clinic_id) DO UPDATE
			SET deleted_at = NULL, is_main = TRUE
	`, spec.StaffID, spec.ClinicID)
	if err != nil {
		return fmt.Errorf("assign clinic for staff %d: %w", spec.StaffID, err)
	}

	var groupID uint64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM permission_groups
		 WHERE clinic_id = $1 AND name = $2 AND deleted_at IS NULL
		 ORDER BY id
		 LIMIT 1
	`, spec.ClinicID, PermissionGroupName).Scan(&groupID)
	if err != nil {
		return fmt.Errorf("permission group %q missing for clinic %d: %w", PermissionGroupName, spec.ClinicID, err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO staff_permission_groups (staff_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT (staff_id, group_id) DO NOTHING
	`, spec.StaffID, groupID)
	if err != nil {
		return fmt.Errorf("assign permission group for staff %d: %w", spec.StaffID, err)
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
