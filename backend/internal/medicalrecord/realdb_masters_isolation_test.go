package medicalrecord

import (
	"context"
	"encoding/json"
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

type namedClinicDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Name     string `json:"name"`
}

type titledClinicDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Title    string `json:"title"`
}

type realDBMastersFixture struct {
	fx            testdb.ClinicGrantFixture
	chiefA        *model.ChiefComplaintType
	chiefB        *model.ChiefComplaintType
	consultA      *model.Consultation
	consultB      *model.Consultation
	diagTypeA     *model.DiagnosisType
	diagTypeB     *model.DiagnosisType
	diagNameA     *model.DiagnosisName
	diagNameB     *model.DiagnosisName
	examTypeA     *model.ExaminationType
	examTypeB     *model.ExaminationType
	inquiryA      *model.InquiryTemplate
	inquiryB      *model.InquiryTemplate
	medicineA     *model.Medicine
	medicineB     *model.Medicine
	procedureA    *model.Procedure
	procedureB    *model.Procedure
	vaccineA      *model.Vaccine
	vaccineB      *model.Vaccine
	cageA         *model.Cage
	cageB         *model.Cage
	hospPlanA     *model.HospitalizationPlan
	hospPlanB     *model.HospitalizationPlan
	checkupTypeA  *model.CheckupType
	checkupTypeB  *model.CheckupType
	checkupFieldA *model.CheckupTypeField
	checkupFieldB *model.CheckupTypeField

	chiefH       *ChiefComplaintHandler
	consultH     *ConsultationHandler
	diagH        *DiagnosisHandler
	examTypeH    *ExamTypeHandler
	inquiryH     *InquiryTemplateHandler
	medicineH    *MedicineHandler
	doseH        *MedicineDoseParamHandler
	procedureH   *ProcedureHandler
	vaccineH     *VaccineHandler
	cageH        *CageHandler
	hospPlanH    *HospitalizationPlanHandler
	checkupTypeH *CheckupTypeHandler
	checkupH     *CheckupHandler
}

func setupRealDBMastersDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.ChiefComplaintType{}, &model.Consultation{}, &model.DiagnosisType{}, &model.DiagnosisName{},
		&model.ExaminationType{}, &model.InquiryTemplate{}, &model.Medicine{}, &model.MedicineDoseParam{},
		&model.Procedure{}, &model.Vaccine{}, &model.Cage{}, &model.HospitalizationPlan{},
		&model.CheckupType{}, &model.CheckupTypeField{},
	))
	testdb.Truncate(t, db,
		"medicine_dose_params", "medicines", "procedures", "vaccines", "consultations",
		"chief_complaint_types", "diagnosis_names", "diagnosis_types", "exam_types",
		"inquiry_templates", "cages", "hospitalization_plans",
		"checkup_type_fields", "checkup_types",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

func wireMastersHandlers(t *testing.T, db *gorm.DB, forbidden uint64) (
	*ChiefComplaintHandler, *ConsultationHandler, *DiagnosisHandler, *ExamTypeHandler,
	*InquiryTemplateHandler, *MedicineHandler, *MedicineDoseParamHandler, *ProcedureHandler,
	*VaccineHandler, *CageHandler, *HospitalizationPlanHandler, *CheckupTypeHandler, *CheckupHandler,
) {
	t.Helper()
	tx := persistenceTx(db)
	chiefSvc := ChiefComplaintTypeService(NewChiefComplaintTypeService(NewChiefComplaintTypeRepository(db)))
	consultSvc := ConsultationService(NewConsultationService(NewConsultationRepository(db)))
	diagTypeSvc := DiagnosisTypeService(NewDiagnosisTypeService(NewDiagnosisTypeRepository(db)))
	diagNameSvc := DiagnosisNameService(NewDiagnosisNameService(NewDiagnosisNameRepository(db), NewDiagnosisTypeRepository(db)))
	examTypeSvc := ExaminationTypeService(NewExamTypeService(NewExamTypeRepository(db), tx))
	inquirySvc := InquiryTemplateService(NewInquiryTemplateService(NewInquiryTemplateRepository(db)))
	medRepo := NewMedicineRepository(db)
	medicineSvc := MedicineService(NewMedicineServiceWithAudit(medRepo, nil, nil, nil))
	doseSvc := MedicineDoseParamService(NewMedicineDoseParamService(NewMedicineDoseParamRepository(db), medRepo, tx, nil))
	procedureSvc := ProcedureService(NewProcedureService(NewProcedureRepository(db), tx))
	vaccineSvc := VaccineService(NewVaccineService(NewVaccineRepository(db)))
	cageSvc := CageService(NewCageService(NewCageRepository(db), tx))
	hospPlanSvc := HospitalizationPlanService(NewHospitalizationPlanService(NewHospitalizationPlanRepository(db)))
	checkupTypeSvc := CheckupTypeService(NewCheckupTypeService(NewCheckupTypeRepository(db)))
	fieldResultSvc := CheckupFieldResultService(NewCheckupFieldResultService(
		NewCheckupRepository(db), NewMedicalRecordRepository(db),
		NewCheckupTypeFieldRepository(db), NewCheckupFieldResultRepository(db), nil, tx,
	))

	if forbidden != 0 {
		g := clinicIDGuard{t: t, forbiddenClinicID: forbidden}
		chiefSvc = &chiefQueryGuard{ChiefComplaintTypeService: chiefSvc, g: g}
		consultSvc = &consultQueryGuard{ConsultationService: consultSvc, g: g}
		diagTypeSvc = &diagTypeQueryGuard{DiagnosisTypeService: diagTypeSvc, g: g}
		diagNameSvc = &diagNameQueryGuard{DiagnosisNameService: diagNameSvc, g: g}
		examTypeSvc = &examTypeQueryGuard{ExaminationTypeService: examTypeSvc, g: g}
		inquirySvc = &inquiryQueryGuard{InquiryTemplateService: inquirySvc, g: g}
		medicineSvc = &medicineQueryGuard{MedicineService: medicineSvc, g: g}
		doseSvc = &doseQueryGuard{MedicineDoseParamService: doseSvc, g: g}
		procedureSvc = &procedureQueryGuard{ProcedureService: procedureSvc, g: g}
		vaccineSvc = &vaccineQueryGuard{VaccineService: vaccineSvc, g: g}
		cageSvc = &cageQueryGuard{CageService: cageSvc, g: g}
		hospPlanSvc = &hospPlanQueryGuard{HospitalizationPlanService: hospPlanSvc, g: g}
		checkupTypeSvc = &checkupTypeQueryGuard{CheckupTypeService: checkupTypeSvc, g: g}
		fieldResultSvc = &checkupFieldQueryGuard{CheckupFieldResultService: fieldResultSvc, g: g}
	}

	return NewChiefComplaintHandler(chiefSvc),
		NewConsultationHandler(consultSvc),
		NewDiagnosisHandler(diagTypeSvc, diagNameSvc),
		NewExamTypeHandler(examTypeSvc),
		NewInquiryTemplateHandler(inquirySvc),
		NewMedicineHandler(medicineSvc),
		NewMedicineDoseParamHandler(doseSvc),
		NewProcedureHandler(procedureSvc),
		NewVaccineHandler(vaccineSvc),
		NewCageHandler(cageSvc),
		NewHospitalizationPlanHandler(hospPlanSvc),
		NewCheckupTypeHandler(checkupTypeSvc),
		NewCheckupHandler(NewCheckupService(NewCheckupRepository(db), NewMedicalRecordRepository(db), NewCheckupTypeRepository(db), nil, nil), fieldResultSvc)
}

type chiefQueryGuard struct {
	ChiefComplaintTypeService
	g clinicIDGuard
}

func (s *chiefQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.ChiefComplaintType, error) {
	s.g.check(clinicID, "chief List")
	return s.ChiefComplaintTypeService.List(ctx, clinicID)
}
func (s *chiefQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.ChiefComplaintType, error) {
	s.g.check(clinicID, "chief GetByID")
	return s.ChiefComplaintTypeService.GetByID(ctx, clinicID, id)
}

type consultQueryGuard struct {
	ConsultationService
	g clinicIDGuard
}

func (s *consultQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	s.g.check(clinicID, "consult List")
	return s.ConsultationService.List(ctx, clinicID)
}
func (s *consultQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	s.g.check(clinicID, "consult GetByID")
	return s.ConsultationService.GetByID(ctx, clinicID, id)
}

type diagTypeQueryGuard struct {
	DiagnosisTypeService
	g clinicIDGuard
}

func (s *diagTypeQueryGuard) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	s.g.check(clinicID, "diagType List")
	return s.DiagnosisTypeService.List(ctx, clinicID, page, limit)
}
func (s *diagTypeQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	s.g.check(clinicID, "diagType GetByID")
	return s.DiagnosisTypeService.GetByID(ctx, clinicID, id)
}

type diagNameQueryGuard struct {
	DiagnosisNameService
	g clinicIDGuard
}

