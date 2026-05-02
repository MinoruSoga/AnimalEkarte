package repository

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"

	"gorm.io/gorm"
)

// Repositories はすべてのリポジトリを保持するDIコンテナ
type Repositories struct {
	db *gorm.DB
	// TransactionFn はテスト時に差し替え可能なトランザクション関数。
	// nil の場合は db.WithContext(ctx).Transaction() を使用する（本番動作）。
	TransactionFn             func(ctx context.Context, fn func(*Repositories) error) error
	Account                   AccountRepository
	StaffClinicAssignment     StaffClinicAssignmentRepository
	AnimalSpecies             AnimalSpeciesRepository
	Owner                     OwnerRepository
	Pet                       PetRepository
	Reservation               ReservationRepository
	MedicalRecord             MedicalRecordRepository
	Hospitalization           HospitalizationRepository
	Accounting                AccountingRepository
	AppointmentTrimmingDetail AppointmentTrimmingDetailRepository
	Inventory                 InventoryRepository
	Staff                     StaffRepository
	Cage                      CageRepository
	Medicine                  MedicineRepository
	Vaccine                   VaccineRepository
	Insurance                 InsuranceRepository
	ReservationType           ReservationTypeRepository
	ReservationTypeGroup      ReservationTypeGroupRepository
	Consultation              ConsultationRepository
	Procedure                 ProcedureRepository
	HospitalizationPlan       HospitalizationPlanRepository
	TrimmingCourse            TrimmingCourseRepository
	TrimmingOption            TrimmingOptionRepository
	ExaminationType           ExamTypeRepository
	DiagnosisType             DiagnosisTypeRepository
	DiagnosisName             DiagnosisNameRepository
	CheckupType               CheckupTypeRepository
	Clinic                    ClinicRepository
	Examination               ExaminationRepository
	Vaccination               VaccinationRepository
	Occupation                OccupationRepository
	ChiefComplaintType        ChiefComplaintTypeRepository
	Inquiry                   InquiryRepository
	InquiryTemplate           InquiryTemplateRepository
	Company                   CompanyRepository
	PermissionGroup           PermissionGroupRepository
	BillingConfirmation       BillingConfirmationRepository
	CarePlanItem              CarePlanItemRepository
	ShiftEntry                ShiftEntryRepository
	ShiftTemplate             ShiftTemplateRepository
	ClinicHoliday             ClinicHolidayRepository
	TreatmentPlan             TreatmentPlanRepository
	Vital                     VitalRepository
	Treatment                 TreatmentRepository
	DailyRecord               DailyRecordRepository
	MedicalRecordImage        MedicalRecordImageRepository
	ClinicalPlan              ClinicalPlanRepository
	Checkup                   CheckupRepository
	Estimate                  EstimateRepository
	MerchandiseItem           MerchandiseItemRepository
	BillingItem               BillingItemRepository
	Refund                    RefundRepository
	Audit                     AuditRepository
	// LINE予約
	LineReservationSetting         LineReservationSettingRepository
	ReservationTypeLiff            ReservationTypeLiffRepository
	ReservationStaff               ReservationStaffRepository
	ReservationSchedule            ReservationScheduleRepository
	ReservationAdmin               ReservationAdminRepository
	LineCustomerMgr                LineCustomerRepository
	ReservationTypeUnavailableTime ReservationTypeUnavailableTimeRepository
	ReservationTypeOccupation      ReservationTypeOccupationRepository
	PasswordResetToken             PasswordResetTokenRepository
	// FEAT-368: 集計・締め機能
	ClinicSettings       ClinicSettingsRepository
	ClosingSpecialPeriod ClosingSpecialPeriodRepository
	PaymentMethodMaster  PaymentMethodMasterRepository
	CashRegisterClose    CashRegisterCloseRepository
	// LSTEP / LINE連携
	LstepSettings     LstepSettingsRepository
	LstepSyncSettings LstepSyncSettingsRepository
	SharedFile        SharedFileRepository
	LstepTagCache     LstepTagCacheRepository
	// LSTEP-BE-009: 処方薬記録
	Prescription PrescriptionRepository
	// LSTEP-BE-010: LTV集計
	Ltv LtvRepository
	// LSTEP-BE-012: 慢性疾患フラグ
	ChronicCondition PetChronicConditionRepository
	// LSTEP-BE-013: LINE個別送信ログ
	LineSendLog LineSendLogRepository
	// LSTEP-BE-021: LINE User ID 紐付けトークン
	LineLinkToken LineLinkTokenRepository
	// LSTEP-BE-004: 健診対象者抽出・一括タグ連携
	CheckupSync CheckupSyncRepository
	// FEAT-375: Lステップ連携エラーカウンター
	LstepSyncErrorCounter LstepSyncErrorCounterRepository
	// FEAT-379: per-clinic コード→タグ マッピング
	LstepTagCodeMapping LstepTagCodeMappingRepository
	// FEAT-383: 自動配信トリガーログ
	LstepDeliveryTriggerLog LstepDeliveryTriggerLogRepository
}

