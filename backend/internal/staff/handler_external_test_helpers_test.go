package staff_test

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
	httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, _ uint64, _, _ string) bool {
		return true
	})
}

func setStaffEditorContext(c *gin.Context) {
	setClinicID(c)
	c.Set("clinic_ids", []uint64{1})
	c.Set("is_system_admin", false)
}

func setStaffID(c *gin.Context) {
	c.Set("user_id", "1")
}

type mockStaffClinicAssignmentService struct{}

func (*mockStaffClinicAssignmentService) FindAllByStaffID(
	_ context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{{StaffID: staffID, ClinicID: 1, IsMain: true}}, nil
}

func (*mockStaffClinicAssignmentService) FindByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	return &model.StaffClinicAssignment{
		StaffID:  staffID,
		ClinicID: clinicID,
		IsMain:   true,
	}, nil
}

func (*mockStaffClinicAssignmentService) Create(
	_ context.Context,
	_ *model.StaffClinicAssignment,
) error {
	return nil
}
