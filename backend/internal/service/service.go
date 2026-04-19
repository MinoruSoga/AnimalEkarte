package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	Account               AccountService
	StaffClinicAssignment StaffClinicAssignmentService
	Audit                 AuditService
	AnimalSpecies         AnimalSpeciesService
	Owner                 OwnerService
	Pet                   PetService
	Reservation           ReservationService
	MedicalRecord         MedicalRecordService
	Hospitalization       HospitalizationService
	Accounting            AccountingService
	Trimming              TrimmingService
	Inventory             InventoryService
	Staff                 StaffService
	Cage                  CageService
	Medicine              MedicineService
	Vaccine               VaccineService
	Insurance             InsuranceService
	ReservationType       ReservationTypeService
	ReservationTypeGroup  ReservationTypeGroupService
	Consultation          ConsultationService
	Procedure             ProcedureService
	HospitalizationPlan   HospitalizationPlanService
	TrimmingCourse        TrimmingCourseService
	TrimmingOption        TrimmingOptionService
	ExaminationType       ExamTypeService
	DiagnosisType         DiagnosisTypeService
	DiagnosisName         DiagnosisNameService
	CheckupType           CheckupTypeService
	Clinic                ClinicService
	Examination           ExaminationService
	Vaccination           VaccinationService
	Occupation            OccupationService
	ChiefComplaintType    ChiefComplaintTypeService
	Inquiry               InquiryService
	InquiryTemplate       InquiryTemplateService
	Company               CompanyService
	PermissionGroup       PermissionGroupService
	BillingConfirmation   BillingConfirmationService
	CarePlanItem          CarePlanItemService
	ShiftEntry            ShiftEntryService
	ShiftTemplate         ShiftTemplateService
	ClinicHoliday         ClinicHolidayService
	TreatmentPlan         TreatmentPlanService
	Vital                 VitalService
	Treatment             TreatmentService
	DailyRecord           DailyRecordService
	MedicalRecordImage    MedicalRecordImageService
	ClinicalPlan          ClinicalPlanService
	Checkup               CheckupService
	Estimate              EstimateService
	MerchandiseItem       MerchandiseItemService
	BillingItem           BillingItemService
	Refund                RefundService
	PasswordReset         PasswordResetService

	// LINE予約
	LineReservationSetting LineReservationSettingService
	ReservationTypeLiff    ReservationTypeLiffService
	ReservationStaff       ReservationStaffService
	ReservationSchedule    ReservationScheduleService
	ReservationAdmin       ReservationAdminService
	LineCustomer           LineCustomerService
	Liff                   LiffService
}

