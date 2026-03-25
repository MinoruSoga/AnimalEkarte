package repository

import (
	"gorm.io/gorm"
)

// Repositories はすべてのリポジトリを保持するDIコンテナ
type Repositories struct {
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
}

// NewRepositories はすべてのリポジトリを初期化して返す
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
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
	}
}
