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
	FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, group *model.PermissionGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	// FindAllEffectivePermissionsByStaffID はスタッフが所属する全権限グループのルールを
	// UNION (bool_or) して実効権限を返す（clinicID スコープ付き）。
	FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
	// FindAllGroupIDsByStaffID はスタッフが所属する権限グループIDリストを返す。
	FindAllGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error)
	// UpdateStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
	UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
}

type permissionGroupRepository struct{ db *gorm.DB }

func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	groups := make([]model.PermissionGroup, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Preload("Rules", "deleted_at IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&groups).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", "")
	}
	return groups, nil
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
	}
	return &group, nil
}

func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	err := r.db.WithContext(ctx).Create(group).Error
	if err != nil {
		return apperrors.FromGORM(err, "permission_group", "")
	}
	return nil
}

func (r *permissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error) {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *permissionGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.PermissionGroup{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", id))
	}
	return nil
}

// UpdateRules はトランザクション内で権限グループの全ルールを置き換える（全削除→再挿入）
func (r *permissionGroupRepository) UpdateRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存ルールを全削除（物理削除してユニーク制約重複を防止）
		if err := tx.Unscoped().Where("group_id = ?", groupID).Delete(&model.PermissionGroupRule{}).Error; err != nil {
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
		return apperrors.Wrap(err, "failed to set permission group rules")
	}
	return nil
}

// CountUsageByGroupID は指定グループを参照しているスタッフ数を返す（削除前の依存チェック用）
// permission_groups テーブルが clinic_id を持つため JOIN でテナント分離を行う
func (r *permissionGroupRepository) CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.StaffPermissionGroup{}).
		Joins("JOIN permission_groups ON permission_groups.id = staff_permission_groups.group_id AND permission_groups.clinic_id = ? AND permission_groups.deleted_at IS NULL", clinicID).
		Joins("JOIN staffs ON staffs.id = staff_permission_groups.staff_id AND staffs.deleted_at IS NULL").
		Where("staff_permission_groups.group_id = ?", groupID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff_permission_group", "")
	}
	return count, nil
}

// FindAllEffectivePermissionsByStaffID はスタッフが所属する全権限グループのルールを
// UNION (bool_or) して実効権限を返す。
// 戻り値の各要素は resource 毎に集約済み（GroupID=0, ID=0）。
// clinicID パラメータで検索範囲を制限し、マルチクリニック昇格を防止（High-7）。
func (r *permissionGroupRepository) FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
	var rules []model.PermissionGroupRule
	// staff_permission_groups → permission_groups (active & not deleted) → permission_group_rules を JOIN し
	// resource 毎に bool_or で集約する。
	// pg.clinic_id = ? で結果を指定クリニックに限定し、所属外クリニックの権限が混入するのを防止。
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
				AND pg.clinic_id = ?
			JOIN permission_group_rules pgr
				ON pgr.group_id = pg.id
			WHERE spg.staff_id = ?
			GROUP BY pgr.resource
		`, clinicID, staffID).
		Scan(&rules).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "effective_permissions", fmt.Sprintf("staff:%d clinic:%d", staffID, clinicID))
	}
	return rules, nil
}

// FindAllGroupIDsByStaffID はスタッフが所属する権限グループIDリストを返す。
func (r *permissionGroupRepository) FindAllGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error) {
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

// UpdateStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
func (r *permissionGroupRepository) UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	// テナント越境 write 防止: 紐付け対象の権限グループIDが呼び出し元クリニックに属することを検証する。
	// staff_permission_groups は自前 clinic_id を持たないため、ここで group_id の所有権を
	// 確認しなければ別クリニックの権限グループIDを紐付けできてしまう（横断テナント書き換え）。
	// スタッフ所有権は呼び出し側（handler verifyStaffClinicMembership）で検証済み。検証は DELETE 前に行い部分書き込みを防ぐ。
	if len(groupIDs) > 0 {
		var count int64
		if err := r.db.WithContext(ctx).
			Model(&model.PermissionGroup{}).
			Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, groupIDs).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group", "")
		}
		if count != int64(len(groupIDs)) {
			return apperrors.WrapInvalidInput("group_ids contains invalid permission group")
		}
	}
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
		return apperrors.Wrap(err, "failed to replace staff permission groups")
	}
	return nil
}

// Reorder は指定されたIDリストの順序でソート順を更新する。
// GORM の論理削除は Model 呼び出しで自動適用されないため、明示的に deleted_at IS NULL を指定する。
func (r *permissionGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.PermissionGroup{}).
				Scopes(clinicScope(clinicID)).Where("id = ? AND deleted_at IS NULL", id).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("permission_group id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder permission groups")
	}
	return nil
}
