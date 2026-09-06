package auth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CurrentAccessStaffIdentity is the preload-free staff identity row required
// for request-time access revalidation.
type CurrentAccessStaffIdentity struct {
	ID        uint64
	AccountID *uint64
	IsActive  bool
	IsDeleted bool
}

// CurrentAccessStaffReader is intentionally not satisfied by the regular
// staff service. PO-005 may classify only this single identity-row read as a
// temporary staff lookup failure; account preloads and ancillary reads must
// never share this method signature.
type CurrentAccessStaffReader interface {
	FindCurrentAccessStaff(
		ctx context.Context,
		staffID uint64,
	) (*CurrentAccessStaffIdentity, error)
}

// CurrentAccessAccountReader resolves the account currently linked to a staff
// record.
type CurrentAccessAccountReader interface {
	GetByID(ctx context.Context, accountID uint64) (*model.Account, error)
}

// CurrentAccessAssignmentReader resolves only active clinic assignments.
type CurrentAccessAssignmentReader interface {
	FindAllByStaffID(
		ctx context.Context,
		staffID uint64,
	) ([]model.StaffClinicAssignment, error)
}

// CurrentAccessClinicReader resolves the current clinic inventory used to
// authorize every staff request. Implementations may return inactive clinics;
// the resolver filters them before producing request authority.
type CurrentAccessClinicReader interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
}

// CurrentAccessActiveClinicIDReader optionally replaces ListClinics for
// non-admin staff. It returns which of the assigned clinic IDs are currently active.
type CurrentAccessActiveClinicIDReader interface {
	ListActiveClinicIDs(ctx context.Context, ids []uint64) ([]uint64, error)
}

// CurrentAccessGinKey is the gin context key for the request's CurrentAccess.
const CurrentAccessGinKey = "current_access"

// CurrentAccess is the current authority snapshot used for one authenticated
// request. Its slices are newly allocated and safe for request-local use.
type CurrentAccess struct {
	StaffID       uint64
	AccountEpoch  int64
	IsSystemAdmin bool
	MainClinicID  string
	ClinicIDs     []uint64
}

// CurrentAccessResolver revalidates mutable access state on every request.
type CurrentAccessResolver interface {
	Resolve(ctx context.Context, staffID uint64) (*CurrentAccess, error)
}

// StaffLookupError identifies only the first staff persistence read. The
// middleware uses this stage marker to keep PO-005's narrowly documented
// temporary staff-validation continuity exception from applying to account or
// clinic-assignment reads.
type StaffLookupError struct {
	cause error
}

func (e *StaffLookupError) Error() string {
	return fmt.Sprintf("failed to resolve current staff: %v", e.cause)
}

func (e *StaffLookupError) Unwrap() error {
	return e.cause
}

type currentAccessResolver struct {
	staff       CurrentAccessStaffReader
	accounts    CurrentAccessAccountReader
	assignments CurrentAccessAssignmentReader
	clinics     CurrentAccessClinicReader
}

// NewCurrentAccessResolver constructs the legacy resolver without clinic
// inventory. Otherwise-valid requests fail closed at clinic authorization;
// new runtime composition must use NewCurrentAccessResolverWithClinics.
func NewCurrentAccessResolver(
	staff CurrentAccessStaffReader,
	accounts CurrentAccessAccountReader,
	assignments CurrentAccessAssignmentReader,
) CurrentAccessResolver {
	return &currentAccessResolver{
		staff:       staff,
		accounts:    accounts,
		assignments: assignments,
	}
}

// NewCurrentAccessResolverWithClinics constructs request-time authority with
// the active-clinic inventory required for every staff account.
func NewCurrentAccessResolverWithClinics(
	staff CurrentAccessStaffReader,
	accounts CurrentAccessAccountReader,
	assignments CurrentAccessAssignmentReader,
	clinics CurrentAccessClinicReader,
) CurrentAccessResolver {
	return &currentAccessResolver{
		staff:       staff,
		accounts:    accounts,
		assignments: assignments,
		clinics:     clinics,
	}
}

