package repository

import (
	"gorm.io/gorm"
)

// Repositories はすべてのリポジトリを保持するDIコンテナ
type Repositories struct {
	db *gorm.DB
	// TransactionFn はテスト時に差し替え可能なトランザクション関数。
	// nil の場合は db.Transaction() を使用する（本番動作）。
	TransactionFn          func(fn func(*Repositories) error) error
	Auth                   AuthRepository
	AnimalSpecies          AnimalSpeciesRepository
	Owner                  OwnerRepository
	Pet                    PetRepository
	Reservation            ReservationRepository
	MedicalRecord          MedicalRecordRepository
	Hospitalization        HospitalizationRepository
	Accounting             AccountingRepository
	Trimming               TrimmingRepository
	Inventory              InventoryRepository
	Staff                  StaffRepository
	Cage                   CageRepository
	Medicine               MedicineRepository
	Vaccine                VaccineRepository
	Insurance              InsuranceRepository
	ServiceType            ServiceTypeRepository
	Consultation           ConsultationRepository
	Procedure              ProcedureRepository
	HospitalizationPlan    HospitalizationPlanRepository
	TrimmingCourse         TrimmingCourseRepository
	TrimmingOption         TrimmingOptionRepository
	ExaminationType        ExamTypeRepository
	DiagnosisCategory      DiagnosisCategoryRepository
	DiagnosisName          DiagnosisNameRepository
	CheckupType            CheckupTypeRepository
	Clinic                 ClinicRepository
	UserAccount            UserAccountRepository
	Examination            ExaminationRepository
	Vaccination            VaccinationRepository
	JobTitle               JobTitleRepository
	ChiefComplaintCategory ChiefComplaintCategoryRepository
	Inquiry                InquiryRepository
	InquiryTemplate        InquiryTemplateRepository
	Company                CompanyRepository
	BillingReview          BillingReviewRepository
	CarePlanItem           CarePlanItemRepository
	ShiftEntry             ShiftEntryRepository
	TreatmentPlan          TreatmentPlanRepository
	Vital                  VitalRepository
	Treatment              TreatmentRepository
	DailyRecord            DailyRecordRepository
	RecordImage            RecordImageRepository
	ClinicalPlan           ClinicalPlanRepository
	Checkup                CheckupRepository
	Estimate               EstimateRepository
	MerchandiseItem        MerchandiseItemRepository
	BillingItem            BillingItemRepository
	Refund                 RefundRepository
	PermissionGroup        PermissionGroupRepository
	Audit                  AuditRepository
}

// NewRepositories はすべてのリポジトリを初期化して返す
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:                     db,
		Auth:                   NewAuthRepository(db),
		AnimalSpecies:          NewAnimalSpeciesRepository(db),
		Owner:                  NewOwnerRepository(db),
		Pet:                    NewPetRepository(db),
		Reservation:            NewReservationRepository(db),
		MedicalRecord:          NewMedicalRecordRepository(db),
		Hospitalization:        NewHospitalizationRepository(db),
		Accounting:             NewAccountingRepository(db),
		Trimming:               NewTrimmingRepository(db),
		Inventory:              NewInventoryRepository(db),
		Staff:                  NewStaffRepository(db),
		Cage:                   NewCageRepository(db),
		Medicine:               NewMedicineRepository(db),
		Vaccine:                NewVaccineRepository(db),
		Insurance:              NewInsuranceRepository(db),
		ServiceType:            NewServiceTypeRepository(db),
		Consultation:           NewConsultationRepository(db),
		Procedure:              NewProcedureRepository(db),
		HospitalizationPlan:    NewHospitalizationPlanRepository(db),
		TrimmingCourse:         NewTrimmingCourseRepository(db),
		TrimmingOption:         NewTrimmingOptionRepository(db),
		ExaminationType:        NewExamTypeRepository(db),
		DiagnosisCategory:      NewDiagnosisCategoryRepository(db),
		DiagnosisName:          NewDiagnosisNameRepository(db),
		CheckupType:            NewCheckupTypeRepository(db),
		Clinic:                 NewClinicRepository(db),
		UserAccount:            NewUserAccountRepository(db),
		Examination:            NewExaminationRepository(db),
		Vaccination:            NewVaccinationRepository(db),
		JobTitle:               NewJobTitleRepository(db),
		ChiefComplaintCategory: NewChiefComplaintCategoryRepository(db),
		Inquiry:                NewInquiryRepository(db),
		InquiryTemplate:        NewInquiryTemplateRepository(db),
		Company:                NewCompanyRepository(db),
		BillingReview:          NewBillingReviewRepository(db),
		CarePlanItem:           NewCarePlanItemRepository(db),
		ShiftEntry:             NewShiftEntryRepository(db),
		TreatmentPlan:          NewTreatmentPlanRepository(db),
		Vital:                  NewVitalRepository(db),
		Treatment:              NewTreatmentRepository(db),
		DailyRecord:            NewDailyRecordRepository(db),
		RecordImage:            NewRecordImageRepository(db),
		ClinicalPlan:           NewClinicalPlanRepository(db),
		Checkup:                NewCheckupRepository(db),
		Estimate:               NewEstimateRepository(db),
		MerchandiseItem:        NewMerchandiseItemRepository(db),
		BillingItem:            NewBillingItemRepository(db),
		Refund:                 NewRefundRepository(db),
		PermissionGroup:        NewPermissionGroupRepository(db),
		Audit:                  NewAuditRepository(db),
	}
}

// Transaction はリポジトリ層のトランザクションを実行する。
// テスト時は TransactionFn に mock を設定することで DB 依存を排除できる。
func (r *Repositories) Transaction(fn func(repos *Repositories) error) error {
	if r.TransactionFn != nil {
		return r.TransactionFn(fn)
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepos := NewRepositories(tx)
		return fn(txRepos)
	})
}
