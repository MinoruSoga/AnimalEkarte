package medicalrecord

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type realDBLabFixture struct {
	fx        testdb.ClinicGrantFixture
	jobA      *model.LabImportJob
	jobB      *model.LabImportJob
	deviceA   *model.LabDevice
	deviceB   *model.LabDevice
	itemA     *model.LabDeviceItemMaster
	itemB     *model.LabDeviceItemMaster
	examA     *model.Examination
	examB     *model.Examination
	examTypeA *model.ExaminationType
	examTypeB *model.ExaminationType
	unlinkedB *model.LabImportJob

	labH    *LabImportHandler
	reportH *LabReportHandler
}

func setupRealDBLabDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	ensureLabImportEnums(t, db)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.LabImportJob{}, &model.LabImportEvent{}, &model.LabImportJobItem{},
		&model.LabDevice{}, &model.LabDeviceItemMaster{}, &model.LabDeviceStationSettings{}, &model.LabDeviceWait{},
		&model.ExaminationType{}, &model.Examination{}, &model.ExamResult{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{},
	))
	require.NoError(t, db.Exec(`ALTER TABLE lab_device_item_masters DROP COLUMN IF EXISTS display_name`).Error)
	testdb.Truncate(t, db,
		"lab_import_events", "lab_import_job_items", "lab_import_jobs",
		"lab_device_waits", "lab_device_station_settings", "lab_device_item_masters", "lab_devices",
		"exam_results", "exams", "exam_types",
		"pets", "animal_species", "owners",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

type labJobQueryGuard struct {
	LabImportJobService
	g clinicIDGuard
}

func (s *labJobQueryGuard) GetJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error) {
	s.g.check(clinicID, "lab GetJob")
	return s.LabImportJobService.GetJob(ctx, clinicID, jobID)
}
func (s *labJobQueryGuard) ListEvents(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	s.g.check(clinicID, "lab ListEvents")
	return s.LabImportJobService.ListEvents(ctx, clinicID, jobID)
}

type labMasterQueryGuard struct {
	LabDeviceItemMasterService
	g clinicIDGuard
}

func (s *labMasterQueryGuard) List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error) {
	s.g.check(clinicID, "labMaster List")
	return s.LabDeviceItemMasterService.List(ctx, clinicID, sourceType)
}
func (s *labMasterQueryGuard) ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error) {
	s.g.check(clinicID, "labMaster ListDevices")
	return s.LabDeviceItemMasterService.ListDevices(ctx, clinicID)
}

type labReceiveQueryGuard struct {
	LabDeviceReceiveService
	g clinicIDGuard
}

func (s *labReceiveQueryGuard) Board(ctx context.Context, clinicID uint64) (*LabDeviceBoard, error) {
	s.g.check(clinicID, "labReceive Board")
	return s.LabDeviceReceiveService.Board(ctx, clinicID)
}
func (s *labReceiveQueryGuard) Unlinked(ctx context.Context, clinicID uint64) ([]LabDeviceJobCard, error) {
	s.g.check(clinicID, "labReceive Unlinked")
	return s.LabDeviceReceiveService.Unlinked(ctx, clinicID)
}
func (s *labReceiveQueryGuard) GetStation(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error) {
	s.g.check(clinicID, "labReceive GetStation")
	return s.LabDeviceReceiveService.GetStation(ctx, clinicID)
}

type labReportQueryGuard struct {
	LabReportQueryService
	g clinicIDGuard
}

func (s *labReportQueryGuard) GetExamReport(ctx context.Context, clinicID, examID uint64) (*model.LabExamReportDetail, error) {
	s.g.check(clinicID, "labReport GetExamReport")
	return s.LabReportQueryService.GetExamReport(ctx, clinicID, examID)
}
func (s *labReportQueryGuard) ListJobReportSummaries(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]model.LabExamReportSummary, error) {
	s.g.check(clinicID, "labReport ListJobReportSummaries")
	return s.LabReportQueryService.ListJobReportSummaries(ctx, clinicID, jobID)
}

