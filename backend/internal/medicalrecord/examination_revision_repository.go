package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	initialExaminationRevisionVersion = uint64(1)
	examinationRevisionSchemaVersion  = int16(1)
	examinationInitialConfirmReason   = "initial_confirmation"
	examinationWorkingUpdateReason    = "working_update"
	examinationWorkingItemsReason     = "working_items_replace"
	examinationReconfirmReason        = "reconfirmation"
)

// ExaminationRevisionRepository is the optional, narrow Slice-A capability implemented
// by the concrete examination repository. Keeping it separate avoids widening legacy
// handler/test doubles while confirmation fails closed when the capability is absent.
type ExaminationRevisionRepository interface {
	AppendOfficialRevision(
		ctx context.Context,
		clinicID, examinationID, actorID uint64,
		changeReason string,
	) (uint64, error)
	ConfirmWithRevisionCAS(
		ctx context.Context,
		clinicID, examinationID uint64,
		expectedStatus model.ExaminationStatus,
		version uint64,
	) (*model.Examination, error)
	FindOfficialByID(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
	// FindPrintSnapshot loads a versioned revision+items print DTO from revision tables only.
	// A nil version resolves to the parent's current_revision_version (fail-closed if unset).
	FindPrintSnapshot(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error)
}

// ExaminationRevisionWorkflowRepository is the Slice-B capability for reopening,
// editing, and reconfirming an already revisioned examination. Every method requires
// the caller's ambient transaction and an exact current-version comparison.
type ExaminationRevisionWorkflowRepository interface {
	ExaminationRevisionRepository
	AppendWorkingRevisionFromOfficial(
		ctx context.Context,
		clinicID, examinationID, officialVersion, actorID uint64,
		changeReason string,
	) (uint64, error)
	AppendWorkingRevisionFromCurrent(
		ctx context.Context,
		clinicID, examinationID, currentVersion, actorID uint64,
		changeReason string,
	) (uint64, error)
	AppendOfficialRevisionFromWorking(
		ctx context.Context,
		clinicID, examinationID, workingVersion, actorID uint64,
		changeReason string,
	) (uint64, error)
	AdvanceRevisionCAS(
		ctx context.Context,
		clinicID, examinationID uint64,
		expectedStatus model.ExaminationStatus,
		expectedVersion uint64,
		nextStatus model.ExaminationStatus,
		nextVersion uint64,
	) (*model.Examination, error)
}

type examinationRevisionSnapshot struct {
	medicalRecordOwnerID *uint64
	petOwnerID           *uint64
	animalSpeciesID      *uint64
	display              model.ExaminationDisplaySnapshot
}

// AppendOfficialRevision validates every mutable identity/master relation in the caller's
// ambient transaction, then appends official v1 and immutable item snapshots. Slice A only
// accepts a nil parent pointer; reconfirming a working revision belongs to Slice B.
func (r *examinationRepository) AppendOfficialRevision(
	ctx context.Context,
	clinicID, examinationID, actorID uint64,
	changeReason string,
) (uint64, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return 0, apperrors.WrapInternalServerError("examination revision append requires an ambient transaction")
	}
	if actorID == 0 {
		return 0, apperrors.WrapInvalidInput("authenticated staff actor is required for examination confirmation")
	}
	if changeReason == "" {
		return 0, apperrors.WrapInvalidInput("examination revision change reason is required")
	}

	var exam model.Examination
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examinationID, clinicID).
		First(&exam).Error; err != nil {
		return 0, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examinationID))
	}
	if exam.CurrentRevisionVersion != nil {
		return 0, apperrors.WrapConflict("revisioned examination requires the reconfirm workflow")
	}
	if !isFirstConfirmSourceStatus(exam.Status) {
		return 0, apperrors.WrapConflict("examination is not eligible for first confirmation")
	}

	var revisionCount int64
	if err := tx.WithContext(ctx).
		Model(&model.ExaminationRevision{}).
		Where("clinic_id = ? AND examination_id = ?", clinicID, examinationID).
		Count(&revisionCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d", examinationID))
	}
	if revisionCount != 0 {
		return 0, apperrors.WrapConflict("examination already has revision history")
	}

	snapshot, items, err := r.loadOfficialRevisionSnapshot(ctx, tx, clinicID, actorID, &exam)
	if err != nil {
		return 0, err
	}
	displayJSON, err := json.Marshal(snapshot.display)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to encode examination display snapshot")
	}

	persistedChangeReason := changeReason
	revision := &model.ExaminationRevision{
		ClinicID:             clinicID,
		ExaminationID:        examinationID,
		Version:              initialExaminationRevisionVersion,
		Kind:                 model.ExaminationRevisionKindOfficial,
		Status:               model.ExaminationStatusConfirmed,
		MedicalRecordID:      cloneOptionalUint64(exam.MedicalRecordID),
		PetID:                cloneOptionalUint64(exam.PetID),
		MedicalRecordOwnerID: cloneOptionalUint64(snapshot.medicalRecordOwnerID),
		PetOwnerID:           cloneOptionalUint64(snapshot.petOwnerID),
		AnimalSpeciesID:      cloneOptionalUint64(snapshot.animalSpeciesID),
		ExamTypeID:           exam.ExamTypeID,
		DoctorID:             cloneOptionalUint64(exam.DoctorID),
		JobID:                cloneOptionalUUID(exam.JobID),
		ActorID:              actorID,
		Date:                 exam.Date,
		ResultSummary:        exam.ResultSummary,
		Machine:              exam.Machine,
		DisplaySnapshot:      displayJSON,
		SchemaVersion:        examinationRevisionSchemaVersion,
		ChangeReason:         &persistedChangeReason,
	}
	if err := tx.WithContext(ctx).Create(revision).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d", examinationID))
	}
	if err := persistOfficialRevisionItems(ctx, tx, clinicID, examinationID, items); err != nil {
		return 0, err
	}
	return initialExaminationRevisionVersion, nil
}

