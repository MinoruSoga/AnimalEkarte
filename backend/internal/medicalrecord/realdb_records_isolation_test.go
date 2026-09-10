package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type realDBRecordsFixture struct {
	fx       testdb.ClinicGrantFixture
	ownerA   *model.Owner
	ownerB   *model.Owner
	petA     *model.Pet
	petB     *model.Pet
	recordA  *model.MedicalRecord
	recordB  *model.MedicalRecord
	checkupA *model.Checkup
	checkupB *model.Checkup
	vaccineA *model.Vaccine
	vaccineB *model.Vaccine
	vaccA    *model.Vaccination
	vaccB    *model.Vaccination

	mrH          *MedicalRecordHandler
	addendumH    *MedicalRecordAddendumHandler
	checkupH     *CheckupHandler
	clinicalH    *ClinicalPlanHandler
	imageH       *MedicalRecordImageHandler
	rxH          *PrescriptionHandler
	treatmentH   *TreatmentHandler
	planH        *TreatmentPlanHandler
	vitalH       *VitalHandler
	vaccinationH *VaccinationHandler
}

func setupRealDBRecordsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.MedicalRecord{},
		&model.ClinicalPlan{}, &model.MedicalRecordAddendum{}, &model.MedicalRecordImage{},
		&model.Prescription{}, &model.Treatment{}, &model.TreatmentPlan{}, &model.VitalRecord{},
		&model.CheckupType{}, &model.Checkup{}, &model.CheckupTypeField{}, &model.CheckupFieldResult{},
		&model.Vaccine{}, &model.Vaccination{},
	))
	testdb.Truncate(t, db,
		"checkup_field_results", "checkups", "checkup_type_fields", "checkup_types",
		"vaccinations", "vaccines", "vital_records", "treatments", "treatment_plans",
		"prescriptions", "medical_record_images", "medical_record_addenda", "clinical_plans",
		"medical_records", "pets", "owners",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

type mrListGuard struct {
	MedicalRecordService
	g clinicIDGuard
}

func (s *mrListGuard) List(ctx context.Context, clinicIDs []uint64, filters MedicalRecordListFilters, page, limit int) ([]model.MedicalRecord, int64, error) {
	for _, id := range clinicIDs {
		s.g.check(id, "medicalRecord List")
	}
	return s.MedicalRecordService.List(ctx, clinicIDs, filters, page, limit)
}

func (s *mrListGuard) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.MedicalRecord, error) {
	for _, cid := range clinicIDs {
		s.g.check(cid, "medicalRecord GetByIDForClinics")
	}
	return s.MedicalRecordService.GetByIDForClinics(ctx, clinicIDs, id)
}

