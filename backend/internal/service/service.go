package service

import (
	"github.com/animal-ekarte/backend/internal/infra/crypto"
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
	MedicineDoseParam              MedicineDoseParamService
	Vaccine                        VaccineService
	Insurance                      InsuranceService
	ReservationType                ReservationTypeCoreService
	ReservationTypeUnavailableTime ReservationTypeUnavailableTimeService
	ReservationTypeAvailableSlot   ReservationTypeAvailableSlotService
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
	CheckupFieldResult             CheckupFieldResultService
	Estimate                       EstimateService
	ManualArticle                  ManualArticleService
	MerchandiseItem                MerchandiseItemService
	BillingItem                    BillingItemService
	Refund                         RefundService
	PasswordReset                  PasswordResetService

	// FEAT-368: 集計・締め機能
	ClosingSettings     ClosingSettingsService
	PaymentMethodMaster PaymentMethodMasterService
	TrimmingCourseType  TrimmingCourseTypeService
	Campaign            CampaignService
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
	// Q23: トリガー優先順位設定
	LstepTriggerPriority LstepTriggerPriorityService
	// FEAT-379: タグコードマッピング設定
	LstepTagCodeMapping LstepTagCodeMappingService
	// 自動管理タグプレフィックス・条件タグ・送信目的タグ設定
	LstepTagConfig LstepTagConfigService
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
	// 認証: refresh_token JTI ブラックリスト
	TokenBlacklist TokenBlacklistService
	// lab import: 外部検査結果インポートジョブ管理 (Phase 3–4)
	LabImportJob    LabImportJobService
	LabResultImport LabResultImportService
	LabAudit        LabAuditLogger
	// lab report: 検査帳票 read-only クエリ (Phase 4B.2)
	LabReportQuery LabReportQueryService
}