func persistOfficialRevisionItems(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, examinationID uint64,
	items []model.ExamResult,
) error {
	revisionItems := make([]model.ExaminationRevisionItem, 0, len(items))
	for i := range items {
		item := items[i]
		assessment := assessExamResult(
			item.InspectionValue,
			item.RefMin,
			item.RefMax,
			item.QualitativeMin,
			item.QualitativeMax,
		)
		revisionItems = append(revisionItems, model.ExaminationRevisionItem{
			ClinicID:        clinicID,
			ExaminationID:   examinationID,
			Version:         initialExaminationRevisionVersion,
			ExamTypeFieldID: cloneOptionalUint64(item.ExamTypeItemID),
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Result:          item.Result,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          cloneOptionalFloat64(item.RefMin),
			RefMax:          cloneOptionalFloat64(item.RefMax),
			QualitativeMin:  cloneOptionalString(item.QualitativeMin),
			QualitativeMax:  cloneOptionalString(item.QualitativeMax),
			IsAssessed:      assessment.isAssessed,
			IsAbnormal:      assessment.isAbnormal,
			Status:          assessment.status,
			SortOrder:       item.SortOrder,
		})
	}
	if len(revisionItems) > 0 {
		if err := tx.WithContext(ctx).Create(&revisionItems).Error; err != nil {
			return apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d", examinationID))
		}
	}
	return nil
}

func isFirstConfirmSourceStatus(status model.ExaminationStatus) bool {
	switch status {
	case model.ExaminationStatusPending,
		model.ExaminationStatusInProgress,
		model.ExaminationStatusResultEntered,
		model.ExaminationStatusCompleted:
		return true
	default:
		return false
	}
}

