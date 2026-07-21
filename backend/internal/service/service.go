package service

import (
	"github.com/animal-ekarte/backend/internal/infra"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/reservation"
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
	// MedicalRecord: BE9-2D ⑦ で実装は internal/medicalrecord へ移動済み。残存 consumer =
	// reservation_handler（AutoCreateFromReservation/DeleteDraftFromReservation）のみのため
	// medicalrecord 型 field として残置し cmd/api/main.go が構築後に代入する（⑤ Hospitalization 先例。
	// 削除 = reservation domain 移行時）。
	MedicalRecord                  medicalrecord.MedicalRecordService
	Accounting                     AccountingService
	Trimming                       TrimmingService
	Inventory                      InventoryService
	Staff                          StaffService
	StaffCore                      StaffCoreService
	StaffAccount                   StaffAccountService
	StaffPermission                StaffPermissionService
	Insurance                      InsuranceService
	ReservationType                reservation.ReservationTypeCoreService
	ReservationTypeUnavailableTime reservation.ReservationTypeUnavailableTimeService
	ReservationTypeAvailableSlot   reservation.ReservationTypeAvailableSlotService
	ReservationTypeOccupation      reservation.ReservationTypeOccupationService
	ReservationTypeGroup           reservation.ReservationTypeGroupService
	TrimmingCourse                 TrimmingCourseService
	TrimmingOption                 TrimmingOptionService
	Clinic                         ClinicService
	Occupation                     OccupationService
	Company                        CompanyService
	PermissionGroup                PermissionGroupService
	EffectivePermission            EffectivePermissionService
	BillingConfirmation            BillingConfirmationService
	ShiftEntry                     ShiftEntryService
	ShiftTemplate                  ShiftTemplateService
	ClinicHoliday                  ClinicHolidayService
	// Vital / MedicalRecordImage / ClinicalPlan（④a）/ Treatment（④b）: BE9-2D — service + handler
	// とも internal/medicalrecord へ移設済み。composition root (cmd/api/main.go) が medicalrecord.NewX
	// で直接構築し medicalrecord.NewHandler へ合成するため、Services には保持しない
	// （treatment-plan / daily-record / estimate は internal/service に残る）。
	Estimate EstimateService
	// ManualArticle: BE9-2B — moved to internal/manualarticle (aggregator-free domain
	// package). No longer constructed here; see cmd/api/main.go.
	MerchandiseItem     MerchandiseItemService
	BillingItem         BillingItemService
	Refund              RefundService
	PasswordReset       PasswordResetService
	ReservationNotifier ReservationNotifier

	// FEAT-368: 集計・締め機能
	ClosingSettings     ClosingSettingsService
	PaymentMethodMaster PaymentMethodMasterService
	TrimmingCourseType  TrimmingCourseTypeService
	Campaign            CampaignService
	CashRegister        CashRegisterService
	AccountingReport    AccountingReportService

	// LINE予約
	LineReservationSetting    LineReservationSettingService
	ReservationTypeLiff       reservation.ReservationTypeLiffService
	ReservationStaff          reservation.ReservationStaffService
	ReservationStaffCore      reservation.ReservationStaffCoreService
	ReservationStaffExclusion reservation.ReservationStaffExclusionService
	ReservationSchedule       reservation.ReservationScheduleService
	ReservationAdmin          ReservationAdminService
	LineCustomer              LineCustomerService
	Liff                      LiffService

	// LSTEP / LINE連携
	LstepSettings  LstepSettingsService
	LstepTagSync   LstepTagSyncService
	LstepLifecycle LstepLifecycleService
	LstepTag       LstepTagService
	SharedFile     SharedFileService
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
	// 認証: JWT 発行・検証
	Token TokenService
	// 認証: ログイン照合・クリニック解決・実効権限計算
	Auth AuthService
	// lab import/report (Phase 3–4): BE9-2D sub-batch③ で実装を internal/medicalrecord へ移動
	// （leaf domain）し、handler ともども cmd/api/main.go で直接構築する（ADR-006 aggregator 非経由）。
	// もはや Services の field ではない。
}

