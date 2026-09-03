package medicalrecord

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRegisterRoutes_Snapshot is medicalrecord's half of the BE9-2B/2C/2D route-snapshot
// regression check (see internal/handler/handler_routes_snapshot_test.go's file header for
// the pattern and internal/manualarticle/routes_snapshot_test.go for the precedent this
// mirrors). Before BE9-2C/2D these 68 routes were captured inside
// internal/handler/testdata/route_snapshot.golden as part of *handler.Handler.RegisterRoutes
// (25 master-CRUD routes via RegisterMasterRoutes in 2C, plus 37 more in 2D — the
// vaccine/checkup-type/inquiry-template masters, /vaccinations, /checkups, and the
// checkup/prescription/inquiry medical-record sub-resources — plus 6 lab saga routes in
// sub-batch③: /lab-imports preview/commit/job/events + /lab-reports summaries/exam, plus 11
// vital/clinical-plan/image medical-record sub-resources in sub-batch④a, for 79 total);
// that golden file was updated to
// drop them (they moved here — same routes, same methods, same handler names, just registered
// by medicalrecord.Handler.RegisterRoutes instead of *handler.Handler). The permission
// arguments are NOT captured here (gin's RouteInfo.Handler exposes only the trailing handler
// name, not the middleware chain) — RBAC parity is enforced by hand-transcription in routes.go.
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewDiagnosisHandler(nil, nil),
		NewExamTypeHandler(nil),
		NewChiefComplaintHandler(nil),
		NewCheckupHandler(nil, nil),
		NewCheckupTypeHandler(nil),
		NewVaccineHandler(nil),
		NewVaccinationHandler(nil),
		NewPrescriptionHandler(nil),
		NewInquiryHandler(nil),
		NewInquiryTemplateHandler(nil),
		NewLabImportHandler(nil, nil, nil),
		NewLabReportHandler(nil),
		NewVitalHandler(nil, nil),
		NewClinicalPlanHandler(nil),
		NewMedicalRecordImageHandler(nil, nil, nil),
		NewTreatmentHandler(nil, nil),
		NewHospitalizationHandler(nil, nil),
		NewHospitalizationPlanHandler(nil),
		NewDailyRecordHandler(nil),
		NewCarePlanItemHandler(nil),
		NewConsultationHandler(nil),
		NewProcedureHandler(nil),
		NewMedicineHandler(nil),
		NewMedicineDoseParamHandler(nil),
		NewCageHandler(nil),
		NewTreatmentPlanHandler(nil, nil, nil, nil),
		NewMedicalRecordHandler(nil),
		NewMedicalRecordAddendumHandler(nil),
		NewExaminationHandler(nil),
		NewCheckupPackageImportHandler(nil),
		noopPermission,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	want := "" +
		"DELETE /api/v1/examinations/:id DeleteExamination\n" +
		"DELETE /api/v1/hospitalizations/:id DeleteHospitalization\n" +
		"DELETE /api/v1/hospitalizations/:id/care-plan-items/:itemId DeleteCarePlanItem\n" +
		"DELETE /api/v1/hospitalizations/:id/treatment-plans/:planId DeleteTreatmentPlanInHospitalization\n" +
		"DELETE /api/v1/lab-device/wait DeleteLabDeviceWait\n" +
		"DELETE /api/v1/masters/cages/:id DeleteCage\n" +
		"DELETE /api/v1/masters/checkup-types/:id DeleteCheckupType\n" +
		"DELETE /api/v1/masters/chief-complaint-types/:id DeleteChiefComplaint\n" +
		"DELETE /api/v1/masters/consultations/:id DeleteConsultation\n" +
		"DELETE /api/v1/masters/diagnosis-names/:id DeleteDiagnosisName\n" +
		"DELETE /api/v1/masters/diagnosis-types/:id DeleteDiagnosisType\n" +
		"DELETE /api/v1/masters/examination-types/:id DeleteExaminationType\n" +
		"DELETE /api/v1/masters/examination-types/:id/fields/:fieldId DeleteExaminationTypeField\n" +
		"DELETE /api/v1/masters/hospitalization-plans/:id DeleteHospitalizationPlan\n" +
		"DELETE /api/v1/masters/inquiry-templates/:id DeleteInquiryTemplate\n" +
		"DELETE /api/v1/masters/medicines/:id DeleteMedicine\n" +
		"DELETE /api/v1/masters/medicines/:id/dose-params/:species DeleteMedicineDoseParam\n" +
		"DELETE /api/v1/masters/procedures/:id DeleteProcedure\n" +
		"DELETE /api/v1/masters/vaccines/:id DeleteVaccine\n" +
		"DELETE /api/v1/medical-records/:id DeleteMedicalRecord\n" +
		"DELETE /api/v1/medical-records/:id/checkups/:checkupId DeleteCheckup\n" +
		"DELETE /api/v1/medical-records/:id/clinical-plan DeleteClinicalPlan\n" +
		"DELETE /api/v1/medical-records/:id/images/:imageId DeleteMedicalRecordImage\n" +
		"DELETE /api/v1/medical-records/:id/prescriptions/:prescriptionId DeletePrescription\n" +
		"DELETE /api/v1/medical-records/:id/treatment-plans/:planId DeleteTreatmentPlanInMedicalRecord\n" +
		"DELETE /api/v1/medical-records/:id/treatments/:treatmentId DeleteTreatment\n" +
		"DELETE /api/v1/medical-records/:id/vitals/:vitalId DeleteVital\n" +
		"DELETE /api/v1/vaccinations/:id DeleteVaccination\n" +
		"GET /api/v1/checkups ListGlobalCheckups\n" +
		"GET /api/v1/checkups/field-results ListPetCheckupResults\n" +
		"GET /api/v1/examinations ListExaminations\n" +
		"GET /api/v1/examinations/:id GetExamination\n" +
		"GET /api/v1/examinations/:id/items ListExaminationItems\n" +
		"GET /api/v1/examinations/:id/print-snapshot GetExaminationPrintSnapshot\n" +
		"GET /api/v1/hospitalizations ListHospitalizations\n" +
		"GET /api/v1/hospitalizations/:id GetHospitalization\n" +
		"GET /api/v1/hospitalizations/:id/care-plan-items ListCarePlanItems\n" +
		"GET /api/v1/hospitalizations/:id/daily-records ListDailyRecords\n" +
		"GET /api/v1/hospitalizations/:id/daily-records/:date GetDailyRecord\n" +
		"GET /api/v1/hospitalizations/:id/treatment-plans ListTreatmentPlansByHospitalization\n" +
		"GET /api/v1/lab-device-item-masters ListLabDeviceItemMasters\n" +
		"GET /api/v1/lab-device/agent-consumer GetLabDeviceAgentConsumer\n" +
		"GET /api/v1/lab-device/board GetLabDeviceBoard\n" +
		"GET /api/v1/lab-device/station GetLabDeviceStation\n" +
		"GET /api/v1/lab-device/unlinked GetLabDeviceUnlinked\n" +
		"GET /api/v1/lab-devices ListLabDevices\n" +
		"GET /api/v1/lab-imports/:job_id GetLabImportJob\n" +
		"GET /api/v1/lab-imports/:job_id/events ListLabImportEvents\n" +
		"GET /api/v1/lab-reports/exams/:exam_id GetLabExamReport\n" +
		"GET /api/v1/lab-reports/jobs/:job_id/summaries GetLabJobReportSummaries\n" +
		"GET /api/v1/masters/cages ListCages\n" +
		"GET /api/v1/masters/cages/:id GetCage\n" +
		"GET /api/v1/masters/checkup-types ListCheckupTypes\n" +
		"GET /api/v1/masters/checkup-types/:id GetCheckupType\n" +
		"GET /api/v1/masters/checkup-types/:id/fields ListCheckupTypeFields\n" +
		"GET /api/v1/masters/chief-complaint-types ListChiefComplaints\n" +
		"GET /api/v1/masters/chief-complaint-types/:id GetChiefComplaint\n" +
		"GET /api/v1/masters/consultations ListConsultations\n" +
		"GET /api/v1/masters/consultations/:id GetConsultation\n" +
		"GET /api/v1/masters/diagnosis-names ListDiagnosisNames\n" +
		"GET /api/v1/masters/diagnosis-names/:id GetDiagnosisName\n" +
		"GET /api/v1/masters/diagnosis-names/all ListDiagnosisNamesAll\n" +
		"GET /api/v1/masters/diagnosis-types ListDiagnosisTypes\n" +
		"GET /api/v1/masters/diagnosis-types/:id GetDiagnosisType\n" +
		"GET /api/v1/masters/examination-types ListExaminationTypes\n" +
		"GET /api/v1/masters/examination-types/:id GetExaminationType\n" +
		"GET /api/v1/masters/hospitalization-plans ListHospitalizationPlans\n" +
		"GET /api/v1/masters/hospitalization-plans/:id GetHospitalizationPlan\n" +
		"GET /api/v1/masters/inquiry-templates ListInquiryTemplates\n" +
		"GET /api/v1/masters/inquiry-templates/:id GetInquiryTemplate\n" +
		"GET /api/v1/masters/medicines ListMedicines\n" +
		"GET /api/v1/masters/medicines/:id GetMedicine\n" +
		"GET /api/v1/masters/medicines/:id/dose-params ListMedicineDoseParams\n" +
		"GET /api/v1/masters/procedures ListProcedures\n" +
		"GET /api/v1/masters/procedures/:id GetProcedure\n" +
		"GET /api/v1/masters/vaccines ListVaccines\n" +
		"GET /api/v1/masters/vaccines/:id GetVaccine\n" +
		"GET /api/v1/medical-records ListMedicalRecords\n" +
		"GET /api/v1/medical-records/:id GetMedicalRecord\n" +
		"GET /api/v1/medical-records/:id/addenda ListMedicalRecordAddenda\n" +
		"GET /api/v1/medical-records/:id/checkups ListCheckups\n" +
		"GET /api/v1/medical-records/:id/checkups/:checkupId/field-results ListCheckupFieldResults\n" +
		"GET /api/v1/medical-records/:id/clinical-plan GetClinicalPlan\n" +
		"GET /api/v1/medical-records/:id/images ListMedicalRecordImages\n" +
		"GET /api/v1/medical-records/:id/prescriptions ListPrescriptions\n" +
		"GET /api/v1/medical-records/:id/treatment-plans ListTreatmentPlansByMedicalRecord\n" +
		"GET /api/v1/medical-records/:id/treatments ListTreatments\n" +
		"GET /api/v1/medical-records/:id/vitals ListVitals\n" +
		"GET /api/v1/pets/:id/treatment-history ListPetTreatmentHistory\n" +
		"GET /api/v1/vaccinations ListVaccinations\n" +
		"GET /api/v1/vaccinations/:id GetVaccination\n" +
		"PATCH /api/v1/examinations/:id UpdateExamination\n" +
		"PATCH /api/v1/hospitalizations/:id UpdateHospitalization\n" +
		"PATCH /api/v1/hospitalizations/:id/care-plan-items/:itemId UpdateCarePlanItem\n" +
		"PATCH /api/v1/hospitalizations/:id/treatment-plans/:planId UpdateTreatmentPlanInHospitalization\n" +
		"PATCH /api/v1/lab-device-item-masters/:id UpdateLabDeviceItemMaster\n" +
		"PATCH /api/v1/lab-devices/:id UpdateLabDevice\n" +
		"PATCH /api/v1/masters/cages/:id UpdateCage\n" +
		"PATCH /api/v1/masters/cages/reorder ReorderCages\n" +
		"PATCH /api/v1/masters/checkup-types/:id UpdateCheckupType\n" +
		"PATCH /api/v1/masters/checkup-types/reorder ReorderCheckupTypes\n" +
		"PATCH /api/v1/masters/chief-complaint-types/:id UpdateChiefComplaint\n" +
		"PATCH /api/v1/masters/chief-complaint-types/reorder ReorderChiefComplaints\n" +
		"PATCH /api/v1/masters/consultations/:id UpdateConsultation\n" +
		"PATCH /api/v1/masters/consultations/reorder ReorderConsultations\n" +
		"PATCH /api/v1/masters/diagnosis-names/:id UpdateDiagnosisName\n" +
		"PATCH /api/v1/masters/diagnosis-names/reorder ReorderDiagnosisNames\n" +
		"PATCH /api/v1/masters/diagnosis-types/:id UpdateDiagnosisType\n" +
		"PATCH /api/v1/masters/diagnosis-types/reorder ReorderDiagnosisTypes\n" +
		"PATCH /api/v1/masters/examination-types/:id UpdateExaminationType\n" +
		"PATCH /api/v1/masters/examination-types/:id/fields/:fieldId UpdateExaminationTypeField\n" +
		"PATCH /api/v1/masters/examination-types/:id/fields/reorder ReorderExaminationTypeFields\n" +
		"PATCH /api/v1/masters/examination-types/reorder ReorderExaminationTypes\n" +
		"PATCH /api/v1/masters/hospitalization-plans/:id UpdateHospitalizationPlan\n" +
		"PATCH /api/v1/masters/hospitalization-plans/reorder ReorderHospitalizationPlans\n" +
		"PATCH /api/v1/masters/inquiry-templates/:id UpdateInquiryTemplate\n" +
		"PATCH /api/v1/masters/inquiry-templates/reorder ReorderInquiryTemplates\n" +
		"PATCH /api/v1/masters/medicines/:id UpdateMedicine\n" +
		"PATCH /api/v1/masters/medicines/reorder ReorderMedicines\n" +
		"PATCH /api/v1/masters/procedures/:id UpdateProcedure\n" +
		"PATCH /api/v1/masters/procedures/reorder ReorderProcedures\n" +
		"PATCH /api/v1/masters/vaccines/:id UpdateVaccine\n" +
		"PATCH /api/v1/masters/vaccines/reorder ReorderVaccines\n" +
		"PATCH /api/v1/medical-records/:id UpdateMedicalRecord\n" +
		"PATCH /api/v1/medical-records/:id/checkups/:checkupId UpdateCheckup\n" +
		"PATCH /api/v1/medical-records/:id/clinical-plan UpdateClinicalPlan\n" +
		"PATCH /api/v1/medical-records/:id/inquiries UpdateInquiry\n" +
		"PATCH /api/v1/medical-records/:id/prescriptions/:prescriptionId UpdatePrescription\n" +
		"PATCH /api/v1/medical-records/:id/recommendation-reason UpdateMedicalRecordRecommendationReason\n" +
		"PATCH /api/v1/medical-records/:id/treatment-plans/:planId UpdateTreatmentPlanInMedicalRecord\n" +
		"PATCH /api/v1/medical-records/:id/treatments/:treatmentId UpdateTreatment\n" +
		"PATCH /api/v1/medical-records/:id/vitals/:vitalId UpdateVital\n" +
		"PATCH /api/v1/vaccinations/:id UpdateVaccination\n" +
		"POST /api/v1/checkup-package-imports ApplyCheckupPackageImport\n" +
		"POST /api/v1/checkup-package-imports/preview PreviewCheckupPackageImport\n" +
		"POST /api/v1/examinations CreateExamination\n" +
		"POST /api/v1/examinations/:id/unconfirm UnconfirmExamination\n" +
		"POST /api/v1/hospitalizations CreateHospitalization\n" +
		"POST /api/v1/hospitalizations/:id/care-plan-items CreateCarePlanItem\n" +
		"POST /api/v1/hospitalizations/:id/daily-records CreateDailyRecord\n" +
		"POST /api/v1/hospitalizations/:id/daily-records/:date/care-logs AddCareLog\n" +
		"POST /api/v1/hospitalizations/:id/daily-records/:date/staff-notes AddStaffNote\n" +
		"POST /api/v1/hospitalizations/:id/daily-records/:date/vitals AddVitalRecord\n" +
		"POST /api/v1/hospitalizations/:id/discharge-with-billing DischargeWithBilling\n" +
		"POST /api/v1/hospitalizations/:id/treatment-plans CreateTreatmentPlanForHospitalization\n" +
		"POST /api/v1/lab-device-item-masters/ensure EnsureLabDeviceItemMasters\n" +
		"POST /api/v1/lab-device/frames ReceiveLabDeviceFrames\n" +
		"POST /api/v1/lab-devices CreateLabDevice\n" +
		"POST /api/v1/lab-imports CommitLabImport\n" +
		"POST /api/v1/lab-imports/:job_id/attach AttachLabDeviceJob\n" +
		"POST /api/v1/lab-imports/:job_id/detach DetachLabDeviceJob\n" +
		"POST /api/v1/lab-imports/:job_id/revert RevertLabImport\n" +
		"POST /api/v1/lab-imports/preview PreviewLabImport\n" +
		"POST /api/v1/masters/cages CreateCage\n" +
		"POST /api/v1/masters/checkup-types CreateCheckupType\n" +
		"POST /api/v1/masters/chief-complaint-types CreateChiefComplaint\n" +
		"POST /api/v1/masters/consultations CreateConsultation\n" +
		"POST /api/v1/masters/diagnosis-names CreateDiagnosisName\n" +
		"POST /api/v1/masters/diagnosis-types CreateDiagnosisType\n" +
		"POST /api/v1/masters/examination-types CreateExaminationType\n" +
		"POST /api/v1/masters/examination-types/:id/fields CreateExaminationTypeField\n" +
		"POST /api/v1/masters/hospitalization-plans CreateHospitalizationPlan\n" +
		"POST /api/v1/masters/inquiry-templates CreateInquiryTemplate\n" +
		"POST /api/v1/masters/medicines CreateMedicine\n" +
		"POST /api/v1/masters/procedures CreateProcedure\n" +
		"POST /api/v1/masters/vaccines CreateVaccine\n" +
		"POST /api/v1/medical-records CreateMedicalRecord\n" +
		"POST /api/v1/medical-records/:id/addenda CreateMedicalRecordAddendum\n" +
		"POST /api/v1/medical-records/:id/checkups CreateCheckup\n" +
		"POST /api/v1/medical-records/:id/images CreateMedicalRecordImage\n" +
		"POST /api/v1/medical-records/:id/images/upload UploadMedicalRecordImage\n" +
		"POST /api/v1/medical-records/:id/prescriptions CreatePrescription\n" +
		"POST /api/v1/medical-records/:id/treatment-plans CreateTreatmentPlanForMedicalRecord\n" +
		"POST /api/v1/medical-records/:id/treatments CreateTreatment\n" +
		"POST /api/v1/medical-records/:id/vitals CreateVital\n" +
		"POST /api/v1/vaccinations CreateVaccination\n" +
		"PUT /api/v1/examinations/:id/items ReplaceExaminationItems\n" +
		"PUT /api/v1/lab-device/station PutLabDeviceStation\n" +
		"PUT /api/v1/lab-device/wait PutLabDeviceWait\n" +
		"PUT /api/v1/lab-devices/:id/configuration SaveLabDeviceConfiguration\n" +
		"PUT /api/v1/masters/examination-types/:id/fields/:fieldId/reference-ranges ReplaceExaminationTypeFieldReferenceRanges\n" +
		"PUT /api/v1/masters/medicines/:id/dose-params/:species UpsertMedicineDoseParam\n" +
		"PUT /api/v1/medical-records/:id/checkups/:checkupId/field-results ReplaceCheckupFieldResults\n" +
		"PUT /api/v1/medical-records/:id/treatments BulkUpdateTreatments\n"

	assert.Equal(t, want, got, "medicalrecord route snapshot drifted from the pre-move baseline "+
		"(internal/handler/testdata/route_snapshot.golden, before these 157 lines were removed)")
}

