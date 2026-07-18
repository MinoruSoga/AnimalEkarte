// Package diagnosistype owns diagnosis type (category) master data access.
package diagnosistype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for diagnosis types (categories).
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error)
	Create(ctx context.Context, category *model.DiagnosisType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// paginate は 1-origin の page と limit を Offset/Limit に変換するローカル Scope。
// repohelpers 未収載のため import cycle 回避のためローカル定義（BE8-4 batch27、checkup 等と同じ理由）。
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisType, int64, error) {
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

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisType, error) {
	var category model.DiagnosisType
	err := r.db.WithContext(ctx).
		Preload("Names", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_type", fmt.Sprintf("%d", id))
	}
	return &category, nil
}

func (r *repository) Create(ctx context.Context, category *model.DiagnosisType) error {
	err := r.db.WithContext(ctx).Create(category).Error
	if err != nil {
		return apperrors.FromGORM(err, "diagnosis_type", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.DiagnosisType{}, "diagnosis_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
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
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
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
func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.DiagnosisType{}, "diagnosis_type", clinicID, ids, "sort_order")
}
