package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateVaccinationInput はワクチン接種作成の入力DTO
type CreateVaccinationInput struct {
	MedicalRecordID  *uint64
	PetID            *uint64
	VaccineID        uint64
	Date             time.Time
	DoctorID         *uint64
	NextDate         *time.Time
	NextScheduleType *model.NextScheduleType
	Supplemental     string
	Lot1             string
	Lot2             string
	Lot3             string
	Lot4             string
	Remarks          string
}

// UpdateVaccinationInput はワクチン接種更新のサービス入力 DTO
type UpdateVaccinationInput struct {
	MedicalRecordID  *uint64
	PetID            *uint64
	VaccineID        *uint64
	Date             *time.Time
	DoctorID         *uint64
	NextDate         *time.Time
	NextScheduleType *model.NextScheduleType
	Supplemental     *string
	Lot1             *string
	Lot2             *string
	Lot3             *string
	Lot4             *string
	Remarks          *string
}

func buildVaccinationUpdate(input *UpdateVaccinationInput) map[string]any {
	fields := make(map[string]any)
	if input.MedicalRecordID != nil {
		fields["medical_record_id"] = *input.MedicalRecordID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.VaccineID != nil {
		fields["vaccine_id"] = *input.VaccineID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.NextDate != nil {
		fields["next_date"] = *input.NextDate
	}
	if input.NextScheduleType != nil {
		fields["next_schedule_type"] = *input.NextScheduleType
	}
	if input.Supplemental != nil {
		fields["supplemental"] = *input.Supplemental
	}
	if input.Lot1 != nil {
		fields["lot1"] = *input.Lot1
	}
	if input.Lot2 != nil {
		fields["lot2"] = *input.Lot2
	}
	if input.Lot3 != nil {
		fields["lot3"] = *input.Lot3
	}
	if input.Lot4 != nil {
		fields["lot4"] = *input.Lot4
	}
	if input.Remarks != nil {
		fields["remarks"] = *input.Remarks
	}
	return fields
}

type VaccinationService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	Create(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type vaccinationService struct {
	repo             VaccinationRepository
	vaccineRepo      VaccineRepository
	tagSyncSvc       vaccinationTagSyncer
	relationVerifier vaccinationRelationVerifier
	medicalRecords   medicalRecordLocker
	transactor       Transactor
}

func NewVaccinationService(
	repo VaccinationRepository,
	vaccineRepo VaccineRepository,
	tagSyncSvc vaccinationTagSyncer,
	relationVerifier vaccinationRelationVerifier,
	medicalRecords medicalRecordLocker,
	transactor Transactor,
) VaccinationService {
	return &vaccinationService{
		repo: repo, vaccineRepo: vaccineRepo, tagSyncSvc: tagSyncSvc,
		relationVerifier: relationVerifier, medicalRecords: medicalRecords, transactor: transactor,
	}
}

func (s *vaccinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, search, page, limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list vaccinations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list vaccinations")
	}
	return items, total, nil
}

func (s *vaccinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get vaccination", "error", err)
		return nil, apperrors.Wrap(err, "failed to get vaccination")
	}
	return result, nil
}

func (s *vaccinationService) Create(ctx context.Context, clinicID uint64, input *CreateVaccinationInput) (*model.Vaccination, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if input.VaccineID == 0 {
		return nil, apperrors.WrapInvalidInput("vaccine_id is required")
	}
	if err := errIfVaccinationDateAfterToday(input.Date); err != nil {
		return nil, err
	}
	if err := errIfNextDateNotAfterVaccinationDate(input.Date, input.NextDate); err != nil {
		return nil, err
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("vaccination transaction dependency is required")
	}
	vaccination := &model.Vaccination{
		ClinicID:         clinicID,
		MedicalRecordID:  input.MedicalRecordID,
		PetID:            input.PetID,
		VaccineID:        input.VaccineID,
		Date:             input.Date,
		DoctorID:         input.DoctorID,
		NextDate:         input.NextDate,
		NextScheduleType: input.NextScheduleType,
		Supplemental:     input.Supplemental,
		Lot1:             input.Lot1,
		Lot2:             input.Lot2,
		Lot3:             input.Lot3,
		Lot4:             input.Lot4,
		Remarks:          input.Remarks,
	}
	var created *model.Vaccination
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.validateRelations(txCtx, clinicID, vaccination.MedicalRecordID, vaccination.PetID, vaccination.DoctorID, vaccination.VaccineID); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, vaccination); err != nil {
			slog.ErrorContext(txCtx, "failed to create vaccination", "error", err)
			return apperrors.Wrap(err, "failed to create vaccination")
		}
		var err error
		created, err = s.repo.FindByID(txCtx, clinicID, vaccination.ID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get vaccination after create", "error", err, "vaccination_id", vaccination.ID)
			return apperrors.Wrap(err, "failed to read vaccination after create")
		}
		if created == nil {
			return apperrors.WrapInternalServerError("vaccination not found after create")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vaccination created",
		slog.Uint64("vaccination_id", vaccination.ID),
		slog.Uint64("clinic_id", vaccination.ClinicID))
	s.syncVaccineTag(ctx, clinicID, created)
	return created, nil
}