func wireRecordsHandlers(t *testing.T, db *gorm.DB, forbidden uint64) realDBRecordsFixture {
	t.Helper()
	tx := persistenceTx(db)
	mrRepo := NewMedicalRecordRepository(db)
	mrSvc := newMedicalRecordReadService(db)
	if forbidden != 0 {
		mrSvc = &mrListGuard{MedicalRecordService: mrSvc, g: clinicIDGuard{t: t, forbiddenClinicID: forbidden}}
	}

	addendumSvc := MedicalRecordAddendumService(NewMedicalRecordAddendumService(NewMedicalRecordAddendumRepository(db), mrRepo, nil, tx))
	imageSvc := MedicalRecordImageService(NewMedicalRecordImageService(NewMedicalRecordImageRepository(db), mrRepo, tx))
	rxSvc := PrescriptionService(NewPrescriptionService(NewPrescriptionRepository(db), mrRepo, nil, tx))
	treatmentSvc := TreatmentService(NewTreatmentServiceWithAudit(
		NewTreatmentRepository(db), mrRepo, nil, nil, nil, nil, NewVitalRepository(db), nil, tx, nil,
	))
	planSvc := TreatmentPlanService(NewTreatmentPlanService(NewTreatmentPlanRepository(db), tx))
	vitalSvc := VitalService(NewVitalService(NewVitalRepository(db), mrRepo, nil, tx))
	checkupSvc := CheckupService(NewCheckupService(NewCheckupRepository(db), mrRepo, NewCheckupTypeRepository(db), nil, nil))
	fieldSvc := CheckupFieldResultService(NewCheckupFieldResultService(
		NewCheckupRepository(db), mrRepo, NewCheckupTypeFieldRepository(db), NewCheckupFieldResultRepository(db), nil, tx,
	))
	clinicalSvc := ClinicalPlanService(NewClinicalPlanService(NewClinicalPlanRepository(db), mrRepo, nil, nil, tx, nil))
	vaccSvc := VaccinationService(NewVaccinationService(NewVaccinationRepository(db), NewVaccineRepository(db), nil, nil, mrRepo, tx))

	if forbidden != 0 {
		g := clinicIDGuard{t: t, forbiddenClinicID: forbidden}
		addendumSvc = &addendumQueryGuard{MedicalRecordAddendumService: addendumSvc, g: g}
		imageSvc = &imageQueryGuard{MedicalRecordImageService: imageSvc, g: g}
		rxSvc = &rxQueryGuard{PrescriptionService: rxSvc, g: g}
		treatmentSvc = &treatmentQueryGuard{TreatmentService: treatmentSvc, g: g}
		planSvc = &treatmentPlanQueryGuard{TreatmentPlanService: planSvc, g: g}
		vitalSvc = &vitalQueryGuard{VitalService: vitalSvc, g: g}
		checkupSvc = &checkupQueryGuard{CheckupService: checkupSvc, g: g}
		fieldSvc = &checkupFieldQueryGuard{CheckupFieldResultService: fieldSvc, g: g}
		clinicalSvc = &clinicalQueryGuard{ClinicalPlanService: clinicalSvc, g: g}
		vaccSvc = &vaccinationQueryGuard{VaccinationService: vaccSvc, g: g}
	}

	return realDBRecordsFixture{
		mrH:          NewMedicalRecordHandler(mrSvc),
		addendumH:    NewMedicalRecordAddendumHandler(addendumSvc),
		checkupH:     NewCheckupHandler(checkupSvc, fieldSvc),
		clinicalH:    NewClinicalPlanHandler(clinicalSvc),
		imageH:       NewMedicalRecordImageHandler(imageSvc, mrSvc, nil),
		rxH:          NewPrescriptionHandler(rxSvc),
		treatmentH:   NewTreatmentHandler(treatmentSvc, nil),
		planH:        NewTreatmentPlanHandler(planSvc, nil, mrSvc, nil),
		vitalH:       NewVitalHandler(vitalSvc, mrSvc),
		vaccinationH: NewVaccinationHandler(vaccSvc),
	}
}

type addendumQueryGuard struct {
	MedicalRecordAddendumService
	g clinicIDGuard
}

func (s *addendumQueryGuard) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	s.g.check(clinicID, "addendum FindByMedicalRecordID")
	return s.MedicalRecordAddendumService.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
}

type imageQueryGuard struct {
	MedicalRecordImageService
	g clinicIDGuard
}

func (s *imageQueryGuard) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	s.g.check(clinicID, "image List")
	return s.MedicalRecordImageService.List(ctx, clinicID, medicalRecordID)
}

type rxQueryGuard struct {
	PrescriptionService
	g clinicIDGuard
}

func (s *rxQueryGuard) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Prescription, error) {
	s.g.check(clinicID, "rx List")
	return s.PrescriptionService.List(ctx, clinicID, medicalRecordID)
}

type treatmentQueryGuard struct {
	TreatmentService
	g clinicIDGuard
}

