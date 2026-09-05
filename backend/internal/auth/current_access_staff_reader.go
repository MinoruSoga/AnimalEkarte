package auth

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type currentAccessStaffReader struct {
	db *gorm.DB
}

type currentAccessGraph struct {
	Staff       CurrentAccessStaffIdentity
	Account     *model.Account
	Assignments []model.StaffClinicAssignment
}

type currentAccessGraphLoader interface {
	loadCurrentAccessGraph(ctx context.Context, staffID uint64) (*currentAccessGraph, error)
}

type currentAccessGraphRow struct {
	StaffID          uint64         `gorm:"column:staff_id"`
	AccountID        *uint64        `gorm:"column:account_id"`
	StaffIsActive    bool           `gorm:"column:staff_is_active"`
	StaffDeletedAt   gorm.DeletedAt `gorm:"column:staff_deleted_at"`
	AccID            *uint64        `gorm:"column:acc_id"`
	AccIsActive      bool           `gorm:"column:acc_is_active"`
	AccIsSystemAdmin bool           `gorm:"column:acc_is_system_admin"`
	AccUpdatedAt     time.Time      `gorm:"column:acc_updated_at"`
	AccDeletedAt     gorm.DeletedAt `gorm:"column:acc_deleted_at"`
	AssignClinicID   *uint64        `gorm:"column:assign_clinic_id"`
	AssignIsMain     bool           `gorm:"column:assign_is_main"`
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

func (r *currentAccessStaffReader) loadCurrentAccessGraph(
	ctx context.Context,
	staffID uint64,
) (*currentAccessGraph, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.WrapInternalServerError(
			"current access staff reader is not configured",
		)
	}

	var rows []currentAccessGraphRow
	err := r.db.WithContext(ctx).
		Table("staffs").
		Select(`
			staffs.id AS staff_id,
			staffs.account_id AS account_id,
			staffs.is_active AS staff_is_active,
			staffs.deleted_at AS staff_deleted_at,
			accounts.id AS acc_id,
			accounts.is_active AS acc_is_active,
			accounts.is_system_admin AS acc_is_system_admin,
			accounts.updated_at AS acc_updated_at,
			accounts.deleted_at AS acc_deleted_at,
			staff_clinic_assignments.clinic_id AS assign_clinic_id,
			staff_clinic_assignments.is_main AS assign_is_main
		`).
		Joins("LEFT JOIN accounts ON accounts.id = staffs.account_id").
		Joins("LEFT JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.deleted_at IS NULL").
		Where("staffs.id = ?", staffID).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	if len(rows) == 0 {
		return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}

	first := rows[0]
	graph := &currentAccessGraph{
		Staff: CurrentAccessStaffIdentity{
			ID:        first.StaffID,
			AccountID: first.AccountID,
			IsActive:  first.StaffIsActive,
			IsDeleted: first.StaffDeletedAt.Valid,
		},
		Assignments: make([]model.StaffClinicAssignment, 0, len(rows)),
	}
	if first.AccID != nil && *first.AccID != 0 {
		graph.Account = &model.Account{
			ID:            *first.AccID,
			IsActive:      first.AccIsActive,
			IsSystemAdmin: first.AccIsSystemAdmin,
			UpdatedAt:     first.AccUpdatedAt,
			DeletedAt:     first.AccDeletedAt,
		}
	}
	seenClinic := make(map[uint64]struct{}, len(rows))
	for i := range rows {
		row := &rows[i]
		if row.AssignClinicID == nil || *row.AssignClinicID == 0 {
			continue
		}
		clinicID := *row.AssignClinicID
		if _, seen := seenClinic[clinicID]; seen {
			continue
		}
		seenClinic[clinicID] = struct{}{}
		graph.Assignments = append(graph.Assignments, model.StaffClinicAssignment{
			StaffID:  staffID,
			ClinicID: clinicID,
			IsMain:   row.AssignIsMain,
		})
	}
	return graph, nil
}