// TestRegisterRoutes_HospitalizationDischargePermissions locks BUG-457 product decision:
// accounting-bearing discharge (POST discharge-with-billing) and no-accounting discharge via
// generic update (PATCH /:id) both require hospitalization:edit — not delete.
// Permission middleware aborts before nil-backed terminal handlers so ServeHTTP never panics.
// Snapshot TestRegisterRoutes_Snapshot cannot observe middleware; this HTTP-driven spy does.
func TestRegisterRoutes_HospitalizationDischargeAndExaminationUnconfirmPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var observed []string
	permSpy := func(resource, action string) gin.HandlerFunc {
		return func(c *gin.Context) {
			observed = append(observed, resource+":"+action)
			c.AbortWithStatus(http.StatusOK)
		}
	}

	h := NewHandler(
		NewDiagnosisHandler(nil, nil),
		NewExamTypeHandler(nil),
		NewChiefComplaintHandler(nil),
		NewCheckupHandler(nil, nil),
		NewCheckupTypeHandler(nil),
		NewVaccineHandler(nil),
		NewVaccinationHandler(nil),
		NewPrescriptionHandler(nil),
		NewInquiryHandler(nil),
		NewInquiryTemplateHandler(nil),
		NewLabImportHandler(nil, nil, nil),
		NewLabReportHandler(nil),
		NewVitalHandler(nil, nil),
		NewClinicalPlanHandler(nil),
		NewMedicalRecordImageHandler(nil, nil, nil),
		NewTreatmentHandler(nil, nil),
		NewHospitalizationHandler(nil, nil),
		NewHospitalizationPlanHandler(nil),
		NewDailyRecordHandler(nil),
		NewCarePlanItemHandler(nil),
		NewConsultationHandler(nil),
		NewProcedureHandler(nil),
		NewMedicineHandler(nil),
		NewMedicineDoseParamHandler(nil),
		NewCageHandler(nil),
		NewTreatmentPlanHandler(nil, nil, nil, nil),
		NewMedicalRecordHandler(nil),
		NewMedicalRecordAddendumHandler(nil),
		NewExaminationHandler(nil),
		NewCheckupPackageImportHandler(nil),
		permSpy,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	// FE accounting-yes branch → POST discharge-with-billing
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/hospitalizations/1/discharge-with-billing", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// FE accounting-no branch → generic PATCH (status discharged)
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/hospitalizations/1", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/examinations/1/unconfirm", http.NoBody)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Assert resource AND action — action-only would pass if another resource's edit were swapped in.
	assert.Equal(t, []string{
		"hospitalization:edit",
		"hospitalization:edit",
		"examination-unconfirm:edit",
	}, observed)
}

// lastHandlerSegment mirrors internal/handler/handler_routes_snapshot_test.go's helper of the
// same name: gin's RouteInfo.Handler returns a fully-qualified function name
// (".../medicalrecord.(*DiagnosisHandler).GetDiagnosisType-fm"); only the trailing method
// name is stable across environments (GOPATH/module cache absolute paths are not).
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
