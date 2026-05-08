package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
)

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	Account                        AccountService
	StaffClinicAssignment          StaffClinicAssignmentService
	Audit                          AuditService
	AnimalSpecies                  AnimalSpeciesService
	Owner                          OwnerService
	Pet                            PetService
	Reservation                    ReservationService
	MedicalRecord                  MedicalRecordService
	MedicalRecordAddendum          MedicalRecordAddendumService
	Hospitalization                HospitalizationService
	Accounting                     AccountingService
	Trimming                       TrimmingService
	Inventory                      InventoryService
	Staff                          StaffService
	StaffCore                      StaffCoreService
	StaffAccount                   StaffAccountService
	StaffPermission                StaffPermissionService
	Cage                           CageService
	Medicine                       MedicineService
	Vaccine                        VaccineService
	Insurance                      InsuranceService
	ReservationType                ReservationTypeCoreService
	ReservationTypeUnavailableTime ReservationTypeUnavailableTimeService
	ReservationTypeOccupation      ReservationTypeOccupationService
	ReservationTypeGroup           ReservationTypeGroupService
	Consultation                   ConsultationService
	Procedure                      ProcedureService
	HospitalizationPlan            HospitalizationPlanService
	TrimmingCourse                 TrimmingCourseService
	TrimmingOption                 TrimmingOptionService
	ExaminationType                ExaminationTypeService
	DiagnosisType                  DiagnosisTypeService
	DiagnosisName                  DiagnosisNameService
	CheckupType                    CheckupTypeService
	Clinic                         ClinicService
	Examination                    ExaminationService
	Vaccination                    VaccinationService
	Occupation                     OccupationService
	ChiefComplaintType             ChiefComplaintTypeService
	Inquiry                        InquiryService
	InquiryTemplate                InquiryTemplateService
	Company                        CompanyService
	PermissionGroup                PermissionGroupService
	EffectivePermission            EffectivePermissionService
	BillingConfirmation            BillingConfirmationService
	CarePlanItem                   CarePlanItemService
	ShiftEntry                     ShiftEntryService
	ShiftTemplate                  ShiftTemplateService
	ClinicHoliday                  ClinicHolidayService
	TreatmentPlan                  TreatmentPlanService
	Vital                          VitalService
	Treatment                      TreatmentService
	DailyRecord                    DailyRecordService
	MedicalRecordImage             MedicalRecordImageService
	ClinicalPlan                   ClinicalPlanService
	Checkup                        CheckupService
	Estimate                       EstimateService
	MerchandiseItem                MerchandiseItemService
	BillingItem                    BillingItemService
	Refund                         RefundService
	PasswordReset                  PasswordResetService

	// FEAT-368: 集計・締め機能
	ClosingSettings     ClosingSettingsService
	PaymentMethodMaster PaymentMethodMasterService
	CashRegister        CashRegisterService
	AccountingReport    AccountingReportService

	// LINE予約
	LineReservationSetting    LineReservationSettingService
	ReservationTypeLiff       ReservationTypeLiffService
	ReservationStaff          ReservationStaffService
	ReservationStaffCore      ReservationStaffCoreService
	ReservationStaffExclusion ReservationStaffExclusionService
	ReservationSchedule       ReservationScheduleService
	ReservationAdmin          ReservationAdminService
	LineCustomer              LineCustomerService
	Liff                      LiffService

	// LSTEP / LINE連携
	LstepSettings  LstepSettingsService
	LstepTagSync   LstepTagSyncService
	LstepLifecycle LstepLifecycleService
	LstepTag       LstepTagService
	SharedFile     SharedFileService
	// LSTEP-BE-009: 処方薬記録
	Prescription PrescriptionService
	// LSTEP-BE-010: LTV集計 → 顧客集計ドメインに統一
	Aggregation AggregationService
	// LSTEP-BE-012: 慢性疾患フラグ
	ChronicCondition ChronicConditionService
	// LSTEP-BE-013: LINE個別送信
	LineSend LineSendService
	// LSTEP-BE-014: ノーショウ検知バッチ
	LstepBatch LstepBatchService
	// FEAT-383: 自動配信トリガー
	LstepDeliveryTrigger LstepDeliveryTriggerService
	// LSTEP-BE-021: LINE User ID 自動取得・飼い主紐付け
	LineLink LineLinkService
	// LSTEP-BE-020: タグ集計・タグ別飼い主検索
	LstepTagSummary LstepTagSummaryService
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	CheckupSync CheckupSyncService
	// FEAT-384: 自動配信トリガー監視
	LstepDeliveryMonitor LstepDeliveryMonitorService
	// FEAT-385: Lステップ CSV インポート・分析
	LstepCsvImport LstepCsvImportService
	LstepAnalytics LstepAnalyticsService
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

	// ClosingSettings は CashRegister より先に初期化する（依存関係のため）
	closingSettingsSvc := NewClosingSettingsService(repos.ClinicSettings, repos.ClosingSpecialPeriod, repos.ClinicHoliday)

	// ReservationTypeService は3サブインターフェースを実装する単一インスタンス。
	// 同一インスタンスを3フィールドに割り当てることで余分な初期化を避ける。
	reservationTypeSvc := NewReservationTypeService(repos.ReservationType, repos.ReservationTypeUnavailableTime, repos.ReservationTypeOccupation, repos.Occupation)

	// permissionGroupSvc は PermissionGroupService と EffectivePermissionService の両方を実装する。
	// 同一インスタンスを2フィールドに割り当てることで余分な初期化を避ける。
	permissionGroupSvc := newPermissionGroupServiceImpl(repos.PermissionGroup)

	// staffSvc は StaffService（StaffCoreService / StaffAccountService / StaffPermissionService の合成）を実装する。
	// 同一インスタンスを4フィールドに割り当てることで余分な初期化を避ける。
	staffSvc := NewStaffService(repos.Staff, repos.Account, repos.StaffClinicAssignment, repos.Reservation, repos.ShiftEntry, repos.PermissionGroup, repos.ReservationStaff, tx)

	// resStaffSvc は ReservationStaffService（Core + Exclusion の合成）を実装する。
	resStaffSvc := NewReservationStaffService(repos.ReservationStaff, tx)

	// LSTEP services initialization with nil cipher (production code in main.go will override with encrypted cipher)
	lstepSettingsSvc := NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, nil, auditSvc, repos.ClinicSettings)
	lstepTagSyncSvc := NewLstepTagSyncService(lstepSettingsSvc, repos.Owner, repos.Vaccination, repos.MedicalRecord, repos.Accounting, repos.LstepTagCache, repos.Pet, repos.Prescription, repos.Checkup, repos.Reservation, repos.LstepSyncErrorCounter, repos.LstepTagCodeMapping, repos.BillingItem)
	lstepLifecycleSvc := NewLstepLifecycleService(lstepSettingsSvc, repos.Owner, repos.Pet, repos.LstepTagCache, lstepTagSyncSvc, auditSvc)

	return &Services{
		Account:                        NewAccountService(repos.Account),
		StaffClinicAssignment:          NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                          auditSvc,
		AnimalSpecies:                  NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                          NewOwnerService(repos.Owner, lstepTagSyncSvc),
		Pet:                            NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord),
		Reservation:                    NewReservationService(repos.Reservation, tx),
		MedicalRecord:                  NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan, repos.LineCustomerMgr, nil, nil, auditSvc),
		MedicalRecordAddendum:          NewMedicalRecordAddendumService(repos.MedicalRecordAddendum, repos.MedicalRecord, auditSvc),
		Hospitalization:                NewHospitalizationService(repos),
		Accounting:                     NewAccountingService(repos.Accounting),
		Trimming:                       NewTrimmingService(repos.Reservation, repos.AppointmentTrimmingDetail, tx),
		Inventory:                      NewInventoryService(repos.Inventory),
		Staff:                          staffSvc,
		StaffCore:                      staffSvc,
		StaffAccount:                   staffSvc,
		StaffPermission:                staffSvc,
		Cage:                           NewCageService(repos.Cage),
		Medicine:                       NewMedicineService(repos.Medicine, repos.Inventory, tx),
		Vaccine:                        NewVaccineService(repos.Vaccine),
		Insurance:                      NewInsuranceService(repos.Insurance),
		ReservationType:                reservationTypeSvc,
		ReservationTypeUnavailableTime: reservationTypeSvc,
		ReservationTypeOccupation:      reservationTypeSvc,
		ReservationTypeGroup:           NewReservationTypeGroupService(repos.ReservationTypeGroup),
		Consultation:                   NewConsultationService(repos.Consultation),
		Procedure:                      NewProcedureService(repos.Procedure),
		HospitalizationPlan:            NewHospitalizationPlanService(repos.HospitalizationPlan),
		TrimmingCourse:                 NewTrimmingCourseService(repos.TrimmingCourse),
		TrimmingOption:                 NewTrimmingOptionService(repos.TrimmingOption),
		ExaminationType:                NewExamTypeService(repos.ExaminationType),
		DiagnosisType:                  NewDiagnosisTypeService(repos.DiagnosisType),
		DiagnosisName:                  NewDiagnosisNameService(repos.DiagnosisName, repos.DiagnosisType),
		CheckupType:                    NewCheckupTypeService(repos.CheckupType),
		Clinic:                         NewClinicService(repos.Clinic, repos.PermissionGroup),
		Examination:                    NewExaminationService(repos.Examination),
		Vaccination:                    NewVaccinationService(repos.Vaccination),
		Occupation:                     NewOccupationService(repos.Occupation),
		ChiefComplaintType:             NewChiefComplaintTypeService(repos.ChiefComplaintType),
		Inquiry:                        NewInquiryService(repos.Inquiry),
		InquiryTemplate:                NewInquiryTemplateService(repos.InquiryTemplate),
		Company:                        NewCompanyService(repos.Company),
		PermissionGroup:                permissionGroupSvc,
		EffectivePermission:            permissionGroupSvc,
		BillingConfirmation:            NewBillingConfirmationService(repos.BillingConfirmation),
		CarePlanItem:                   NewCarePlanItemService(repos.CarePlanItem),
		ShiftEntry:                     NewShiftEntryService(repos.ShiftEntry),
		ShiftTemplate:                  NewShiftTemplateService(repos.ShiftTemplate),
		ClinicHoliday:                  NewClinicHolidayService(repos.ClinicHoliday),
		TreatmentPlan:                  NewTreatmentPlanService(repos.TreatmentPlan),
		Vital:                          NewVitalService(repos.Vital, repos.MedicalRecord, auditSvc),
		Treatment:                      NewTreatmentService(repos),
		DailyRecord:                    NewDailyRecordService(repos.DailyRecord),
		MedicalRecordImage:             NewMedicalRecordImageService(repos.MedicalRecordImage),
		ClinicalPlan:                   NewClinicalPlanService(repos.ClinicalPlan),
		Checkup:                        NewCheckupService(repos.Checkup, nil),
		Estimate:                       NewEstimateService(repos.Estimate),
		MerchandiseItem:                NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:                    NewBillingItemService(repos.BillingItem, repos.Accounting, repos.Treatment),
		Refund:                         NewRefundService(repos.Refund, repos.Accounting),
		PasswordReset:                  NewPasswordResetService(&pwResetCfg, repos.Account, repos.PasswordResetToken),
		// FEAT-368: 集計・締め機能
		ClosingSettings:           closingSettingsSvc,
		PaymentMethodMaster:       NewPaymentMethodMasterService(repos.PaymentMethodMaster),
		CashRegister:              NewCashRegisterService(repos.CashRegisterClose, repos.Accounting, closingSettingsSvc, repos.PaymentMethodMaster),
		AccountingReport:          NewAccountingReportService(repos.Accounting, repos.PaymentMethodMaster, repos.ClinicHoliday),
		LineReservationSetting:    NewLineReservationSettingService(repos.LineReservationSetting),
		ReservationTypeLiff:       NewReservationTypeLiffService(repos.ReservationTypeLiff, repos.ReservationAdmin, repos.Reservation),
		ReservationStaff:          resStaffSvc,
		ReservationStaffCore:      resStaffSvc,
		ReservationStaffExclusion: resStaffSvc,
		ReservationSchedule:       NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:          NewReservationAdminService(repos.ReservationAdmin, repos.Reservation, tx),
		LineCustomer:              NewLineCustomerService(repos.LineCustomerMgr),
		Prescription:              NewPrescriptionService(repos.Prescription, repos.MedicalRecord),
		Aggregation:               NewAggregationService(repos.Ltv, repos.LstepTagCache),
		LstepSettings:             lstepSettingsSvc,
		LstepTagSync:              lstepTagSyncSvc,
		LstepLifecycle:            lstepLifecycleSvc,
		LstepTag:                  NewLstepTagService(lstepSettingsSvc, repos.Owner, repos.LstepTagCache, auditSvc),
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
