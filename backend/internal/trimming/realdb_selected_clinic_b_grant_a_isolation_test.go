package trimming

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 trimming package (8 routes)
//
// Proves clinic-fixed trimming master + appointment list/detail isolation through
// real repository + service + HTTP handler paths. Offline `go test -short` SKIPs
// via testdb.SetupTestDB.

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

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBCourseTypeSeedA = "D3trim-realdb-course-type-A"
	realDBCourseTypeSeedB = "D3trim-realdb-course-type-B"
	realDBCourseSeedA     = "D3trim-realdb-course-A"
	realDBCourseSeedB     = "D3trim-realdb-course-B"
	realDBOptionSeedA     = "D3trim-realdb-option-A"
	realDBOptionSeedB     = "D3trim-realdb-option-B"
)

type trimmingQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *trimmingQueryGuardService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("trimming List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.List(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (s *trimmingQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("trimming GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.Service.GetByID(ctx, clinicID, id)
}

type courseQueryGuardService struct {
	TrimmingCourseService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *courseQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("course List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingCourseService.List(ctx, clinicID)
}

func (s *courseQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("course GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingCourseService.GetByID(ctx, clinicID, id)
}

type courseTypeQueryGuardService struct {
	TrimmingCourseTypeService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *courseTypeQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("course-type List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingCourseTypeService.List(ctx, clinicID)
}

func (s *courseTypeQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("course-type GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingCourseTypeService.GetByID(ctx, clinicID, id)
}

type optionQueryGuardService struct {
	TrimmingOptionService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *optionQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("option List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingOptionService.List(ctx, clinicID)
}

func (s *optionQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("option GetByID must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.TrimmingOptionService.GetByID(ctx, clinicID, id)
}

type realDBTrimmingFixture struct {
	fx          testdb.ClinicGrantFixture
	courseTypeA *model.TrimmingCourseType
	courseTypeB *model.TrimmingCourseType
	courseA     *model.TrimmingCourse
	courseB     *model.TrimmingCourse
	optionA     *model.TrimmingOption
	optionB     *model.TrimmingOption
	apptA       *model.Reservation
	apptB       *model.Reservation
	handler     *Handler
}

func setupRealDBTrimmingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// Warm shared DBs may keep orphan option rows that block AutoMigrate FK add.
	if db.Migrator().HasTable("appointment_trimming_options") {
		testdb.Truncate(t, db, "appointment_trimming_options")
	}
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.ReservationType{},
		&model.Reservation{},
		&model.AppointmentTrimmingDetail{},
		&model.TrimmingCourseType{},
		&model.TrimmingCourse{},
		&model.TrimmingOption{},
	))
	testdb.Truncate(t, db,
		"appointment_trimming_options",
		"appointment_trimming_details",
		"appointments",
		"trimming_courses",
		"trimming_options",
		"trimming_course_types",
		"reservation_types",
		"pets",
		"owners",
		"animal_species",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

func newRealDBTrimmingHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	reservationRepo := reservation.NewReservationRepository(db)
	detailRepo := NewAppointmentTrimmingDetailRepository(db)
	courseRepo := NewTrimmingCourseRepository(db)
	courseTypeRepo := NewTrimmingCourseTypeRepository(db)
	optionRepo := NewTrimmingOptionRepository(db)

	trimmingSvc := Service(NewService(reservationRepo, nil, nil, nil, detailRepo, courseRepo, optionRepo, nil))
	courseSvc := TrimmingCourseService(NewTrimmingCourseService(courseRepo, courseTypeRepo, nil))
	courseTypeSvc := TrimmingCourseTypeService(NewTrimmingCourseTypeService(courseTypeRepo, nil))
	optionSvc := TrimmingOptionService(NewTrimmingOptionService(optionRepo, nil))

	if forbiddenClinicID != 0 {
		trimmingSvc = &trimmingQueryGuardService{Service: trimmingSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		courseSvc = &courseQueryGuardService{TrimmingCourseService: courseSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		courseTypeSvc = &courseTypeQueryGuardService{TrimmingCourseTypeService: courseTypeSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		optionSvc = &optionQueryGuardService{TrimmingOptionService: optionSvc, t: t, forbiddenClinicID: forbiddenClinicID}
	}
	return NewHandler(trimmingSvc, courseSvc, courseTypeSvc, optionSvc)
}

func seedTrimmingAppointment(t *testing.T, db *gorm.DB, clinicID uint64, start time.Time) *model.Reservation {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID, Name: fmt.Sprintf("D3trim-rt-%d", clinicID),
		Category: model.ReservationTypeCategoryTrimming, IsActive: true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	appt := &model.Reservation{
		ClinicID: clinicID, StartTime: start, EndTime: start.Add(30 * time.Minute),
		ReservationTypeID: rt.ID, VisitType: model.VisitTypeRevisit,
		Status: model.ReservationStatusPending, Source: model.ReservationSourceManual,
		CustomerFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(appt).Error)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.AppointmentTrimmingDetail{
		ClinicID: clinicID, AppointmentID: appt.ID, StyleRequest: fmt.Sprintf("D3trim-style-%d", clinicID),
	}).Error)
	return appt
}

func seedRealDBTrimmingFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBTrimmingFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3trim realDB")
	ctx := context.Background()

	courseTypeA := &model.TrimmingCourseType{ClinicID: fx.ClinicA, Name: realDBCourseTypeSeedA, IsActive: true}
	courseTypeB := &model.TrimmingCourseType{ClinicID: fx.ClinicB, Name: realDBCourseTypeSeedB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(courseTypeA).Error)
	require.NoError(t, db.WithContext(ctx).Create(courseTypeB).Error)

	courseA := &model.TrimmingCourse{ClinicID: fx.ClinicA, Name: realDBCourseSeedA, IsActive: true}
	courseB := &model.TrimmingCourse{ClinicID: fx.ClinicB, Name: realDBCourseSeedB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(courseA).Error)
	require.NoError(t, db.WithContext(ctx).Create(courseB).Error)

	optionA := &model.TrimmingOption{ClinicID: fx.ClinicA, Name: realDBOptionSeedA, IsActive: true}
	optionB := &model.TrimmingOption{ClinicID: fx.ClinicB, Name: realDBOptionSeedB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(optionA).Error)
	require.NoError(t, db.WithContext(ctx).Create(optionB).Error)

	start := time.Date(2026, 9, 20, 10, 0, 0, 0, time.Local)
	apptA := seedTrimmingAppointment(t, db, fx.ClinicA, start)
	apptB := seedTrimmingAppointment(t, db, fx.ClinicB, start.Add(time.Hour))

	return realDBTrimmingFixture{
		fx: fx, courseTypeA: courseTypeA, courseTypeB: courseTypeB,
		courseA: courseA, courseB: courseB, optionA: optionA, optionB: optionB,
		apptA: apptA, apptB: apptB, handler: newRealDBTrimmingHandler(t, db, forbiddenClinicID),
	}
}

func configureTrimmingGrant(fx realDBTrimmingFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}

func withTrimIDParam(configure func(*gin.Context), id uint64) func(*gin.Context) {
	return func(c *gin.Context) {
		configure(c)
		c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", id)}}
	}
}

func TestRealDB_SelectedClinicBGrantAIsolation_Trimming(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("course_types_list_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-course-types",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA))
		fx.handler.ListTrimmingCourseTypes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBCourseTypeSeedA, realDBCourseTypeSeedB)
	})

	t.Run("course_types_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-course-types",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB))
		fx.handler.ListTrimmingCourseTypes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []trimmingCourseTypeResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBCourseTypeSeedA, item.Name)
			if item.ID == fx.courseTypeB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("course_types_get_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-course-types/%d", fx.courseTypeB.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA), fx.courseTypeB.ID))
		fx.handler.GetTrimmingCourseType(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("course_types_get_A_id_404", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-course-types/%d", fx.courseTypeA.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB), fx.courseTypeA.ID))
		fx.handler.GetTrimmingCourseType(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBCourseTypeSeedA)
	})

	t.Run("courses_list_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-courses",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA))
		fx.handler.ListTrimmingCourses(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBCourseSeedA, realDBCourseSeedB)
	})

	t.Run("courses_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-courses",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB))
		fx.handler.ListTrimmingCourses(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []trimmingCourseResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBCourseSeedA, item.Name)
			if item.ID == fx.courseB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("courses_get_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-courses/%d", fx.courseB.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA), fx.courseB.ID))
		fx.handler.GetTrimmingCourse(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("courses_get_A_id_404", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-courses/%d", fx.courseA.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB), fx.courseA.ID))
		fx.handler.GetTrimmingCourse(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBCourseSeedA)
	})

	t.Run("options_list_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-options",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA))
		fx.handler.ListTrimmingOptions(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBOptionSeedA, realDBOptionSeedB)
	})

	t.Run("options_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/trimming-options",
			configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB))
		fx.handler.ListTrimmingOptions(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []trimmingOptionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBOptionSeedA, item.Name)
			if item.ID == fx.optionB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("options_get_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-options/%d", fx.optionB.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicA), fx.optionB.ID))
		fx.handler.GetTrimmingOption(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("options_get_A_id_404", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/masters/trimming-options/%d", fx.optionA.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceMasterTrimming), fx.fx.ClinicB), fx.optionA.ID))
		fx.handler.GetTrimmingOption(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBOptionSeedA)
	})

	t.Run("trimmings_list_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/trimmings?page=1&limit=50",
			configureTrimmingGrant(fx, string(model.ResourceTrimming), fx.fx.ClinicA))
		fx.handler.ListTrimmings(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB})
	})

	t.Run("trimmings_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/trimmings?page=1&limit=50",
			configureTrimmingGrant(fx, string(model.ResourceTrimming), fx.fx.ClinicB))
		fx.handler.ListTrimmings(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]TrimmingResponse]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		foundB := false
		for _, item := range listed.Data {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, fx.apptA.ID, item.ID)
			if item.ID == fx.apptB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA})
	})

	t.Run("trimmings_get_grantA_403", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		fx.handler = newRealDBTrimmingHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/trimmings/%d", fx.apptB.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceTrimming), fx.fx.ClinicA), fx.apptB.ID))
		fx.handler.GetTrimming(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("trimmings_get_A_id_404", func(t *testing.T) {
		db := setupRealDBTrimmingIsolationTestDB(t)
		fx := seedRealDBTrimmingFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/trimmings/%d", fx.apptA.ID),
			withTrimIDParam(configureTrimmingGrant(fx, string(model.ResourceTrimming), fx.fx.ClinicB), fx.apptA.ID))
		fx.handler.GetTrimming(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB})
	})
}
