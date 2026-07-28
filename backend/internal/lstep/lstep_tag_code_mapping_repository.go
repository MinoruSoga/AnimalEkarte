package lstep

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LstepTagCodeMappingRepository は per-clinic コード→タグ マッピングのリポジトリ。
type LstepTagCodeMappingRepository interface {
	FindAllByClinicID(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	FindByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error)
	Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error
	SoftDelete(ctx context.Context, clinicID, id uint64) error
	// SoftDeleteByClinicIDAndTagName は指定タグ名に紐づく全 mapping を一括ソフトデリートする（PUT replace 用）。
	SoftDeleteByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) error
}

type lstepTagCodeMappingRepository struct {
	db *gorm.DB
}

// NewLstepTagCodeMappingRepository はリポジトリを生成する。
func NewLstepTagCodeMappingRepository(db *gorm.DB) LstepTagCodeMappingRepository {
	return &lstepTagCodeMappingRepository{db: db}
}

func (r *lstepTagCodeMappingRepository) FindAllByClinicID(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	var mappings []*model.LstepTagCodeMapping
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Find(&mappings).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("clinic:%d", clinicID))
	}
	return mappings, nil
}

func (r *lstepTagCodeMappingRepository) FindByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error) {
	var mappings []*model.LstepTagCodeMapping
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND tag_name = ? AND deleted_at IS NULL", clinicID, tagName).
		Find(&mappings).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("clinic:%d tag:%s", clinicID, tagName))
	}
	return mappings, nil
}

func (r *lstepTagCodeMappingRepository) Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(mapping).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_tag_code_mapping", "create")
	}
	return nil
}

func (r *lstepTagCodeMappingRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	now := time.Now()
	result := persistence.DBOrTx(ctx, r.db).Model(&model.LstepTagCodeMapping{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_tag_code_mapping", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_tag_code_mapping", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *lstepTagCodeMappingRepository) SoftDeleteByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) error {
	now := time.Now()
	err := persistence.DBOrTx(ctx, r.db).Model(&model.LstepTagCodeMapping{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("tag_name = ? AND deleted_at IS NULL", tagName).
		Update("deleted_at", now).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("clinic:%d tag:%s", clinicID, tagName))
	}
	return nil
}
