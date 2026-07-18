// Package repository provides data access implementations for DiagnosisType and DiagnosisName entities.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/diagnosistype"
)

// DiagnosisTypeRepository is a facade alias to diagnosistype.Repository (BE8-4 batch27).
type DiagnosisTypeRepository = diagnosistype.Repository

// NewDiagnosisTypeRepository constructs a DiagnosisTypeRepository.
func NewDiagnosisTypeRepository(db *gorm.DB) DiagnosisTypeRepository {
	return diagnosistype.New(db)
}

// ---- DiagnosisName ----

type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindAllByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error)
}

type diagnosisNameRepository struct{ db *gorm.DB }

func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(clinicScope(clinicID))
	}
	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	names := make([]model.DiagnosisName, 0)
	if err := buildBase().
		Scopes(paginate(page, limit)).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, total, nil
}

func (r *diagnosisNameRepository) FindAllByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).
			Scopes(clinicScope(clinicID)).
			Where("diagnosis_type_id = ?", categoryID)
	}
	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	names := make([]model.DiagnosisName, 0)
	if err := buildBase().
		Scopes(paginate(page, limit)).
		Order("sort_order ASC, name ASC").
		Find(&names).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, total, nil
}

// FindAll はページネーションなしで全件取得する（#418: ListNames 用）。
// typeID が非 nil の場合は該当カテゴリのみ、nil の場合はクリニック全件を返す。
// is_active = true のレコードのみを返す（CODE-QUALITY-232）。
func (r *diagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	q := r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(clinicScope(clinicID)).
		Where("is_active = ?", true)
	if typeID != nil {
		q = q.Where("diagnosis_type_id = ?", *typeID)
	}
	names := make([]model.DiagnosisName, 0)
	if err := q.Order("sort_order ASC, name ASC").Find(&names).Error; err != nil {
		return nil, apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return names, nil
}

func (r *diagnosisNameRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error) {
	return findByIDScoped[model.DiagnosisName](ctx, r.db, "diagnosis_name", clinicID, id)
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	err := r.db.WithContext(ctx).Create(name).Error
	if err != nil {
		return apperrors.FromGORM(err, "diagnosis_name", "")
	}
	return nil
}

func (r *diagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.DiagnosisName, error) {
	if err := updateScopedByID(ctx, r.db, &model.DiagnosisName{}, "diagnosis_name", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.DiagnosisName{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("diagnosis_name", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByDiagnosisNameID は診断名を参照している clinical_plans の件数を返す（BUG-113）
// diagnosis_name_id および diagnosis_2_name_id 両方をカウントする
// clinical_plans は直接 clinic_id を持たないため medical_records を JOIN してテナント分離する
func (r *diagnosisNameRepository) CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Scopes(medicalRecordTenantScope("clinical_plans", clinicID)).
		Where("(clinical_plans.diagnosis_name_id = ? OR clinical_plans.diagnosis_2_name_id = ?) AND clinical_plans.deleted_at IS NULL", diagnosisNameID, diagnosisNameID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "clinical_plan", "")
	}
	return count, nil
}

// Reorder はトランザクション内で診断名の sort_order を ids の順序で更新する (#019)
func (r *diagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.DiagnosisName{}, "diagnosis_name", clinicID, ids)
}