// NewServices はリポジトリからすべてのサービスを初期化して返す。
// cipher は LINE 認証情報（line_channel_secret / line_access_token）の暗号化に使う（H-4）。
// nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
// lstep 連携と同一の cipher を再利用する。
func NewServices(repos *repository.Repositories, notifCfg *ReservationNotificationConfig, cipher *crypto.AESGCMCipher) *Services {
	notifier := NewReservationNotificationService(notifCfg, repos.LineReservationSetting, cipher)
	auditSvc := NewAuditService(repos.Audit)
	// auditTxLogger: 具象 *auditService は tx 内監査の LogEntryTx も実装する（#211）。
	// AuditService インターフェース自体は広げず（既存サービス/モックへ非波及）、tx 内監査を要する
	// checkup のみ narrow な AuditTxLogger に依存させる。コンパイル時保証は audit_service.go の
	// `var _ AuditTxLogger = (*auditService)(nil)`。配線時の comma-ok で、将来 NewAuditService が
	// AuditTxLogger 非実装のラッパを返すよう変わった場合に原因の分かる panic を出す。
	auditTxLogger, ok := auditSvc.(AuditTxLogger)
	if !ok {
		panic("DI wiring error: AuditService concrete does not implement AuditTxLogger (#211 tx-internal audit); check NewAuditService return type")
	}
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

	// ReservationTypeService はサブインターフェースを実装する単一インスタンス。
	// 同一インスタンスを複数フィールドに割り当てることで余分な初期化を避ける。
	reservationTypeSvc := NewReservationTypeService(
		repos.ReservationType,
		repos.ReservationTypeUnavailableTime,
		repos.ReservationTypeOccupation,
		repos.Occupation,
		repos.ReservationTypeAvailableSlot,
	)

	// permissionGroupSvc は PermissionGroupService と EffectivePermissionService の両方を実装する。
	// 同一インスタンスを2フィールドに割り当てることで余分な初期化を避ける。
	permissionGroupSvc := newPermissionGroupServiceImpl(repos.PermissionGroup)

	// staffSvc は StaffService（StaffCoreService / StaffAccountService / StaffPermissionService の合成）を実装する。
	// 同一インスタンスを4フィールドに割り当てることで余分な初期化を避ける。
	staffSvc := NewStaffService(repos.Staff, repos.Account, repos.StaffClinicAssignment, repos.Reservation, repos.ShiftEntry, repos.PermissionGroup, repos.ReservationStaff, tx)

	// resStaffSvc は ReservationStaffService（Core + Exclusion の合成）を実装する。
	resStaffSvc := NewReservationStaffService(repos.ReservationStaff, tx)

	// LSTEP services initialization: LINE予約設定と同一の cipher を再利用する（X-2）。
	lstepSettingsSvc := NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, cipher, auditSvc, repos.ClinicSettings)
	lstepTagSyncSvc := NewLstepTagSyncService(lstepSettingsSvc, repos.Owner, repos.Vaccination, repos.MedicalRecord, repos.Accounting, repos.LstepTagCache, repos.Pet, repos.Prescription, repos.Checkup, repos.Reservation, repos.LstepSyncErrorCounter, repos.LstepTagCodeMapping, repos.BillingItem, repos.LstepTagConfig)
	lstepLifecycleSvc := NewLstepLifecycleService(lstepSettingsSvc, repos.Owner, repos.Pet, repos.LstepTagCache, lstepTagSyncSvc, auditSvc, repos.LstepTagConfig)

	// lab import (Phase 3): 同一 jobSvc インスタンスを LabImportJob/LabResultImport で共有する。
	labImportJobSvc := NewLabImportJobService(repos.LabImportJob, repos.LabImportEvent)

	return &Services{
		Account:               NewAccountService(repos.Account),
		StaffClinicAssignment: NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                 auditSvc,
		AnimalSpecies:         NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                 NewOwnerService(repos.Owner, lstepTagSyncSvc, auditSvc),
		Pet:                   NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord, lstepTagSyncSvc),
		Reservation:           NewReservationServiceWithAvailabilityAndType(repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		MedicalRecord:         NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan, repos.LineCustomerMgr, repos.Reservation, nil, auditSvc, lstepTagSyncSvc),
		MedicalRecordAddendum: NewMedicalRecordAddendumService(repos.MedicalRecordAddendum, repos.MedicalRecord, auditSvc),
		Hospitalization:       NewHospitalizationService(repos),
		Accounting:            NewAccountingService(repos.Accounting, lstepTagSyncSvc, tx, auditTxLogger, repos.PaymentMethodMaster),
		Trimming: NewTrimmingService(
			repos.Reservation,
			repos.ReservationType,
			repos.ReservationStaff,
			repos.ReservationTypeUnavailableTime,
			repos.ReservationTypeAvailableSlot,
			repos.AppointmentTrimmingDetail,
			tx,
		),
		Inventory:                      NewInventoryService(repos.Inventory),
		Staff:                          staffSvc,
		StaffCore:                      staffSvc,
		StaffAccount:                   staffSvc,
		StaffPermission:                staffSvc,
		Cage:                           NewCageService(repos.Cage),
		Medicine:                       NewMedicineServiceWithAudit(repos.Medicine, repos.Inventory, tx, auditTxLogger),
		MedicineDoseParam:              NewMedicineDoseParamService(repos.MedicineDoseParam, repos.Medicine, tx, auditTxLogger),
		Vaccine:                        NewVaccineService(repos.Vaccine),
		Insurance:                      NewInsuranceService(repos.Insurance),
		ReservationType:                reservationTypeSvc,
		ReservationTypeUnavailableTime: reservationTypeSvc,
		ReservationTypeAvailableSlot:   reservationTypeSvc,
		ReservationTypeOccupation:      reservationTypeSvc,
		ReservationTypeGroup:           NewReservationTypeGroupService(repos.ReservationTypeGroup),
		Consultation:                   NewConsultationService(repos.Consultation),
		Procedure:                      NewProcedureService(repos.Procedure),
		HospitalizationPlan:            NewHospitalizationPlanService(repos.HospitalizationPlan),
		TrimmingCourse:                 NewTrimmingCourseService(repos.TrimmingCourse, repos.TrimmingCourseType),
		TrimmingOption:                 NewTrimmingOptionService(repos.TrimmingOption),
		ExaminationType:                NewExamTypeService(repos.ExaminationType),
		DiagnosisType:                  NewDiagnosisTypeService(repos.DiagnosisType),
		DiagnosisName:                  NewDiagnosisNameService(repos.DiagnosisName, repos.DiagnosisType),
		CheckupType:                    NewCheckupTypeService(repos.CheckupType),
		Clinic:                         NewClinicService(repos.Clinic, repos.PermissionGroup, tx),
		Examination:                    NewExaminationService(repos.Examination, repos.MedicalRecord, repos.ExaminationType, auditTxLogger, tx),
		Vaccination:                    NewVaccinationService(repos.Vaccination, repos.Vaccine, lstepTagSyncSvc),
		Occupation:                     NewOccupationService(repos.Occupation),
		ChiefComplaintType:             NewChiefComplaintTypeService(repos.ChiefComplaintType),
		Inquiry:                        NewInquiryService(repos.Inquiry),
		InquiryTemplate:                NewInquiryTemplateService(repos.InquiryTemplate),
		Company:                        NewCompanyService(repos.Company),
		PermissionGroup:                permissionGroupSvc,
		EffectivePermission:            permissionGroupSvc,
		BillingConfirmation:            NewBillingConfirmationService(repos.BillingConfirmation, repos.MedicalRecord),
		CarePlanItem:                   NewCarePlanItemService(repos.CarePlanItem, repos.Hospitalization, repos.Medicine, repos.Procedure),
		ShiftEntry:                     NewShiftEntryService(repos.ShiftEntry),
		ShiftTemplate:                  NewShiftTemplateService(repos.ShiftTemplate),
		ClinicHoliday:                  NewClinicHolidayService(repos.ClinicHoliday),
		TreatmentPlan:                  NewTreatmentPlanService(repos.TreatmentPlan),
		Vital:                          NewVitalService(repos.Vital, repos.MedicalRecord, auditSvc),
		Treatment:                      NewTreatmentServiceWithAudit(repos, auditSvc),
		DailyRecord:                    NewDailyRecordService(repos.DailyRecord),
		MedicalRecordImage:             NewMedicalRecordImageService(repos.MedicalRecordImage, repos.MedicalRecord),
		ClinicalPlan:                   NewClinicalPlanService(repos.ClinicalPlan, repos.MedicalRecord, repos.DiagnosisType, repos.DiagnosisName),
		Checkup:                        NewCheckupService(repos.Checkup, repos.MedicalRecord, repos.CheckupType, nil, lstepTagSyncSvc),
		CheckupFieldResult:             NewCheckupFieldResultService(repos.Checkup, repos.MedicalRecord, repos.CheckupTypeField, repos.CheckupFieldResult, auditTxLogger, tx),
		Estimate:                       NewEstimateService(repos.Estimate),
		ManualArticle:                  NewManualArticleService(repos.ManualArticle),
		MerchandiseItem:                NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:                    NewBillingItemServiceWithCampaign(repos.BillingItem, repos.Accounting, repos.Treatment, tx, repos.Campaign, repos.Owner),
		Refund:                         NewRefundService(repos.Refund, repos.Accounting, auditTxLogger, tx),
		PasswordReset:                  NewPasswordResetService(&pwResetCfg, repos.Account, repos.PasswordResetToken),
		// FEAT-368: 集計・締め機能
		ClosingSettings:           closingSettingsSvc,
		PaymentMethodMaster:       NewPaymentMethodMasterService(repos.PaymentMethodMaster),
		TrimmingCourseType:        NewTrimmingCourseTypeService(repos.TrimmingCourseType),
		Campaign:                  NewCampaignService(repos.Campaign),
		CashRegister:              NewCashRegisterService(repos.CashRegisterClose, repos.Accounting, closingSettingsSvc, repos.PaymentMethodMaster, repos.Clinic),
		AccountingReport:          NewAccountingReportService(repos.Accounting, repos.PaymentMethodMaster, repos.ClinicHoliday, repos.Clinic),
		LineReservationSetting:    NewLineReservationSettingService(repos.LineReservationSetting, cipher),
		ReservationTypeLiff:       NewReservationTypeLiffService(repos.ReservationTypeLiff, repos.ReservationAdmin, repos.Reservation),
		ReservationStaff:          resStaffSvc,
		ReservationStaffCore:      resStaffSvc,
		ReservationStaffExclusion: resStaffSvc,
		ReservationSchedule:       NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:          NewReservationAdminServiceWithAvailabilityAndType(repos.ReservationAdmin, repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		LineCustomer:              NewLineCustomerService(repos.LineCustomerMgr),
		Prescription:              NewPrescriptionService(repos.Prescription, repos.MedicalRecord, lstepTagSyncSvc),
		Aggregation:               NewAggregationService(repos.Ltv, repos.LstepTagCache, repos.LstepTagConfig, lstepSettingsSvc),
		LstepSettings:             lstepSettingsSvc,
		LstepTagSync:              lstepTagSyncSvc,
		LstepLifecycle:            lstepLifecycleSvc,
		LstepTag:                  NewLstepTagService(lstepSettingsSvc, repos.Owner, repos.LstepTagCache, auditSvc, repos.LstepTagConfig),
		LstepTagCodeMapping:       NewLstepTagCodeMappingService(repos.LstepTagCodeMapping),
		LstepTagConfig:            NewLstepTagConfigService(repos.LstepTagConfig),
		Liff: NewLiffServiceWithType(
			repos.LineReservationSetting,
			repos.ReservationTypeLiff,
			repos.ReservationType,
			repos.ReservationStaff,
			repos.ReservationSchedule,
			repos.ReservationAdmin,
			repos.LineCustomerMgr,
			repos.Owner,
			tx,
			repos.Reservation,
			notifier,
			repos.ReservationTypeUnavailableTime,
			repos.ReservationTypeAvailableSlot,
			repos.ReservationTypeOccupation,
			repos.TrimmingCourse,
			repos.TrimmingOption,
			repos.AppointmentTrimmingDetail,
		),
		TokenBlacklist: NewTokenBlacklistService(repos.TokenBlacklist),
		// lab import (Phase 3–4)
		LabImportJob: labImportJobSvc,
		LabResultImport: NewLabResultImportService(
			labImportJobSvc,
			NewLabImportExaminationService(
				repos.Examination,
				repository.NewLabImportDuplicateCheckerDB(repos.DB()),
			),
		),
		LabAudit:       NewLabAuditLogger(auditSvc),
		LabReportQuery: NewLabReportQueryService(repos.Examination),
	}
}
