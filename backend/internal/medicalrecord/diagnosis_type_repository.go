package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// DiagnosisTypeRepository is the data access interface for diagnosis types (categories).
// Moved from internal/repository/diagnosistype (BE8-4 batch27) — BE9-2C roll-up. Renamed
// from that subpackage's generic "Repository" to this entity-specific name only because
// medicalrecord holds multiple repository interfaces in one package; every external caller
// only ever saw this name via the internal/repository facade, so no call site changes.
type DiagnosisTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	Create(ctx context.Context, category *model.DiagnosisType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error)
}

type diagnosisTypeRepository struct{ db *gorm.DB }

// NewDiagnosisTypeRepository constructs a DiagnosisTypeRepository.
func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return &diagnosisTypeRepository{db: db}
}

func (r *diagnosisTypeRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisType{}).Scopes(repohelpers.ClinicScope(clinicID))
	}
	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_type", "")
	}
	categories := make([]model.DiagnosisType, 0)
	if err := buildBase().
		Preload("Names", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(paginate(page, limit)).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_type", "")
	}
	return categories, total, nil
}

func (r *diagnosisTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	var category model.DiagnosisType
	err := r.db.WithContext(ctx).
		Preload("Names", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *diagnosisTypeRepository) Create(ctx context.Context, category *model.DiagnosisType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return apperrors.FromGORM(err, "diagnosis_type", "")
	}
	return nil
}

func (r *diagnosisTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.DiagnosisType{}, "diagnosis_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *diagnosisTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.DiagnosisType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountChildrenByParentID は指定カテゴリに属する diagnosis_names の件数を返す（BUG-113 補足）
// diagnosis_names テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *diagnosisTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.DiagnosisName{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("diagnosis_type_id = ? AND deleted_at IS NULL", categoryID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return count, nil
}

// Reorder はトランザクション内でカテゴリの sort_order を ids の順序で更新する (#019)
func (r *diagnosisTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.DiagnosisType{}, "diagnosis_type", clinicID, ids, "sort_order")
}
