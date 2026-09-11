package staff_test

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 staff package (12 routes)
//
// Proves clinic-fixed staff/occupation/shift list/detail isolation through real
// repository + service + HTTP handler paths. Offline `go test -short` SKIPs via
// testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBStaffSeedNameA       = "D3staff-realdb-staff-A"
	realDBStaffSeedNameB       = "D3staff-realdb-staff-B"
	realDBOccupationSeedNameA  = "D3staff-realdb-occupation-A"
	realDBOccupationSeedNameB  = "D3staff-realdb-occupation-B"
	realDBShiftTemplateSeedA   = "D3staff-realdb-shift-template-A"
	realDBShiftTemplateSeedB   = "D3staff-realdb-shift-template-B"
	realDBPermissionGroupSeedB = "D3staff-realdb-permission-group-B"
	realDBShiftYearMonth       = "2026-09"
	realDBOnDutyDate           = "2026-09-15"
)

type staffNameDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type occupationDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Name     string `json:"name"`
}

type shiftTemplateDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Name     string `json:"name"`
}

type shiftDTO struct {
	ClinicID  string `json:"clinic_id"`
	StaffID   string `json:"staff_id"`
	StaffName string `json:"staff_name"`
}

type groupIDsDTO struct {
	GroupIDs []uint64 `json:"group_ids"`
}

type clinicIDsDTO struct {
	ClinicIDs []uint64 `json:"clinic_ids"`
}

type reservationTypeIDsDTO struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids"`
}

type staffQueryGuardService struct {
	staffdomain.Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *staffQueryGuardService) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("staff List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.List(ctx, clinicID, page, limit)
}

func (s *staffQueryGuardService) GetByIDInClinic(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetByIDInClinic must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetByIDInClinic(ctx, clinicID, id)
}

func (s *staffQueryGuardService) GetPermissionGroupIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetPermissionGroupIDs must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetPermissionGroupIDs(ctx, clinicID, staffID)
}

func (s *staffQueryGuardService) GetCapableReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetCapableReservationTypeIDs must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetCapableReservationTypeIDs(ctx, clinicID, staffID)
}

func (s *staffQueryGuardService) GetExcludedReservationTypeIDs(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetExcludedReservationTypeIDs must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetExcludedReservationTypeIDs(ctx, clinicID, staffID)
}