func (s *diagNameQueryGuard) List(ctx context.Context, clinicID uint64, typeID *uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	s.g.check(clinicID, "diagName List")
	return s.DiagnosisNameService.List(ctx, clinicID, typeID, page, limit)
}
func (s *diagNameQueryGuard) ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	s.g.check(clinicID, "diagName ListNames")
	return s.DiagnosisNameService.ListNames(ctx, clinicID, typeID)
}
func (s *diagNameQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	s.g.check(clinicID, "diagName GetByID")
	return s.DiagnosisNameService.GetByID(ctx, clinicID, id)
}

type examTypeQueryGuard struct {
	ExaminationTypeService
	g clinicIDGuard
}

func (s *examTypeQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	s.g.check(clinicID, "examType List")
	return s.ExaminationTypeService.List(ctx, clinicID)
}
func (s *examTypeQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	s.g.check(clinicID, "examType GetByID")
	return s.ExaminationTypeService.GetByID(ctx, clinicID, id)
}

type inquiryQueryGuard struct {
	InquiryTemplateService
	g clinicIDGuard
}

func (s *inquiryQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	s.g.check(clinicID, "inquiry List")
	return s.InquiryTemplateService.List(ctx, clinicID)
}
func (s *inquiryQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	s.g.check(clinicID, "inquiry GetByID")
	return s.InquiryTemplateService.GetByID(ctx, clinicID, id)
}

type medicineQueryGuard struct {
	MedicineService
	g clinicIDGuard
}

func (s *medicineQueryGuard) List(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	s.g.check(clinicID, "medicine List")
	return s.MedicineService.List(ctx, clinicID, page, limit)
}
func (s *medicineQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	s.g.check(clinicID, "medicine GetByID")
	return s.MedicineService.GetByID(ctx, clinicID, id)
}

type doseQueryGuard struct {
	MedicineDoseParamService
	g clinicIDGuard
}

func (s *doseQueryGuard) List(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	s.g.check(clinicID, "dose List")
	return s.MedicineDoseParamService.List(ctx, clinicID, medicineID)
}

type procedureQueryGuard struct {
	ProcedureService
	g clinicIDGuard
}

func (s *procedureQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
	s.g.check(clinicID, "procedure List")
	return s.ProcedureService.List(ctx, clinicID)
}
func (s *procedureQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error) {
	s.g.check(clinicID, "procedure GetByID")
	return s.ProcedureService.GetByID(ctx, clinicID, id)
}

type vaccineQueryGuard struct {
	VaccineService
	g clinicIDGuard
}

func (s *vaccineQueryGuard) List(ctx context.Context, clinicID uint64, species *string) ([]model.Vaccine, error) {
	s.g.check(clinicID, "vaccine List")
	return s.VaccineService.List(ctx, clinicID, species)
}
func (s *vaccineQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error) {
	s.g.check(clinicID, "vaccine GetByID")
	return s.VaccineService.GetByID(ctx, clinicID, id)
}

type cageQueryGuard struct {
	CageService
	g clinicIDGuard
}

func (s *cageQueryGuard) List(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	s.g.check(clinicID, "cage List")
	return s.CageService.List(ctx, clinicID, cageType)
}
func (s *cageQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	s.g.check(clinicID, "cage GetByID")
	return s.CageService.GetByID(ctx, clinicID, id)
}

type hospPlanQueryGuard struct {
	HospitalizationPlanService
	g clinicIDGuard
}

func (s *hospPlanQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.HospitalizationPlan, error) {
	s.g.check(clinicID, "hospPlan List")
	return s.HospitalizationPlanService.List(ctx, clinicID)
}
func (s *hospPlanQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.HospitalizationPlan, error) {
	s.g.check(clinicID, "hospPlan GetByID")
	return s.HospitalizationPlanService.GetByID(ctx, clinicID, id)
}

type checkupTypeQueryGuard struct {
	CheckupTypeService
	g clinicIDGuard
}

func (s *checkupTypeQueryGuard) List(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	s.g.check(clinicID, "checkupType List")
	return s.CheckupTypeService.List(ctx, clinicID)
}
func (s *checkupTypeQueryGuard) GetByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	s.g.check(clinicID, "checkupType GetByID")
	return s.CheckupTypeService.GetByID(ctx, clinicID, id)
}

type checkupFieldQueryGuard struct {
	CheckupFieldResultService
	g clinicIDGuard
}

func (s *checkupFieldQueryGuard) ListFields(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	s.g.check(clinicID, "checkupField ListFields")
	return s.CheckupFieldResultService.ListFields(ctx, clinicID, checkupTypeID)
}

