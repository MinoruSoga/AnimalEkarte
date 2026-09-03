package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ExamTypeRepository is the data access interface for examination types. Moved from
// internal/repository/examtype — BE9-2C roll-up (see diagnosis_type_repository.go's header
// for the rename rationale).
type ExamTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error)
	Create(ctx context.Context, exType *model.ExaminationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	CreateField(ctx context.Context, field *model.ExamTypeField) error
	LockFieldByID(ctx context.Context, clinicID, examTypeID, fieldID uint64) (*model.ExamTypeField, error)
	UpdateField(ctx context.Context, clinicID, examTypeID, fieldID uint64, fields map[string]any) (*model.ExamTypeField, error)
	DeleteField(ctx context.Context, clinicID, examTypeID, fieldID uint64) error
	ReorderFields(ctx context.Context, clinicID, examTypeID uint64, ids []uint64) error
	CountExamResultsByFieldID(ctx context.Context, fieldID uint64) (int64, error)
	CountReferenceRangesByFieldID(ctx context.Context, fieldID uint64) (int64, error)
	AnimalSpeciesExists(ctx context.Context, speciesID uint64) (bool, error)
	ReplaceReferenceRanges(ctx context.Context, clinicID, fieldID uint64, ranges []model.ExamReferenceRange) error
	FindReferenceRangesByFieldIDs(ctx context.Context, clinicID uint64, fieldIDs []uint64) (map[uint64][]model.ExamReferenceRange, error)
}

type examTypeRepository struct{ db *gorm.DB }

// NewExamTypeRepository constructs an ExamTypeRepository.
func NewExamTypeRepository(db *gorm.DB) ExamTypeRepository {
	return &examTypeRepository{db: db}
}

func (r *examTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ExaminationType, error) {
	exTypes := make([]model.ExaminationType, 0)
	// G2F-11: vaccine/procedure/consultation と同型の master list safety Limit（unbounded Find 防止）。
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Preload("Items", "clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Limit(persistence.MaxMasterListRows).
		Find(&exTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", "")
	}
	return exTypes, nil
}

func (r *examTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ExaminationType, error) {
	var exType model.ExaminationType
	db := persistence.DBOrTx(ctx, r.db)
	lockForValidation := persistence.TxFromContext(ctx) != nil
	if lockForValidation {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.Preload("Items", func(itemsDB *gorm.DB) *gorm.DB {
		itemsDB = itemsDB.Where("clinic_id = ?", clinicID)
		if lockForValidation {
			itemsDB = itemsDB.Clauses(clause.Locking{Strength: "SHARE"})
		}
		return itemsDB
	}).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&exType).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", id))
	}
	return &exType, nil
}

func (r *examTypeRepository) Create(ctx context.Context, exType *model.ExaminationType) error {
	// MRB-08: ambient tx 参加（examTypeService.Create の WithTx 内 parent 検証と同一コネクション）。
	db := persistence.DBOrTx(ctx, r.db)
	wantActive := exType.IsActive
	if err := db.Create(exType).Error; err != nil {
		return apperrors.FromGORM(err, "examination_type", "")
	}
	if !wantActive {
		if err := db.Model(exType).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "examination_type", fmt.Sprintf("%d", exType.ID))
		}
		exType.IsActive = false
	}
	return nil
}

func (r *examTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ExaminationType, error) {
	if err := persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.ExaminationType{}, "examination_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *examTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM examination_types children
			WHERE children.parent_id = examination_types.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM examinations
			WHERE examinations.exam_type_id = examination_types.id
			  AND examinations.clinic_id = ?
			  AND examinations.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.ExaminationType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "examination_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeExamTypeDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *examTypeRepository) normalizeExamTypeDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
	}
	return apperrors.WrapConflict("この検査種別は検査記録で使用中のため削除できません")
}

func (r *examTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, persistence.DBOrTx(ctx, r.db), &model.ExaminationType{}, "examination_type", clinicID, ids, "sort_order")
}

// CountUsageByExamTypeID returns exam references (BUG-107).
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Examination{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("exam_type_id = ? AND deleted_at IS NULL", examTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "exam_type", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child examination-type count before parent delete.
func (r *examTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ExaminationType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "examination_type", "")
	}
	return count, nil
}

func (r *examTypeRepository) CreateField(ctx context.Context, field *model.ExamTypeField) error {
	err := persistence.DBOrTx(ctx, r.db).Create(field).Error
	return apperrors.FromGORM(err, "exam_type_field", "")
}

