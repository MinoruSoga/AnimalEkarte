package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepTagCodeMappingRepository は per-clinic コード→タグ マッピングのリポジトリ。
type LstepTagCodeMappingRepository interface {
	FindAllByClinicID(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	FindByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error)
	Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.LstepTagCodeMapping, error)
	SoftDelete(ctx context.Context, clinicID, id uint64) error
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
	if err := r.db.WithContext(ctx).Create(mapping).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_tag_code_mapping", "create")
	}
	return nil
}

func (r *lstepTagCodeMappingRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.LstepTagCodeMapping, error) {
	err := r.db.WithContext(ctx).Model(&model.LstepTagCodeMapping{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("%d", id))
	}
	var updated model.LstepTagCodeMapping
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&updated).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("%d", id))
	}
	return &updated, nil
}

func (r *lstepTagCodeMappingRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&model.LstepTagCodeMapping{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_code_mapping", fmt.Sprintf("%d", id))
	}
	return nil
}
