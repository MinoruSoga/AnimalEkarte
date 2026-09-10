package clinic

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 clinic package (5 routes)
//
// Proves clinic-fixed clinic/holiday/closing-settings isolation through real
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

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBHolidayReasonA   = "D3clinic-realdb-holiday-A"
	realDBHolidayReasonB   = "D3clinic-realdb-holiday-B"
	realDBSpecialNoteA     = "D3clinic-realdb-special-A"
	realDBSpecialNoteB     = "D3clinic-realdb-special-B"
	realDBHolidayYearMonth = "2026-09"
	realDBHolidayDateB     = "2026-09-18"
	realDBBoundaryA        = "13:00:00"
	realDBBoundaryB        = "15:00:00"
)

type clinicQueryGuardService struct {
	Service
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *clinicQueryGuardService) GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	if id == s.forbiddenClinicID {
		s.t.Fatalf("GetClinicByID must not query clinic %d without selected-clinic grant", id)
	}
	return s.Service.GetClinicByID(ctx, id)
}

type holidayQueryGuardService struct {
	ClinicHolidayService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *holidayQueryGuardService) List(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("holiday List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ClinicHolidayService.List(ctx, clinicID, yearMonth)
}

type closingQueryGuardService struct {
	ClosingSettingsService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *closingQueryGuardService) Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("closing Get must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ClosingSettingsService.Get(ctx, clinicID)
}

func (s *closingQueryGuardService) ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("ListSpecialPeriods must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.ClosingSettingsService.ListSpecialPeriods(ctx, clinicID)
}

type realDBClinicFixture struct {
	fx       testdb.ClinicGrantFixture
	holidayA *model.ClinicHoliday
	holidayB *model.ClinicHoliday
	periodA  *model.ClosingSpecialPeriod
	periodB  *model.ClosingSpecialPeriod
	handler  *Handler
}

func setupRealDBClinicIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ClinicHoliday{},
		&model.ClosingSpecialPeriod{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	// AutoMigrate maps AmPmBoundary/PmEnd as timestamptz; coerce to time like production.
	require.NoError(t, db.Exec(`ALTER TABLE closing_special_periods ALTER COLUMN am_pm_boundary TYPE time USING am_pm_boundary::time`).Error)
	require.NoError(t, db.Exec(`ALTER TABLE closing_special_periods ALTER COLUMN pm_end TYPE time USING pm_end::time`).Error)
	testdb.Truncate(t, db,
		"closing_special_periods",
		"clinic_settings",
		"clinic_holidays",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

func newRealDBClinicHandler(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) *Handler {
	t.Helper()
	clinicRepo := NewClinicRepository(db)
	holidayRepo := NewClinicHolidayRepository(db)
	settingsRepo := NewClinicSettingsRepository(db)
	periodRepo := NewClosingSpecialPeriodRepository(db)

	clinicSvc := Service(NewService(clinicRepo, nil, nil))
	holidaySvc := ClinicHolidayService(NewClinicHolidayService(holidayRepo))
	closingSvc := ClosingSettingsService(NewClosingSettingsService(settingsRepo, periodRepo, holidayRepo, nil))

	if forbiddenClinicID != 0 {
		clinicSvc = &clinicQueryGuardService{Service: clinicSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		holidaySvc = &holidayQueryGuardService{ClinicHolidayService: holidaySvc, t: t, forbiddenClinicID: forbiddenClinicID}
		closingSvc = &closingQueryGuardService{ClosingSettingsService: closingSvc, t: t, forbiddenClinicID: forbiddenClinicID}
	}
	return NewHandler(clinicSvc, holidaySvc, closingSvc, nil, nil)
}

func seedRealDBClinicFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBClinicFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3clinic realDB")
	ctx := context.Background()

	settingsA := model.ClinicSettings{ClinicID: fx.ClinicA, ClosingAmPmBoundary: realDBBoundaryA, ClosingWeekdayEnd: "18:30", ClosingSundayEnd: "17:30"}
	settingsB := model.ClinicSettings{ClinicID: fx.ClinicB, ClosingAmPmBoundary: realDBBoundaryB, ClosingWeekdayEnd: "18:30", ClosingSundayEnd: "17:30"}
	require.NoError(t, db.WithContext(ctx).Create(&settingsA).Error)
	require.NoError(t, db.WithContext(ctx).Create(&settingsB).Error)

	dateA := time.Date(2026, 9, 10, 0, 0, 0, 0, time.Local)
	dateB, err := time.ParseInLocation(time.DateOnly, realDBHolidayDateB, time.Local)
	require.NoError(t, err)
	holidayA := &model.ClinicHoliday{ClinicID: fx.ClinicA, Date: dateA, Reason: realDBHolidayReasonA}
	holidayB := &model.ClinicHoliday{ClinicID: fx.ClinicB, Date: dateB, Reason: realDBHolidayReasonB}
	require.NoError(t, db.WithContext(ctx).Create(holidayA).Error)
	require.NoError(t, db.WithContext(ctx).Create(holidayB).Error)

	periodA := &model.ClosingSpecialPeriod{
		ClinicID: fx.ClinicA, StartDate: time.Date(2026, 12, 29, 0, 0, 0, 0, time.Local),
		EndDate:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.Local),
		AmPmBoundary: "12:00:00", PmEnd: "16:00:00", Note: realDBSpecialNoteA,
	}
	periodB := &model.ClosingSpecialPeriod{
		ClinicID: fx.ClinicB, StartDate: time.Date(2026, 12, 29, 0, 0, 0, 0, time.Local),
		EndDate:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.Local),
		AmPmBoundary: "12:00:00", PmEnd: "16:00:00", Note: realDBSpecialNoteB,
	}
	require.NoError(t, db.WithContext(ctx).Create(periodA).Error)
	require.NoError(t, db.WithContext(ctx).Create(periodB).Error)

	return realDBClinicFixture{
		fx: fx, holidayA: holidayA, holidayB: holidayB, periodA: periodA, periodB: periodB,
		handler: newRealDBClinicHandler(t, db, forbiddenClinicID),
	}
}

func configureClinicGrant(fx realDBClinicFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}

func TestRealDB_SelectedClinicBGrantAIsolation_Clinic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get_clinic_grantA_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		fx.handler = newRealDBClinicHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/clinics/%d", fx.fx.ClinicB),
			func(c *gin.Context) {
				configureClinicGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA)(c)
				c.Params = gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)}}
			})
		fx.handler.GetClinic(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, fx.fx.ClinicAName, fx.fx.ClinicBName)
	})

	t.Run("get_clinic_grantB_ok", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/clinics/%d", fx.fx.ClinicB),
			func(c *gin.Context) {
				configureClinicGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB)(c)
				c.Params = gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)}}
			})
		fx.handler.GetClinic(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fx.fx.ClinicBName)
		assert.NotContains(t, w.Body.String(), fx.fx.ClinicAName)
	})

	t.Run("get_clinic_A_id_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			fmt.Sprintf("/api/v1/clinics/%d", fx.fx.ClinicA),
			func(c *gin.Context) {
				configureClinicGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB)(c)
				c.Params = gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicA)}}
			})
		fx.handler.GetClinic(c)
		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), nil, fx.fx.ClinicAName, fx.fx.ClinicBName)
	})

	t.Run("holidays_grantA_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		fx.handler = newRealDBClinicHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/clinic-holidays?year_month="+realDBHolidayYearMonth,
			configureClinicGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA))
		fx.handler.ListClinicHolidays(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBHolidayReasonA, realDBHolidayReasonB)
	})

	t.Run("holidays_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/clinic-holidays?year_month="+realDBHolidayYearMonth,
			configureClinicGrant(fx, string(model.ResourceShifts), fx.fx.ClinicB))
		fx.handler.ListClinicHolidays(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []ClinicHolidayResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBHolidayReasonA, item.Reason)
			if item.ID == fx.holidayB.ID {
				foundB = true
				assert.Equal(t, realDBHolidayReasonB, item.Reason)
			}
		}
		require.True(t, foundB)
	})

	t.Run("closing_settings_holidays_grantA_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		fx.handler = newRealDBClinicHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/closing-settings/holidays?year_month="+realDBHolidayYearMonth,
			configureClinicGrant(fx, string(model.ResourceShifts), fx.fx.ClinicA))
		fx.handler.ListClinicHolidays(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBHolidayReasonA, realDBHolidayReasonB)
	})

	t.Run("closing_settings_grantA_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		fx.handler = newRealDBClinicHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/closing-settings",
			configureClinicGrant(fx, string(model.ResourceClosingSettings), fx.fx.ClinicA))
		fx.handler.GetClosingSettings(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBSpecialNoteA, realDBSpecialNoteB, realDBBoundaryA, realDBBoundaryB)
	})

	t.Run("closing_settings_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/closing-settings",
			configureClinicGrant(fx, string(model.ResourceClosingSettings), fx.fx.ClinicB))
		fx.handler.GetClosingSettings(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got ClosingSettingsFullResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, fx.fx.ClinicB, got.Settings.ClinicID)
		assert.Equal(t, realDBBoundaryB, got.Settings.ClosingAmPmBoundary)
		require.NotEmpty(t, got.SpecialPeriods)
		foundB := false
		for _, item := range got.SpecialPeriods {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBSpecialNoteA, item.Note)
			if item.ID == fx.periodB.ID {
				foundB = true
				assert.Equal(t, realDBSpecialNoteB, item.Note)
			}
		}
		require.True(t, foundB)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBSpecialNoteA, realDBBoundaryA)
	})

	t.Run("special_periods_grantA_403", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		fx.handler = newRealDBClinicHandler(t, db, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/closing-settings/special-periods",
			configureClinicGrant(fx, string(model.ResourceClosingSettings), fx.fx.ClinicA))
		fx.handler.ListSpecialPeriods(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBSpecialNoteA, realDBSpecialNoteB)
	})

	t.Run("special_periods_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBClinicIsolationTestDB(t)
		fx := seedRealDBClinicFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet,
			"/api/v1/closing-settings/special-periods",
			configureClinicGrant(fx, string(model.ResourceClosingSettings), fx.fx.ClinicB))
		fx.handler.ListSpecialPeriods(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []ClosingSpecialPeriodResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBSpecialNoteA, item.Note)
			if item.ID == fx.periodB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})
}