func seedRealDBMastersFixture(t *testing.T, db *gorm.DB, forbidden uint64) realDBMastersFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr masters realDB")
	ctx := context.Background()

	chiefA := &model.ChiefComplaintType{ClinicID: fx.ClinicA, Name: realDBChiefA, IsActive: true}
	chiefB := &model.ChiefComplaintType{ClinicID: fx.ClinicB, Name: realDBChiefB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(chiefA).Error)
	require.NoError(t, db.WithContext(ctx).Create(chiefB).Error)

	consultA := &model.Consultation{ClinicID: fx.ClinicA, Name: realDBConsultA, IsActive: true}
	consultB := &model.Consultation{ClinicID: fx.ClinicB, Name: realDBConsultB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(consultA).Error)
	require.NoError(t, db.WithContext(ctx).Create(consultB).Error)

	diagTypeA := &model.DiagnosisType{ClinicID: fx.ClinicA, Name: realDBDiagTypeA, IsActive: true}
	diagTypeB := &model.DiagnosisType{ClinicID: fx.ClinicB, Name: realDBDiagTypeB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(diagTypeA).Error)
	require.NoError(t, db.WithContext(ctx).Create(diagTypeB).Error)
	diagNameA := &model.DiagnosisName{ClinicID: fx.ClinicA, Name: realDBDiagNameA, DiagnosisTypeID: diagTypeA.ID, IsActive: true}
	diagNameB := &model.DiagnosisName{ClinicID: fx.ClinicB, Name: realDBDiagNameB, DiagnosisTypeID: diagTypeB.ID, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(diagNameA).Error)
	require.NoError(t, db.WithContext(ctx).Create(diagNameB).Error)

	examTypeA := &model.ExaminationType{ClinicID: fx.ClinicA, Name: realDBExamTypeA, IsActive: true}
	examTypeB := &model.ExaminationType{ClinicID: fx.ClinicB, Name: realDBExamTypeB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(examTypeA).Error)
	require.NoError(t, db.WithContext(ctx).Create(examTypeB).Error)

	inquiryA := &model.InquiryTemplate{ClinicID: fx.ClinicA, Title: realDBInquiryA, IsActive: true}
	inquiryB := &model.InquiryTemplate{ClinicID: fx.ClinicB, Title: realDBInquiryB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(inquiryA).Error)
	require.NoError(t, db.WithContext(ctx).Create(inquiryB).Error)

	medicineA := makeMedicine(t, db, fx.ClinicA, realDBMedicineA)
	medicineB := makeMedicine(t, db, fx.ClinicB, realDBMedicineB)
	require.NoError(t, db.WithContext(ctx).Create(&model.MedicineDoseParam{
		ClinicID: fx.ClinicA, MedicineID: medicineA.ID, Species: model.MedicineDoseSpeciesDog,
		DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 1.0,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.MedicineDoseParam{
		ClinicID: fx.ClinicB, MedicineID: medicineB.ID, Species: model.MedicineDoseSpeciesDog,
		DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 2.0,
	}).Error)

	procedureA := &model.Procedure{ClinicID: fx.ClinicA, Name: realDBProcedureA, IsActive: true}
	procedureB := &model.Procedure{ClinicID: fx.ClinicB, Name: realDBProcedureB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(procedureA).Error)
	require.NoError(t, db.WithContext(ctx).Create(procedureB).Error)

	vaccineA := &model.Vaccine{ClinicID: fx.ClinicA, Name: realDBVaccineA, IsActive: true}
	vaccineB := &model.Vaccine{ClinicID: fx.ClinicB, Name: realDBVaccineB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(vaccineA).Error)
	require.NoError(t, db.WithContext(ctx).Create(vaccineB).Error)

	cageA := &model.Cage{ClinicID: fx.ClinicA, Name: realDBCageA, CageType: model.CageTypeGeneral, CageSize: model.CageSizeMedium, IsActive: true}
	cageB := &model.Cage{ClinicID: fx.ClinicB, Name: realDBCageB, CageType: model.CageTypeGeneral, CageSize: model.CageSizeMedium, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(cageA).Error)
	require.NoError(t, db.WithContext(ctx).Create(cageB).Error)

	hospPlanA := &model.HospitalizationPlan{ClinicID: fx.ClinicA, Name: realDBHospPlanA, IsActive: true}
	hospPlanB := &model.HospitalizationPlan{ClinicID: fx.ClinicB, Name: realDBHospPlanB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(hospPlanA).Error)
	require.NoError(t, db.WithContext(ctx).Create(hospPlanB).Error)

	checkupTypeA := &model.CheckupType{ClinicID: fx.ClinicA, Name: realDBCheckupTypeA, IsActive: true}
	checkupTypeB := &model.CheckupType{ClinicID: fx.ClinicB, Name: realDBCheckupTypeB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(checkupTypeA).Error)
	require.NoError(t, db.WithContext(ctx).Create(checkupTypeB).Error)
	checkupFieldA := ensureCheckupFieldOptions(&model.CheckupTypeField{
		ClinicID: fx.ClinicA, CheckupTypeID: checkupTypeA.ID, Name: realDBCheckupFieldA, FieldType: model.CheckupFieldTypeText,
	})
	checkupFieldB := ensureCheckupFieldOptions(&model.CheckupTypeField{
		ClinicID: fx.ClinicB, CheckupTypeID: checkupTypeB.ID, Name: realDBCheckupFieldB, FieldType: model.CheckupFieldTypeText,
	})
	require.NoError(t, db.WithContext(ctx).Create(checkupFieldA).Error)
	require.NoError(t, db.WithContext(ctx).Create(checkupFieldB).Error)

	chiefH, consultH, diagH, examTypeH, inquiryH, medicineH, doseH, procedureH, vaccineH, cageH, hospPlanH, checkupTypeH, checkupH := wireMastersHandlers(t, db, forbidden)

	return realDBMastersFixture{
		fx: fx, chiefA: chiefA, chiefB: chiefB, consultA: consultA, consultB: consultB,
		diagTypeA: diagTypeA, diagTypeB: diagTypeB, diagNameA: diagNameA, diagNameB: diagNameB,
		examTypeA: examTypeA, examTypeB: examTypeB, inquiryA: inquiryA, inquiryB: inquiryB,
		medicineA: medicineA, medicineB: medicineB, procedureA: procedureA, procedureB: procedureB,
		vaccineA: vaccineA, vaccineB: vaccineB, cageA: cageA, cageB: cageB,
		hospPlanA: hospPlanA, hospPlanB: hospPlanB, checkupTypeA: checkupTypeA, checkupTypeB: checkupTypeB,
		checkupFieldA: checkupFieldA, checkupFieldB: checkupFieldB,
		chiefH: chiefH, consultH: consultH, diagH: diagH, examTypeH: examTypeH, inquiryH: inquiryH,
		medicineH: medicineH, doseH: doseH, procedureH: procedureH, vaccineH: vaccineH,
		cageH: cageH, hospPlanH: hospPlanH, checkupTypeH: checkupTypeH, checkupH: checkupH,
	}
}

