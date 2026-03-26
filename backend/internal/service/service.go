package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	AnimalSpecies          AnimalSpeciesService
	Owner                  OwnerService
	Pet                    PetService
	Reservation            ReservationService
	MedicalRecord          MedicalRecordService
	Hospitalization        HospitalizationService
	Accounting             AccountingService
	Trimming               TrimmingService
	Inventory              InventoryService
	Staff                  StaffService
	Cage                   CageService
	Medicine               MedicineService
	Vaccine                VaccineService
	Insurance              InsuranceService
	ServiceType            ServiceTypeService
	Consultation           ConsultationService
	Procedure              ProcedureService
	HospitalizationPlan    HospitalizationPlanService
	TrimmingCourse         TrimmingCourseService
	TrimmingOption         TrimmingOptionService
	ExaminationType        ExamTypeService
	DiagnosisCategory      DiagnosisCategoryService
	DiagnosisName          DiagnosisNameService
	CheckupType            CheckupTypeService
	Clinic                 ClinicService
	UserAccount            UserAccountService
	Examination            ExaminationService
	Vaccination            VaccinationService
	JobTitle               JobTitleService
	ChiefComplaintCategory ChiefComplaintCategoryService
	Inquiry                InquiryService
	InquiryTemplate        InquiryTemplateService
	Company                CompanyService
	BillingReview          BillingReviewService
	CarePlanItem           CarePlanItemService
	ShiftEntry             ShiftEntryService
	TreatmentPlan          TreatmentPlanService
	Vital                  VitalService
	Treatment              TreatmentService
	DailyRecord            DailyRecordService
	RecordImage            RecordImageService
	ClinicalPlan           ClinicalPlanService
	Checkup                CheckupService
	Estimate               EstimateService
	MerchandiseItem        MerchandiseItemService
	BillingItem            BillingItemService
	Refund                 RefundService
	PermissionGroup        PermissionGroupService
}

// NewServices はリポジトリからすべてのサービスを初期化して返す
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		AnimalSpecies:          NewAnimalSpeciesService(repos.AnimalSpecies),
		Owner:                  NewOwnerService(repos.Owner),
		Pet:                    NewPetService(repos.Pet, repos.Owner, repos.Insurance),
		Reservation:            NewReservationService(repos.Reservation),
		MedicalRecord:          NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet),
		Hospitalization:        NewHospitalizationService(repos.Hospitalization),
		Accounting:             NewAccountingService(repos.Accounting),
		Trimming:               NewTrimmingService(repos.Trimming),
		Inventory:              NewInventoryService(repos.Inventory),
		Staff:                  NewStaffService(repos.Staff),
		Cage:                   NewCageService(repos.Cage),
		Medicine:               NewMedicineService(repos.Medicine),
		Vaccine:                NewVaccineService(repos.Vaccine),
		Insurance:              NewInsuranceService(repos.Insurance),
		ServiceType:            NewServiceTypeService(repos.ServiceType),
		Consultation:           NewConsultationService(repos.Consultation),
		Procedure:              NewProcedureService(repos.Procedure),
		HospitalizationPlan:    NewHospitalizationPlanService(repos.HospitalizationPlan),
		TrimmingCourse:         NewTrimmingCourseService(repos.TrimmingCourse),
		TrimmingOption:         NewTrimmingOptionService(repos.TrimmingOption),
		ExaminationType:        NewExamTypeService(repos.ExaminationType),
		DiagnosisCategory:      NewDiagnosisCategoryService(repos.DiagnosisCategory),
		DiagnosisName:          NewDiagnosisNameService(repos.DiagnosisName, repos.DiagnosisCategory),
		CheckupType:            NewCheckupTypeService(repos.CheckupType),
		Clinic:                 NewClinicService(repos.Clinic),
		UserAccount:            NewUserAccountService(repos.UserAccount),
		Examination:            NewExaminationService(repos.Examination),
		Vaccination:            NewVaccinationService(repos.Vaccination),
		JobTitle:               NewJobTitleService(repos.JobTitle),
		ChiefComplaintCategory: NewChiefComplaintCategoryService(repos.ChiefComplaintCategory),
		Inquiry:                NewInquiryService(repos.Inquiry),
		InquiryTemplate:        NewInquiryTemplateService(repos.InquiryTemplate),
		Company:                NewCompanyService(repos.Company),
		BillingReview:          NewBillingReviewService(repos.BillingReview),
		CarePlanItem:           NewCarePlanItemService(repos.CarePlanItem),
		ShiftEntry:             NewShiftEntryService(repos.ShiftEntry),
		TreatmentPlan:          NewTreatmentPlanService(repos.TreatmentPlan),
		Vital:                  NewVitalService(repos.Vital),
		Treatment:              NewTreatmentService(repos.Treatment),
		DailyRecord:            NewDailyRecordService(repos.DailyRecord),
		RecordImage:            NewRecordImageService(repos.RecordImage),
		ClinicalPlan:           NewClinicalPlanService(repos.ClinicalPlan),
		Checkup:                NewCheckupService(repos.Checkup),
		Estimate:               NewEstimateService(repos.Estimate),
		MerchandiseItem:        NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:            NewBillingItemService(repos.BillingItem),
		Refund:                 NewRefundService(repos.Refund, repos.Accounting),
		PermissionGroup:        NewPermissionGroupService(repos.PermissionGroup),
	}
}