// NewServices はリポジトリからすべてのサービスを初期化して返す
func NewServices(repos *repository.Repositories, notifCfg *ReservationNotificationConfig) *Services {
	notifier := NewReservationNotificationService(notifCfg, repos.LineReservationSetting)
	auditSvc := NewAuditService(repos.Audit)
	tx := repository.NewTransactor(repos.DB())
	pwResetCfg := PasswordResetConfig{
		SMTPHost:    notifCfg.SMTPHost,
		SMTPPort:    notifCfg.SMTPPort,
		SMTPUser:    notifCfg.SMTPUser,
		SMTPPass:    notifCfg.SMTPPass,
		SMTPFrom:    notifCfg.SMTPFrom,
		FrontendURL: notifCfg.FrontendURL,
	}

	return &Services{
		Account:                NewAccountService(repos.Account),
		StaffClinicAssignment:  NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                  auditSvc,
		AnimalSpecies:          NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                  NewOwnerService(repos.Owner),
		Pet:                    NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord),
		Reservation:            NewReservationService(repos.Reservation, tx),
		MedicalRecord:          NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan, repos.LineCustomerMgr),
		Hospitalization:        NewHospitalizationService(repos),
		Accounting:             NewAccountingService(repos.Accounting),
		Trimming:               NewTrimmingService(repos.Reservation, repos.AppointmentTrimmingDetail, tx),
		Inventory:              NewInventoryService(repos.Inventory),
		Staff:                  NewStaffService(repos.Staff, repos.Account, repos.StaffClinicAssignment, repos.Reservation, repos.ShiftEntry, repos.PermissionGroup, repos.ReservationStaff, tx),
		Cage:                   NewCageService(repos.Cage, repos.Hospitalization),
		Medicine:               NewMedicineService(repos.Medicine, repos.Inventory),
		Vaccine:                NewVaccineService(repos.Vaccine),
		Insurance:              NewInsuranceService(repos.Insurance),
		ReservationType:        NewReservationTypeService(repos.ReservationType, repos.Reservation, repos.ReservationTypeUnavailableTime, repos.ReservationTypeOccupation, repos.Occupation),
		ReservationTypeGroup:   NewReservationTypeGroupService(repos.ReservationTypeGroup),
		Consultation:           NewConsultationService(repos.Consultation),
		Procedure:              NewProcedureService(repos.Procedure),
		HospitalizationPlan:    NewHospitalizationPlanService(repos.HospitalizationPlan),
		TrimmingCourse:         NewTrimmingCourseService(repos.TrimmingCourse),
		TrimmingOption:         NewTrimmingOptionService(repos.TrimmingOption),
		ExaminationType:        NewExamTypeService(repos.ExaminationType),
		DiagnosisType:          NewDiagnosisTypeService(repos.DiagnosisType),
		DiagnosisName:          NewDiagnosisNameService(repos.DiagnosisName, repos.DiagnosisType),
		CheckupType:            NewCheckupTypeService(repos.CheckupType),
		Clinic:                 NewClinicService(repos.Clinic),
		Examination:            NewExaminationService(repos.Examination),
		Vaccination:            NewVaccinationService(repos.Vaccination),
		Occupation:             NewOccupationService(repos.Occupation),
		ChiefComplaintType:     NewChiefComplaintTypeService(repos.ChiefComplaintType, repos.Inquiry),
		Inquiry:                NewInquiryService(repos.Inquiry),
		InquiryTemplate:        NewInquiryTemplateService(repos.InquiryTemplate),
		Company:                NewCompanyService(repos.Company),
		PermissionGroup:        NewPermissionGroupService(repos.PermissionGroup),
		BillingConfirmation:    NewBillingConfirmationService(repos.BillingConfirmation),
		CarePlanItem:           NewCarePlanItemService(repos.CarePlanItem),
		ShiftEntry:             NewShiftEntryService(repos.ShiftEntry),
		ShiftTemplate:          NewShiftTemplateService(repos.ShiftTemplate),
		ClinicHoliday:          NewClinicHolidayService(repos.ClinicHoliday),
		TreatmentPlan:          NewTreatmentPlanService(repos.TreatmentPlan),
		Vital:                  NewVitalService(repos.Vital),
		Treatment:              NewTreatmentService(repos),
		DailyRecord:            NewDailyRecordService(repos.DailyRecord),
		MedicalRecordImage:     NewMedicalRecordImageService(repos.MedicalRecordImage),
		ClinicalPlan:           NewClinicalPlanService(repos.ClinicalPlan),
		Checkup:                NewCheckupService(repos.Checkup),
		Estimate:               NewEstimateService(repos.Estimate),
		MerchandiseItem:        NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:            NewBillingItemService(repos.BillingItem),
		Refund:                 NewRefundService(repos.Refund, repos.Accounting),
		PasswordReset:          NewPasswordResetService(&pwResetCfg, repos.Account, repos.PasswordResetToken),
		LineReservationSetting: NewLineReservationSettingService(repos.LineReservationSetting),
		ReservationTypeLiff:    NewReservationTypeLiffService(repos.ReservationTypeLiff, repos.ReservationAdmin, repos.Reservation),
		ReservationStaff:       NewReservationStaffService(repos.ReservationStaff),
		ReservationSchedule:    NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:       NewReservationAdminService(repos.ReservationAdmin, repos.Reservation, tx),
		LineCustomer:           NewLineCustomerService(repos.LineCustomerMgr),
		Liff: NewLiffService(
			repos.LineReservationSetting,
			repos.ReservationTypeLiff,
			repos.ReservationStaff,
			repos.ReservationSchedule,
			repos.ReservationAdmin,
			repos.LineCustomerMgr,
			repos.Owner,
			tx,
			repos.Reservation,
			notifier,
			repos.ReservationTypeUnavailableTime,
			repos.ReservationTypeOccupation,
			repos.TrimmingCourse,
			repos.TrimmingOption,
			repos.AppointmentTrimmingDetail,
		),
	}
}