func rewireMastersGuarded(t *testing.T, db *gorm.DB, fx *realDBMastersFixture) {
	t.Helper()
	chiefH, consultH, diagH, examTypeH, inquiryH, medicineH, doseH, procedureH, vaccineH, cageH, hospPlanH, checkupTypeH, checkupH := wireMastersHandlers(t, db, fx.fx.ClinicB)
	fx.chiefH, fx.consultH, fx.diagH, fx.examTypeH, fx.inquiryH = chiefH, consultH, diagH, examTypeH, inquiryH
	fx.medicineH, fx.doseH, fx.procedureH, fx.vaccineH = medicineH, doseH, procedureH, vaccineH
	fx.cageH, fx.hospPlanH, fx.checkupTypeH, fx.checkupH = cageH, hospPlanH, checkupTypeH, checkupH
}

func assertNamedListHasB(t *testing.T, body []byte, clinicB uint64, nameB, nameA string, idB uint64) {
	t.Helper()
	var listed []namedClinicDTO
	require.NoError(t, json.Unmarshal(body, &listed))
	require.NotEmpty(t, listed)
	found := false
	for _, item := range listed {
		assert.Equal(t, clinicB, item.ClinicID)
		assert.NotEqual(t, nameA, item.Name)
		if item.ID == idB {
			found = true
			assert.Equal(t, nameB, item.Name)
		}
	}
	require.True(t, found)
}

