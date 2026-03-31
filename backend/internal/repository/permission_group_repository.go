package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionGroupRepository は権限グループのデータアクセスインターフェース
type PermissionGroupRepository interface {
	FindByCompanyID(ctx context.Context, companyID uint64) ([]model.PermissionGroup, error)
	FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, group *model.PermissionGroup) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	HasPermission(ctx context.Context, userID uint64, resource model.Resource, action string) (bool, error)
	VerifyGroupExists(ctx context.Context, groupID uint64) error
}

type permissionGroupRepository struct {
	db *gorm.DB
}

// NewPermissionGroupRepository はPermissionGroupRepositoryを初期化して返す
func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindByCompanyID(ctx context.Context, companyID uint64) ([]model.PermissionGroup, error) {
	var groups []model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		Where("company_id = ? AND deleted_at IS NULL", companyID).
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
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		// BUG-032: 同一 company 内のグループ名重複は 409 に変換する
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("permission_group", group.Name)
		}
		return apperrors.Wrap(err, "create permission group")
	}
	return nil
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
		return apperrors.Wrap(err, "set permission group rules")
	}
	return nil
}

// VerifyGroupExists は権限グループが存在するかどうかを確認する（BUG-031: 存在しない場合はErrNotFoundを返す）
func (r *permissionGroupRepository) VerifyGroupExists(ctx context.Context, groupID uint64) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PermissionGroup{}).
		Where("id = ? AND deleted_at IS NULL", groupID).
		Count(&count).Error
	if err != nil {
		return apperrors.Wrap(err, "verify group exists")
	}
	if count == 0 {
		return apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", groupID))
	}
	return nil
}

// HasPermission はユーザーが指定リソースに対して指定アクションの権限を持つかどうかを確認する
func (r *permissionGroupRepository) HasPermission(ctx context.Context, userID uint64, resource model.Resource, action string) (bool, error) {
	type result struct {
		HasPerm bool
	}
	var res result

	column := ""
	switch action {
	case "view":
		column = "bool_or(pgr.can_view)"
	case "create":
		column = "bool_or(pgr.can_create)"
	case "edit":
		column = "bool_or(pgr.can_edit)"
	case "delete":
		column = "bool_or(pgr.can_delete)"
	default:
		return false, nil
	}

	err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT COALESCE(%s, false) AS has_perm
		FROM user_permission_groups upg
		JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
		JOIN permission_group_rules pgr ON pgr.group_id = pg.id AND pgr.resource = ?
		WHERE upg.user_id = ?
	`, column), string(resource), userID).Scan(&res).Error
	if err != nil {
		return false, apperrors.Wrap(err, "check permission")
	}
	return res.HasPerm, nil
}
