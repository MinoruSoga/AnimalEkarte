package medicalrecord

import (
	"context"
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

type realDBHospFixture struct {
	fx     testdb.ClinicGrantFixture
	ownerA *model.Owner
	ownerB *model.Owner
	petA   *model.Pet
	petB   *model.Pet
	hospA  *model.Hospitalization
	hospB  *model.Hospitalization
	dailyB *model.DailyRecord

	hospH  *HospitalizationHandler
	careH  *CarePlanItemHandler
	dailyH *DailyRecordHandler
	planH  *TreatmentPlanHandler
}

func setupRealDBHospDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{},
		&model.Hospitalization{}, &model.CarePlanItem{}, &model.DailyRecord{},
		&model.VitalRecord{}, &model.CareLog{}, &model.StaffNote{}, &model.TreatmentPlan{},
	))
	testdb.Truncate(t, db,
		"staff_notes", "care_logs", "vital_records",
		"treatment_plans", "daily_records", "care_plan_items", "hospitalizations",
		"pets", "owners", "staff_clinic_assignments", "staffs",
	)
	return db
}

type hospQueryGuard struct {
	HospitalizationService
	g clinicIDGuard
}

func (s *hospQueryGuard) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	s.g.check(clinicID, "hosp List")
	return s.HospitalizationService.List(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}
func (s *hospQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	s.g.check(clinicID, "hosp GetByID")
	return s.HospitalizationService.GetByID(ctx, clinicID, id)
}

type careQueryGuard struct {
	CarePlanItemService
	g clinicIDGuard
}

func (s *careQueryGuard) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	s.g.check(clinicID, "care List")
	return s.CarePlanItemService.List(ctx, clinicID, hospitalizationID)
}

type dailyQueryGuard struct {
	DailyRecordService
	g clinicIDGuard
}

func (s *dailyQueryGuard) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	s.g.check(clinicID, "daily List")
	return s.DailyRecordService.List(ctx, clinicID, hospitalizationID)
}
func (s *dailyQueryGuard) GetByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	s.g.check(clinicID, "daily GetByDate")
	return s.DailyRecordService.GetByDate(ctx, clinicID, hospitalizationID, date)
}

type hospPlanListGuard struct {
	TreatmentPlanService
	g clinicIDGuard
}

func (s *hospPlanListGuard) ListByHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.TreatmentPlan, error) {
	s.g.check(clinicID, "treatmentPlan ListByHospitalization")
	return s.TreatmentPlanService.ListByHospitalization(ctx, clinicID, hospitalizationID)
}

func wireHospHandlers(t *testing.T, db *gorm.DB, forbidden uint64) (
	*HospitalizationHandler, *CarePlanItemHandler, *DailyRecordHandler, *TreatmentPlanHandler,
) {
	t.Helper()
	tx := persistenceTx(db)
	hospRepo := NewHospitalizationRepository(db)
	hospSvc := HospitalizationService(NewHospitalizationService(hospRepo, nil, nil, nil, nil, nil, nil, tx))
	careSvc := CarePlanItemService(NewCarePlanItemService(NewCarePlanItemRepository(db), hospRepo, nil, nil, nil, tx, nil))
	dailySvc := DailyRecordService(NewDailyRecordService(NewDailyRecordRepository(db), hospRepo, reservationOwnerPet(db), tx))
	planSvc := TreatmentPlanService(NewTreatmentPlanService(NewTreatmentPlanRepository(db), tx))
	if forbidden != 0 {
		g := clinicIDGuard{t: t, forbiddenClinicID: forbidden}
		hospSvc = &hospQueryGuard{HospitalizationService: hospSvc, g: g}
		careSvc = &careQueryGuard{CarePlanItemService: careSvc, g: g}
		dailySvc = &dailyQueryGuard{DailyRecordService: dailySvc, g: g}
		planSvc = &hospPlanListGuard{TreatmentPlanService: planSvc, g: g}
	}
	return NewHospitalizationHandler(hospSvc, nil),
		NewCarePlanItemHandler(careSvc),
		NewDailyRecordHandler(dailySvc),
		NewTreatmentPlanHandler(planSvc, hospSvc, newMedicalRecordReadService(db), nil)
}

