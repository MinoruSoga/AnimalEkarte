package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (r *examinationRepository) AppendWorkingRevisionFromOfficial(
	ctx context.Context,
	clinicID, examinationID, officialVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return 0, apperrors.WrapInternalServerError("examination revision workflow requires an ambient transaction")
	}
	exam, err := lockRevisionWorkflowParent(ctx, tx, clinicID, examinationID, officialVersion)
	if err != nil {
		return 0, err
	}
	if exam.Status != model.ExaminationStatusConfirmed {
		return 0, apperrors.WrapConflict("only a confirmed examination can be unconfirmed")
	}
	if err := validateRevisionWorkflowActorAndReason(ctx, tx, clinicID, actorID, changeReason); err != nil {
		return 0, err
	}
	source, items, err := loadRevisionForWorkflow(
		ctx,
		tx,
		clinicID,
		examinationID,
		officialVersion,
		model.ExaminationRevisionKindOfficial,
	)
	if err != nil {
		return 0, err
	}
	if source.Status != model.ExaminationStatusConfirmed {
		return 0, apperrors.WrapConflict("selected official examination revision is invalid")
	}
	if err := validateRevisionRestoreRelations(ctx, tx, clinicID, source, items); err != nil {
		return 0, err
	}
	nextVersion := officialVersion + 1
	if err := createCopiedExaminationRevision(
		ctx,
		tx,
		source,
		items,
		nextVersion,
		model.ExaminationRevisionKindWorking,
		model.ExaminationStatusCompleted,
		actorID,
		changeReason,
	); err != nil {
		return 0, err
	}
	if err := restoreMutableExaminationFromRevision(
		ctx,
		tx,
		clinicID,
		examinationID,
		officialVersion,
		exam.Status,
		source,
		items,
	); err != nil {
		return 0, err
	}
	return nextVersion, nil
}

func (r *examinationRepository) AppendWorkingRevisionFromCurrent(
	ctx context.Context,
	clinicID, examinationID, currentVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return 0, apperrors.WrapInternalServerError("examination revision workflow requires an ambient transaction")
	}
	exam, err := lockRevisionWorkflowParent(ctx, tx, clinicID, examinationID, currentVersion)
	if err != nil {
		return 0, err
	}
	if exam.Status == model.ExaminationStatusConfirmed {
		return 0, apperrors.WrapConflict("confirmed examination cannot append a working revision")
	}
	if err := validateRevisionWorkflowActorAndReason(ctx, tx, clinicID, actorID, changeReason); err != nil {
		return 0, err
	}
	_, _, err = loadRevisionForWorkflow(
		ctx,
		tx,
		clinicID,
		examinationID,
		currentVersion,
		model.ExaminationRevisionKindWorking,
	)
	if err != nil {
		return 0, err
	}
	snapshot, items, err := r.loadOfficialRevisionSnapshot(ctx, tx, clinicID, actorID, exam)
	if err != nil {
		return 0, err
	}
	displayJSON, err := json.Marshal(snapshot.display)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to encode examination display snapshot")
	}
	nextVersion := currentVersion + 1
	persistedReason := changeReason
	revision := &model.ExaminationRevision{
		ClinicID:             clinicID,
		ExaminationID:        examinationID,
		Version:              nextVersion,
		Kind:                 model.ExaminationRevisionKindWorking,
		Status:               exam.Status,
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
		ChangeReason:         &persistedReason,
	}
	if err := tx.WithContext(ctx).Create(revision).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d", examinationID))
	}
	revisionItems := buildExaminationRevisionItems(clinicID, examinationID, nextVersion, items)
	if len(revisionItems) > 0 {
		if err := tx.WithContext(ctx).Create(&revisionItems).Error; err != nil {
			return 0, apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d", examinationID))
		}
	}
	return nextVersion, nil
}

func (r *examinationRepository) AppendOfficialRevisionFromWorking(
	ctx context.Context,
	clinicID, examinationID, workingVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return 0, apperrors.WrapInternalServerError("examination revision workflow requires an ambient transaction")
	}
	exam, err := lockRevisionWorkflowParent(ctx, tx, clinicID, examinationID, workingVersion)
	if err != nil {
		return 0, err
	}
	if exam.Status == model.ExaminationStatusConfirmed {
		return 0, apperrors.WrapConflict("confirmed examination cannot be reconfirmed")
	}
	if err := validateRevisionWorkflowActorAndReason(ctx, tx, clinicID, actorID, changeReason); err != nil {
		return 0, err
	}
	source, items, err := loadRevisionForWorkflow(
		ctx,
		tx,
		clinicID,
		examinationID,
		workingVersion,
		model.ExaminationRevisionKindWorking,
	)
	if err != nil {
		return 0, err
	}
	if source.Status != exam.Status {
		return 0, apperrors.WrapConflict("selected working examination revision is stale")
	}
	if err := validateRevisionRestoreRelations(ctx, tx, clinicID, source, items); err != nil {
		return 0, err
	}
	nextVersion := workingVersion + 1
	if err := createCopiedExaminationRevision(
		ctx,
		tx,
		source,
		items,
		nextVersion,
		model.ExaminationRevisionKindOfficial,
		model.ExaminationStatusConfirmed,
		actorID,
		changeReason,
	); err != nil {
		return 0, err
	}
	if err := restoreMutableExaminationFromRevision(
		ctx,
		tx,
		clinicID,
		examinationID,
		workingVersion,
		exam.Status,
		source,
		items,
	); err != nil {
		return 0, err
	}
	return nextVersion, nil
}