func wireLabHandlers(t *testing.T, db *gorm.DB, forbidden uint64) (*LabImportHandler, *LabReportHandler) {
	t.Helper()
	tx := persistenceTx(db)
	jobSvc := LabImportJobService(NewLabImportJobService(NewLabImportJobRepository(db), NewLabImportEventRepository(db)))
	masterSvc := LabDeviceItemMasterService(NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db), tx))
	receiveSvc := LabDeviceReceiveService(NewLabDeviceReceiveService(
		NewLabDeviceReceiveRepository(db), masterSvc, nil, tx, nil, nil,
	))
	reportSvc := LabReportQueryService(NewLabReportQueryService(NewExaminationRepository(db)))
	if forbidden != 0 {
		g := clinicIDGuard{t: t, forbiddenClinicID: forbidden}
		jobSvc = &labJobQueryGuard{LabImportJobService: jobSvc, g: g}
		masterSvc = &labMasterQueryGuard{LabDeviceItemMasterService: masterSvc, g: g}
		receiveSvc = &labReceiveQueryGuard{LabDeviceReceiveService: receiveSvc, g: g}
		reportSvc = &labReportQueryGuard{LabReportQueryService: reportSvc, g: g}
	}
	labH := NewLabImportHandler(nil, jobSvc, nil).WithDeviceMasters(masterSvc).WithDeviceReceive(receiveSvc)
	return labH, NewLabReportHandler(reportSvc)
}

func seedRealDBLabFixture(t *testing.T, db *gorm.DB, forbidden uint64) realDBLabFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr lab realDB")
	ctx := context.Background()

	jobA := makeLabImportJob(t, db, fx.ClinicA, model.LabImportJobStatusReceived)
	jobB := makeLabImportJob(t, db, fx.ClinicB, model.LabImportJobStatusReceived)
	toStatus := model.LabImportJobStatusReceived
	require.NoError(t, db.WithContext(ctx).Create(&model.LabImportEvent{
		ClinicID: fx.ClinicA, JobID: jobA.ID, EventType: model.LabImportEventTypeStatusTransition, ToStatus: &toStatus, RowCount: 1,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.LabImportEvent{
		ClinicID: fx.ClinicB, JobID: jobB.ID, EventType: model.LabImportEventTypeStatusTransition, ToStatus: &toStatus, RowCount: 2,
	}).Error)

	deviceA := &model.LabDevice{ClinicID: fx.ClinicA, SourceType: "fuji_nx600", Name: realDBLabDeviceA, IsActive: true}
	deviceB := &model.LabDevice{ClinicID: fx.ClinicB, SourceType: "fuji_nx600", Name: realDBLabDeviceB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(deviceA).Error)
	require.NoError(t, db.WithContext(ctx).Create(deviceB).Error)

	itemA := &model.LabDeviceItemMaster{ClinicID: fx.ClinicA, SourceType: "fuji_nx600", DeviceItemCode: realDBLabItemCodeA, ValueShape: "number", IsActive: true}
	itemB := &model.LabDeviceItemMaster{ClinicID: fx.ClinicB, SourceType: "fuji_nx600", DeviceItemCode: realDBLabItemCodeB, ValueShape: "number", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(itemA).Error)
	require.NoError(t, db.WithContext(ctx).Create(itemB).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.LabDeviceStationSettings{
		ClinicID: fx.ClinicB, WaitTTLSeconds: 1800, SlotsJSON: realDBLabStationSlotsB,
	}).Error)

	unlinkedB := &model.LabImportJob{
		ClinicID: fx.ClinicB, SourceType: model.LabImportSourceType("fuji_nx600"),
		Status: model.LabImportJobStatusReceived, SourceFingerprint: realDBMRPrefix + "-unlinked-B",
	}
	require.NoError(t, db.WithContext(ctx).Create(unlinkedB).Error)

	examTypeA := makeExamTypeMaster(t, db, fx.ClinicA, realDBExamTypeA)
	examTypeB := makeExamTypeMaster(t, db, fx.ClinicB, realDBExamTypeB)
	examA := &model.Examination{ClinicID: fx.ClinicA, ExamTypeID: examTypeA.ID, Date: realDBNow(), Status: model.ExaminationStatusPending, Machine: realDBExamMachineA, JobID: &jobA.ID}
	examB := &model.Examination{ClinicID: fx.ClinicB, ExamTypeID: examTypeB.ID, Date: realDBNow(), Status: model.ExaminationStatusPending, Machine: realDBExamMachineB, JobID: &jobB.ID}
	require.NoError(t, db.WithContext(ctx).Create(examA).Error)
	require.NoError(t, db.WithContext(ctx).Create(examB).Error)

	labH, reportH := wireLabHandlers(t, db, forbidden)
	return realDBLabFixture{
		fx: fx, jobA: jobA, jobB: jobB, deviceA: deviceA, deviceB: deviceB,
		itemA: itemA, itemB: itemB, examA: examA, examB: examB,
		examTypeA: examTypeA, examTypeB: examTypeB, unlinkedB: unlinkedB,
		labH: labH, reportH: reportH,
	}
}