func (r *examinationRepository) loadOfficialRevisionSnapshot(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, actorID uint64,
	exam *model.Examination,
) (*examinationRevisionSnapshot, []model.ExamResult, error) {
	if _, err := lockRevisionStaff(ctx, tx, clinicID, actorID, false); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to verify examination revision actor")
	}

	var examType model.ExaminationType
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", exam.ExamTypeID, clinicID).
		First(&examType).Error; err != nil {
		return nil, nil, apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", exam.ExamTypeID))
	}

	snapshot := &examinationRevisionSnapshot{
		display: model.ExaminationDisplaySnapshot{ExamTypeName: examType.Name},
	}
	var record *model.MedicalRecord
	if exam.MedicalRecordID != nil {
		var locked model.MedicalRecord
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *exam.MedicalRecordID, clinicID).
			First(&locked).Error; err != nil {
			return nil, nil, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", *exam.MedicalRecordID))
		}
		record = &locked
		snapshot.display.MedicalRecordNo = locked.RecordNo
		if err := copyRevisionRecordOwner(ctx, tx, clinicID, &locked, snapshot); err != nil {
			return nil, nil, err
		}
	}

	var pet *model.Pet
	if exam.PetID != nil {
		locked, owner, species, err := lockRevisionPetGraph(ctx, tx, clinicID, *exam.PetID)
		if err != nil {
			return nil, nil, err
		}
		pet = locked
		snapshot.petOwnerID = cloneUint64(owner.ID)
		snapshot.animalSpeciesID = cloneUint64(species.ID)
		snapshot.display.PetName = locked.Name
		snapshot.display.PetOwnerName = owner.Name
		snapshot.display.SpeciesName = species.Name
	}
	if record != nil && record.PetID != nil {
		if pet == nil || pet.ID != *record.PetID {
			return nil, nil, apperrors.WrapNotFound("medical_record", "pet relation")
		}
	}

	if exam.DoctorID != nil {
		doctor, err := lockRevisionStaff(ctx, tx, clinicID, *exam.DoctorID, true)
		if err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to verify examination doctor")
		}
		snapshot.display.DoctorName = doctor.Name
	}
	if exam.JobID != nil {
		var job model.LabImportJob
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ? AND clinic_id = ?", *exam.JobID, clinicID).
			First(&job).Error; err != nil {
			return nil, nil, apperrors.FromGORM(err, "lab_import_job", exam.JobID.String())
		}
	}

	items := make([]model.ExamResult, 0)
	if err := tx.WithContext(ctx).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exams.id = exam_results.exam_id AND exams.clinic_id = ? AND exams.deleted_at IS NULL", clinicID).
		Where("exam_results.exam_id = ?", exam.ID).
		Order("exam_results.sort_order ASC, exam_results.id ASC").
		Clauses(clause.Locking{Strength: "SHARE"}).
		Find(&items).Error; err != nil {
		return nil, nil, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("examination_id=%d", exam.ID))
	}
	if err := lockRevisionItemFields(ctx, tx, clinicID, exam.ExamTypeID, items); err != nil {
		return nil, nil, err
	}
	return snapshot, items, nil
}

func lockRevisionOwner(ctx context.Context, tx *gorm.DB, clinicID, ownerID uint64) (*model.Owner, error) {
	var owner model.Owner
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", ownerID, clinicID).
		First(&owner).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", ownerID))
	}
	return &owner, nil
}

func copyRevisionRecordOwner(
	ctx context.Context,
	tx *gorm.DB,
	clinicID uint64,
	locked *model.MedicalRecord,
	snapshot *examinationRevisionSnapshot,
) error {
	if locked.OwnerID == nil {
		return nil
	}
	owner, err := lockRevisionOwner(ctx, tx, clinicID, *locked.OwnerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify medical record owner")
	}
	snapshot.medicalRecordOwnerID = cloneOptionalUint64(locked.OwnerID)
	snapshot.display.MedicalRecordOwnerName = owner.Name
	return nil
}

func assertRevisionRecordOwner(ctx context.Context, tx *gorm.DB, clinicID uint64, locked *model.MedicalRecord) error {
	if locked.OwnerID == nil {
		return nil
	}
	if _, err := lockRevisionOwner(ctx, tx, clinicID, *locked.OwnerID); err != nil {
		return apperrors.Wrap(err, "failed to verify medical record owner")
	}
	return nil
}

