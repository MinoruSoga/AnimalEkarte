package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CheckupTypeRepository is the data access interface for checkup types.
// Moved from internal/repository/checkuptype (BE8-4 batch5) — BE9-2D roll-up. Renamed from that
// subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type CheckupTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error)
	Create(ctx context.Context, checkupType *model.CheckupType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type checkupTypeRepository struct{ db *gorm.DB }

// NewCheckupTypeRepository constructs a CheckupTypeRepository.
func NewCheckupTypeRepository(db *gorm.DB) CheckupTypeRepository {
	return &checkupTypeRepository{db: db}
}

func (r *checkupTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.CheckupType, error) {
	checkupTypes := make([]model.CheckupType, 0)
	err := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type", "")
	}
	return checkupTypes, nil
}

func (r *checkupTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CheckupType, error) {
	return persistence.FindByIDScoped[model.CheckupType](ctx, r.db, "checkup_type", clinicID, id)
}

func (r *checkupTypeRepository) Create(ctx context.Context, checkupType *model.CheckupType) error {
	db := r.db.WithContext(ctx)
	wantActive := checkupType.IsActive
	if err := db.Create(checkupType).Error; err != nil {
		return apperrors.FromGORM(err, "checkup_type", "")
	}
	if !wantActive {
		if err := db.Model(checkupType).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "checkup_type", fmt.Sprintf("%d", checkupType.ID))
		}
		checkupType.IsActive = false
	}
	return nil
}

func (r *checkupTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.CheckupType, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *checkupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM checkup_types children
			WHERE children.parent_id = checkup_types.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM checkups
			WHERE checkups.checkup_type_id = checkup_types.id
			  AND checkups.clinic_id = ?
			  AND checkups.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.CheckupType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeCheckupTypeDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *checkupTypeRepository) normalizeCheckupTypeDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この定期健診種別にはサブ種別が登録されているため削除できません")
	}
	return apperrors.WrapConflict("この定期健診種別は健診記録で使用中のため削除できません")
}

func (r *checkupTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.CheckupType{}, "checkup_type", clinicID, ids, "sort_order")
}

// CountUsageByCheckupTypeID は定期健診種別を参照している checkups の件数を返す（BUG-107）
// checkups テーブルは直接 clinic_id を持つためテナント分離を直接適用する
func (r *checkupTypeRepository) CountUsageByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Checkup{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("checkup_type_id = ? AND deleted_at IS NULL", checkupTypeID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}

// CountChildrenByParentID は指定した親 ID を持つ子健診種別の件数を返す。
func (r *checkupTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.CheckupType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "checkup_type", "")
	}
	return count, nil
}