func TestRealDB_LabSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	res := string(model.ResourceLabImport)

	run := func(name string, guarded bool, fn func(*testing.T, realDBLabFixture)) {
		t.Run(name, func(t *testing.T) {
			db := setupRealDBLabDB(t)
			fx := seedRealDBLabFixture(t, db, 0)
			if guarded {
				labH, reportH := wireLabHandlers(t, db, fx.fx.ClinicB)
				fx.labH, fx.reportH = labH, reportH
			}
			fn(t, fx)
		})
	}

	run("item_masters_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device-item-masters", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.labH.ListLabDeviceItemMasters(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("item_masters_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device-item-masters", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.labH.ListLabDeviceItemMasters(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBLabItemCodeB)
		assert.NotContains(t, w.Body.String(), realDBLabItemCodeA)
	})

	run("devices_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-devices", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.labH.ListLabDevices(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("devices_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-devices", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.labH.ListLabDevices(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBLabDeviceB)
		assert.NotContains(t, w.Body.String(), realDBLabDeviceA)
	})

	run("board_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/board", configureGrantAction(fx.fx, res, "create", fx.fx.ClinicA))
		fx.labH.GetLabDeviceBoard(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("board_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/board", configureGrantAction(fx.fx, res, "create", fx.fx.ClinicB))
		fx.labH.GetLabDeviceBoard(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "d3b")
	})

	run("station_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/station", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.labH.GetLabDeviceStation(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("station_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/station", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.labH.GetLabDeviceStation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "d3b")
	})

	run("unlinked_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/unlinked", configureGrant(fx.fx, res, fx.fx.ClinicA))
		fx.labH.GetLabDeviceUnlinked(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("unlinked_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/unlinked", configureGrant(fx.fx, res, fx.fx.ClinicB))
		fx.labH.GetLabDeviceUnlinked(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fx.unlinkedB.ID.String())
	})

	run("agent_consumer_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/agent-consumer", configureGrantAction(fx.fx, res, "create", fx.fx.ClinicA))
		fx.labH.GetLabDeviceAgentConsumer(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("agent_consumer_grantB_ok", false, func(t *testing.T, fx realDBLabFixture) {
		t.Setenv("LAB_DEVICE_AGENT_CONSUMER_TOKEN", realDBAgentToken)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/lab-device/agent-consumer", configureGrantAction(fx.fx, res, "create", fx.fx.ClinicB))
		fx.labH.GetLabDeviceAgentConsumer(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBAgentToken)
		_ = os.Unsetenv
	})

	run("import_job_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), "job_id", fx.jobB.ID))
		fx.labH.GetLabImportJob(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("import_job_grantB_ok", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), "job_id", fx.jobB.ID))
		fx.labH.GetLabImportJob(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fx.jobB.ID.String())
	})
	run("import_job_A_id_404", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s", fx.jobA.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), "job_id", fx.jobA.ID))
		fx.labH.GetLabImportJob(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("import_events_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s/events", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), "job_id", fx.jobB.ID))
		fx.labH.ListLabImportEvents(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("import_events_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s/events", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), "job_id", fx.jobB.ID))
		fx.labH.ListLabImportEvents(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NotEqual(t, "[]", w.Body.String())
	})
	run("import_events_A_id_404", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-imports/%s/events", fx.jobA.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), "job_id", fx.jobA.ID))
		fx.labH.ListLabImportEvents(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("report_exam_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-reports/exams/%d", fx.examB.ID), withParams(configureGrant(fx.fx, res, fx.fx.ClinicA), gin.Param{Key: "exam_id", Value: fmt.Sprintf("%d", fx.examB.ID)}))
		fx.reportH.GetLabExamReport(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("report_exam_grantB_ok", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-reports/exams/%d", fx.examB.ID), withParams(configureGrant(fx.fx, res, fx.fx.ClinicB), gin.Param{Key: "exam_id", Value: fmt.Sprintf("%d", fx.examB.ID)}))
		fx.reportH.GetLabExamReport(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBExamMachineB)
		assert.NotContains(t, w.Body.String(), realDBExamMachineA)
	})
	run("report_exam_A_id_404", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-reports/exams/%d", fx.examA.ID), withParams(configureGrant(fx.fx, res, fx.fx.ClinicB), gin.Param{Key: "exam_id", Value: fmt.Sprintf("%d", fx.examA.ID)}))
		fx.reportH.GetLabExamReport(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	run("report_job_summaries_grantA_403", true, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-reports/jobs/%s/summaries", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), "job_id", fx.jobB.ID))
		fx.reportH.GetLabJobReportSummaries(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("report_job_summaries_grantB_nonempty", false, func(t *testing.T, fx realDBLabFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/lab-reports/jobs/%s/summaries", fx.jobB.ID), withUUIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), "job_id", fx.jobB.ID))
		fx.reportH.GetLabJobReportSummaries(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBExamMachineB)
		assert.NotContains(t, w.Body.String(), realDBExamMachineA)
	})
}