func (r *currentAccessResolver) Resolve(
	ctx context.Context,
	staffID uint64,
) (*CurrentAccess, error) {
	if staffID == 0 {
		return nil, apperrors.WrapUnauthorized("invalid staff identity")
	}
	if r.staff == nil || r.accounts == nil || r.assignments == nil {
		return nil, apperrors.WrapInternalServerError(
			"current access dependencies are not configured",
		)
	}

	var account *model.Account
	var accountEpoch int64
	var assignments []model.StaffClinicAssignment
	if loader, ok := r.staff.(currentAccessGraphLoader); ok {
		graph, err := loader.loadCurrentAccessGraph(ctx, staffID)
		if err != nil {
			return nil, &StaffLookupError{cause: err}
		}
		_, account, accountEpoch, err = r.validateCurrentAccessIdentity(staffID, &graph.Staff, graph.Account)
		if err != nil {
			return nil, err
		}
		assignments = graph.Assignments
	} else {
		_, loadedAccount, epoch, err := r.loadCurrentAccessIdentity(ctx, staffID)
		if err != nil {
			return nil, err
		}
		account = loadedAccount
		accountEpoch = epoch
		assignments, err = r.assignments.FindAllByStaffID(ctx, staffID)
		if err != nil {
			return nil, apperrors.Wrap(
				err,
				"failed to resolve current clinic assignments",
			)
		}
	}
	clinicIDs, mainClinicID, err := currentClinicAccess(staffID, assignments)
	if err != nil {
		return nil, err
	}
	clinicIDs, mainClinicID, err = r.resolveCurrentAccessClinics(
		ctx,
		account,
		clinicIDs,
		mainClinicID,
	)
	if err != nil {
		return nil, err
	}

	return &CurrentAccess{
		StaffID:       staffID,
		AccountEpoch:  accountEpoch,
		IsSystemAdmin: account.IsSystemAdmin,
		MainClinicID:  mainClinicID,
		ClinicIDs:     clinicIDs,
	}, nil
}

func (r *currentAccessResolver) loadCurrentAccessIdentity(
	ctx context.Context,
	staffID uint64,
) (*CurrentAccessStaffIdentity, *model.Account, int64, error) {
	staff, err := r.staff.FindCurrentAccessStaff(ctx, staffID)
	if err != nil {
		return nil, nil, 0, &StaffLookupError{cause: err}
	}
	if staff == nil || staff.ID != staffID || !staff.IsActive ||
		staff.IsDeleted {
		return nil, nil, 0, apperrors.WrapForbidden(
			"staff account is no longer active",
		)
	}
	if staff.AccountID == nil || *staff.AccountID == 0 {
		return nil, nil, 0, apperrors.WrapUnauthorized(
			"staff account link is unavailable",
		)
	}

	accountID := *staff.AccountID
	account, err := r.accounts.GetByID(ctx, accountID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, nil, 0, apperrors.WrapUnauthorized(
				"linked account is no longer active",
			)
		}
		return nil, nil, 0, apperrors.Wrap(err, "failed to resolve linked account")
	}
	return r.validateCurrentAccessIdentity(staffID, staff, account)
}

func (r *currentAccessResolver) validateCurrentAccessIdentity(
	staffID uint64,
	staff *CurrentAccessStaffIdentity,
	account *model.Account,
) (*CurrentAccessStaffIdentity, *model.Account, int64, error) {
	if staff == nil || staff.ID != staffID || !staff.IsActive ||
		staff.IsDeleted {
		return nil, nil, 0, apperrors.WrapForbidden(
			"staff account is no longer active",
		)
	}
	if staff.AccountID == nil || *staff.AccountID == 0 {
		return nil, nil, 0, apperrors.WrapUnauthorized(
			"staff account link is unavailable",
		)
	}

	accountID := *staff.AccountID
	if account == nil || account.ID != accountID || !account.IsActive ||
		account.DeletedAt.Valid {
		return nil, nil, 0, apperrors.WrapUnauthorized(
			"linked account is no longer active",
		)
	}
	accountEpoch := account.UpdatedAt.UnixNano()
	if accountEpoch <= 0 {
		return nil, nil, 0, apperrors.WrapUnauthorized(
			"linked account session epoch is unavailable",
		)
	}
	return staff, account, accountEpoch, nil
}

