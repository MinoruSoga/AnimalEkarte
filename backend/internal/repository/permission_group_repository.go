// Package repository provides data access implementations for PermissionGroup entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- PermissionGroup ----

type PermissionGroupRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, group *model.PermissionGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	CountStaffsByGroupID(ctx context.Context, groupID uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	// GetEffectivePermissionsByStaffID はスタッフが所属する全権限グループのルールを
	// UNION (bool_or) して実効権限を返す。
	GetEffectivePermissionsByStaffID(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error)
	// GetGroupIDsByStaffID はスタッフが所属する権限グループIDリストを返す。
	GetGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error)
	// SetStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
	SetStaffGroups(ctx context.Context, staffID uint64, groupIDs []uint64) error
}

type permissionGroupRepository struct{ db *gorm.DB }

func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	groups := make([]model.PermissionGroup, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Rules").
		Order("sort_order ASC, name ASC").
		Find(&groups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", "")
	}
	return groups, nil
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		First(&group, "id = ? AND deleted_at IS NULL", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
	}
	return &group, nil
}

func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	err := r.db.WithContext(ctx).Create(group).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "permission_group", "")
	}
	return nil
}

func (r *permissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		if isUniqueConstraintErr(result.Error) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *permissionGroupRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("now()"))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", id))
	}
	return nil
}

// SetRules はトランザクション内で権限グループの全ルールを置き換える（全削除→再挿入）
func (r *permissionGroupRepository) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存ルールを全削除
		if err := tx.Where("group_id = ?", groupID).Delete(&model.PermissionGroupRule{}).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group_rule", "")
		}

		// 新しいルールを一括作成
		if len(rules) > 0 {
			for i := range rules {
				rules[i].GroupID = groupID
			}
			if err := tx.CreateInBatches(rules, 100).Error; err != nil {
				return apperrors.FromGORM(err, "permission_group_rule", "")
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

// CountStaffsByGroupID は指定グループを参照しているスタッフ数を返す（削除前の依存チェック用）
func (r *permissionGroupRepository) CountStaffsByGroupID(ctx context.Context, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.StaffPermissionGroup{}).
		Where("group_id = ?", groupID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff_permission_group", "")
	}
	return count, nil
}

// GetEffectivePermissionsByStaffID はスタッフが所属する全権限グループのルールを
// UNION (bool_or) して実効権限を返す。
// 戻り値の各要素は resource 毎に集約済み（GroupID=0, ID=0）。
func (r *permissionGroupRepository) GetEffectivePermissionsByStaffID(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
	var rules []model.PermissionGroupRule
	// staff_permission_groups → permission_groups (active & not deleted) → permission_group_rules を JOIN し
	// resource 毎に bool_or で集約する
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT
				pgr.resource,
				bool_or(pgr.can_view)   AS can_view,
				bool_or(pgr.can_create) AS can_create,
				bool_or(pgr.can_edit)   AS can_edit,
				bool_or(pgr.can_delete) AS can_delete
			FROM staff_permission_groups spg
			JOIN permission_groups pg
				ON pg.id = spg.group_id
				AND pg.deleted_at IS NULL
				AND pg.is_active = true
			JOIN permission_group_rules pgr
				ON pgr.group_id = pg.id
			WHERE spg.staff_id = ?
			GROUP BY pgr.resource
		`, staffID).
		Scan(&rules).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "effective_permissions", fmt.Sprintf("staff:%d", staffID))
	}
	return rules, nil
}

// GetGroupIDsByStaffID はスタッフが所属する権限グループIDリストを返す。
func (r *permissionGroupRepository) GetGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error) {
	var rows []struct {
		GroupID uint64
	}
	if err := r.db.WithContext(ctx).
		Model(&model.StaffPermissionGroup{}).
		Select("group_id").
		Where("staff_id = ?", staffID).
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.GroupID)
	}
	return ids, nil
}

// SetStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
func (r *permissionGroupRepository) SetStaffGroups(ctx context.Context, staffID uint64, groupIDs []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存の紐付けを全削除
		if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffPermissionGroup{}).Error; err != nil {
			return apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
		}
		// 新しい紐付けを挿入
		if len(groupIDs) > 0 {
			rows := make([]model.StaffPermissionGroup, 0, len(groupIDs))
			for _, gid := range groupIDs {
				rows = append(rows, model.StaffPermissionGroup{StaffID: staffID, GroupID: gid})
			}
			if err := tx.CreateInBatches(rows, 100).Error; err != nil {
				return apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Reorder は指定されたIDリストの順序でソート順を更新する
func (r *permissionGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.PermissionGroup{}, "permission_group", clinicID, ids)
}
