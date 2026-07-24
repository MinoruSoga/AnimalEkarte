package repository

// This file contains the smallest test-only compatibility surface required by
// the cross-domain isolation, transaction, race, RLS, and security gates that
// remain in this directory during BE9-2F. It is intentionally unavailable to
// production importers; production code must use the owning domain packages.

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
)

type targetStaffAccountStore struct {
	auth.AccountRepository
	passwordResetTokens auth.PasswordResetTokenRepository
}

func (s *targetStaffAccountStore) DeletePasswordResetTokens(
	ctx context.Context,
	accountID uint64,
) error {
	return s.passwordResetTokens.DeleteByAccountID(ctx, accountID)
}

func NewAccountRepository(db *gorm.DB) *targetStaffAccountStore {
	return &targetStaffAccountStore{
		AccountRepository:   auth.NewAccountRepository(db),
		passwordResetTokens: auth.NewPasswordResetTokenRepository(db),
	}
}

type AccountingRepository = billing.AccountingRepository

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return billing.NewAccountingRepository(db)
}

type ReservationAdminRepository = reservation.ReservationAdminRepository

func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return reservation.NewReservationAdminRepository(db)
}

type AuditRepository = audit.Repository

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return audit.NewRepository(db)
}

func NewCarePlanItemRepository(db *gorm.DB) medicalrecord.CarePlanItemRepository {
	return medicalrecord.NewCarePlanItemRepository(db)
}

func NewCheckupRepository(db *gorm.DB) medicalrecord.CheckupRepository {
	return medicalrecord.NewCheckupRepository(db)
}

type ClinicRepository = clinic.ClinicRepository

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return clinic.NewClinicRepository(db)
}

func NewClinicalPlanRepository(db *gorm.DB) medicalrecord.ClinicalPlanRepository {
	return medicalrecord.NewClinicalPlanRepository(db)
}

type ClosingSpecialPeriodRepository = clinic.ClosingSpecialPeriodRepository

func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return clinic.NewClosingSpecialPeriodRepository(db)
}

type (
	DiagnosisTypeRepository = medicalrecord.DiagnosisTypeRepository
	DiagnosisNameRepository = medicalrecord.DiagnosisNameRepository
)

func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return medicalrecord.NewDiagnosisTypeRepository(db)
}

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return medicalrecord.NewDiagnosisNameRepository(db)
}

func NewEstimateRepository(db *gorm.DB) billing.EstimateRepository {
	return billing.NewEstimateRepository(db)
}

func NewExaminationRepository(db *gorm.DB) medicalrecord.ExaminationRepository {
	return medicalrecord.NewExaminationRepository(db)
}

func NewHospitalizationRepository(db *gorm.DB) medicalrecord.HospitalizationRepository {
	return medicalrecord.NewHospitalizationRepository(db)
}

type InsuranceRepository = billing.InsuranceRepository

func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return billing.NewInsuranceRepository(db)
}

func NewMedicalRecordImageRepository(db *gorm.DB) medicalrecord.MedicalRecordImageRepository {
	return medicalrecord.NewMedicalRecordImageRepository(db)
}

func NewMedicalRecordRepository(db *gorm.DB) medicalrecord.MedicalRecordRepository {
	return medicalrecord.NewMedicalRecordRepository(db)
}

type OccupationRepository = staff.OccupationRepository

func NewOccupationRepository(db *gorm.DB) OccupationRepository {
	return staff.NewOccupationRepository(db)
}

type OwnerRepository = owner.Repository

func NewOwnerRepository(db *gorm.DB) OwnerRepository {
	writer := pet.NewWriter(db)
	return owner.NewRepository(db, pet.NewOwnerRegistrationAdapter(writer))
}

func NewOwnerRepositoryWithPetWriter(
	db *gorm.DB,
	writer pet.OwnerRegistrationWriter,
) OwnerRepository {
	return owner.NewRepository(db, pet.NewOwnerRegistrationAdapter(writer))
}

func NewPaymentMethodMasterRepository(db *gorm.DB) billing.PaymentMethodMasterRepository {
	return billing.NewPaymentMethodMasterRepository(db)
}

func NewPermissionGroupRepository(db *gorm.DB) auth.PermissionGroupRepository {
	return auth.NewPermissionGroupRepository(db)
}

type PetRepository = pet.Repository

func NewPetRepository(db *gorm.DB) PetRepository {
	return pet.NewRepository(db)
}

func NewPetRepositoryWithWriter(db *gorm.DB, writer pet.Creator) PetRepository {
	return pet.NewRepositoryWithWriter(db, writer)
}

func NewReservationRepository(db *gorm.DB) reservation.ReservationStore {
	return reservation.NewReservationRepository(db)
}

type ReservationScheduleRepository = reservation.ReservationScheduleRepository

func NewReservationScheduleRepository(db *gorm.DB) ReservationScheduleRepository {
	return reservation.NewReservationScheduleRepository(db, staff.NewShiftEntryRepository(db))
}

type ReservationStaffRepository = reservation.ReservationStaffRepository

func NewReservationStaffRepository(db *gorm.DB) ReservationStaffRepository {
	return reservation.NewReservationStaffRepository(db, staff.NewStaffRepository(db))
}

type ShiftEntryRepository = staff.ShiftEntryRepository

func NewShiftEntryRepository(db *gorm.DB) ShiftEntryRepository {
	return staff.NewShiftEntryRepository(db)
}

type ShiftTemplateRepository = staff.ShiftTemplateRepository

func NewShiftTemplateRepository(db *gorm.DB) ShiftTemplateRepository {
	return staff.NewShiftTemplateRepository(db)
}

func NewStaffClinicAssignmentRepository(db *gorm.DB) staff.StaffClinicAssignmentRepository {
	return staff.NewStaffClinicAssignmentRepository(db)
}

func NewStaffRepository(db *gorm.DB) staff.StaffRepository {
	return staff.NewStaffRepository(db)
}

func NewTransactor(db *gorm.DB) persistence.Transactor {
	return persistence.NewTransactor(db)
}

func txFromContext(ctx context.Context) *gorm.DB {
	return persistence.TxFromContext(ctx)
}
