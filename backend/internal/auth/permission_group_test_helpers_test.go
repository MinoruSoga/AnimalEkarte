package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func makePermissionGroupTestClinic(t *testing.T, db *gorm.DB, name string) *model.Clinic {
	t.Helper()

	ctx := context.Background()
	company := &model.Company{Name: name + " 法人"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	clinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      name,
	}
	require.NoError(t, db.WithContext(ctx).Create(clinic).Error)
	return clinic
}

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()

	staff := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	return staff
}

func makeDoctorAssignedToClinic(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()

	staff := makeDoctor(t, db, clinicID, name)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   true,
	}).Error)
	return staff
}
