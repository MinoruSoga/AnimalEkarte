package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/staffclinicassignment"
)

// StaffClinicAssignmentRepository is a stable facade alias for the
// staffclinicassignment domain package (BE8-4). Service/handler imports keep
// using repository.* so the split does not churn all importers.
type StaffClinicAssignmentRepository = staffclinicassignment.Repository

// NewStaffClinicAssignmentRepository constructs the staff clinic assignment repository.
func NewStaffClinicAssignmentRepository(db *gorm.DB) StaffClinicAssignmentRepository {
	return staffclinicassignment.New(db)
}
