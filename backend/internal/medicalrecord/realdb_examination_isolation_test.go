package medicalrecord

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type realDBExamFixture struct {
	fx        testdb.ClinicGrantFixture
	examTypeA *model.ExaminationType
	examTypeB *model.ExaminationType
	examA     *model.Examination
	examB     *model.Examination
	handler   *ExaminationHandler
}

func setupRealDBExamDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.ExaminationType{}, &model.Examination{}, &model.ExamResult{},
		&model.ExaminationRevision{}, &model.ExaminationRevisionItem{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{},
	))
	testdb.Truncate(t, db,
		"examination_revision_items", "examination_revisions",
		"exam_results", "exams", "exam_types",
		"pets", "animal_species", "owners",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

type examQueryGuard struct {
	ExaminationService
	g clinicIDGuard
}

func (s *examQueryGuard) List(ctx context.Context, clinicID uint64, petID, ownerID, medicalRecordID *uint64, status, startDate, endDate *string, page, limit int, includeItems bool) ([]model.Examination, int64, error) {
	s.g.check(clinicID, "exam List")
	return s.ExaminationService.List(ctx, clinicID, petID, ownerID, medicalRecordID, status, startDate, endDate, page, limit, includeItems)
}
func (s *examQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	s.g.check(clinicID, "exam GetByID")
	return s.ExaminationService.GetByID(ctx, clinicID, id)
}
func (s *examQueryGuard) ListItems(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	s.g.check(clinicID, "exam ListItems")
	return s.ExaminationService.ListItems(ctx, clinicID, examID)
}
func (s *examQueryGuard) GetPrintSnapshot(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error) {
	s.g.check(clinicID, "exam GetPrintSnapshot")
	return s.ExaminationService.GetPrintSnapshot(ctx, clinicID, examinationID, version)
}

func wireExamHandler(t *testing.T, db *gorm.DB, forbidden uint64) *ExaminationHandler {
	t.Helper()
	tx := persistenceTx(db)
	svc := ExaminationService(NewExaminationService(
		NewExaminationRepository(db),
		NewMedicalRecordRepository(db),
		NewExamTypeRepository(db),
		nil,
		tx,
	))
	if forbidden != 0 {
		svc = &examQueryGuard{ExaminationService: svc, g: clinicIDGuard{t: t, forbiddenClinicID: forbidden}}
	}
	return NewExaminationHandler(svc)
}

func seedRealDBExamFixture(t *testing.T, db *gorm.DB, forbidden uint64) realDBExamFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr exam realDB")
	ctx := context.Background()

	examTypeA := makeExamTypeMaster(t, db, fx.ClinicA, realDBExamTypeA)
	examTypeB := makeExamTypeMaster(t, db, fx.ClinicB, realDBExamTypeB)

	actorA := makeExaminationActor(t, db, fx.ClinicA, realDBMRPrefix+"-actor-A")
	actorB := makeExaminationActor(t, db, fx.ClinicB, realDBMRPrefix+"-actor-B")

	svc := NewExaminationService(
		NewExaminationRepository(db),
		NewMedicalRecordRepository(db),
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistenceTx(db),
	)
	itemsA := []UpsertExamItemInput{{Name: "WBC-A", InspectionValue: "1.0", Unit: "x", SortOrder: 1}}
	itemsB := []UpsertExamItemInput{{Name: "WBC-B", InspectionValue: "2.0", Unit: "x", SortOrder: 1}}
	examA, err := svc.Create(ctx, fx.ClinicA, &CreateExaminationInput{
		ExamTypeID: examTypeA.ID, Date: realDBNow(), Status: model.ExaminationStatusConfirmed, ActorID: &actorA, Items: &itemsA, Machine: realDBExamMachineA,
	})
	require.NoError(t, err)
	examB, err := svc.Create(ctx, fx.ClinicB, &CreateExaminationInput{
		ExamTypeID: examTypeB.ID, Date: realDBNow(), Status: model.ExaminationStatusConfirmed, ActorID: &actorB, Items: &itemsB, Machine: realDBExamMachineB,
	})
	require.NoError(t, err)

	return realDBExamFixture{
		fx: fx, examTypeA: examTypeA, examTypeB: examTypeB, examA: examA, examB: examB,
		handler: wireExamHandler(t, db, forbidden),
	}
}

func TestRealDB_ExaminationSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	res := string(model.ResourceExaminations)

	run := func(name string, guarded bool, fn func(*testing.T, realDBExamFixture)) {
		t.Run(name, func(t *testing.T) {
			db := setupRealDBExamDB(t)
			fx := seedRealDBExamFixture(t, db, 0)
			if guarded {
				fx.handler = wireExamHandler(t, db, fx.fx.ClinicB)
			}
			fn(t, fx)
		})
	}

	run("exam_list_grantA_403", true, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/examinations?page=1&limit=10", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.handler.ListExaminations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("exam_list_grantB_nonempty", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/examinations?page=1&limit=10", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.handler.ListExaminations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBExamMachineB)
		assert.NotContains(t, w.Body.String(), realDBExamMachineA)
	})
	run("exam_get_grantA_403", true, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.examB.ID))
		fx.handler.GetExamination(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("exam_get_grantB_ok", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examB.ID))
		fx.handler.GetExamination(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBExamMachineB)
	})
	run("exam_get_A_id_404", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d", fx.examA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examA.ID))
		fx.handler.GetExamination(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("exam_items_grantA_403", true, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/items", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.examB.ID))
		fx.handler.ListExaminationItems(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("exam_items_grantB_nonempty", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/items", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examB.ID))
		fx.handler.ListExaminationItems(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "WBC-B")
		assert.NotContains(t, w.Body.String(), "WBC-A")
	})
	run("exam_items_A_id_404", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/items", fx.examA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examA.ID))
		fx.handler.ListExaminationItems(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("exam_print_grantA_403", true, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/print-snapshot", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), fx.examB.ID))
		fx.handler.GetExaminationPrintSnapshot(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("exam_print_grantB_ok", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/print-snapshot", fx.examB.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examB.ID))
		fx.handler.GetExaminationPrintSnapshot(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "WBC-B")
		assert.NotContains(t, w.Body.String(), "WBC-A")
	})
	run("exam_print_A_id_404", false, func(t *testing.T, fx realDBExamFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/examinations/%d/print-snapshot", fx.examA.ID), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), fx.examA.ID))
		fx.handler.GetExaminationPrintSnapshot(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}