func (s *treatmentQueryGuard) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	s.g.check(clinicID, "treatment List")
	return s.TreatmentService.List(ctx, clinicID, medicalRecordID)
}
func (s *treatmentQueryGuard) ListPetHistory(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	s.g.check(clinicID, "treatment ListPetHistory")
	return s.TreatmentService.ListPetHistory(ctx, clinicID, petID, filter, page, limit)
}

type treatmentPlanQueryGuard struct {
	TreatmentPlanService
	g clinicIDGuard
}

func (s *treatmentPlanQueryGuard) ListByMedicalRecord(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error) {
	s.g.check(clinicID, "treatmentPlan ListByMedicalRecord")
	return s.TreatmentPlanService.ListByMedicalRecord(ctx, clinicID, medicalRecordID)
}

type vitalQueryGuard struct {
	VitalService
	g clinicIDGuard
}

func (s *vitalQueryGuard) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	s.g.check(clinicID, "vital List")
	return s.VitalService.List(ctx, clinicID, medicalRecordID)
}

type checkupQueryGuard struct {
	CheckupService
	g clinicIDGuard
}

func (s *checkupQueryGuard) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	s.g.check(clinicID, "checkup List")
	return s.CheckupService.List(ctx, clinicID, medicalRecordID)
}
func (s *checkupQueryGuard) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
	s.g.check(input.ClinicID, "checkup ListByClinic")
	return s.CheckupService.ListByClinic(ctx, input)
}

type clinicalQueryGuard struct {
	ClinicalPlanService
	g clinicIDGuard
}

func (s *clinicalQueryGuard) GetOrCreate(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	s.g.check(clinicID, "clinical GetOrCreate")
	return s.ClinicalPlanService.GetOrCreate(ctx, clinicID, medicalRecordID)
}

type vaccinationQueryGuard struct {
	VaccinationService
	g clinicIDGuard
}

func (s *vaccinationQueryGuard) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error) {
	s.g.check(clinicID, "vaccination List")
	return s.VaccinationService.List(ctx, clinicID, petID, ownerID, startDate, endDate, search, page, limit)
}
func (s *vaccinationQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	s.g.check(clinicID, "vaccination GetByID")
	return s.VaccinationService.GetByID(ctx, clinicID, id)
}