func (r *examinationRepository) AdvanceRevisionCAS(
	ctx context.Context,
	clinicID, examinationID uint64,
	expectedStatus model.ExaminationStatus,
	expectedVersion uint64,
	nextStatus model.ExaminationStatus,
	nextVersion uint64,
) (*model.Examination, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("examination revision CAS requires an ambient transaction")
	}
	if nextVersion != expectedVersion+1 {
		return nil, apperrors.WrapInvalidInput("examination revision version must advance by one")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Examination{}).
		Where(
			"clinic_id = ? AND id = ? AND deleted_at IS NULL AND status = ? AND current_revision_version = ?",
			clinicID,
			examinationID,
			expectedStatus,
			expectedVersion,
		).
		Updates(map[string]any{
			"status":                   nextStatus,
			"current_revision_version": nextVersion,
		})
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", examinationID))
	}
	if result.RowsAffected != 1 {
		return nil, apperrors.WrapConflict("examination revision update was stale")
	}
	return r.FindByID(ctx, clinicID, examinationID)
}

func lockRevisionWorkflowParent(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, examinationID, expectedVersion uint64,
) (*model.Examination, error) {
	if expectedVersion == 0 {
		return nil, apperrors.WrapConflict("examination revision pointer is missing")
	}
	var exam model.Examination
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examinationID, clinicID).
		First(&exam).Error; err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examinationID))
	}
	if exam.CurrentRevisionVersion == nil || *exam.CurrentRevisionVersion != expectedVersion {
		return nil, apperrors.WrapConflict("examination revision pointer is stale")
	}
	return &exam, nil
}

func validateRevisionWorkflowActorAndReason(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, actorID uint64,
	changeReason string,
) error {
	if actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for examination revision")
	}
	if changeReason == "" {
		return apperrors.WrapInvalidInput("examination revision change reason is required")
	}
	if _, err := lockRevisionStaff(ctx, tx, clinicID, actorID, false); err != nil {
		return apperrors.Wrap(err, "failed to verify examination revision actor")
	}
	return nil
}

func loadRevisionForWorkflow(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, examinationID, version uint64,
	kind model.ExaminationRevisionKind,
) (*model.ExaminationRevision, []model.ExaminationRevisionItem, error) {
	var revision model.ExaminationRevision
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where(
			"clinic_id = ? AND examination_id = ? AND version = ? AND kind = ?",
			clinicID,
			examinationID,
			version,
			kind,
		).
		First(&revision).Error; err != nil {
		return nil, nil, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d version=%d", examinationID, version))
	}
	items := make([]model.ExaminationRevisionItem, 0)
	if err := tx.WithContext(ctx).
		Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicID, examinationID, version).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, nil, apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d version=%d", examinationID, version))
	}
	return &revision, items, nil
}

func createCopiedExaminationRevision(
	ctx context.Context,
	tx *gorm.DB,
	source *model.ExaminationRevision,
	items []model.ExaminationRevisionItem,
	nextVersion uint64,
	kind model.ExaminationRevisionKind,
	status model.ExaminationStatus,
	actorID uint64,
	changeReason string,
) error {
	copied := *source
	copied.ID = 0
	copied.Version = nextVersion
	copied.Kind = kind
	copied.Status = status
	copied.ActorID = actorID
	copied.DisplaySnapshot = append(json.RawMessage(nil), source.DisplaySnapshot...)
	copied.SchemaVersion = examinationRevisionSchemaVersion
	copied.ChangeReason = cloneOptionalString(&changeReason)
	if err := tx.WithContext(ctx).Omit("CreatedAt").Create(&copied).Error; err != nil {
		return apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d", source.ExaminationID))
	}
	copiedItems := make([]model.ExaminationRevisionItem, 0, len(items))
	for i := range items {
		item := items[i]
		item.ID = 0
		item.Version = nextVersion
		copiedItems = append(copiedItems, item)
	}
	if len(copiedItems) > 0 {
		if err := tx.WithContext(ctx).Omit("CreatedAt").Create(&copiedItems).Error; err != nil {
			return apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d", source.ExaminationID))
		}
	}
	return nil
}