type occupationQueryGuardService struct {
	staffdomain.OccupationService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *occupationQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("occupation List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.OccupationService.List(ctx, clinicID)
}

func (s *occupationQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("occupation GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.OccupationService.GetByID(ctx, clinicID, id)
}

type shiftEntryQueryGuardService struct {
	staffdomain.ShiftEntryService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *shiftEntryQueryGuardService) List(ctx context.Context, clinicID uint64, yearMonth string, staffID *uint64) ([]model.ShiftEntry, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("shift List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ShiftEntryService.List(ctx, clinicID, yearMonth, staffID)
}

func (s *shiftEntryQueryGuardService) GetOnDutyStaffs(ctx context.Context, clinicID uint64, date time.Time) ([]model.Staff, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetOnDutyStaffs must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ShiftEntryService.GetOnDutyStaffs(ctx, clinicID, date)
}

type shiftTemplateQueryGuardService struct {
	staffdomain.ShiftTemplateService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *shiftTemplateQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("shift-template List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ShiftTemplateService.List(ctx, clinicID)
}

func (s *shiftTemplateQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("shift-template GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ShiftTemplateService.GetByID(ctx, clinicID, id)
}

type assignmentQueryGuardService struct {
	staffdomain.StaffClinicAssignmentService
	t     *testing.T
	armed bool
}

func (s *assignmentQueryGuardService) FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	if s.armed {
		s.t.Fatalf("FindAllByStaffID must not run without selected-clinic grant (staffID=%d)", staffID)
	}
	return s.StaffClinicAssignmentService.FindAllByStaffID(ctx, staffID)
}

type realDBStaffFixture struct {
	fx            testdb.ClinicGrantFixture
	staffA        *model.Staff
	staffB        *model.Staff
	occupationA   *model.Occupation
	occupationB   *model.Occupation
	templateA     *model.ShiftTemplate
	templateB     *model.ShiftTemplate
	groupB        *model.PermissionGroup
	typeCapableB  *model.ReservationType
	typeExcludedB *model.ReservationType
	typeA         *model.ReservationType
	handler       *staffdomain.Handler
}

func setupRealDBStaffIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Occupation{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
		&model.ReservationType{},
		&model.StaffReservationCapability{},
		&model.StaffReservationExclusion{},
		&model.ShiftTemplate{},
		&model.ShiftTemplateBreak{},
		&model.ShiftEntry{},
		&model.ShiftEntryBreak{},
	))
	testdb.EnsureStaffPermissionGroupsCreatedAt(t, db)
	testdb.Truncate(t, db,
		"staff_reservation_capabilities",
		"staff_reservation_exclusions",
		"staff_permission_groups",
		"permission_group_rules",
		"permission_groups",
		"shift_entry_breaks",
		"shift_entries",
		"shift_template_breaks",
		"shift_templates",
		"staff_clinic_assignments",
		"staffs",
		"occupations",
		"reservation_types",
	)
	return db
}

func newRealDBStaffHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *staffdomain.Handler {
	t.Helper()
	staffRepo := staffdomain.NewRepository(db)
	assignRepo := staffdomain.NewStaffClinicAssignmentRepository(db)
	occRepo := staffdomain.NewOccupationRepository(db)
	shiftRepo := staffdomain.NewShiftEntryRepository(db)
	tplRepo := staffdomain.NewShiftTemplateRepository(db)
	pgRepo := auth.NewPermissionGroupRepository(db)
	resStaffRepo := reservation.NewReservationStaffRepository(db, staffRepo)

	staffSvc := staffdomain.Service(staffdomain.NewService(
		staffRepo, nil, assignRepo, nil, shiftRepo, pgRepo, resStaffRepo, occRepo, nil, nil,
	))
	occSvc := staffdomain.OccupationService(staffdomain.NewOccupationService(occRepo))
	shiftSvc := staffdomain.ShiftEntryService(staffdomain.NewShiftEntryService(shiftRepo, staffRepo, assignRepo, nil))
	tplSvc := staffdomain.ShiftTemplateService(staffdomain.NewShiftTemplateService(tplRepo))
	assignSvc := staffdomain.StaffClinicAssignmentService(staffdomain.NewStaffClinicAssignmentService(assignRepo))

	if forbiddenClinicID != 0 {
		staffSvc = &staffQueryGuardService{Service: staffSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		occSvc = &occupationQueryGuardService{OccupationService: occSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		shiftSvc = &shiftEntryQueryGuardService{ShiftEntryService: shiftSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		tplSvc = &shiftTemplateQueryGuardService{ShiftTemplateService: tplSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		assignSvc = &assignmentQueryGuardService{StaffClinicAssignmentService: assignSvc, t: t, armed: true}
	}
	return staffdomain.NewHandler(staffSvc, assignSvc, occSvc, shiftSvc, tplSvc, nil)
}

func seedClinicStaff(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	staff := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicID, IsMain: true,
	}).Error)
	return staff
}

func seedRealDBStaffFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBStaffFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3staff realDB")
	ctx := context.Background()

	occupationA := &model.Occupation{ClinicID: fx.ClinicA, Name: realDBOccupationSeedNameA, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(occupationA).Error)
	occupationB := &model.Occupation{ClinicID: fx.ClinicB, Name: realDBOccupationSeedNameB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(occupationB).Error)

	staffA := seedClinicStaff(t, db, fx.ClinicA, realDBStaffSeedNameA)
	staffB := seedClinicStaff(t, db, fx.ClinicB, realDBStaffSeedNameB)

	// Omit StartTime/EndTime: AutoMigrate doubles time columns as timestamptz and
	// rejects "HH:MM" literals. Nil times are enough for isolation observations.
	templateA := &model.ShiftTemplate{
		ClinicID: fx.ClinicA, Name: realDBShiftTemplateSeedA, ShiftType: model.ShiftTypeFull, IsActive: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(templateA).Error)
	templateB := &model.ShiftTemplate{
		ClinicID: fx.ClinicB, Name: realDBShiftTemplateSeedB, ShiftType: model.ShiftTypeFull, IsActive: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(templateB).Error)

	onDutyDay, err := time.ParseInLocation(time.DateOnly, realDBOnDutyDate, time.Local)
	require.NoError(t, err)
	require.NoError(t, db.WithContext(ctx).Create(&model.ShiftEntry{
		ClinicID: fx.ClinicA, StaffID: staffA.ID, Date: onDutyDay, ShiftType: model.ShiftTypeFull,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ShiftEntry{
		ClinicID: fx.ClinicB, StaffID: staffB.ID, Date: onDutyDay, ShiftType: model.ShiftTypeFull,
	}).Error)

	groupB := &model.PermissionGroup{ClinicID: fx.ClinicB, Name: realDBPermissionGroupSeedB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(groupB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staffB.ID, GroupID: groupB.ID}).Error)

	typeA := &model.ReservationType{ClinicID: fx.ClinicA, Name: "D3staff-realdb-type-A", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(typeA).Error)
	typeCapableB := &model.ReservationType{ClinicID: fx.ClinicB, Name: "D3staff-realdb-type-capable-B", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(typeCapableB).Error)
	typeExcludedB := &model.ReservationType{ClinicID: fx.ClinicB, Name: "D3staff-realdb-type-excluded-B", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(typeExcludedB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffReservationCapability{ClinicID: fx.ClinicB, StaffID: staffB.ID, ReservationTypeID: typeCapableB.ID}).Error)

	return realDBStaffFixture{
		fx: fx, staffA: staffA, staffB: staffB, occupationA: occupationA, occupationB: occupationB,
		templateA: templateA, templateB: templateB, groupB: groupB,
		typeCapableB: typeCapableB, typeExcludedB: typeExcludedB, typeA: typeA,
		handler: newRealDBStaffHandler(t, db, forbiddenClinicID),
	}
}

func configureStaffGrant(fx realDBStaffFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}

func withIDParam(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}
}

// TestRealDB_SelectedClinicBGrantAIsolation covers all 12 clinic-fixed staff routes.
func TestRealDB_SelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("staffs_list_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/staffs", configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA))
		fx.handler.ListStaffs(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBStaffSeedNameA, realDBStaffSeedNameB)
	})

	t.Run("staffs_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/staffs", configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB))
		fx.handler.ListStaffs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []staffNameDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.NotEqual(t, realDBStaffSeedNameA, item.Name)
			if item.ID == fx.staffB.ID {
				foundB = true
				assert.Equal(t, realDBStaffSeedNameB, item.Name)
			}
		}
		require.True(t, foundB)
	})

	t.Run("staffs_get_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.staffB.ID))
		fx.handler.GetStaff(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBStaffSeedNameA, realDBStaffSeedNameB)
	})

	t.Run("staffs_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffB.ID))
		fx.handler.GetStaff(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBStaffSeedNameB)
		assert.NotContains(t, w.Body.String(), realDBStaffSeedNameA)
	})

	t.Run("staffs_get_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d", fx.staffA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffA.ID))
		fx.handler.GetStaff(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, realDBStaffSeedNameA, realDBStaffSeedNameB)
	})

	t.Run("occupations_list_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/occupations", configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA))
		fx.handler.ListOccupations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBOccupationSeedNameA, realDBOccupationSeedNameB)
	})

	t.Run("occupations_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/occupations", configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB))
		fx.handler.ListOccupations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []occupationDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBOccupationSeedNameA, item.Name)
			if item.ID == fx.occupationB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("occupations_get_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/occupations/%d", fx.occupationB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.occupationB.ID))
		fx.handler.GetOccupation(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("occupations_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/occupations/%d", fx.occupationB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.occupationB.ID))
		fx.handler.GetOccupation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBOccupationSeedNameB)
		assert.NotContains(t, w.Body.String(), realDBOccupationSeedNameA)
	})

	t.Run("occupations_get_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/occupations/%d", fx.occupationA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.occupationA.ID))
		fx.handler.GetOccupation(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBOccupationSeedNameA)
	})

	t.Run("permission_groups_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/permission-groups", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.staffB.ID))
		fx.handler.GetStaffPermissionGroups(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("permission_groups_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/permission-groups", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffB.ID))
		fx.handler.GetStaffPermissionGroups(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got groupIDsDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Contains(t, got.GroupIDs, fx.groupB.ID)
	})

	t.Run("permission_groups_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/permission-groups", fx.staffA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffA.ID))
		fx.handler.GetStaffPermissionGroups(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("clinics_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/clinics", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.staffB.ID))
		fx.handler.GetStaffClinicAssignments(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("clinics_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/clinics", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffB.ID))
		fx.handler.GetStaffClinicAssignments(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got clinicIDsDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Contains(t, got.ClinicIDs, fx.fx.ClinicB)
		assert.NotContains(t, got.ClinicIDs, fx.fx.ClinicA)
	})

	t.Run("clinics_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/clinics", fx.staffA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffA.ID))
		fx.handler.GetStaffClinicAssignments(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("capable_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/capable-reservation-types", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.staffB.ID))
		fx.handler.GetStaffCapableReservationTypes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("capable_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/capable-reservation-types", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffB.ID))
		fx.handler.GetStaffCapableReservationTypes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got reservationTypeIDsDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Contains(t, got.ReservationTypeIDs, fx.typeCapableB.ID)
		assert.NotContains(t, got.ReservationTypeIDs, fx.typeA.ID)
	})

	t.Run("capable_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/capable-reservation-types", fx.staffA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffA.ID))
		fx.handler.GetStaffCapableReservationTypes(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("excluded_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/excluded-reservation-types", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), fx.staffB.ID))
		fx.handler.GetStaffExcludedReservationTypes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("excluded_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/excluded-reservation-types", fx.staffB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffB.ID))
		fx.handler.GetStaffExcludedReservationTypes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got reservationTypeIDsDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Contains(t, got.ReservationTypeIDs, fx.typeExcludedB.ID)
		assert.NotContains(t, got.ReservationTypeIDs, fx.typeCapableB.ID)
		assert.NotContains(t, got.ReservationTypeIDs, fx.typeA.ID)
	})

	t.Run("excluded_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/staffs/%d/excluded-reservation-types", fx.staffA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), fx.staffA.ID))
		fx.handler.GetStaffExcludedReservationTypes(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("shift_templates_list_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shift-templates", configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA))
		fx.handler.ListShiftTemplates(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBShiftTemplateSeedA, realDBShiftTemplateSeedB)
	})

	t.Run("shift_templates_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shift-templates", configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB))
		fx.handler.ListShiftTemplates(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []shiftTemplateDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBShiftTemplateSeedA, item.Name)
			if item.ID == fx.templateB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("shift_templates_get_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/shift-templates/%d", fx.templateB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA), fx.templateB.ID))
		fx.handler.GetShiftTemplate(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("shift_templates_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/shift-templates/%d", fx.templateB.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB), fx.templateB.ID))
		fx.handler.GetShiftTemplate(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBShiftTemplateSeedB)
		assert.NotContains(t, w.Body.String(), realDBShiftTemplateSeedA)
	})

	t.Run("shift_templates_get_A_id_404", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/shift-templates/%d", fx.templateA.ID), withIDParam(configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB), fx.templateA.ID))
		fx.handler.GetShiftTemplate(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBShiftTemplateSeedA)
	})

	t.Run("shifts_list_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shifts?date="+realDBShiftYearMonth, configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA))
		fx.handler.ListShiftEntries(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBStaffSeedNameA)
		assert.NotContains(t, w.Body.String(), realDBStaffSeedNameB)
	})

	t.Run("shifts_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shifts?date="+realDBShiftYearMonth, configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB))
		fx.handler.ListShiftEntries(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []shiftDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fmt.Sprintf("%d", fx.fx.ClinicB), item.ClinicID)
			assert.NotEqual(t, fmt.Sprintf("%d", fx.staffA.ID), item.StaffID)
			assert.NotEqual(t, realDBStaffSeedNameA, item.StaffName)
			if item.StaffID == fmt.Sprintf("%d", fx.staffB.ID) {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("on_duty_grantA_403", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		fx.handler = newRealDBStaffHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shifts/on-duty-staffs?date="+realDBOnDutyDate, configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA))
		fx.handler.GetOnDutyStaffs(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBStaffSeedNameA)
		assert.NotContains(t, w.Body.String(), realDBStaffSeedNameB)
	})

	t.Run("on_duty_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBStaffIsolationTestDB(t)
		fx := seedRealDBStaffFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/shifts/on-duty-staffs?date="+realDBOnDutyDate, configureStaffGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB))
		fx.handler.GetOnDutyStaffs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []staffNameDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.NotEqual(t, realDBStaffSeedNameA, item.Name)
			if item.ID == fx.staffB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})
}