// TestRealDB_MastersSelectedClinicBGrantAIsolation covers master-medical,
// master-hospitalization, and checkup-type clinic-fixed routes.
func TestRealDB_MastersSelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resMedical := string(model.ResourceMasterMedical)
	resHosp := string(model.ResourceMasterHospitalization)
	resCheckups := string(model.ResourceCheckups)

	type caseFn func(t *testing.T, db *gorm.DB, fx realDBMastersFixture)

	run := func(name string, fn caseFn) {
		t.Run(name, func(t *testing.T) {
			db := setupRealDBMastersDB(t)
			fx := seedRealDBMastersFixture(t, db, 0)
			fn(t, db, fx)
		})
	}

	// ---- chief complaints ----
	run("chief_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/chief-complaint-types", configureGrant(fx.fx, resMedical, fx.fx.ClinicA))
		fx.chiefH.ListChiefComplaints(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBChiefA, realDBChiefB)
	})
	run("chief_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/chief-complaint-types", configureGrant(fx.fx, resMedical, fx.fx.ClinicB))
		fx.chiefH.ListChiefComplaints(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assertNamedListHasB(t, w.Body.Bytes(), fx.fx.ClinicB, realDBChiefB, realDBChiefA, fx.chiefB.ID)
	})
	run("chief_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/chief-complaint-types/%d", fx.chiefB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicA), fx.chiefB.ID))
		fx.chiefH.GetChiefComplaint(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("chief_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/chief-complaint-types/%d", fx.chiefB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.chiefB.ID))
		fx.chiefH.GetChiefComplaint(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBChiefB)
		assert.NotContains(t, w.Body.String(), realDBChiefA)
	})
	run("chief_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/chief-complaint-types/%d", fx.chiefA.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.chiefA.ID))
		fx.chiefH.GetChiefComplaint(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBChiefA)
	})

	pairListGet := func(prefix, path string, res string, idB, idA uint64, nameB, nameA string, listFn func(*realDBMastersFixture, *gin.Context), getFn func(*realDBMastersFixture, *gin.Context), listAssert func(*testing.T, []byte, realDBMastersFixture)) {
		run(prefix+"_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			rewireMastersGuarded(t, db, &fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureGrant(fx.fx, res, fx.fx.ClinicA))
			listFn(&fx, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NotContains(t, w.Body.String(), nameA)
			assert.NotContains(t, w.Body.String(), nameB)
		})
		run(prefix+"_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureGrant(fx.fx, res, fx.fx.ClinicB))
			listFn(&fx, c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			listAssert(t, w.Body.Bytes(), fx)
		})
		run(prefix+"_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			rewireMastersGuarded(t, db, &fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", path, idB), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicA), idB))
			getFn(&fx, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
		run(prefix+"_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", path, idB), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), idB))
			getFn(&fx, c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), nameB)
			assert.NotContains(t, w.Body.String(), nameA)
		})
		run(prefix+"_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", path, idA), withIDParam(configureGrant(fx.fx, res, fx.fx.ClinicB), idA))
			getFn(&fx, c)
			require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
			assert.NotContains(t, w.Body.String(), nameA)
		})
	}

	// consultations
	run("consult_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/consultations", configureGrant(fx.fx, resMedical, fx.fx.ClinicA))
		fx.consultH.ListConsultations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("consult_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/consultations", configureGrant(fx.fx, resMedical, fx.fx.ClinicB))
		fx.consultH.ListConsultations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assertNamedListHasB(t, w.Body.Bytes(), fx.fx.ClinicB, realDBConsultB, realDBConsultA, fx.consultB.ID)
	})
	run("consult_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/consultations/%d", fx.consultB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicA), fx.consultB.ID))
		fx.consultH.GetConsultation(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("consult_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/consultations/%d", fx.consultB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.consultB.ID))
		fx.consultH.GetConsultation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBConsultB)
	})
	run("consult_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/consultations/%d", fx.consultA.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.consultA.ID))
		fx.consultH.GetConsultation(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	_ = pairListGet // keep helper available for readability; remaining resources use explicit cases below
	_ = resHosp
	_ = resCheckups

	// diagnosis types
	run("diag_types_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-types", configureGrant(fx.fx, resMedical, fx.fx.ClinicA))
		fx.diagH.ListDiagnosisTypes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("diag_types_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-types", configureGrant(fx.fx, resMedical, fx.fx.ClinicB))
		fx.diagH.ListDiagnosisTypes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBDiagTypeB)
		assert.NotContains(t, w.Body.String(), realDBDiagTypeA)
	})
	run("diag_types_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-types/%d", fx.diagTypeB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicA), fx.diagTypeB.ID))
		fx.diagH.GetDiagnosisType(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("diag_types_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-types/%d", fx.diagTypeB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.diagTypeB.ID))
		fx.diagH.GetDiagnosisType(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBDiagTypeB)
	})
	run("diag_types_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-types/%d", fx.diagTypeA.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.diagTypeA.ID))
		fx.diagH.GetDiagnosisType(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	// diagnosis names + all
	run("diag_names_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-names", configureGrant(fx.fx, resMedical, fx.fx.ClinicA))
		fx.diagH.ListDiagnosisNames(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("diag_names_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-names", configureGrant(fx.fx, resMedical, fx.fx.ClinicB))
		fx.diagH.ListDiagnosisNames(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBDiagNameB)
		assert.NotContains(t, w.Body.String(), realDBDiagNameA)
	})
	run("diag_names_all_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-names/all", configureGrant(fx.fx, resMedical, fx.fx.ClinicA))
		fx.diagH.ListDiagnosisNamesAll(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("diag_names_all_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/diagnosis-names/all", configureGrant(fx.fx, resMedical, fx.fx.ClinicB))
		fx.diagH.ListDiagnosisNamesAll(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBDiagNameB)
		assert.NotContains(t, w.Body.String(), realDBDiagNameA)
	})
	run("diag_names_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-names/%d", fx.diagNameB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicA), fx.diagNameB.ID))
		fx.diagH.GetDiagnosisName(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("diag_names_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-names/%d", fx.diagNameB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.diagNameB.ID))
		fx.diagH.GetDiagnosisName(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBDiagNameB)
	})
	run("diag_names_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/diagnosis-names/%d", fx.diagNameA.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.diagNameA.ID))
		fx.diagH.GetDiagnosisName(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	// Remaining master resources: exam types, inquiry, medicines, dose-params, procedures, vaccines, cages, hosp plans, checkup types/fields
	resources := []struct {
		prefix string
		path   string
		res    string
		list   func(*realDBMastersFixture, *gin.Context)
		get    func(*realDBMastersFixture, *gin.Context)
		idB    func(realDBMastersFixture) uint64
		idA    func(realDBMastersFixture) uint64
		nameB  string
		nameA  string
		title  bool
	}{
		{"exam_types", "/api/v1/masters/examination-types", resMedical, func(fx *realDBMastersFixture, c *gin.Context) { fx.examTypeH.ListExaminationTypes(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.examTypeH.GetExaminationType(c) }, func(fx realDBMastersFixture) uint64 { return fx.examTypeB.ID }, func(fx realDBMastersFixture) uint64 { return fx.examTypeA.ID }, realDBExamTypeB, realDBExamTypeA, false},
		{"inquiry", "/api/v1/masters/inquiry-templates", resMedical, func(fx *realDBMastersFixture, c *gin.Context) { fx.inquiryH.ListInquiryTemplates(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.inquiryH.GetInquiryTemplate(c) }, func(fx realDBMastersFixture) uint64 { return fx.inquiryB.ID }, func(fx realDBMastersFixture) uint64 { return fx.inquiryA.ID }, realDBInquiryB, realDBInquiryA, true},
		{"medicines", "/api/v1/masters/medicines", resMedical, func(fx *realDBMastersFixture, c *gin.Context) { fx.medicineH.ListMedicines(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.medicineH.GetMedicine(c) }, func(fx realDBMastersFixture) uint64 { return fx.medicineB.ID }, func(fx realDBMastersFixture) uint64 { return fx.medicineA.ID }, realDBMedicineB, realDBMedicineA, false},
		{"procedures", "/api/v1/masters/procedures", resMedical, func(fx *realDBMastersFixture, c *gin.Context) { fx.procedureH.ListProcedures(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.procedureH.GetProcedure(c) }, func(fx realDBMastersFixture) uint64 { return fx.procedureB.ID }, func(fx realDBMastersFixture) uint64 { return fx.procedureA.ID }, realDBProcedureB, realDBProcedureA, false},
		{"vaccines", "/api/v1/masters/vaccines", resMedical, func(fx *realDBMastersFixture, c *gin.Context) { fx.vaccineH.ListVaccines(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.vaccineH.GetVaccine(c) }, func(fx realDBMastersFixture) uint64 { return fx.vaccineB.ID }, func(fx realDBMastersFixture) uint64 { return fx.vaccineA.ID }, realDBVaccineB, realDBVaccineA, false},
		{"cages", "/api/v1/masters/cages", resHosp, func(fx *realDBMastersFixture, c *gin.Context) { fx.cageH.ListCages(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.cageH.GetCage(c) }, func(fx realDBMastersFixture) uint64 { return fx.cageB.ID }, func(fx realDBMastersFixture) uint64 { return fx.cageA.ID }, realDBCageB, realDBCageA, false},
		{"hosp_plans", "/api/v1/masters/hospitalization-plans", resHosp, func(fx *realDBMastersFixture, c *gin.Context) { fx.hospPlanH.ListHospitalizationPlans(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.hospPlanH.GetHospitalizationPlan(c) }, func(fx realDBMastersFixture) uint64 { return fx.hospPlanB.ID }, func(fx realDBMastersFixture) uint64 { return fx.hospPlanA.ID }, realDBHospPlanB, realDBHospPlanA, false},
		{"checkup_types", "/api/v1/masters/checkup-types", resCheckups, func(fx *realDBMastersFixture, c *gin.Context) { fx.checkupTypeH.ListCheckupTypes(c) }, func(fx *realDBMastersFixture, c *gin.Context) { fx.checkupTypeH.GetCheckupType(c) }, func(fx realDBMastersFixture) uint64 { return fx.checkupTypeB.ID }, func(fx realDBMastersFixture) uint64 { return fx.checkupTypeA.ID }, realDBCheckupTypeB, realDBCheckupTypeA, false},
	}

	for _, rc := range resources {
		rc := rc
		run(rc.prefix+"_list_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			rewireMastersGuarded(t, db, &fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, rc.path, configureGrant(fx.fx, rc.res, fx.fx.ClinicA))
			rc.list(&fx, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.NotContains(t, w.Body.String(), rc.nameA)
			assert.NotContains(t, w.Body.String(), rc.nameB)
		})
		run(rc.prefix+"_list_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, rc.path, configureGrant(fx.fx, rc.res, fx.fx.ClinicB))
			rc.list(&fx, c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), rc.nameB)
			assert.NotContains(t, w.Body.String(), rc.nameA)
		})
		run(rc.prefix+"_get_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			rewireMastersGuarded(t, db, &fx)
			id := rc.idB(fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", rc.path, id), withIDParam(configureGrant(fx.fx, rc.res, fx.fx.ClinicA), id))
			rc.get(&fx, c)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
		run(rc.prefix+"_get_grantB_ok", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			id := rc.idB(fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", rc.path, id), withIDParam(configureGrant(fx.fx, rc.res, fx.fx.ClinicB), id))
			rc.get(&fx, c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), rc.nameB)
			assert.NotContains(t, w.Body.String(), rc.nameA)
		})
		run(rc.prefix+"_get_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
			id := rc.idA(fx)
			c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("%s/%d", rc.path, id), withIDParam(configureGrant(fx.fx, rc.res, fx.fx.ClinicB), id))
			rc.get(&fx, c)
			require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
			assert.NotContains(t, w.Body.String(), rc.nameA)
		})
	}

	// dose-params under medicine id
	run("dose_params_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/medicines/%d/dose-params", fx.medicineB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicA), fx.medicineB.ID))
		fx.doseH.ListMedicineDoseParams(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("dose_params_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/medicines/%d/dose-params", fx.medicineB.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.medicineB.ID))
		fx.doseH.ListMedicineDoseParams(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NotEqual(t, "[]", w.Body.String())
		assert.NotContains(t, w.Body.String(), fmt.Sprintf(`"medicine_id":%d`, fx.medicineA.ID))
	})
	run("dose_params_A_id_404", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/medicines/%d/dose-params", fx.medicineA.ID), withIDParam(configureGrant(fx.fx, resMedical, fx.fx.ClinicB), fx.medicineA.ID))
		fx.doseH.ListMedicineDoseParams(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	// checkup type fields
	run("checkup_fields_grantA_403", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		rewireMastersGuarded(t, db, &fx)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/checkup-types/%d/fields", fx.checkupTypeB.ID), withIDParam(configureGrant(fx.fx, resCheckups, fx.fx.ClinicA), fx.checkupTypeB.ID))
		fx.checkupH.ListCheckupTypeFields(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	run("checkup_fields_grantB_nonempty", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/checkup-types/%d/fields", fx.checkupTypeB.ID), withIDParam(configureGrant(fx.fx, resCheckups, fx.fx.ClinicB), fx.checkupTypeB.ID))
		fx.checkupH.ListCheckupTypeFields(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBCheckupFieldB)
		assert.NotContains(t, w.Body.String(), realDBCheckupFieldA)
	})
	run("checkup_fields_A_id_empty_no_leak", func(t *testing.T, db *gorm.DB, fx realDBMastersFixture) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/checkup-types/%d/fields", fx.checkupTypeA.ID), withIDParam(configureGrant(fx.fx, resCheckups, fx.fx.ClinicB), fx.checkupTypeA.ID))
		fx.checkupH.ListCheckupTypeFields(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "[]", w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBCheckupFieldA)
	})
}
