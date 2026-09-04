package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/seedlogin"
)

// AuthEffectivePermissions is the resource-keyed permission map returned to
// the HTTP compatibility surface.
type AuthEffectivePermissions map[string]AuthResourcePermission

// AuthResourcePermission is the effective CRUD permission for one resource.
type AuthResourcePermission struct {
	View   bool
	Create bool
	Edit   bool
	Delete bool
}

// AuthService provides login authentication and effective-permission
// calculation.
type AuthService interface {
	AuthenticateUser(ctx context.Context, email, password string) (*model.Account, *model.Staff, error)
	ResolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64)
	ResolveSystemAdminMainClinicID(mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) string
	CalculateEffectivePermissions(ctx context.Context, isSystemAdmin bool, staffID, clinicID uint64) AuthEffectivePermissions
}

// StaffAccountFinder is the auth-owned view of staff account lookup.
type StaffAccountFinder interface {
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
}

type authService struct {
	account             AccountService
	staff               StaffAccountFinder
	effectivePermission EffectivePermissionService
	comparePassword     passwordCompareFunc
}

type wrongPasswordError struct {
	accountID uint64
	err       error
}

const invalidCredentialsMessage = "メールアドレスまたはパスワードが正しくありません" //nolint:gosec // generic UI text, not a credential

func invalidCredentialsError() error {
	return apperrors.WrapUnauthorized(invalidCredentialsMessage)
}

func (e *wrongPasswordError) Error() string {
	return e.err.Error()
}

func (e *wrongPasswordError) Unwrap() error {
	return e.err
}

// IsAuthenticateWrongPassword reports whether err represents a password
// mismatch and returns the authenticated account ID for audit attribution.
func IsAuthenticateWrongPassword(err error) (uint64, bool) {
	var wrongPassword *wrongPasswordError
	if errors.As(err, &wrongPassword) {
		return wrongPassword.accountID, true
	}
	return 0, false
}

// NewAuthService constructs the login and permission use case.
func NewAuthService(
	account AccountService,
	staff StaffAccountFinder,
	effectivePermission EffectivePermissionService,
) AuthService {
	return &authService{
		account:             account,
		staff:               staff,
		effectivePermission: effectivePermission,
		comparePassword:     defaultPasswordCompare,
	}
}

func (s *authService) AuthenticateUser(
	ctx context.Context,
	email, password string,
) (*model.Account, *model.Staff, error) {
	account, err := s.account.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			performDummyPasswordComparison(s.comparePassword, password)
			slog.InfoContext(ctx, "skip audit log for login failure: account not found")
			return nil, nil, invalidCredentialsError()
		}
		return nil, nil, apperrors.Wrap(err, "failed to find account by email")
	}
	if account == nil || !account.IsActive || account.DeletedAt.Valid {
		performDummyPasswordComparison(s.comparePassword, password)
		return nil, nil, invalidCredentialsError()
	}
	if err := normalizePasswordComparer(s.comparePassword)(
		[]byte(account.PasswordHash),
		[]byte(password),
	); err != nil {
		if !seedlogin.AcceptSharedPassword(os.Getenv("APP_ENV"), email, password) {
			return nil, nil, &wrongPasswordError{
				accountID: account.ID,
				err:       invalidCredentialsError(),
			}
		}
	}
	staff, err := s.staff.FindByAccountID(ctx, account.ID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil, invalidCredentialsError()
		}
		return nil, nil, apperrors.Wrap(err, "スタッフ情報の取得に失敗しました")
	}
	if staff == nil || !staff.IsActive || staff.DeletedAt.Valid {
		return nil, nil, invalidCredentialsError()
	}
	return account, staff, nil
}

func (s *authService) ResolveClinicInfo(
	assignments []model.StaffClinicAssignment,
) (mainClinicID string, clinicIDs []uint64) {
	clinicIDs = make([]uint64, 0, len(assignments))
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.IsMain && mainClinicID == "" {
			mainClinicID = strconv.FormatUint(assignment.ClinicID, 10)
		}
		clinicIDs = append(clinicIDs, assignment.ClinicID)
	}
	if mainClinicID == "" && len(assignments) > 0 {
		mainClinicID = strconv.FormatUint(assignments[0].ClinicID, 10)
	}
	return mainClinicID, clinicIDs
}

func (s *authService) ResolveSystemAdminMainClinicID(
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) string {
	if !isSystemAdmin {
		return mainClinicID
	}
	activeClinicIDs := activeSystemAdminClinicIDs(allClinics)
	if len(activeClinicIDs) == 0 {
		return ""
	}
	return resolveSystemAdminMainClinicID(mainClinicID, activeClinicIDs)
}

func resolveSystemAdminMainClinicID(
	preferredMainClinicID string,
	activeClinicIDs []uint64,
) string {
	preferredID, err := strconv.ParseUint(preferredMainClinicID, 10, 64)
	if err == nil && preferredID != 0 {
		for _, clinicID := range activeClinicIDs {
			if clinicID == preferredID {
				return strconv.FormatUint(preferredID, 10)
			}
		}
	}
	return strconv.FormatUint(activeClinicIDs[0], 10)
}

func activeSystemAdminClinicIDs(allClinics []model.Clinic) []uint64 {
	clinicIDs := make([]uint64, 0, len(allClinics))
	seen := make(map[uint64]struct{}, len(allClinics))
	for i := range allClinics {
		clinic := &allClinics[i]
		if clinic.ID == 0 || !clinic.IsActive {
			continue
		}
		if _, duplicate := seen[clinic.ID]; duplicate {
			continue
		}
		seen[clinic.ID] = struct{}{}
		clinicIDs = append(clinicIDs, clinic.ID)
	}
	return clinicIDs
}

func buildAuthAllPermissions() AuthEffectivePermissions {
	permissions := make(AuthEffectivePermissions, len(model.AllResources))
	for _, resource := range model.AllResources {
		permissions[string(resource)] = AuthResourcePermission{
			View:   true,
			Create: true,
			Edit:   true,
			Delete: true,
		}
	}
	return permissions
}

func (s *authService) CalculateEffectivePermissions(
	ctx context.Context,
	isSystemAdmin bool,
	staffID, clinicID uint64,
) AuthEffectivePermissions {
	if isSystemAdmin {
		return buildAuthAllPermissions()
	}

	rules, err := s.effectivePermission.GetEffectivePermissions(ctx, staffID, clinicID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to get effective permissions",
			"staff_id", staffID,
			"clinic_id", clinicID,
			"error", err,
		)
		return make(AuthEffectivePermissions)
	}

	permissions := make(AuthEffectivePermissions, len(rules))
	for i := range rules {
		rule := &rules[i]
		permissions[rule.Resource] = AuthResourcePermission{
			View:   rule.CanView,
			Create: rule.CanCreate,
			Edit:   rule.CanEdit,
			Delete: rule.CanDelete,
		}
	}
	return permissions
}
