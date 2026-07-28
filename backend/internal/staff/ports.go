package staff

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// Transactor is the transaction boundary consumed by staff use cases.
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PermissionGroupRepository is the staff-owned consumer view of permission
// group membership persistence.
type PermissionGroupRepository interface {
	FindAllGroupIDsByStaffID(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
}

// ReservationStaffRepository is the staff-owned consumer view of reservation
// capability and exclusion persistence.
type ReservationStaffRepository interface {
	FindAllExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error)
	UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
}
