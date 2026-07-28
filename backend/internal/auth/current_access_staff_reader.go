package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type currentAccessStaffReader struct {
	db *gorm.DB
}

// NewCurrentAccessStaffReader constructs the dedicated preload-free identity
// reader used by request-time access revalidation.
func NewCurrentAccessStaffReader(db *gorm.DB) CurrentAccessStaffReader {
	return &currentAccessStaffReader{db: db}
}

func (r *currentAccessStaffReader) FindCurrentAccessStaff(
	ctx context.Context,
	staffID uint64,
) (*CurrentAccessStaffIdentity, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.WrapInternalServerError(
			"current access staff reader is not configured",
		)
	}

	var staff model.Staff
	err := r.db.WithContext(ctx).
		Select("id", "account_id", "is_active", "deleted_at").
		Where("id = ?", staffID).
		First(&staff).
		Error
	if err != nil {
		return nil, apperrors.FromGORM(
			err,
			"staff",
			fmt.Sprintf("%d", staffID),
		)
	}
	return &CurrentAccessStaffIdentity{
		ID:        staff.ID,
		AccountID: staff.AccountID,
		IsActive:  staff.IsActive,
		IsDeleted: staff.DeletedAt.Valid,
	}, nil
}
