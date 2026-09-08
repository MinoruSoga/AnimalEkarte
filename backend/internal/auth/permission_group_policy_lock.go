package auth

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LockPermissionPolicy serializes permission mutations for one clinic through
// commit, including effective-permission checks and audit writes. Lock this
// before any staff, assignment or group row: locking only the changed group
// allows two transactions to each remove a different remaining admin grant.
func (r *permissionGroupRepository) LockPermissionPolicy(ctx context.Context, clinicID uint64) error {
	if clinicID == 0 {
		return apperrors.WrapInvalidInput("permission policy clinic is required")
	}
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("permission policy lock requires an ambient transaction")
	}
	if err := tx.WithContext(ctx).Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		fmt.Sprintf("permission-policy:clinic:%d", clinicID),
	).Error; err != nil {
		return apperrors.Wrap(err, "failed to acquire permission policy lock")
	}
	return nil
}
