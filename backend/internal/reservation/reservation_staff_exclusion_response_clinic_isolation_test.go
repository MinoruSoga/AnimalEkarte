package reservation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type exclusionResponseStaffWriter struct{}

func (exclusionResponseStaffWriter) CreateForReservation(
	context.Context,
	*model.Staff,
	uint64,
) error {
	return nil
}

func (exclusionResponseStaffWriter) UpdateForReservation(
	context.Context,
	uint64,
	uint64,
	staffpkg.ReservationStaffUpdate,
) error {
	return nil
}

func (exclusionResponseStaffWriter) SwapSortOrderForReservation(
	context.Context,
	uint64,
	uint64,
	string,
) error {
	return nil
}

func TestListReservationStaffs_DoesNotExposeOtherClinicExclusionIDOrName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ReservationType{},
		&model.StaffReservationExclusion{},
		&model.StaffReservationCapability{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_reservation_capabilities, staff_reservation_exclusions, staff_clinic_assignments, reservation_types, staffs CASCADE",
	).Error)

	company := &model.Company{Name: "予約スタッフ除外レスポンステスト法人"}
	require.NoError(t, db.Create(company).Error)
	clinicA := &model.Clinic{CompanyID: company.ID, Name: "医院A", IsActive: true}
	clinicB := &model.Clinic{CompanyID: company.ID, Name: "医院B", IsActive: true}
	require.NoError(t, db.Create(clinicA).Error)
	require.NoError(t, db.Create(clinicB).Error)

	sharedStaff := &model.Staff{
		ClinicID:  clinicA.ID,
		Name:      "共有予約スタッフ",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(sharedStaff).Error)
	require.NoError(t, db.Create([]model.StaffClinicAssignment{
		{StaffID: sharedStaff.ID, ClinicID: clinicA.ID, IsMain: true},
		{StaffID: sharedStaff.ID, ClinicID: clinicB.ID, IsMain: false},
	}).Error)

	typeA := &model.ReservationType{
		ClinicID: clinicA.ID,
		Name:     "医院A専用コース",
		Category: model.ReservationTypeCategoryGeneral,
	}
	typeB := &model.ReservationType{
		ClinicID: clinicB.ID,
		Name:     "医院B秘匿コース",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.Create(typeA).Error)
	require.NoError(t, db.Create(typeB).Error)
	// Stage B: no capabilities ⇒ clinic A derived excluded = {typeA} only (universe scoped).
	// Legacy exclusion rows (if any) are ignored; do not seed them.

	repo := NewReservationStaffRepository(db, exclusionResponseStaffWriter{})
	service := NewReservationStaffService(repo, nil, nil)
	handler := NewReservationStaffHandler(service)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/reservation-staffs", http.NoBody)
	c.Set("clinic_id", strconv.FormatUint(clinicA.ID, 10))

	handler.ListReservationStaffs(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response []reservationStaffResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response, 1)
	require.Len(t, response[0].ExcludedCourses, 1)
	assert.Equal(t, typeA.ID, response[0].ExcludedCourses[0].ID)
	assert.Equal(t, typeA.Name, response[0].ExcludedCourses[0].Name)
	assert.NotEqual(t, typeB.ID, response[0].ExcludedCourses[0].ID)
	assert.NotEqual(t, typeB.Name, response[0].ExcludedCourses[0].Name)
	assert.Empty(t, response[0].CapableCourses, "empty capabilities surface as empty capable_courses")
}
