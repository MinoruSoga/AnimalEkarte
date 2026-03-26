package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionGroupRepository は権限グループのデータアクセスインターフェース
type PermissionGroupRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, group *model.PermissionGroup) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
}

type permissionGroupRepository struct {
	db *gorm.DB
}

// NewPermissionGroupRepository はPermissionGroupRepositoryを初期化して返す
func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	var groups []model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("id ASC").
		Find(&groups).Error
	return groups, err
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *permissionGroupRepository) UpdateFields(ctx context.Context, id uint64, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error
}

func (r *permissionGroupRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&model.PermissionGroup{}).Error
}

// SetRules はトランザクション内で既存ルールを全削除→新規一括挿入する
func (r *permissionGroupRepository) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&model.PermissionGroupRule{}).Error; err != nil {
			return err
		}
		if len(rules) == 0 {
			return nil
		}
		return tx.Create(&rules).Error
	}); err != nil {
		return fmt.Errorf("set permission group rules: %w", err)
	}
	return nil
}