func validateRevisionRestoreRelations(
	ctx context.Context,
	tx *gorm.DB,
	clinicID uint64,
	revision *model.ExaminationRevision,
	items []model.ExaminationRevisionItem,
) error {
	var examType model.ExaminationType
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", revision.ExamTypeID, clinicID).
		First(&examType).Error; err != nil {
		return apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", revision.ExamTypeID))
	}

	var record *model.MedicalRecord
	if revision.MedicalRecordID != nil {
		var locked model.MedicalRecord
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *revision.MedicalRecordID, clinicID).
			First(&locked).Error; err != nil {
			return apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", *revision.MedicalRecordID))
		}
		if locked.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの検査確定状態は変更できません")
		}
		if locked.OwnerID != nil {
			if _, err := lockRevisionOwner(ctx, tx, clinicID, *locked.OwnerID); err != nil {
				return apperrors.Wrap(err, "failed to verify medical record owner")
			}
		}
		record = &locked
	}

	var pet *model.Pet
	if revision.PetID != nil {
		locked, _, _, err := lockRevisionPetGraph(ctx, tx, clinicID, *revision.PetID)
		if err != nil {
			return err
		}
		pet = locked
	}
	if record != nil && record.PetID != nil {
		if pet == nil || pet.ID != *record.PetID {
			return apperrors.WrapNotFound("medical_record", "pet relation")
		}
	}
	if revision.DoctorID != nil {
		if _, err := lockRevisionStaff(ctx, tx, clinicID, *revision.DoctorID, true); err != nil {
			return apperrors.Wrap(err, "failed to verify examination doctor")
		}
	}
	if revision.JobID != nil {
		var job model.LabImportJob
		if err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Where("id = ? AND clinic_id = ?", *revision.JobID, clinicID).
			First(&job).Error; err != nil {
			return apperrors.FromGORM(err, "lab_import_job", revision.JobID.String())
		}
	}
	fieldIDs := make([]uint64, 0, len(items))
	fieldSet := make(map[uint64]struct{}, len(items))
	for i := range items {
		if items[i].ExamTypeFieldID == nil {
			continue
		}
		fieldID := *items[i].ExamTypeFieldID
		if _, seen := fieldSet[fieldID]; seen {
			continue
		}
		fieldSet[fieldID] = struct{}{}
		fieldIDs = append(fieldIDs, fieldID)
	}
	if len(fieldIDs) == 0 {
		return nil
	}
	var fields []model.ExamTypeField
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id IN ? AND clinic_id = ? AND exam_type_id = ?", fieldIDs, clinicID, revision.ExamTypeID).
		Find(&fields).Error; err != nil {
		return apperrors.FromGORM(err, "exam_type_field", "revision restore")
	}
	if len(fields) != len(fieldIDs) {
		return apperrors.WrapInvalidInput("examination revision references an invalid clinic-scoped exam type field")
	}
	return nil
}

func restoreMutableExaminationFromRevision(
	ctx context.Context,
	tx *gorm.DB,
	clinicID, examinationID, expectedVersion uint64,
	expectedStatus model.ExaminationStatus,
	revision *model.ExaminationRevision,
	items []model.ExaminationRevisionItem,
) error {
	result := tx.WithContext(ctx).
		Model(&model.Examination{}).
		Where(
			"clinic_id = ? AND id = ? AND deleted_at IS NULL AND status = ? AND current_revision_version = ?",
			clinicID,
			examinationID,
			expectedStatus,
			expectedVersion,
		).
		Updates(map[string]any{
			"medical_record_id": revision.MedicalRecordID,
			"pet_id":            revision.PetID,
			"exam_type_id":      revision.ExamTypeID,
			"doctor_id":         revision.DoctorID,
			"job_id":            revision.JobID,
			"date":              revision.Date,
			"result_summary":    revision.ResultSummary,
			"machine":           revision.Machine,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", examinationID))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("examination revision restore was stale")
	}
	deleteResult := tx.WithContext(ctx).
		Where(
			`exam_id = ? AND EXISTS (
				SELECT 1 FROM exams scoped_exam
				WHERE scoped_exam.id = exam_results.exam_id
				  AND scoped_exam.clinic_id = ?
				  AND scoped_exam.deleted_at IS NULL
			)`,
			examinationID,
			clinicID,
		).
		Delete(&model.ExamResult{})
	if deleteResult.Error != nil {
		return apperrors.FromGORM(deleteResult.Error, "exam_item", fmt.Sprintf("exam_id=%d", examinationID))
	}
	mutableItems := make([]model.ExamResult, 0, len(items))
	for i := range items {
		item := items[i]
		mutableItems = append(mutableItems, model.ExamResult{
			ExamID:          examinationID,
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
		})
	}
	if len(mutableItems) > 0 {
		if err := tx.WithContext(ctx).Create(&mutableItems).Error; err != nil {
			return apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examinationID))
		}
	}
	return nil
}

func buildExaminationRevisionItems(
	clinicID, examinationID, version uint64,
	items []model.ExamResult,
) []model.ExaminationRevisionItem {
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
			Version:         version,
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
	return revisionItems
}