func (s *vaccinationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccinationInput) (*model.Vaccination, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildVaccinationUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if input.Date != nil {
		if err := errIfVaccinationDateAfterToday(*input.Date); err != nil {
			return nil, err
		}
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("vaccination transaction dependency is required")
	}
	var vaccination *model.Vaccination
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Resolve and lock parent/master relations before locking the child row. This keeps
		// vaccination updates in the same parent-before-child order as create and master writes.
		snapshot, err := s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to find vaccination", "error", err)
			return apperrors.Wrap(err, "failed to find vaccination")
		}
		if snapshot == nil {
			return apperrors.WrapInternalServerError("vaccination not found during update validation")
		}
		medicalRecordID, petID, doctorID, vaccineID := effectiveVaccinationRelations(snapshot, input)
		if err := s.validateRelations(txCtx, clinicID, medicalRecordID, petID, doctorID, vaccineID); err != nil {
			return err
		}

		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to lock vaccination", "error", err)
			return apperrors.Wrap(err, "failed to lock vaccination")
		}
		lockedMedicalRecordID, lockedPetID, lockedDoctorID, lockedVaccineID := effectiveVaccinationRelations(locked, input)
		if !sameVaccinationRelations(
			medicalRecordID, petID, doctorID, vaccineID,
			lockedMedicalRecordID, lockedPetID, lockedDoctorID, lockedVaccineID,
		) {
			return apperrors.WrapConflict("vaccination relations changed concurrently")
		}
		effectiveDate := locked.Date
		if input.Date != nil {
			effectiveDate = *input.Date
		}
		effectiveNextDate := locked.NextDate
		if input.NextDate != nil {
			effectiveNextDate = input.NextDate
		}
		if err := errIfNextDateNotAfterVaccinationDate(effectiveDate, effectiveNextDate); err != nil {
			return err
		}
		vaccination, err = s.repo.Update(txCtx, clinicID, id, fields)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to update vaccination", "error", err)
			return apperrors.Wrap(err, "failed to update vaccination")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vaccination updated",
		slog.Uint64("vaccination_id", id),
		slog.Uint64("clinic_id", clinicID))
	s.resyncOwnerVaccineTags(ctx, clinicID, vaccination)
	return vaccination, nil
}
func (s *vaccinationService) Delete(ctx context.Context, clinicID, id uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find vaccination")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete vaccination", "error", err, "clinic_id", clinicID, "vaccination_id", id)
		return apperrors.Wrap(err, "failed to delete vaccination")
	}
	slog.InfoContext(ctx, "vaccination deleted",
		slog.Uint64("vaccination_id", id),
		slog.Uint64("clinic_id", clinicID))
	s.resyncOwnerVaccineTags(ctx, clinicID, existing)
	return nil
}

func (s *vaccinationService) syncVaccineTag(ctx context.Context, clinicID uint64, vaccination *model.Vaccination) {
	ownerID, ok := safeVaccinationOwnerID(clinicID, vaccination)
	if s.tagSyncSvc == nil || !ok {
		return
	}
	if err := s.tagSyncSvc.SyncVaccineTag(ctx, clinicID, ownerID, vaccination.ID); err != nil {
		slog.ErrorContext(ctx, "failed to sync vaccine tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "vaccination_id", vaccination.ID)
	}
}

func (s *vaccinationService) resyncOwnerVaccineTags(ctx context.Context, clinicID uint64, vaccination *model.Vaccination) {
	ownerID, ok := safeVaccinationOwnerID(clinicID, vaccination)
	if s.tagSyncSvc == nil || !ok {
		return
	}
	if err := s.tagSyncSvc.ResyncOwnerVaccineTags(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to resync owner vaccine tags", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "vaccination_id", vaccination.ID)
	}
}

func errIfVaccinationDateAfterToday(date time.Time) error {
	now := time.Now()
	dateJST := date.In(config.JST)
	nowJST := now.In(config.JST)
	dateDay := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, config.JST)
	today := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, config.JST)
	if dateDay.After(today) {
		return apperrors.WrapInvalidInput("接種日は今日以前の日付を入力してください")
	}
	return nil
}

