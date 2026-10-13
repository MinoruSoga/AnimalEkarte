package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
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
	// Version は楽観的ロック用 expectedVersion（nil=照合スキップ・後方互換）。
	// UpdateClinicalPlanInput.Version と同型: repository Update の WHERE version=? 述語に
	// そのまま渡り、一致時のみ version+1 で書き戻す。0/負値はどの行にも一致せず Conflict。
	Version *int
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
	auditTx          AuditTxLogger
}

func NewVaccinationService(
	repo VaccinationRepository,
	vaccineRepo VaccineRepository,
	tagSyncSvc vaccinationTagSyncer,
	relationVerifier vaccinationRelationVerifier,
	medicalRecords medicalRecordLocker,
	transactor Transactor,
	auditTx AuditTxLogger,
) VaccinationService {
	return &vaccinationService{
		repo: repo, vaccineRepo: vaccineRepo, tagSyncSvc: tagSyncSvc,
		relationVerifier: relationVerifier, medicalRecords: medicalRecords, transactor: transactor,
		auditTx: auditTx,
	}
}

func (s *vaccinationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, search, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list vaccinations")
	}
	return items, total, nil
}

func (s *vaccinationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
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
		// EMR-323: validateRelations locks the linked parent row but never checks its
		// status. Re-check the already-held row lock in the same tx so a finalized
		// parent rejects the create (Update/Delete と同じ lockDraftMedicalRecord 規約)。
		if vaccination.MedicalRecordID != nil {
			if err := s.lockVaccinationParentDraft(txCtx, clinicID, *vaccination.MedicalRecordID,
				"確定済みカルテのワクチン接種は編集できません"); err != nil {
				return err
			}
		}
		if err := s.repo.Create(txCtx, vaccination); err != nil {
			return apperrors.Wrap(err, "failed to create vaccination")
		}
		var err error
		created, err = s.repo.FindByID(txCtx, clinicID, vaccination.ID)
		if err != nil {
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
			return apperrors.Wrap(err, "failed to find vaccination")
		}
		if snapshot == nil {
			return apperrors.WrapInternalServerError("vaccination not found during update validation")
		}
		medicalRecordID, petID, doctorID, vaccineID := effectiveVaccinationRelations(snapshot, input)
		// EMR-323 / P323-LOCKS: guard BOTH the original source and the effective
		// destination parent. Unique parent IDs are locked FOR UPDATE in deterministic
		// ascending order BEFORE validateRelations may take the destination lock, so
		// opposite-direction relinks cannot deadlock. A finalized source or
		// destination rejects with Conflict (treatment/vital/checkup の
		// lockDraftMedicalRecord 規約と対称化)。
		for _, recordID := range uniqueSortedVaccinationParentIDs(snapshot.MedicalRecordID, medicalRecordID) {
			if err := s.lockVaccinationParentDraft(txCtx, clinicID, recordID,
				"確定済みカルテのワクチン接種は編集できません"); err != nil {
				return err
			}
		}
		if err := s.validateRelations(txCtx, clinicID, medicalRecordID, petID, doctorID, vaccineID); err != nil {
			return err
		}

		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock vaccination")
		}
		// Compare the ORIGINAL snapshot tuple against the locked row: request
		// overrides must not hide a concurrent relink committed between the
		// snapshot read and this row lock (P323-LOCKS).
		if locked == nil || locked.ID != id || locked.ClinicID != clinicID ||
			!sameVaccinationRelations(
				snapshot.MedicalRecordID, snapshot.PetID, snapshot.DoctorID, snapshot.VaccineID,
				locked.MedicalRecordID, locked.PetID, locked.DoctorID, locked.VaccineID,
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
		vaccination, err = s.repo.Update(txCtx, clinicID, id, *input)
		if err != nil {
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
	// EMR-323: every write-side dependency must exist before any read or write —
	// a missing repository, transaction boundary, or audit sink fails closed.
	if s.repo == nil {
		return apperrors.WrapInternalServerError("vaccination repository dependency is required")
	}
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("vaccination transaction dependency is required")
	}
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("vaccination audit dependency is required")
	}
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find vaccination")
	}
	if existing == nil {
		return apperrors.WrapInternalServerError("vaccination not found during delete")
	}
	// The parent locker is required only for linked vaccinations; standalone
	// (unlinked) rows stay deletable without it.
	if existing.MedicalRecordID != nil && s.medicalRecords == nil {
		return apperrors.WrapInternalServerError("vaccination medical record dependency is required")
	}
	var current *model.Vaccination
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// Lock the linked parent before the child row so a concurrent finalize
		// serializes against this delete (X-11); a finalized parent rejects with
		// Conflict and produces no child delete, audit, or tag resync.
		if existing.MedicalRecordID != nil {
			if err := s.lockVaccinationParentDraft(txCtx, clinicID, *existing.MedicalRecordID,
				"確定済みカルテのワクチン接種は削除できません"); err != nil {
				return err
			}
		}
		locked, err := s.repo.LockByIDForUpdate(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock vaccination")
		}
		// Revalidate the ORIGINAL snapshot against the locked row: a concurrent
		// relink committed between the snapshot read and this lock must reject
		// (P323-LOCKS — effective/request fields are never compared here).
		if locked == nil || locked.ID != id || locked.ClinicID != clinicID ||
			!sameVaccinationRelations(
				existing.MedicalRecordID, existing.PetID, existing.DoctorID, existing.VaccineID,
				locked.MedicalRecordID, locked.PetID, locked.DoctorID, locked.VaccineID,
			) {
			return apperrors.WrapConflict("vaccination relations changed concurrently")
		}
		// The bare locked row has no Pet/Owner preloads; reread clinic-scoped
		// inside the same transaction for the current linkage and tag-resync
		// owner data (P323-TAGS), and verify it still agrees with the locked row.
		current, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to reread vaccination")
		}
		if current == nil || current.ID != id || current.ClinicID != clinicID ||
			!sameVaccinationRelations(
				locked.MedicalRecordID, locked.PetID, locked.DoctorID, locked.VaccineID,
				current.MedicalRecordID, current.PetID, current.DoctorID, current.VaccineID,
			) {
			return apperrors.WrapConflict("vaccination relations changed concurrently")
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to delete vaccination")
		}
		// Fail-closed audit inside the same transaction: an audit write failure
		// propagates and rolls the delete back (P323-AUDIT).
		return s.auditVaccinationDeleteTx(txCtx, clinicID, id, locked)
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "vaccination deleted",
		slog.Uint64("vaccination_id", id),
		slog.Uint64("clinic_id", clinicID))
	// Tag resync runs only after a successful commit — never on rejection, audit
	// failure, or commit failure — and uses the current scoped reread, not the
	// stale pretransaction snapshot.
	s.resyncOwnerVaccineTags(ctx, clinicID, current)
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