func (r *examTypeRepository) LockFieldByID(
	ctx context.Context,
	clinicID, examTypeID, fieldID uint64,
) (*model.ExamTypeField, error) {
	var field model.ExamTypeField
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND exam_type_id = ? AND id = ?", clinicID, examTypeID, fieldID).
		First(&field).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	return &field, nil
}

func (r *examTypeRepository) UpdateField(
	ctx context.Context,
	clinicID, examTypeID, fieldID uint64,
	fields map[string]any,
) (*model.ExamTypeField, error) {
	result := persistence.DBOrTx(ctx, r.db).Model(&model.ExamTypeField{}).
		Where("clinic_id = ? AND exam_type_id = ? AND id = ?", clinicID, examTypeID, fieldID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	if result.RowsAffected != 1 {
		return nil, apperrors.WrapNotFound("exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	var field model.ExamTypeField
	if err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND exam_type_id = ? AND id = ?", clinicID, examTypeID, fieldID).
		First(&field).Error; err != nil {
		return nil, apperrors.FromGORM(err, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	return &field, nil
}

func (r *examTypeRepository) DeleteField(ctx context.Context, clinicID, examTypeID, fieldID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND exam_type_id = ? AND id = ?", clinicID, examTypeID, fieldID).
		Delete(&model.ExamTypeField{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapNotFound("exam_type_field", fmt.Sprintf("%d", fieldID))
	}
	return nil
}

func (r *examTypeRepository) ReorderFields(
	ctx context.Context,
	clinicID, examTypeID uint64,
	ids []uint64,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	var fields []model.ExamTypeField
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ? AND exam_type_id = ? AND id IN ?", clinicID, examTypeID, ids).
		Find(&fields).Error; err != nil {
		return apperrors.FromGORM(err, "exam_type_field", "")
	}
	if len(fields) != len(ids) {
		return apperrors.WrapNotFound("exam_type_field", "")
	}
	for position, id := range ids {
		result := db.Model(&model.ExamTypeField{}).
			Where("clinic_id = ? AND exam_type_id = ? AND id = ?", clinicID, examTypeID, id).
			Update("sort_order", position+1)
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "exam_type_field", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *examTypeRepository) CountExamResultsByFieldID(ctx context.Context, fieldID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.ExamResult{}).
		Where("exam_type_field_id = ?", fieldID).Count(&count).Error
	return count, apperrors.FromGORM(err, "exam_result", "")
}

func (r *examTypeRepository) CountReferenceRangesByFieldID(ctx context.Context, fieldID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.ExamReferenceRange{}).
		Where("exam_type_field_id = ?", fieldID).Count(&count).Error
	return count, apperrors.FromGORM(err, "exam_reference_range", "")
}

func (r *examTypeRepository) AnimalSpeciesExists(ctx context.Context, speciesID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.AnimalSpecies{}).
		Where("id = ?", speciesID).Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "animal_species", fmt.Sprintf("%d", speciesID))
	}
	return count == 1, nil
}

func (r *examTypeRepository) ReplaceReferenceRanges(
	ctx context.Context,
	clinicID, fieldID uint64,
	ranges []model.ExamReferenceRange,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Where("clinic_id = ? AND exam_type_field_id = ?", clinicID, fieldID).
		Delete(&model.ExamReferenceRange{}).Error; err != nil {
		return apperrors.FromGORM(err, "exam_reference_range", "")
	}
	if len(ranges) == 0 {
		return nil
	}
	return apperrors.FromGORM(db.Create(&ranges).Error, "exam_reference_range", "")
}

func (r *examTypeRepository) FindReferenceRangesByFieldIDs(
	ctx context.Context,
	clinicID uint64,
	fieldIDs []uint64,
) (map[uint64][]model.ExamReferenceRange, error) {
	result := make(map[uint64][]model.ExamReferenceRange, len(fieldIDs))
	if len(fieldIDs) == 0 {
		return result, nil
	}
	var ranges []model.ExamReferenceRange
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND exam_type_field_id IN ?", clinicID, fieldIDs).
		Order("animal_species_id ASC").
		Find(&ranges).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_reference_range", "")
	}
	for i := range ranges {
		fieldID := ranges[i].ExamTypeFieldID
		result[fieldID] = append(result[fieldID], ranges[i])
	}
	return result, nil
}
