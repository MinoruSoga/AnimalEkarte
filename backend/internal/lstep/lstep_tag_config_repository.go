package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type LstepTagConfigRepository interface {
	FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	CreateAutoManagedPrefix(ctx context.Context, m *model.LstepAutoManagedPrefix) error
	DeleteAutoManagedPrefix(ctx context.Context, id uint64) error

	FindAllConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	CreateConditionTagMapping(ctx context.Context, m *model.LstepConditionTagMapping) error
	DeleteConditionTagMapping(ctx context.Context, id uint64) error

	FindAllSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
	CreateSendPurposeTagPrefix(ctx context.Context, m *model.LstepSendPurposeTagPrefix) error
	DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error
}

type lstepTagConfigRepository struct {
	db *gorm.DB
}

func NewLstepTagConfigRepository(db *gorm.DB) LstepTagConfigRepository {
	return &lstepTagConfigRepository{db: db}
}

func (r *lstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	var prefixes []*model.LstepAutoManagedPrefix
	if err := r.db.WithContext(ctx).Order("category, prefix").Find(&prefixes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_auto_managed_prefix", "find_all")
	}
	return prefixes, nil
}

func (r *lstepTagConfigRepository) CreateAutoManagedPrefix(ctx context.Context, m *model.LstepAutoManagedPrefix) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_auto_managed_prefix", "create")
	}
	return nil
}

func (r *lstepTagConfigRepository) DeleteAutoManagedPrefix(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.LstepAutoManagedPrefix{}, id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_auto_managed_prefix", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_auto_managed_prefix", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *lstepTagConfigRepository) FindAllConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	var mappings []*model.LstepConditionTagMapping
	if err := r.db.WithContext(ctx).Order("condition_code").Find(&mappings).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_condition_tag_mapping", "find_all")
	}
	return mappings, nil
}

func (r *lstepTagConfigRepository) CreateConditionTagMapping(ctx context.Context, m *model.LstepConditionTagMapping) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_condition_tag_mapping", "create")
	}
	return nil
}

func (r *lstepTagConfigRepository) DeleteConditionTagMapping(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.LstepConditionTagMapping{}, id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_condition_tag_mapping", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_condition_tag_mapping", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *lstepTagConfigRepository) FindAllSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	var prefixes []*model.LstepSendPurposeTagPrefix
	if err := r.db.WithContext(ctx).Order("purpose").Find(&prefixes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "lstep_send_purpose_tag_prefix", "find_all")
	}
	return prefixes, nil
}

func (r *lstepTagConfigRepository) CreateSendPurposeTagPrefix(ctx context.Context, m *model.LstepSendPurposeTagPrefix) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_send_purpose_tag_prefix", "create")
	}
	return nil
}

func (r *lstepTagConfigRepository) DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.LstepSendPurposeTagPrefix{}, id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lstep_send_purpose_tag_prefix", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lstep_send_purpose_tag_prefix", fmt.Sprintf("%d", id))
	}
	return nil
}