func lockRevisionPetGraph(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, petID uint64,
) (*model.Pet, *model.Owner, *model.AnimalSpecies, error) {
	var pet model.Pet
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", petID, clinicID).
		First(&pet).Error; err != nil {
		return nil, nil, nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	owner, err := lockRevisionOwner(ctx, tx, clinicID, pet.OwnerID)
	if err != nil {
		return nil, nil, nil, apperrors.Wrap(err, "failed to verify current pet owner")
	}
	var species model.AnimalSpecies
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ?", pet.AnimalSpeciesID).
		First(&species).Error; err != nil {
		return nil, nil, nil, apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", pet.AnimalSpeciesID))
	}
	return &pet, owner, &species, nil
}

func lockRevisionStaff(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, staffID uint64,
	requireDoctor bool,
) (*model.Staff, error) {
	var staff model.Staff
	query := tx.WithContext(ctx).
		Model(&model.Staff{}).
		Select("staffs.*").
		Joins(`JOIN staff_clinic_assignments revision_assignment
			ON revision_assignment.staff_id = staffs.id
			AND revision_assignment.clinic_id = ?
			AND revision_assignment.deleted_at IS NULL`, clinicID).
		Where("staffs.id = ? AND staffs.deleted_at IS NULL AND staffs.is_active = TRUE", staffID)
	if requireDoctor {
		query = query.Where("staffs.staff_type = ?", model.StaffTypeDoctor)
	}
	if err := query.Clauses(clause.Locking{Strength: "SHARE"}).First(&staff).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	return &staff, nil
}

func lockRevisionItemFields(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, examTypeID uint64,
	items []model.ExamResult,
) error {
	fieldSet := make(map[uint64]struct{})
	for i := range items {
		if items[i].ExamTypeItemID != nil {
			fieldSet[*items[i].ExamTypeItemID] = struct{}{}
		}
	}
	if len(fieldSet) == 0 {
		return nil
	}
	fieldIDs := make([]uint64, 0, len(fieldSet))
	for fieldID := range fieldSet {
		fieldIDs = append(fieldIDs, fieldID)
	}
	var fields []model.ExamTypeField
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id IN ? AND clinic_id = ? AND exam_type_id = ?", fieldIDs, clinicID, examTypeID).
		Find(&fields).Error; err != nil {
		return apperrors.FromGORM(err, "exam_type_field", "revision snapshot")
	}
	if len(fields) != len(fieldIDs) {
		return apperrors.WrapInvalidInput("examination item references an invalid clinic-scoped exam type field")
	}
	return nil
}

// ConfirmWithRevisionCAS advances both parent status and revision pointer atomically from
// the actual locked pre-confirm status. A stale status or non-nil pointer is a conflict.
func (r *examinationRepository) ConfirmWithRevisionCAS(
	ctx context.Context,
	clinicID, examinationID uint64,
	expectedStatus model.ExaminationStatus,
	version uint64,
) (*model.Examination, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("examination revision CAS requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Examination{}).
		Where(
			"clinic_id = ? AND id = ? AND deleted_at IS NULL AND status = ? AND current_revision_version IS NULL",
			clinicID,
			examinationID,
			expectedStatus,
		).
		Updates(map[string]any{
			"status":                   model.ExaminationStatusConfirmed,
			"current_revision_version": version,
		})
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", examinationID))
	}
	if result.RowsAffected != 1 {
		return nil, apperrors.WrapConflict("examination confirmation was stale")
	}
	return r.FindByID(ctx, clinicID, examinationID)
}

// FindOfficialByID reads only immutable revision tables. It never joins mutable exams,
// exam_results, or current identity/master rows, and therefore fails closed for legacy
// confirmed rows without an official revision.
func (r *examinationRepository) FindOfficialByID(
	ctx context.Context,
	clinicID, examinationID uint64,
) (*ExaminationOfficialProjection, error) {
	db := persistence.DBOrTx(ctx, r.db)
	var revision model.ExaminationRevision
	if err := db.
		Where(
			"clinic_id = ? AND examination_id = ? AND kind = ? AND status = ?",
			clinicID,
			examinationID,
			model.ExaminationRevisionKindOfficial,
			model.ExaminationStatusConfirmed,
		).
		Order("version DESC").
		First(&revision).Error; err != nil {
		return nil, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d", examinationID))
	}

	items := make([]model.ExaminationRevisionItem, 0)
	if err := db.
		Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicID, examinationID, revision.Version).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d", examinationID))
	}
	var display model.ExaminationDisplaySnapshot
	if err := json.Unmarshal(revision.DisplaySnapshot, &display); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode examination display snapshot")
	}
	if revision.PetID != nil && (revision.PetOwnerID == nil || revision.AnimalSpeciesID == nil) {
		return nil, apperrors.WrapInternalServerError("official examination revision has an incomplete patient snapshot")
	}

	exam := projectOfficialExamination(revision, display, items)
	return &ExaminationOfficialProjection{
		Examination:     exam,
		OfficialVersion: revision.Version,
	}, nil
}

