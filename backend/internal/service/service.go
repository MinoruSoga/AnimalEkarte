package service

import (
	"context"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// LegacyLstepDependencies is the typed compatibility input for real legacy consumers.
// It does not retain LSTEP services in Services; remove it with the service aggregator in BE9-2F.
type LegacyLstepDependencies struct {
	TagSync           LstepTagSyncService
	LineCustomers     reservation.LiffLineCustomerRepository
	EncryptCredential func(value string) (string, error)
	DecryptCredential func(ctx context.Context, value string) string
	NewLinePusher     func(channelToken string) reservation.LinePusher
}

// Services はすべてのサービスを保持するDIコンテナ
type Services struct {
	Account               AccountService
	StaffClinicAssignment StaffClinicAssignmentService
	Audit                 AuditService
	AnimalSpecies         AnimalSpeciesService
	Owner                 OwnerService
	Pet                   PetService
	Reservation           reservation.ReservationService
	// MedicalRecord: BE9-2D ⑦ で実装は internal/medicalrecord へ移動済み。残存 consumer =
	// reservation_handler（AutoCreateFromReservation/DeleteDraftFromReservation）のみのため
	// medicalrecord 型 field として残置し cmd/api/main.go が構築後に代入する（⑤ Hospitalization 先例。
	// 削除 = reservation domain 移行時）。
	MedicalRecord                  medicalrecord.MedicalRecordService
	Accounting                     billing.AccountingService
	Trimming                       TrimmingService
	Inventory                      InventoryService
	Staff                          StaffService
	StaffCore                      StaffCoreService
	StaffAccount                   StaffAccountService
	StaffPermission                StaffPermissionService
	Insurance                      billing.InsuranceService
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
	BillingConfirmation            billing.BillingConfirmationService
	ShiftEntry                     ShiftEntryService
	ShiftTemplate                  ShiftTemplateService
	ClinicHoliday                  ClinicHolidayService
	// Vital / MedicalRecordImage / ClinicalPlan（④a）/ Treatment（④b）: BE9-2D — service + handler
	// とも internal/medicalrecord へ移設済み。composition root (cmd/api/main.go) が medicalrecord.NewX
	// で直接構築し medicalrecord.NewHandler へ合成するため、Services には保持しない
	// （treatment-plan / daily-record / estimate は internal/service に残る）。
	Estimate billing.EstimateService
	// ManualArticle: BE9-2B — moved to internal/manualarticle (aggregator-free domain
	// package). No longer constructed here; see cmd/api/main.go.
	MerchandiseItem     MerchandiseItemService
	BillingItem         billing.BillingItemService
	Refund              billing.RefundService
	PasswordReset       PasswordResetService
	ReservationNotifier reservation.ReservationNotifier

	// FEAT-368: 集計・締め機能
	ClosingSettings     ClosingSettingsService
	PaymentMethodMaster billing.PaymentMethodMasterService
	TrimmingCourseType  TrimmingCourseTypeService
	Campaign            billing.CampaignService
	CashRegister        billing.CashRegisterService
	AccountingReport    billing.AccountingReportService

	// LINE予約
	LineReservationSetting    reservation.LineReservationSettingService
	ReservationTypeLiff       reservation.ReservationTypeLiffService
	ReservationStaff          reservation.ReservationStaffService
	ReservationStaffCore      reservation.ReservationStaffCoreService
	ReservationStaffExclusion reservation.ReservationStaffExclusionService
	ReservationSchedule       reservation.ReservationScheduleService
	ReservationAdmin          reservation.ReservationAdminService
	Liff                      reservation.LiffService

	// LSTEP-BE-012: 慢性疾患フラグ。BE9-2Eまでtarget TagSyncを注入して残置。
	ChronicCondition ChronicConditionService
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
// LSTEP-owned implementations arrive only through the typed compatibility input; this
// legacy aggregator neither constructs nor stores LSTEP services.
func NewServices(
	repos *repository.Repositories,
	notifCfg *reservation.ReservationNotificationConfig,
	jwtSecret string,
	auditSvc AuditKernel,
	lstepDeps *LegacyLstepDependencies,
	lineReservationSettings reservation.LineReservationSettingRepository,
) *Services {
	notifier := reservation.NewReservationNotificationService(notifCfg, lineReservationSettings,
		lstepDeps.DecryptCredential,
		lstepDeps.NewLinePusher,
		smtpSendAdapter)
	auditTxLogger := AuditTxLogger(auditSvc)
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

	// LSTEP-BE-012: 慢性疾患フラグ
	chronicConditionSvc := NewChronicConditionService(repos.ChronicCondition, repos.Pet, lstepDeps.TagSync)

	tokenBlacklistSvc := NewTokenBlacklistService(repos.TokenBlacklist)

	return &Services{
		Account:               NewAccountService(repos.Account),
		StaffClinicAssignment: NewStaffClinicAssignmentService(repos.StaffClinicAssignment),
		Audit:                 auditSvc,
		AnimalSpecies:         NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
		Owner:                 NewOwnerService(repos.Owner, repos.Insurance, lstepDeps.TagSync, auditSvc),
		Pet:                   NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord, lstepDeps.TagSync),
		Reservation:           reservation.NewReservationServiceWithAvailabilityAndType(repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		Accounting:            billing.NewAccountingService(repos.Accounting, repos.MedicalRecord, repos.Hospitalization, repos.Reservation, lstepDeps.TagSync, tx, billingAuditTxAdapter{inner: auditTxLogger}, repos.PaymentMethodMaster),
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
		Insurance:                      billing.NewInsuranceService(repos.Insurance),
		ReservationType:                reservationTypeSvc,
		ReservationTypeUnavailableTime: reservationTypeSvc,
		ReservationTypeAvailableSlot:   reservationTypeSvc,
		ReservationTypeOccupation:      reservationTypeSvc,
		ReservationTypeGroup:           reservation.NewReservationTypeGroupService(repos.ReservationTypeGroup),
		TrimmingCourse:                 NewTrimmingCourseService(repos.TrimmingCourse, repos.TrimmingCourseType, tx),
		TrimmingOption:                 NewTrimmingOptionService(repos.TrimmingOption, tx),
		Clinic:                         NewClinicService(repos.Clinic, repos.PermissionGroup, tx),
		Occupation:                     NewOccupationService(repos.Occupation),
		Company:                        NewCompanyService(repos.Company),
		PermissionGroup:                permissionGroupSvc,
		EffectivePermission:            permissionGroupSvc,
		BillingConfirmation:            billing.NewBillingConfirmationService(repos.BillingConfirmation, repos.MedicalRecord, tx),
		ShiftEntry:                     NewShiftEntryService(repos.ShiftEntry),
		ShiftTemplate:                  NewShiftTemplateService(repos.ShiftTemplate),
		ClinicHoliday:                  NewClinicHolidayService(repos.ClinicHoliday),
		// Vital / MedicalRecordImage / ClinicalPlan: BE9-2D sub-batch④a — service + handler とも
		// internal/medicalrecord へ移設済み。composition root (cmd/api/main.go) が medicalrecord.NewX
		// で直接構築する（Services には保持しない）。
		Estimate:            billing.NewEstimateService(repos.Estimate, repos.MedicalRecord, repos.Reservation, repos.StaffClinicAssignment, billingAuditAdapter{inner: auditSvc}, tx),
		MerchandiseItem:     NewMerchandiseItemService(repos.MerchandiseItem),
		BillingItem:         billing.NewBillingItemServiceWithCampaign(repos.BillingItem, repos.Accounting, repos.Treatment, tx, repos.TrimmingCourse, repos.TrimmingOption, repos.Campaign, repos.Owner),
		Refund:              billing.NewRefundService(repos.Refund, repos.Accounting, billingAuditTxAdapter{inner: auditTxLogger}, tx),
		PasswordReset:       NewPasswordResetService(&pwResetCfg, repos.Account, repos.PasswordResetToken),
		ReservationNotifier: notifier,
		// FEAT-368: 集計・締め機能
		ClosingSettings:     closingSettingsSvc,
		PaymentMethodMaster: billing.NewPaymentMethodMasterService(repos.PaymentMethodMaster),
		TrimmingCourseType:  NewTrimmingCourseTypeService(repos.TrimmingCourseType, tx),
		Campaign:            billing.NewCampaignService(repos.Campaign, repos.MerchandiseItem),
		CashRegister:        billing.NewCashRegisterService(repos.CashRegisterClose, repos.Accounting, closingSettingsSvc, repos.PaymentMethodMaster, repos.Clinic),
		AccountingReport:    billing.NewAccountingReportService(repos.Accounting, repos.PaymentMethodMaster, repos.ClinicHoliday, repos.Clinic),
		LineReservationSetting: reservation.NewLineReservationSettingService(lineReservationSettings,
			lstepDeps.EncryptCredential,
			lstepDeps.DecryptCredential),
		ReservationTypeLiff:       reservation.NewReservationTypeLiffService(repos.ReservationTypeLiff, repos.Reservation),
		ReservationStaff:          resStaffSvc,
		ReservationStaffCore:      resStaffSvc,
		ReservationStaffExclusion: resStaffSvc,
		ReservationSchedule:       reservation.NewReservationScheduleService(repos.ReservationSchedule),
		ReservationAdmin:          reservation.NewReservationAdminServiceWithAvailabilityAndType(repos.ReservationAdmin, repos.Reservation, repos.ReservationType, tx, repos.ReservationStaff, repos.ReservationTypeUnavailableTime, repos.ReservationTypeAvailableSlot),
		ChronicCondition:          chronicConditionSvc,
		Liff: reservation.NewLiffServiceWithType(
			lineReservationSettings,
			repos.ReservationTypeLiff,
			repos.ReservationType,
			repos.ReservationStaff,
			repos.ReservationSchedule,
			repos.ReservationAdmin,
			lstepDeps.LineCustomers,
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
