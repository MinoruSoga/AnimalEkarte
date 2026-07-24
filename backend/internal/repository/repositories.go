package repository

import "gorm.io/gorm"

// Repositories はすべてのリポジトリを保持するDIコンテナ
type Repositories struct {
	db                    *gorm.DB
	Account               AccountRepository
	StaffClinicAssignment StaffClinicAssignmentRepository
	AnimalSpecies         AnimalSpeciesRepository
	Owner                 OwnerRepository
	Pet                   PetRepository
	Reservation           ReservationRepository
	MedicalRecord         MedicalRecordRepository
	MedicalRecordAddendum MedicalRecordAddendumRepository
	Hospitalization       HospitalizationRepository
	Accounting            AccountingRepository
	Staff                 StaffRepository
	Cage                  CageRepository
	Medicine              MedicineRepository
	MedicineDoseParam     MedicineDoseParamRepository
	Vaccine               VaccineRepository
	Insurance             InsuranceRepository
	ReservationType       ReservationTypeRepository
	ReservationTypeGroup  ReservationTypeGroupRepository
	Consultation          ConsultationRepository
	Procedure             ProcedureRepository
	HospitalizationPlan   HospitalizationPlanRepository
	ExaminationType       ExamTypeRepository
	DiagnosisType         DiagnosisTypeRepository
	DiagnosisName         DiagnosisNameRepository
	CheckupType           CheckupTypeRepository
	Clinic                ClinicRepository
	Examination           ExaminationRepository
	Vaccination           VaccinationRepository
	Occupation            OccupationRepository
	ChiefComplaintType    ChiefComplaintTypeRepository
	Inquiry               InquiryRepository
	InquiryTemplate       InquiryTemplateRepository
	Company               CompanyRepository
	PermissionGroup       PermissionGroupRepository
	BillingConfirmation   BillingConfirmationRepository
	CarePlanItem          CarePlanItemRepository
	ShiftEntry            ShiftEntryRepository
	ShiftTemplate         ShiftTemplateRepository
	ClinicHoliday         ClinicHolidayRepository
	TreatmentPlan         TreatmentPlanRepository
	Vital                 VitalRepository
	Treatment             TreatmentRepository
	DailyRecord           DailyRecordRepository
	MedicalRecordImage    MedicalRecordImageRepository
	ClinicalPlan          ClinicalPlanRepository
	Checkup               CheckupRepository
	CheckupTypeField      CheckupTypeFieldRepository
	CheckupFieldResult    CheckupFieldResultRepository
	Estimate              EstimateRepository
	// ManualArticle: BE9-2B — moved to internal/manualarticle (aggregator-free domain
	// package). No longer constructed here; see cmd/api/main.go.
	BillingItem BillingItemRepository
	Refund      RefundRepository
	Audit       AuditRepository
	// LINE予約
	ReservationTypeLiff            ReservationTypeLiffRepository
	ReservationStaff               ReservationStaffRepository
	ReservationSchedule            ReservationScheduleRepository
	ReservationAdmin               ReservationAdminRepository
	ReservationTypeUnavailableTime ReservationTypeUnavailableTimeRepository
	ReservationTypeAvailableSlot   ReservationTypeAvailableSlotRepository
	ReservationTypeOccupation      ReservationTypeOccupationRepository
	PasswordResetToken             PasswordResetTokenRepository
	// FEAT-368: 集計・締め機能
	ClinicSettings       ClinicSettingsRepository
	ClosingSpecialPeriod ClosingSpecialPeriodRepository
	PaymentMethodMaster  PaymentMethodMasterRepository
	Campaign             CampaignRepository
	CashRegisterClose    CashRegisterCloseRepository
	// LSTEP-BE-009: 処方薬記録
	Prescription PrescriptionRepository
	// LSTEP-BE-012: 慢性疾患フラグ
	ChronicCondition PetChronicConditionRepository
	// 認証: refresh_token JTI ブラックリスト
	TokenBlacklist TokenBlacklistRepository
	// lab import: BE9-2D sub-batch③ で internal/medicalrecord へ移動（leaf domain, no facade）。
	// 構築は medicalrecord.NewLabImport*Repository が担う（cmd/api/main.go / NewServices）。
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
		MedicalRecordAddendum:          NewMedicalRecordAddendumRepository(db),
		Hospitalization:                NewHospitalizationRepository(db),
		Accounting:                     NewAccountingRepository(db),
		Staff:                          NewStaffRepository(db),
		Cage:                           NewCageRepository(db),
		Medicine:                       NewMedicineRepository(db),
		MedicineDoseParam:              NewMedicineDoseParamRepository(db),
		Vaccine:                        NewVaccineRepository(db),
		Insurance:                      NewInsuranceRepository(db),
		ReservationType:                NewReservationTypeRepository(db),
		ReservationTypeGroup:           NewReservationTypeGroupRepository(db),
		Consultation:                   NewConsultationRepository(db),
		Procedure:                      NewProcedureRepository(db),
		HospitalizationPlan:            NewHospitalizationPlanRepository(db),
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
		CheckupTypeField:               NewCheckupTypeFieldRepository(db),
		CheckupFieldResult:             NewCheckupFieldResultRepository(db),
		Estimate:                       NewEstimateRepository(db),
		BillingItem:                    NewBillingItemRepository(db),
		Refund:                         NewRefundRepository(db),
		Audit:                          NewAuditRepository(db),
		ReservationTypeLiff:            NewReservationTypeLiffRepository(db),
		ReservationStaff:               NewReservationStaffRepository(db),
		ReservationSchedule:            NewReservationScheduleRepository(db),
		ReservationAdmin:               NewReservationAdminRepository(db),
		ReservationTypeUnavailableTime: NewReservationTypeUnavailableTimeRepository(db),
		ReservationTypeAvailableSlot:   NewReservationTypeAvailableSlotRepository(db),
		ReservationTypeOccupation:      NewReservationTypeOccupationRepository(db),
		PasswordResetToken:             NewPasswordResetTokenRepository(db),
		// FEAT-368: 集計・締め機能
		ClinicSettings:       NewClinicSettingsRepository(db),
		ClosingSpecialPeriod: NewClosingSpecialPeriodRepository(db),
		PaymentMethodMaster:  NewPaymentMethodMasterRepository(db),
		Campaign:             NewCampaignRepository(db),
		CashRegisterClose:    NewCashRegisterCloseRepository(db),
		Prescription:         NewPrescriptionRepository(db),
		ChronicCondition:     NewPetChronicConditionRepository(db),
		TokenBlacklist:       NewTokenBlacklistRepository(db),
		// lab import: BE9-2D sub-batch③ — moved to internal/medicalrecord (leaf domain, no facade).
	}
}

// DB はGORMのDBインスタンスを返す（バリデーター等の直接DB操作に使用）。
func (r *Repositories) DB() *gorm.DB { return r.db }