func projectOfficialExamination(
	revision model.ExaminationRevision,
	display model.ExaminationDisplaySnapshot,
	items []model.ExaminationRevisionItem,
) model.Examination {
	exam := model.Examination{
		ID:              revision.ExaminationID,
		ClinicID:        revision.ClinicID,
		MedicalRecordID: cloneOptionalUint64(revision.MedicalRecordID),
		PetID:           cloneOptionalUint64(revision.PetID),
		ExamTypeID:      revision.ExamTypeID,
		DoctorID:        cloneOptionalUint64(revision.DoctorID),
		JobID:           cloneOptionalUUID(revision.JobID),
		Date:            revision.Date,
		ResultSummary:   revision.ResultSummary,
		Machine:         revision.Machine,
		Status:          revision.Status,
		CreatedAt:       revision.CreatedAt,
		UpdatedAt:       revision.CreatedAt,
		ExaminationType: &model.ExaminationType{
			ID: revision.ExamTypeID, ClinicID: revision.ClinicID, Name: display.ExamTypeName,
		},
	}
	if revision.MedicalRecordID != nil {
		exam.MedicalRecord = &model.MedicalRecord{
			ID:       *revision.MedicalRecordID,
			ClinicID: revision.ClinicID,
			RecordNo: display.MedicalRecordNo,
			OwnerID:  cloneOptionalUint64(revision.MedicalRecordOwnerID),
			PetID:    cloneOptionalUint64(revision.PetID),
		}
		if revision.MedicalRecordOwnerID != nil {
			exam.MedicalRecord.Owner = &model.Owner{
				ID: *revision.MedicalRecordOwnerID, ClinicID: revision.ClinicID, Name: display.MedicalRecordOwnerName,
			}
		}
	}
	if revision.PetID != nil {
		exam.Pet = &model.Pet{
			ID:              *revision.PetID,
			ClinicID:        revision.ClinicID,
			OwnerID:         *revision.PetOwnerID,
			AnimalSpeciesID: *revision.AnimalSpeciesID,
			Name:            display.PetName,
			Owner: &model.Owner{
				ID: *revision.PetOwnerID, ClinicID: revision.ClinicID, Name: display.PetOwnerName,
			},
			AnimalSpecies: &model.AnimalSpecies{ID: *revision.AnimalSpeciesID, Name: display.SpeciesName},
		}
	}
	if revision.DoctorID != nil {
		exam.Doctor = &model.Staff{
			ID: *revision.DoctorID, ClinicID: revision.ClinicID, Name: display.DoctorName,
		}
	}
	exam.Items = make([]model.ExamResult, 0, len(items))
	for i := range items {
		item := items[i]
		exam.Items = append(exam.Items, model.ExamResult{
			ID:              item.ID,
			ExamID:          revision.ExaminationID,
			ExamTypeItemID:  cloneOptionalUint64(item.ExamTypeFieldID),
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Result:          item.Result,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          cloneOptionalFloat64(item.RefMin),
			RefMax:          cloneOptionalFloat64(item.RefMax),
			QualitativeMin:  cloneOptionalString(item.QualitativeMin),
			QualitativeMax:  cloneOptionalString(item.QualitativeMax),
			IsAbnormal:      item.IsAbnormal,
			Status:          item.Status,
			SortOrder:       item.SortOrder,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.CreatedAt,
		})
	}
	return exam
}

func cloneOptionalUint64(value *uint64) *uint64 {
	if value == nil {
		return nil
	}
	return cloneUint64(*value)
}

func cloneUint64(value uint64) *uint64 {
	cloned := value
	return &cloned
}

func cloneOptionalUUID(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