// NewRepositories はすべてのリポジトリを初期化して返す
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:                             db,
		Account:                        NewAccountRepository(db),
		StaffClinicAssignment:          NewStaffClinicAssignmentRepository(db),
		AnimalSpecies:                  NewAnimalSpeciesRepository(db),
		Owner:                          NewOwnerRepository(db),
		Pet:                            NewPetRepository(db),
		Reservation:                    NewReservationRepository(db),
		MedicalRecord:                  NewMedicalRecordRepository(db),
		Hospitalization:                NewHospitalizationRepository(db),
		Accounting:                     NewAccountingRepository(db),
		AppointmentTrimmingDetail:      NewAppointmentTrimmingDetailRepository(db),
		Inventory:                      NewInventoryRepository(db),
		Staff:                          NewStaffRepository(db),
		Cage:                           NewCageRepository(db),
		Medicine:                       NewMedicineRepository(db),
		Vaccine:                        NewVaccineRepository(db),
		Insurance:                      NewInsuranceRepository(db),
		ReservationType:                NewReservationTypeRepository(db),
		ReservationTypeGroup:           NewReservationTypeGroupRepository(db),
		Consultation:                   NewConsultationRepository(db),
		Procedure:                      NewProcedureRepository(db),
		HospitalizationPlan:            NewHospitalizationPlanRepository(db),
		TrimmingCourse:                 NewTrimmingCourseRepository(db),
		TrimmingOption:                 NewTrimmingOptionRepository(db),
		ExaminationType:                NewExamTypeRepository(db),
		DiagnosisType:                  NewDiagnosisTypeRepository(db),
		DiagnosisName:                  NewDiagnosisNameRepository(db),
		CheckupType:                    NewCheckupTypeRepository(db),
		Clinic:                         NewClinicRepository(db),
		Examination:                    NewExaminationRepository(db),
		Vaccination:                    NewVaccinationRepository(db),
		Occupation:                     NewOccupationRepository(db),
		ChiefComplaintType:             NewChiefComplaintTypeRepository(db),
		Inquiry:                        NewInquiryRepository(db),
		InquiryTemplate:                NewInquiryTemplateRepository(db),
		Company:                        NewCompanyRepository(db),
		PermissionGroup:                NewPermissionGroupRepository(db),
		BillingConfirmation:            NewBillingConfirmationRepository(db),
		CarePlanItem:                   NewCarePlanItemRepository(db),
		ShiftEntry:                     NewShiftEntryRepository(db),
		ShiftTemplate:                  NewShiftTemplateRepository(db),
		ClinicHoliday:                  NewClinicHolidayRepository(db),
		TreatmentPlan:                  NewTreatmentPlanRepository(db),
		Vital:                          NewVitalRepository(db),
		Treatment:                      NewTreatmentRepository(db),
		DailyRecord:                    NewDailyRecordRepository(db),
		MedicalRecordImage:             NewMedicalRecordImageRepository(db),
		ClinicalPlan:                   NewClinicalPlanRepository(db),
		Checkup:                        NewCheckupRepository(db),
		Estimate:                       NewEstimateRepository(db),
		MerchandiseItem:                NewMerchandiseItemRepository(db),
		BillingItem:                    NewBillingItemRepository(db),
		Refund:                         NewRefundRepository(db),
		Audit:                          NewAuditRepository(db),
		LineReservationSetting:         NewLineReservationSettingRepository(db),
		ReservationTypeLiff:            NewReservationTypeLiffRepository(db),
		ReservationStaff:               NewReservationStaffRepository(db),
		ReservationSchedule:            NewReservationScheduleRepository(db),
		ReservationAdmin:               NewReservationAdminRepository(db),
		LineCustomerMgr:                NewLineCustomerRepository(db),
		ReservationTypeUnavailableTime: NewReservationTypeUnavailableTimeRepository(db),
		ReservationTypeOccupation:      NewReservationTypeOccupationRepository(db),
		PasswordResetToken:             NewPasswordResetTokenRepository(db),
		// FEAT-368: 集計・締め機能
		ClinicSettings:       NewClinicSettingsRepository(db),
		ClosingSpecialPeriod: NewClosingSpecialPeriodRepository(db),
		PaymentMethodMaster:  NewPaymentMethodMasterRepository(db),
		CashRegisterClose:    NewCashRegisterCloseRepository(db),
		// LSTEP / LINE連携
		LstepSettings:           NewLstepSettingsRepository(db),
		LstepSyncSettings:       NewLstepSyncSettingsRepository(db),
		SharedFile:              NewSharedFileRepository(db),
		LstepTagCache:           NewLstepTagCacheRepository(db),
		Prescription:            NewPrescriptionRepository(db),
		Ltv:                     NewLtvRepository(db),
		ChronicCondition:        NewPetChronicConditionRepository(db),
		LineSendLog:             NewLineSendLogRepository(db),
		LineLinkToken:           NewLineLinkTokenRepository(db),
		CheckupSync:             NewCheckupSyncRepository(db),
		LstepSyncErrorCounter:   NewLstepSyncErrorCounterRepository(db),
		LstepTagCodeMapping:     NewLstepTagCodeMappingRepository(db),
		LstepDeliveryTriggerLog: NewLstepDeliveryTriggerLogRepository(db),
	}
}

// DB はGORMのDBインスタンスを返す（バリデーター等の直接DB操作に使用）。
func (r *Repositories) DB() *gorm.DB { return r.db }

// Transaction はリポジトリ層のトランザクションを実行する。
// テスト時は TransactionFn に mock を設定することで DB 依存を排除できる。
func (r *Repositories) Transaction(ctx context.Context, fn func(repos *Repositories) error) error {
	if r.TransactionFn != nil {
		return r.TransactionFn(ctx, fn)
	}
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := NewRepositories(tx)
		return fn(txRepos)
	}); err != nil {
		return apperrors.Wrap(err, "transaction failed")
	}
	return nil
}