func (r *currentAccessResolver) resolveCurrentAccessClinics(
	ctx context.Context,
	account *model.Account,
	clinicIDs []uint64,
	mainClinicID string,
) ([]uint64, string, error) {
	if r.clinics == nil {
		return nil, "", apperrors.WrapInternalServerError(
			"current clinic authority is not configured",
		)
	}
	if account != nil && !account.IsSystemAdmin {
		if scoped, ok := r.clinics.(CurrentAccessActiveClinicIDReader); ok {
			activeIDs, clinicErr := scoped.ListActiveClinicIDs(ctx, clinicIDs)
			if clinicErr != nil {
				return nil, "", apperrors.Wrap(
					clinicErr,
					"failed to resolve active clinic authority",
				)
			}
			return currentStaffClinicAccessFromActiveIDs(
				mainClinicID,
				clinicIDs,
				activeIDs,
			)
		}
	}
	clinics, clinicErr := r.clinics.ListClinics(ctx)
	if clinicErr != nil {
		return nil, "", apperrors.Wrap(
			clinicErr,
			"failed to resolve active clinic authority",
		)
	}
	var err error
	if account.IsSystemAdmin {
		clinicIDs, mainClinicID, err = currentSystemAdminClinicAccess(
			mainClinicID,
			clinics,
		)
	} else {
		clinicIDs, mainClinicID, err = currentStaffClinicAccess(
			mainClinicID,
			clinicIDs,
			clinics,
		)
	}
	return clinicIDs, mainClinicID, err
}

func currentClinicAccess(
	staffID uint64,
	assignments []model.StaffClinicAssignment,
) ([]uint64, string, error) {
	clinicIDs := make([]uint64, 0, len(assignments))
	seen := make(map[uint64]struct{}, len(assignments))
	var mainClinicID string
	for i := range assignments {
		assignment := &assignments[i]
		if assignment.DeletedAt.Valid {
			continue
		}
		if assignment.StaffID != staffID || assignment.ClinicID == 0 {
			return nil, "", apperrors.WrapInternalServerError(
				"current clinic assignment is invalid",
			)
		}
		if _, duplicate := seen[assignment.ClinicID]; duplicate {
			continue
		}
		seen[assignment.ClinicID] = struct{}{}
		clinicIDs = append(clinicIDs, assignment.ClinicID)
		if assignment.IsMain && mainClinicID == "" {
			mainClinicID = strconv.FormatUint(assignment.ClinicID, 10)
		}
	}
	if mainClinicID == "" && len(clinicIDs) > 0 {
		mainClinicID = strconv.FormatUint(clinicIDs[0], 10)
	}
	return clinicIDs, mainClinicID, nil
}

func currentSystemAdminClinicAccess(
	preferredMainClinicID string,
	clinics []model.Clinic,
) ([]uint64, string, error) {
	clinicIDs := activeSystemAdminClinicIDs(clinics)
	if len(clinicIDs) == 0 {
		return nil, "", apperrors.WrapForbidden(
			"no active clinic access is available",
		)
	}
	mainClinicID := resolveSystemAdminMainClinicID(
		preferredMainClinicID,
		clinicIDs,
	)
	return clinicIDs, mainClinicID, nil
}

func currentStaffClinicAccess(
	preferredMainClinicID string,
	assignedClinicIDs []uint64,
	clinics []model.Clinic,
) ([]uint64, string, error) {
	return currentStaffClinicAccessFromActiveIDs(
		preferredMainClinicID,
		assignedClinicIDs,
		activeSystemAdminClinicIDs(clinics),
	)
}

func currentStaffClinicAccessFromActiveIDs(
	preferredMainClinicID string,
	assignedClinicIDs []uint64,
	activeClinicIDs []uint64,
) ([]uint64, string, error) {
	activeClinics := make(map[uint64]struct{}, len(activeClinicIDs))
	for _, clinicID := range activeClinicIDs {
		if clinicID == 0 {
			continue
		}
		activeClinics[clinicID] = struct{}{}
	}

	clinicIDs := make([]uint64, 0, len(assignedClinicIDs))
	for _, clinicID := range assignedClinicIDs {
		if _, active := activeClinics[clinicID]; active {
			clinicIDs = append(clinicIDs, clinicID)
		}
	}
	if len(clinicIDs) == 0 {
		return nil, "", apperrors.WrapForbidden(
			"no active clinic assignment is available",
		)
	}
	return clinicIDs, resolveSystemAdminMainClinicID(
		preferredMainClinicID,
		clinicIDs,
	), nil
}