func errIfNextDateNotAfterVaccinationDate(vaccinationDate time.Time, nextDate *time.Time) error {
	if nextDate == nil {
		return nil
	}
	dateJST := vaccinationDate.In(config.JST)
	nextJST := nextDate.In(config.JST)
	dateDay := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, config.JST)
	nextDay := time.Date(nextJST.Year(), nextJST.Month(), nextJST.Day(), 0, 0, 0, 0, config.JST)
	if !nextDay.After(dateDay) {
		return apperrors.WrapInvalidInput("次回予定日は接種日より後の日付を入力してください")
	}
	return nil
}

func effectiveVaccinationRelations(existing *model.Vaccination, input *UpdateVaccinationInput) (medicalRecordID, petID, doctorID *uint64, vaccineID uint64) {
	medicalRecordID, petID, doctorID, vaccineID = existing.MedicalRecordID, existing.PetID, existing.DoctorID, existing.VaccineID
	if input.MedicalRecordID != nil {
		medicalRecordID = input.MedicalRecordID
	}
	if input.PetID != nil {
		petID = input.PetID
	}
	if input.DoctorID != nil {
		doctorID = input.DoctorID
	}
	if input.VaccineID != nil {
		vaccineID = *input.VaccineID
	}
	return medicalRecordID, petID, doctorID, vaccineID
}

func sameVaccinationRelations(
	medicalRecordIDA, petIDA, doctorIDA *uint64,
	vaccineIDA uint64,
	medicalRecordIDB, petIDB, doctorIDB *uint64,
	vaccineIDB uint64,
) bool {
	return vaccineIDA == vaccineIDB &&
		sameOptionalUint64(medicalRecordIDA, medicalRecordIDB) &&
		sameOptionalUint64(petIDA, petIDB) &&
		sameOptionalUint64(doctorIDA, doctorIDB)
}

func sameOptionalUint64(a, b *uint64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (s *vaccinationService) validateRelations(ctx context.Context, clinicID uint64, medicalRecordID, petID, doctorID *uint64, vaccineID uint64) error {
	if s.vaccineRepo == nil || s.relationVerifier == nil || s.medicalRecords == nil {
		return apperrors.WrapInternalServerError("vaccination relation validation dependencies are required")
	}
	if _, err := s.vaccineRepo.FindByID(ctx, clinicID, vaccineID); err != nil {
		return apperrors.Wrap(err, "failed to verify vaccine ownership")
	}

	if petID != nil {
		// Clinic-scoped pet existence stays required; owner equality after pet ID match is unreachable.
		if _, err := s.relationVerifier.FindPetOwnerInClinic(ctx, clinicID, *petID); err != nil {
			return apperrors.Wrap(err, "failed to verify vaccination pet ownership")
		}
	}

	if medicalRecordID != nil {
		record, err := s.medicalRecords.LockByIDForUpdate(ctx, clinicID, *medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to verify vaccination medical record ownership")
		}
		if err := s.assertVaccinationRecordOwner(ctx, clinicID, record); err != nil {
			return err
		}
		if err := s.assertVaccinationRecordPet(ctx, clinicID, record); err != nil {
			return err
		}
		if petID != nil && (record.PetID == nil || *record.PetID != *petID) {
			return apperrors.WrapNotFound("medical_record", "relation")
		}
	}

	if doctorID != nil {
		if err := s.relationVerifier.AssertMedicalRecordDoctorInClinic(ctx, clinicID, *doctorID); err != nil {
			return apperrors.Wrap(err, "failed to verify vaccination doctor ownership")
		}
	}
	return nil
}

func (s *vaccinationService) assertVaccinationRecordOwner(ctx context.Context, clinicID uint64, record *model.MedicalRecord) error {
	if record.OwnerID == nil {
		return nil
	}
	if err := s.relationVerifier.AssertOwnerInClinic(ctx, clinicID, *record.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify medical record owner ownership")
	}
	return nil
}

func (s *vaccinationService) assertVaccinationRecordPet(ctx context.Context, clinicID uint64, record *model.MedicalRecord) error {
	if record.PetID == nil {
		return nil
	}
	if _, err := s.relationVerifier.FindPetOwnerInClinic(ctx, clinicID, *record.PetID); err != nil {
		return apperrors.Wrap(err, "failed to verify medical record pet ownership")
	}
	return nil
}

func safeVaccinationOwnerID(clinicID uint64, vaccination *model.Vaccination) (uint64, bool) {
	if vaccination == nil || vaccination.PetID == nil || vaccination.Pet == nil || vaccination.Pet.Owner == nil {
		return 0, false
	}
	pet := vaccination.Pet
	owner := pet.Owner
	if pet.ID != *vaccination.PetID || pet.ClinicID != clinicID || owner.ID != pet.OwnerID || owner.ClinicID != clinicID {
		return 0, false
	}
	return owner.ID, true
}
