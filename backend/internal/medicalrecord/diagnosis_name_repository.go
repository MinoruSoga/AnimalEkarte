package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// DiagnosisNameRepository is the data access interface for diagnosis names. Moved from
// internal/repository/diagnosisname (BE8-4 batch28) — BE9-2C roll-up (see
// diagnosis_type_repository.go's header for the rename rationale).
type DiagnosisNameRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindAllByCategoryID(ctx context.Context, clinicID, categoryID uint64, page, limit int) ([]model.DiagnosisName, int64, error)
	FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.DiagnosisName, error)
	Create(ctx context.Context, name *model.DiagnosisName) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateDiagnosisNameInput) (*model.DiagnosisName, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error)
}

type diagnosisNameRepository struct{ db *gorm.DB }

// NewDiagnosisNameRepository constructs a DiagnosisNameRepository.
func NewDiagnosisNameRepository(db *gorm.DB) DiagnosisNameRepository {
	return &diagnosisNameRepository{db: db}
}

func (r *diagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.DiagnosisName, int64, error) {
	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(persistence.ClinicScope(clinicID))
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
			Scopes(persistence.ClinicScope(clinicID)).
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

// FindAllByFilter はページネーションなしで全件取得する（#418: ListNames 用）。
// typeID が非 nil の場合は該当カテゴリのみ、nil の場合はクリニック全件を返す。
// is_active = true のレコードのみを返す（CODE-QUALITY-232）。
func (r *diagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {
	q := r.db.WithContext(ctx).Model(&model.DiagnosisName{}).Scopes(persistence.ClinicScope(clinicID)).
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
	return persistence.FindByIDScoped[model.DiagnosisName](ctx, r.db, "diagnosis_name", clinicID, id)
}

func (r *diagnosisNameRepository) Create(ctx context.Context, name *model.DiagnosisName) error {
	db := r.db.WithContext(ctx)
	wantActive := name.IsActive
	if err := db.Create(name).Error; err != nil {
		return apperrors.FromGORM(err, "diagnosis_name", "")
	}
	if !wantActive {
		if err := db.Model(name).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "diagnosis_name", fmt.Sprintf("%d", name.ID))
		}
		name.IsActive = false
	}
	return nil
}

func (r *diagnosisNameRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
	if err := r.update(ctx, clinicID, id, buildDiagnosisNameUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *diagnosisNameRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, r.db, &model.DiagnosisName{}, "diagnosis_name", clinicID, id, fields)
}

func (r *diagnosisNameRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM clinical_plans
			JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id
			  AND medical_records.clinic_id = ?
			  AND medical_records.deleted_at IS NULL
			WHERE (clinical_plans.diagnosis_name_id = diagnosis_names.id
			    OR clinical_plans.diagnosis_2_name_id = diagnosis_names.id)
			  AND clinical_plans.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.DiagnosisName{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "diagnosis_name", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDiagnosisNameDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *diagnosisNameRepository) normalizeDiagnosisNameDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この診断名は診療記録で使用中のため削除できません")
}

// CountUsageByDiagnosisNameID は診断名を参照している clinical_plans の件数を返す（BUG-113）
// diagnosis_name_id および diagnosis_2_name_id 両方をカウントする
// clinical_plans は直接 clinic_id を持たないため medical_records を JOIN してテナント分離する
func (r *diagnosisNameRepository) CountUsageByDiagnosisNameID(ctx context.Context, clinicID, diagnosisNameID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ClinicalPlan{}).
		Scopes(persistence.MedicalRecordTenantScope("clinical_plans", clinicID)).
		Where("(clinical_plans.diagnosis_name_id = ? OR clinical_plans.diagnosis_2_name_id = ?) AND clinical_plans.deleted_at IS NULL", diagnosisNameID, diagnosisNameID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "clinical_plan", "")
	}
	return count, nil
}

// Reorder はトランザクション内で診断名の sort_order を ids の順序で更新する (#019)
func (r *diagnosisNameRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.DiagnosisName{}, "diagnosis_name", clinicID, ids, "sort_order")
}