// NewServices はリポジトリからすべてのサービスを初期化して返す。
// cipher は LINE 認証情報（line_channel_secret / line_access_token）の暗号化に使う（H-4）。
// nil の場合は暗号化なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
// lstep 連携と同一の cipher を再利用する。
func NewServices(repos *repository.Repositories, notifCfg *ReservationNotificationConfig, cipher *crypto.AESGCMCipher, sharedStorage infra.FileStorage, jwtSecret string) *Services {
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
	reservationTypeSvc := reservation.NewReservationTypeService(
		repos.ReservationType,
		repos.ReservationTypeUnavailableTime,
		repos.ReservationTypeOccupation,
		repos.Occupation,
		repos.ReservationTypeGroup,
		repos.ReservationTypeAvailableSlot,
	)

	// permissionGroupSvc は PermissionGroupService と EffectivePermissionService の両方を実装する。
	// 同一インスタンスを2フィールドに割り当てることで余分な初期化を避ける。
	permissionGroupSvc := newPermissionGroupServiceImpl(repos.PermissionGroup)

	// staffSvc は StaffService（StaffCoreService / StaffAccountService / StaffPermissionService の合成）を実装する。
	// 同一インスタンスを4フィールドに割り当てることで余分な初期化を避ける。
	staffSvc := NewStaffService(repos.Staff, repos.Account, repos.StaffClinicAssignment, repos.Reservation, repos.ShiftEntry, repos.PermissionGroup, repos.ReservationStaff, repos.Occupation, tx)

	// resStaffSvc は ReservationStaffService（Core + Exclusion の合成）を実装する。
	resStaffSvc := reservation.NewReservationStaffService(repos.ReservationStaff, tx)

	// LSTEP services initialization: LINE予約設定と同一の cipher を再利用する（X-2）。
	lstepSettingsSvc := NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, cipher, auditSvc, repos.ClinicSettings)
	lstepTagSyncSvc := NewLstepTagSyncFromRepos(repos, lstepSettingsSvc)
	lstepLifecycleSvc := NewLstepLifecycleService(lstepSettingsSvc, repos.Owner, repos.Pet, repos.LstepTagCache, lstepTagSyncSvc, auditSvc, repos.LstepTagConfig, tx, auditTxLogger)

	// G9-1: 旧 main.go 二段階DI（NewServices 呼び出し後の再構築ブロック）をここに統合。
	// main.go 由来の元の構築順序をそのまま保持する。
	sharedFileSvc := NewSharedFileService(repos.SharedFile, repos.Owner, sharedStorage)
	// LSTEP-BE-012: 慢性疾患フラグ
	chronicConditionSvc := NewChronicConditionService(repos.ChronicCondition, repos.Pet, lstepTagSyncSvc)
	// LSTEP-BE-013: LINE個別送信
	lineSendSvc := NewLineSendService(lstepSettingsSvc, repos.Owner, sharedFileSvc, repos.LstepTagCache, auditSvc, repos.LineSendLog, repos.LstepTagConfig)
	// LSTEP-BE-021: LINE User ID 自動取得・飼い主紐付け
	lineLinkSvc := NewLineLinkService(repos.Owner, repos.LineLinkToken, repos.LineReservationSetting, auditSvc, cipher)
	// LSTEP-BE-020: タグ集計・タグ別飼い主検索
	lstepTagSummarySvc := NewLstepTagSummaryService(repos.LstepTagCache)
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	checkupSyncSvc := NewCheckupSyncService(repos.CheckupSync, repos.Owner, repos.Pet, repos.LstepTagCache, lstepSettingsSvc, auditSvc)
	// FEAT-384: 自動配信トリガー監視
	lstepDeliveryMonitorSvc := NewLstepDeliveryMonitorService(repos.LstepDeliveryTriggerLog)
	// Q23: トリガー優先順位設定
	lstepTriggerPrioritySvc := NewLstepTriggerPriorityService(repos.LstepTriggerPriority)
	// FEAT-383: 自動配信トリガー（LstepBatch / MedicalRecord / Checkup より先に初期化）
	lstepDeliveryTriggerSvc := NewLstepDeliveryTriggerService(repos.Owner, repos.MedicalRecord, repos.Vaccination, repos.BillingItem, repos.Pet, repos.LstepTagCache, repos.LstepDeliveryTriggerLog, lstepSettingsSvc, lstepTriggerPrioritySvc)
	// FEAT-383: イベントフック注入（LstepDeliveryTrigger 確定後に構築）
	// BE9-2D: checkup/checkup-field-result/checkup-type/vaccine/vaccination/inquiry/
	// inquiry-template/prescription services moved to internal/medicalrecord and are now
	// constructed directly in cmd/api/main.go (ADR-006 aggregator 非経由) — no longer fields
	// on Services. The LSTEP tag-sync / delivery-trigger deps they need are exposed to main.go
	// via svcs.LstepTagSync / svcs.LstepDeliveryTrigger.
	// LSTEP-BE-014: ノーショウ検知バッチ（LstepDeliveryTrigger 確定後に初期化）
	lstepBatchSvc := NewLstepBatchService(repos.Reservation, lstepTagSyncSvc, repos.Clinic, repos.MedicalRecord, auditSvc, lstepSettingsSvc, lstepDeliveryTriggerSvc)
	// FEAT-385: Lステップ CSV インポート・分析
	lstepCsvImportSvc := NewLstepCsvImportService(repos.DB(), repos.LstepCsvImport, repos.LstepFriendAttributeSnapshot, repos.Owner)
	lstepAnalyticsSvc := NewLstepAnalyticsService(repos.Owner, repos.LstepDeliveryTriggerLog, repos.LstepFriendAttributeSnapshot)

	tokenBlacklistSvc := NewTokenBlacklistService(repos.TokenBlacklist)

	return &Services{
		Account:               NewAccountService(repos.Account),
		StaffClinicAssignment: NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                 auditSvc,
		AnimalSpecies:         NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                 NewOwnerService(repos.Owner, repos.Insurance, lstepTagSyncSvc, auditSvc),
		Pet:                   NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord, lstepTagSyncSvc),
		Reservation:           NewReservationServiceWithAvailabilityAndType(repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		Accounting:            NewAccountingService(repos.Accounting, repos.MedicalRecord, repos.Hospitalization, repos.Reservation, lstepTagSyncSvc, tx, auditTxLogger, repos.PaymentMethodMaster),
		Trimming: NewTrimmingService(
			repos.Reservation,
			repos.ReservationType,
			repos.ReservationStaff,
			repos.ReservationTypeUnavailableTime,
			repos.ReservationTypeAvailableSlot,
			repos.AppointmentTrimmingDetail,
			repos.TrimmingCourse,
			repos.TrimmingOption,
			tx,
		),
		Inventory:                      NewInventoryService(repos.Inventory),
		Staff:                          staffSvc,
		StaffCore:                      staffSvc,
		StaffAccount:                   staffSvc,
		StaffPermission:                staffSvc,
		Insurance:                      NewInsuranceService(repos.Insurance),
		ReservationType:                reservationTypeSvc,
		ReservationTypeUnavailableTime: reservationTypeSvc,
		ReservationTypeAvailableSlot:   reservationTypeSvc,
		ReservationTypeOccupation:      reservationTypeSvc,
		ReservationTypeGroup:           reservation.NewReservationTypeGroupService(repos.ReservationTypeGroup),
		TrimmingCourse:                 NewTrimmingCourseService(repos.TrimmingCourse, repos.TrimmingCourseType),
		TrimmingOption:                 NewTrimmingOptionService(repos.TrimmingOption),
		Clinic:                         NewClinicService(repos.Clinic, repos.PermissionGroup, tx),
		Occupation:                     NewOccupationService(repos.Occupation),
		Company:                        NewCompanyService(repos.Company),
		PermissionGroup:                permissionGroupSvc,
		EffectivePermission:            permissionGroupSvc,
		BillingConfirmation:            NewBillingConfirmationService(repos.BillingConfirmation, repos.MedicalRecord, tx),
		ShiftEntry:                     NewShiftEntryService(repos.ShiftEntry),
		ShiftTemplate:                  NewShiftTemplateService(repos.ShiftTemplate),
		ClinicHoliday:                  NewClinicHolidayService(repos.ClinicHoliday),
		// Vital / MedicalRecordImage / ClinicalPlan: BE9-2D sub-batch④a — service + handler とも
		// internal/medicalrecord へ移設済み。composition root (cmd/api/main.go) が medicalrecord.NewX
		// で直接構築する（Services には保持しない）。
		Estimate:            NewEstimateService(repos.Estimate, repos.MedicalRecord, repos.Reservation, repos.StaffClinicAssignment, auditSvc, tx),
		MerchandiseItem:     NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:         NewBillingItemServiceWithCampaign(repos.BillingItem, repos.Accounting, repos.Treatment, tx, repos.TrimmingCourse, repos.TrimmingOption, repos.Campaign, repos.Owner),
		Refund:              NewRefundService(repos.Refund, repos.Accounting, auditTxLogger, tx),
		PasswordReset:       NewPasswordResetService(&pwResetCfg, repos.Account, repos.PasswordResetToken),
		ReservationNotifier: notifier,
		// FEAT-368: 集計・締め機能
		ClosingSettings:           closingSettingsSvc,
		PaymentMethodMaster:       NewPaymentMethodMasterService(repos.PaymentMethodMaster),
		TrimmingCourseType:        NewTrimmingCourseTypeService(repos.TrimmingCourseType),
		Campaign:                  NewCampaignService(repos.Campaign, repos.MerchandiseItem),
		CashRegister:              NewCashRegisterService(repos.CashRegisterClose, repos.Accounting, closingSettingsSvc, repos.PaymentMethodMaster, repos.Clinic),
		AccountingReport:          NewAccountingReportService(repos.Accounting, repos.PaymentMethodMaster, repos.ClinicHoliday, repos.Clinic),
		LineReservationSetting:    NewLineReservationSettingService(repos.LineReservationSetting, cipher),
		ReservationTypeLiff:       reservation.NewReservationTypeLiffService(repos.ReservationTypeLiff, repos.Reservation),
		ReservationStaff:          resStaffSvc,
		ReservationStaffCore:      resStaffSvc,
		ReservationStaffExclusion: resStaffSvc,
		ReservationSchedule:       reservation.NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:          NewReservationAdminServiceWithAvailabilityAndType(repos.ReservationAdmin, repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		LineCustomer:              NewLineCustomerService(repos.LineCustomerMgr, repos.Owner),
		Aggregation:               NewAggregationService(repos.Ltv, repos.LstepTagCache, repos.LstepTagConfig, lstepSettingsSvc),
		LstepSettings:             lstepSettingsSvc,
		LstepTagSync:              lstepTagSyncSvc,
		LstepLifecycle:            lstepLifecycleSvc,
		LstepTag:                  NewLstepTagService(lstepSettingsSvc, repos.Owner, repos.LstepTagCache, auditSvc, repos.LstepTagConfig),
		SharedFile:                sharedFileSvc,
		ChronicCondition:          chronicConditionSvc,
		LineSend:                  lineSendSvc,
		LstepBatch:                lstepBatchSvc,
		LstepDeliveryTrigger:      lstepDeliveryTriggerSvc,
		LstepTriggerPriority:      lstepTriggerPrioritySvc,
		LstepTagCodeMapping:       NewLstepTagCodeMappingService(repos.LstepTagCodeMapping),
		LstepTagConfig:            NewLstepTagConfigService(repos.LstepTagConfig),
		LineLink:                  lineLinkSvc,
		LstepTagSummary:           lstepTagSummarySvc,
		CheckupSync:               checkupSyncSvc,
		LstepDeliveryMonitor:      lstepDeliveryMonitorSvc,
		LstepCsvImport:            lstepCsvImportSvc,
		LstepAnalytics:            lstepAnalyticsSvc,
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
			repos.Vaccination,
		),
		TokenBlacklist: tokenBlacklistSvc,
		Token:          NewTokenService(jwtSecret, tokenBlacklistSvc),
		Auth:           NewAuthService(NewAccountService(repos.Account), staffSvc, permissionGroupSvc),
	}
}