func seedRealDBHospFixture(t *testing.T, db *gorm.DB, forbidden uint64) realDBHospFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr hosp realDB")
	ctx := context.Background()
	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBOwnerNameA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBOwnerNameB)
	petA := testdb.MakeSpeciesAndPet(t, db, fx.ClinicA, ownerA.ID, realDBPetNameA)
	petB := testdb.MakeSpeciesAndPet(t, db, fx.ClinicB, ownerB.ID, realDBPetNameB)
	hospA := makeHospitalizationFixture(t, db, fx.ClinicA, ownerA.ID, petA.ID, nil)
	hospB := makeHospitalizationFixture(t, db, fx.ClinicB, ownerB.ID, petB.ID, nil)

	require.NoError(t, db.WithContext(ctx).Create(&model.CarePlanItem{
		HospitalizationID: hospA.ID, Type: model.CarePlanTypeInstruction, Name: realDBCarePlanA,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.CarePlanItem{
		HospitalizationID: hospB.ID, Type: model.CarePlanTypeInstruction, Name: realDBCarePlanB,
	}).Error)

	dailyDate := realDBNow()
	dailyA := &model.DailyRecord{ClinicID: fx.ClinicA, HospitalizationID: hospA.ID, Date: dailyDate}
	dailyB := &model.DailyRecord{ClinicID: fx.ClinicB, HospitalizationID: hospB.ID, Date: dailyDate}
	require.NoError(t, db.WithContext(ctx).Create(dailyA).Error)
	require.NoError(t, db.WithContext(ctx).Create(dailyB).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.TreatmentPlan{
		ClinicID: fx.ClinicA, HospitalizationID: &hospA.ID, TreatmentContent: realDBTreatmentPlanA,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.TreatmentPlan{
		ClinicID: fx.ClinicB, HospitalizationID: &hospB.ID, TreatmentContent: realDBTreatmentPlanB,
	}).Error)

	hospH, careH, dailyH, planH := wireHospHandlers(t, db, forbidden)
	return realDBHospFixture{
		fx: fx, ownerA: ownerA, ownerB: ownerB, petA: petA, petB: petB,
		hospA: hospA, hospB: hospB, dailyB: dailyB,
		hospH: hospH, careH: careH, dailyH: dailyH, planH: planH,
	}
}

func TestRealDB_HospSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	res := string(model.ResourceHospitalization)

	run := func(name string, guarded bool, fn func(*testing.T, realDBHospFixture)) {
		t.Run(name, func(t *testing.T) {
			db := setupRealDBHospDB(t)
			fx := seedRealDBHospFixture(t, db, 0)
			if guarded {
				hospH, careH, dailyH, planH := wireHospHandlers(t, db, fx.fx.ClinicB)
				fx.hospH, fx.careH, fx.dailyH, fx.planH = hospH, careH, dailyH, planH
			}
			fn(t, fx)
		})
	}

	run("hosp_list_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/hospitalizations?page=1&limit=10", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.hospH.ListHospitalizations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("hosp_list_grantB_nonempty", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/hospitalizations?page=1&limit=10", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.hospH.ListHospitalizations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fmt.Sprintf(`"id":%d`, fx.hospB.ID))
		assert.NotContains(t, w.Body.String(), fmt.Sprintf(`"id":%d`, fx.hospA.ID))
	})
	run("hosp_get_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.hospB.ID))
		fx.hospH.GetHospitalization(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("hosp_get_grantB_ok", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospB.ID))
		fx.hospH.GetHospitalization(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fmt.Sprintf(`"id":%d`, fx.hospB.ID))
	})
	run("hosp_get_A_id_404", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d", fx.hospA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospA.ID))
		fx.hospH.GetHospitalization(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("care_plan_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/care-plan-items", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.hospB.ID))
		fx.careH.ListCarePlanItems(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("care_plan_grantB_nonempty", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/care-plan-items", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospB.ID))
		fx.careH.ListCarePlanItems(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBCarePlanB)
		assert.NotContains(t, w.Body.String(), realDBCarePlanA)
	})
	run("care_plan_A_id_empty_no_leak", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/care-plan-items", fx.hospA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospA.ID))
		fx.careH.ListCarePlanItems(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBCarePlanA)
	})

	run("daily_list_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.hospB.ID))
		fx.dailyH.ListDailyRecords(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("daily_list_grantB_nonempty", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospB.ID))
		fx.dailyH.ListDailyRecords(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fmt.Sprintf(`"id":"%d"`, fx.dailyB.ID))
	})
	run("daily_get_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", fx.hospB.ID, realDBDailyDate), withParams(configureGrant(fx.fx, res, fx.fx.ClinicA), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.hospB.ID)}, gin.Param{Key: "date", Value: realDBDailyDate}))
		fx.dailyH.GetDailyRecord(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("daily_get_grantB_ok", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", fx.hospB.ID, realDBDailyDate), withParams(configureGrant(fx.fx, res, fx.fx.ClinicB), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.hospB.ID)}, gin.Param{Key: "date", Value: realDBDailyDate}))
		fx.dailyH.GetDailyRecord(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fmt.Sprintf(`"id":"%d"`, fx.dailyB.ID))
	})
	run("daily_get_A_id_404", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/daily-records/%s", fx.hospA.ID, realDBDailyDate), withParams(configureGrant(fx.fx, res, fx.fx.ClinicB), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.hospA.ID)}, gin.Param{Key: "date", Value: realDBDailyDate}))
		fx.dailyH.GetDailyRecord(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("hosp_treatment_plans_grantA_403", true, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/treatment-plans", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.hospB.ID))
		fx.planH.ListTreatmentPlansByHospitalization(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("hosp_treatment_plans_grantB_nonempty", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/treatment-plans", fx.hospB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospB.ID))
		fx.planH.ListTreatmentPlansByHospitalization(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBTreatmentPlanB)
		assert.NotContains(t, w.Body.String(), realDBTreatmentPlanA)
	})
	run("hosp_treatment_plans_A_id_404", false, func(t *testing.T, fx realDBHospFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/hospitalizations/%d/treatment-plans", fx.hospA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.hospA.ID))
		fx.planH.ListTreatmentPlansByHospitalization(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}
