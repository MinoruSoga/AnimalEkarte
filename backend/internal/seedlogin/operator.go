package seedlogin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	operatorEnvEmail = "SEEDLOGIN_OPERATOR_EMAIL"
	operatorEnvName  = "SEEDLOGIN_OPERATOR_NAME"
	// Env var name only — the secret value is never stored in git.
	operatorEnvPassword = "SEEDLOGIN_OPERATOR_PASSWORD" //nolint:gosec // G101: identifier, not a credential
)

type operatorSpec struct {
	Email    string
	Name     string
	Password string
}

func operatorFromEnv() (operatorSpec, bool, error) {
	email := strings.ToLower(strings.TrimSpace(os.Getenv(operatorEnvEmail)))
	name := strings.TrimSpace(os.Getenv(operatorEnvName))
	password := os.Getenv(operatorEnvPassword)
	if email == "" && name == "" && password == "" {
		return operatorSpec{}, false, nil
	}
	if email == "" || name == "" || password == "" {
		return operatorSpec{}, false, fmt.Errorf(
			"operator bootstrap requires %s, %s, and %s",
			operatorEnvEmail, operatorEnvName, operatorEnvPassword,
		)
	}
	if !strings.Contains(email, "@") {
		return operatorSpec{}, false, fmt.Errorf("operator bootstrap email is invalid")
	}
	if IsCatalogEmail(email) {
		return operatorSpec{}, false, fmt.Errorf("operator bootstrap email collides with demo catalog")
	}
	if err := validateOperatorPassword(password); err != nil {
		return operatorSpec{}, false, err
	}
	return operatorSpec{Email: email, Name: name, Password: password}, true, nil
}

func validateOperatorPassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return fmt.Errorf("operator bootstrap password must be at least 8 characters")
	}
	if len([]byte(password)) > 72 {
		return fmt.Errorf("operator bootstrap password must be at most 72 bytes")
	}
	var hasLetter bool
	var hasDigit bool
	for _, character := range password {
		hasLetter = hasLetter || unicode.IsLetter(character)
		hasDigit = hasDigit || unicode.IsDigit(character)
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("operator bootstrap password must contain letters and digits")
	}
	return nil
}

func upsertOperatorFromEnv(ctx context.Context, tx *sql.Tx) error {
	spec, ok, err := operatorFromEnv()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	return upsertOperator(ctx, tx, spec)
}

func upsertOperator(ctx context.Context, tx *sql.Tx, spec operatorSpec) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(spec.Password), config.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash operator login password: %w", err)
	}

	var accountID uint64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO accounts (email, password_hash, is_active, is_system_admin)
		VALUES ($1, $2, TRUE, TRUE)
		ON CONFLICT (email) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    is_active = TRUE,
			    is_system_admin = TRUE,
			    deleted_at = NULL
		RETURNING id
	`, spec.Email, string(hash)).Scan(&accountID)
	if err != nil {
		return fmt.Errorf("upsert operator account: %w", err)
	}

	clinics := catalogClinicIDs()
	if len(clinics) == 0 {
		return fmt.Errorf("operator bootstrap catalog clinics are empty")
	}
	homeClinicID := clinics[0]

	var staffID uint64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM staffs WHERE account_id = $1 ORDER BY id LIMIT 1
	`, accountID).Scan(&staffID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		err = tx.QueryRowContext(ctx, `
			INSERT INTO staffs (
				clinic_id, name, is_active, license_number, staff_type,
				reservation_display_name, reservation_visible, reservation_comment, reservation_image_url,
				account_id
			) VALUES ($1, $2, TRUE, '', $3, '', FALSE, '', '', $4)
			RETURNING id
		`, homeClinicID, spec.Name, model.StaffTypeNurse, accountID).Scan(&staffID)
		if err != nil {
			return fmt.Errorf("insert operator staff: %w", err)
		}
	case err != nil:
		return fmt.Errorf("lookup operator staff: %w", err)
	default:
		_, err = tx.ExecContext(ctx, `
			UPDATE staffs
			   SET name = $1,
			       clinic_id = $2,
			       is_active = TRUE,
			       deleted_at = NULL,
			       account_id = $3,
			       reservation_visible = FALSE
			 WHERE id = $4
		`, spec.Name, homeClinicID, accountID, staffID)
		if err != nil {
			return fmt.Errorf("update operator staff %d: %w", staffID, err)
		}
	}

	assignmentSpec := AccountSpec{
		StaffID:                 staffID,
		ClinicID:                homeClinicID,
		Name:                    spec.Name,
		StaffType:               model.StaffTypeNurse,
		Email:                   spec.Email,
		PermissionGroupName:     PermissionGroupExecutive,
		AssignAllCatalogClinics: true,
	}
	if err := replaceClinicAssignments(ctx, tx, assignmentSpec); err != nil {
		return err
	}
	if err := replacePermissionGroups(ctx, tx, assignmentSpec); err != nil {
		return err
	}
	return nil
}
