package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	Account                AccountService
	StaffClinicAssignment  StaffClinicAssignmentService
	Audit                  AuditService
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
	Examination            ExaminationService
	Vaccination            VaccinationService
	Occupation             OccupationService
	ChiefComplaintCategory ChiefComplaintCategoryService
	Inquiry                InquiryService
	InquiryTemplate        InquiryTemplateService
	Company                CompanyService
	PermissionGroup        PermissionGroupService
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
	// LINE予約
	ReservationSetting  ReservationSettingService
	ReservationCourse   ReservationCourseService
	ReservationStaff    ReservationStaffService
	ReservationSchedule ReservationScheduleService
	ReservationAdmin    ReservationAdminService
	ReservationCustomer ReservationCustomerService
	Liff                LiffService
}

// NewServices はリポジトリからすべてのサービスを初期化して返す
func NewServices(repos *repository.Repositories, notifCfg ReservationNotificationConfig) *Services {
	notifier := NewReservationNotificationService(notifCfg, repos.ReservationSetting)
	auditSvc := NewAuditService(repos.Audit)

	return &Services{
		Account:                NewAccountService(repos.Account),
		StaffClinicAssignment:  NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                  auditSvc,
		AnimalSpecies:          NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                  NewOwnerService(repos.Owner),
		Pet:                    NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord),
		Reservation:            NewReservationService(repos.Reservation, repos.DB()),
		MedicalRecord:          NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan),
		Hospitalization:        NewHospitalizationService(repos),
		Accounting:             NewAccountingService(repos.Accounting),
		Trimming:               NewTrimmingService(repos.Trimming),
		Inventory:              NewInventoryService(repos.Inventory),
		Staff:                  NewStaffService(repos.Staff, repos.Account, repos.StaffClinicAssignment, repos.Reservation, repos.ShiftEntry, repos.PermissionGroup, repos.ReservationStaff),
		Cage:                   NewCageService(repos.Cage, repos.Hospitalization),
		Medicine:               NewMedicineService(repos.Medicine),
		Vaccine:                NewVaccineService(repos.Vaccine),
		Insurance:              NewInsuranceService(repos.Insurance),
		ServiceType:            NewServiceTypeService(repos.ServiceType, repos.Reservation),
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
		Examination:            NewExaminationService(repos.Examination),
		Vaccination:            NewVaccinationService(repos.Vaccination),
		Occupation:             NewOccupationService(repos.Occupation),
		ChiefComplaintCategory: NewChiefComplaintCategoryService(repos.ChiefComplaintCategory, repos.Inquiry),
		Inquiry:                NewInquiryService(repos.Inquiry),
		InquiryTemplate:        NewInquiryTemplateService(repos.InquiryTemplate),
		Company:                NewCompanyService(repos.Company),
		PermissionGroup:        NewPermissionGroupService(repos.PermissionGroup, auditSvc),
		BillingReview:          NewBillingReviewService(repos.BillingReview),
		CarePlanItem:           NewCarePlanItemService(repos.CarePlanItem),
		ShiftEntry:             NewShiftEntryService(repos.ShiftEntry),
		TreatmentPlan:          NewTreatmentPlanService(repos.TreatmentPlan),
		Vital:                  NewVitalService(repos.Vital),
		Treatment:              NewTreatmentService(repos),
		DailyRecord:            NewDailyRecordService(repos.DailyRecord),
		RecordImage:            NewRecordImageService(repos.RecordImage),
		ClinicalPlan:           NewClinicalPlanService(repos.ClinicalPlan),
		Checkup:                NewCheckupService(repos.Checkup),
		Estimate:               NewEstimateService(repos.Estimate),
		MerchandiseItem:        NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:            NewBillingItemService(repos.BillingItem),
		Refund:                 NewRefundService(repos.Refund, repos.Accounting),
		ReservationSetting:     NewReservationSettingService(repos.ReservationSetting),
		ReservationCourse:      NewReservationCourseService(repos.ReservationCourse, repos.ReservationAdmin, repos.Reservation),
		ReservationStaff:       NewReservationStaffService(repos.ReservationStaff),
		ReservationSchedule:    NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:       NewReservationAdminService(repos.ReservationAdmin, repos.DB()),
		ReservationCustomer:    NewReservationCustomerService(repos.ReservationCustomerMgr),
		Liff: NewLiffService(
			repos.ReservationSetting,
			repos.ReservationCourse,
			repos.ReservationStaff,
			repos.ReservationSchedule,
			repos.ReservationAdmin,
			repos.ReservationCustomerMgr,
			repos.Owner,
			repos.DB(),
			notifier,
		),
	}
}