// lockVaccinationParentDraft locks the parent medical record FOR UPDATE inside the
// ambient transaction and rejects a finalized record with Conflict — the EMR-323
// extension of the X-11 draft-parent invariant (sharedkernel.LockDraftMedicalRecord
// と同じ写像) with an added identity check so nil / foreign-clinic / wrong-ID
// parents also fail closed instead of being trusted by scope alone.
func (s *vaccinationService) lockVaccinationParentDraft(txCtx context.Context, clinicID, recordID uint64, conflictMsg string) error {
	if s.medicalRecords == nil {
		return apperrors.WrapInternalServerError("vaccination medical record dependency is required")
	}
	parent, err := s.medicalRecords.LockByIDForUpdate(txCtx, clinicID, recordID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify medical record ownership")
	}
	if parent == nil || parent.ID != recordID || parent.ClinicID != clinicID {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", recordID))
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict(conflictMsg)
	}
	return nil
}

// uniqueSortedVaccinationParentIDs returns the deduplicated, ascending-ordered set
// of parent medical record IDs touched by a mutation — the deterministic lock
// order required for source+destination relinks (P323-LOCKS).
func uniqueSortedVaccinationParentIDs(ids ...*uint64) []uint64 {
	seen := map[uint64]bool{}
	unique := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == nil || seen[*id] {
			continue
		}
		seen[*id] = true
		unique = append(unique, *id)
	}
	slices.Sort(unique)
	return unique
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