func seedRealDBRecordsFixture(t *testing.T, db *gorm.DB, forbidden uint64) realDBRecordsFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr records realDB")
	ctx := context.Background()

	ownerA, petA, recordA := seedOwnerPetRecord(t, db, fx.ClinicA, realDBOwnerNameA, realDBPetNameA, realDBRecordNoA)
	ownerB, petB, recordB := seedOwnerPetRecord(t, db, fx.ClinicB, realDBOwnerNameB, realDBPetNameB, realDBRecordNoB)

	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicalPlan{MedicalRecordID: recordA.ID, PhysicalExam: realDBClinicalPlanA}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicalPlan{MedicalRecordID: recordB.ID, PhysicalExam: realDBClinicalPlanB}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecordAddendum{
		MedicalRecordID: recordA.ID, ClinicID: fx.ClinicA, AuthorUserID: fx.StaffID, AfterText: realDBAddendumA, Reason: "d3-a",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecordAddendum{
		MedicalRecordID: recordB.ID, ClinicID: fx.ClinicB, AuthorUserID: fx.StaffID, AfterText: realDBAddendumB, Reason: "d3-b",
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecordImage{
		MedicalRecordID: recordA.ID, ImageURL: realDBImageURLA, ImageType: model.MedicalImageTypePhoto, FileName: "a.png",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.MedicalRecordImage{
		MedicalRecordID: recordB.ID, ImageURL: realDBImageURLB, ImageType: model.MedicalImageTypePhoto, FileName: "b.png",
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.Prescription{
		ClinicID: fx.ClinicA, OwnerID: ownerA.ID, PetID: &petA.ID, MedicalRecordID: &recordA.ID, PrescribedAt: realDBNow(),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Prescription{
		ClinicID: fx.ClinicB, OwnerID: ownerB.ID, PetID: &petB.ID, MedicalRecordID: &recordB.ID, PrescribedAt: realDBNow(),
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.Treatment{
		MedicalRecordID: recordA.ID, ItemType: model.TreatmentItemTypeOther, Content: realDBTreatmentA,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.Treatment{
		MedicalRecordID: recordB.ID, ItemType: model.TreatmentItemTypeOther, Content: realDBTreatmentB,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&model.TreatmentPlan{
		ClinicID: fx.ClinicA, MedicalRecordID: &recordA.ID, TreatmentContent: realDBTreatmentPlanA,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.TreatmentPlan{
		ClinicID: fx.ClinicB, MedicalRecordID: &recordB.ID, TreatmentContent: realDBTreatmentPlanB,
	}).Error)

	tempA, tempB := 38.5, 39.1
	recordAID, recordBID := recordA.ID, recordB.ID
	require.NoError(t, db.WithContext(ctx).Create(&model.VitalRecord{
		ClinicID: fx.ClinicA, MedicalRecordID: &recordAID, PetID: petA.ID, Temperature: &tempA, Notes: realDBVitalNotesA, RecordedAt: realDBNow(),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.VitalRecord{
		ClinicID: fx.ClinicB, MedicalRecordID: &recordBID, PetID: petB.ID, Temperature: &tempB, Notes: realDBVitalNotesB, RecordedAt: realDBNow(),
	}).Error)

	ctA := &model.CheckupType{ClinicID: fx.ClinicA, Name: realDBCheckupTypeA, IsActive: true}
	ctB := &model.CheckupType{ClinicID: fx.ClinicB, Name: realDBCheckupTypeB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(ctA).Error)
	require.NoError(t, db.WithContext(ctx).Create(ctB).Error)
	checkupA := &model.Checkup{ClinicID: fx.ClinicA, MedicalRecordID: recordA.ID, PetID: &petA.ID, CheckupTypeID: ctA.ID, Date: realDBNow(), Result: realDBCheckupFieldA}
	checkupB := &model.Checkup{ClinicID: fx.ClinicB, MedicalRecordID: recordB.ID, PetID: &petB.ID, CheckupTypeID: ctB.ID, Date: realDBNow(), Result: realDBCheckupFieldB}
	require.NoError(t, db.WithContext(ctx).Create(checkupA).Error)
	require.NoError(t, db.WithContext(ctx).Create(checkupB).Error)
	fieldB := ensureCheckupFieldOptions(&model.CheckupTypeField{
		ClinicID: fx.ClinicB, CheckupTypeID: ctB.ID, Name: realDBCheckupFieldB, FieldType: model.CheckupFieldTypeText,
	})
	require.NoError(t, db.WithContext(ctx).Create(fieldB).Error)
	fieldBID := fieldB.ID
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: fx.ClinicB, CheckupID: checkupB.ID, CheckupTypeFieldID: &fieldBID, FieldName: realDBCheckupFieldB, FieldType: model.CheckupFieldTypeText, ValueText: "ok-B",
	}).Error)

	vaccineA := &model.Vaccine{ClinicID: fx.ClinicA, Name: realDBVaccineA, IsActive: true}
	vaccineB := &model.Vaccine{ClinicID: fx.ClinicB, Name: realDBVaccineB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(vaccineA).Error)
	require.NoError(t, db.WithContext(ctx).Create(vaccineB).Error)
	vaccA := &model.Vaccination{ClinicID: fx.ClinicA, MedicalRecordID: &recordA.ID, PetID: &petA.ID, VaccineID: vaccineA.ID, Date: realDBNow(), Lot1: realDBVaccLotA}
	vaccB := &model.Vaccination{ClinicID: fx.ClinicB, MedicalRecordID: &recordB.ID, PetID: &petB.ID, VaccineID: vaccineB.ID, Date: realDBNow(), Lot1: realDBVaccLotB}
	require.NoError(t, db.WithContext(ctx).Create(vaccA).Error)
	require.NoError(t, db.WithContext(ctx).Create(vaccB).Error)

	handlers := wireRecordsHandlers(t, db, forbidden)
	handlers.fx = fx
	handlers.ownerA, handlers.ownerB = ownerA, ownerB
	handlers.petA, handlers.petB = petA, petB
	handlers.recordA, handlers.recordB = recordA, recordB
	handlers.checkupA, handlers.checkupB = checkupA, checkupB
	handlers.vaccineA, handlers.vaccineB = vaccineA, vaccineB
	handlers.vaccA, handlers.vaccB = vaccA, vaccB
	return handlers
}

func TestRealDB_RecordsSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resMR := string(model.ResourceMedicalRecords)
	resCheckups := string(model.ResourceCheckups)
	resVacc := string(model.ResourceVaccinations)

	run := func(name string, guarded bool, fn func(*testing.T, realDBRecordsFixture)) {
		t.Run(name, func(t *testing.T) {
			db := setupRealDBRecordsDB(t)
			forbidden := uint64(0)
			if guarded {
				// seed first without guard, then rebuild handlers with guard using clinic B
				fx := seedRealDBRecordsFixture(t, db, 0)
				handlers := wireRecordsHandlers(t, db, fx.fx.ClinicB)
				fx.mrH, fx.addendumH, fx.checkupH, fx.clinicalH = handlers.mrH, handlers.addendumH, handlers.checkupH, handlers.clinicalH
				fx.imageH, fx.rxH, fx.treatmentH, fx.planH = handlers.imageH, handlers.rxH, handlers.treatmentH, handlers.planH
				fx.vitalH, fx.vaccinationH = handlers.vitalH, handlers.vaccinationH
				fn(t, fx)
				return
			}
			_ = forbidden
			fx := seedRealDBRecordsFixture(t, db, 0)
			fn(t, fx)
		})
	}

	// ---- cross-clinic list/detail ----
	run("cross_list_grantA_selectedB_default_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/medical-records?page=1&limit=10", configureGrant(fx.fx, resMR, fx.fx.ClinicA))
		fx.mrH.ListMedicalRecords(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBRecordNoA)
		assert.NotContains(t, w.Body.String(), realDBRecordNoB)
	})
	run("cross_list_grantA_clinic_ids_AB_returns_A", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records?page=1&limit=10&clinic_ids=%d,%d", fx.fx.ClinicA, fx.fx.ClinicB), configureGrant(fx.fx, resMR, fx.fx.ClinicA))
		fx.mrH.ListMedicalRecords(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBRecordNoA)
		assert.NotContains(t, w.Body.String(), realDBRecordNoB)
	})
	run("cross_list_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/medical-records?page=1&limit=10", configureGrant(fx.fx, resMR, fx.fx.ClinicB))
		fx.mrH.ListMedicalRecords(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBRecordNoB)
		assert.NotContains(t, w.Body.String(), realDBRecordNoA)
	})
	run("cross_get_grantA_returns_A", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d", fx.recordA.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicA), fx.recordA.ID))
		fx.mrH.GetMedicalRecord(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBRecordNoA)
		assert.NotContains(t, w.Body.String(), realDBRecordNoB)
	})
	run("cross_get_grantB_A_id_404", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d", fx.recordA.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.recordA.ID))
		fx.mrH.GetMedicalRecord(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBRecordNoA)
	})
	run("cross_get_grantB_ok", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d", fx.recordB.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.recordB.ID))
		fx.mrH.GetMedicalRecord(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBRecordNoB)
		assert.NotContains(t, w.Body.String(), realDBRecordNoA)
	})

	nested := []struct {
		name   string
		path   func(realDBRecordsFixture) string
		invoke func(realDBRecordsFixture, *gin.Context)
		marker string
		forbid string
	}{
		{"addenda", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/addenda", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.addendumH.ListMedicalRecordAddenda(c) }, realDBAddendumB, realDBAddendumA},
		{"checkups", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/checkups", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.checkupH.ListCheckups(c) }, realDBCheckupFieldB, realDBCheckupFieldA},
		{"clinical_plan", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/clinical-plan", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.clinicalH.GetClinicalPlan(c) }, realDBClinicalPlanB, realDBClinicalPlanA},
		{"images", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/images", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.imageH.ListMedicalRecordImages(c) }, realDBImageURLB, realDBImageURLA},
		{"prescriptions", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/prescriptions", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.rxH.ListPrescriptions(c) }, fmt.Sprintf(`"owner_id":%d`, 0), ""}, // filled later
		{"treatment_plans", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/treatment-plans", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.planH.ListTreatmentPlansByMedicalRecord(c) }, realDBTreatmentPlanB, realDBTreatmentPlanA},
		{"treatments", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/treatments", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.treatmentH.ListTreatments(c) }, realDBTreatmentB, realDBTreatmentA},
		{"vitals", func(fx realDBRecordsFixture) string {
			return fmt.Sprintf("/api/v1/medical-records/%d/vitals", fx.recordB.ID)
		}, func(fx realDBRecordsFixture, c *gin.Context) { fx.vitalH.ListVitals(c) }, realDBVitalNotesB, realDBVitalNotesA},
	}

	for _, nc := range nested {
		nc := nc
		run(nc.name+"_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, nc.path(fx), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicA), fx.recordB.ID))
			nc.invoke(fx, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
			if nc.forbid != "" {
				assert.NotContains(t, w.Body.String(), nc.forbid)
			}
			assert.NotContains(t, w.Body.String(), nc.marker)
		})
		run(nc.name+"_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
			path := nc.path(fx)
			marker := nc.marker
			if nc.name == "prescriptions" {
				marker = fmt.Sprintf(`"owner_id":"%d"`, fx.ownerB.ID)
			}
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.recordB.ID))
			nc.invoke(fx, c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), marker)
			if nc.forbid != "" {
				assert.NotContains(t, w.Body.String(), nc.forbid)
			}
		})
		run(nc.name+"_A_id_no_leak", false, func(t *testing.T, fx realDBRecordsFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, stringsReplace(nc.path(fx), fx.recordB.ID, fx.recordA.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.recordA.ID))
			nc.invoke(fx, c)
			switch nc.name {
			case "checkups", "prescriptions", "treatments":
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				body := w.Body.String()
				assert.NotContains(t, body, realDBRecordNoA)
				assert.True(t, body == "[]" || body == "null" || len(w.Body.Bytes()) <= 2)
			default:
				require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
			}
			if nc.forbid != "" {
				assert.NotContains(t, w.Body.String(), nc.forbid)
			}
			assert.NotContains(t, w.Body.String(), nc.marker)
		})
	}

	// field-results under checkup
	run("checkup_field_results_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d/checkups/%d/field-results", fx.recordB.ID, fx.checkupB.ID), withParams(configureGrant(fx.fx, resMR, fx.fx.ClinicA), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.recordB.ID)}, gin.Param{Key: "checkupId", Value: fmt.Sprintf("%d", fx.checkupB.ID)}))
		fx.checkupH.ListCheckupFieldResults(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("checkup_field_results_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d/checkups/%d/field-results", fx.recordB.ID, fx.checkupB.ID), withParams(configureGrant(fx.fx, resMR, fx.fx.ClinicB), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.recordB.ID)}, gin.Param{Key: "checkupId", Value: fmt.Sprintf("%d", fx.checkupB.ID)}))
		fx.checkupH.ListCheckupFieldResults(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBCheckupFieldB)
	})
	run("checkup_field_results_A_id_404", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-records/%d/checkups/%d/field-results", fx.recordA.ID, fx.checkupA.ID), withParams(configureGrant(fx.fx, resMR, fx.fx.ClinicB), gin.Param{Key: "id", Value: fmt.Sprintf("%d", fx.recordA.ID)}, gin.Param{Key: "checkupId", Value: fmt.Sprintf("%d", fx.checkupA.ID)}))
		fx.checkupH.ListCheckupFieldResults(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	// global checkups + pet field-results
	run("global_checkups_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/checkups?page=1&limit=10", configureGrant(fx.fx, resCheckups, fx.fx.ClinicA))
		fx.checkupH.ListGlobalCheckups(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("global_checkups_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/checkups?page=1&limit=10", configureGrant(fx.fx, resCheckups, fx.fx.ClinicB))
		fx.checkupH.ListGlobalCheckups(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBCheckupFieldB)
		assert.NotContains(t, w.Body.String(), realDBCheckupFieldA)
	})
	run("pet_checkup_results_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/checkups/field-results?pet_id=%d", fx.petB.ID), configureGrant(fx.fx, resCheckups, fx.fx.ClinicA))
		fx.checkupH.ListPetCheckupResults(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("pet_checkup_results_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/checkups/field-results?pet_id=%d", fx.petB.ID), configureGrant(fx.fx, resCheckups, fx.fx.ClinicB))
		fx.checkupH.ListPetCheckupResults(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBCheckupFieldB)
	})

	// pet treatment history
	run("pet_treatment_history_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/pets/%d/treatment-history?page=1&limit=10", fx.petB.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicA), fx.petB.ID))
		fx.treatmentH.ListPetTreatmentHistory(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("pet_treatment_history_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/pets/%d/treatment-history?page=1&limit=10", fx.petB.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.petB.ID))
		fx.treatmentH.ListPetTreatmentHistory(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBTreatmentB)
		assert.NotContains(t, w.Body.String(), realDBTreatmentA)
	})
	run("pet_treatment_history_A_id_empty_no_leak", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/pets/%d/treatment-history?page=1&limit=10", fx.petA.ID), withIDParam(configureGrant(fx.fx, resMR, fx.fx.ClinicB), fx.petA.ID))
		fx.treatmentH.ListPetTreatmentHistory(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBTreatmentA)
		var payload struct {
			Total int             `json:"total"`
			Data  json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err == nil {
			assert.Equal(t, 0, payload.Total)
		}
	})

	// vaccinations
	run("vaccinations_list_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/vaccinations?page=1&limit=10", configureGrant(fx.fx, resVacc, fx.fx.ClinicA))
		fx.vaccinationH.ListVaccinations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("vaccinations_list_grantB_nonempty", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/vaccinations?page=1&limit=10", configureGrant(fx.fx, resVacc, fx.fx.ClinicB))
		fx.vaccinationH.ListVaccinations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBVaccLotB)
		assert.NotContains(t, w.Body.String(), realDBVaccLotA)
	})
	run("vaccinations_get_grantA_403", true, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/vaccinations/%d", fx.vaccB.ID), withIDParam(configureGrant(fx.fx, resVacc, fx.fx.ClinicA), fx.vaccB.ID))
		fx.vaccinationH.GetVaccination(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("vaccinations_get_grantB_ok", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/vaccinations/%d", fx.vaccB.ID), withIDParam(configureGrant(fx.fx, resVacc, fx.fx.ClinicB), fx.vaccB.ID))
		fx.vaccinationH.GetVaccination(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBVaccLotB)
	})
	run("vaccinations_get_A_id_404", false, func(t *testing.T, fx realDBRecordsFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/vaccinations/%d", fx.vaccA.ID), withIDParam(configureGrant(fx.fx, resVacc, fx.fx.ClinicB), fx.vaccA.ID))
		fx.vaccinationH.GetVaccination(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}

func stringsReplace(path string, from, to uint64) string {
	return strings.Replace(path, fmt.Sprintf("%d", from), fmt.Sprintf("%d", to), 1)
}
