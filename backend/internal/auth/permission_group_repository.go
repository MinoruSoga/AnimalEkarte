package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// PermissionGroupRepository provides clinic-scoped permission-group persistence.
type PermissionGroupRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, group *model.PermissionGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	DeleteSoftDeletedByClinicID(ctx context.Context, clinicID uint64) error
	UpdateRules(ctx context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error
	CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	// FindAllEffectivePermissionsByStaffID はスタッフが所属する全権限グループのルールを
	// UNION (bool_or) して実効権限を返す（clinicID スコープ付き）。
	FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
	// FindAllGroupIDsByStaffID は認証済みクリニック内でスタッフが所属する権限グループIDを返す。
	FindAllGroupIDsByStaffID(ctx context.Context, clinicID, staffID uint64) ([]uint64, error)
	// UpdateStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
	UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
}

// PermissionGroupRulesAtomicWriter persists a permission-group parent and its
// complete rule set as one atomic aggregate write.
type PermissionGroupRulesAtomicWriter interface {
	CreateWithRules(
		ctx context.Context,
		group *model.PermissionGroup,
		rules []model.PermissionGroupRule,
	) (*model.PermissionGroup, error)
	UpdateWithRules(
		ctx context.Context,
		clinicID, id uint64,
		fields map[string]any,
		rules []model.PermissionGroupRule,
	) (*model.PermissionGroup, error)
}

type permissionGroupRepository struct{ db *gorm.DB }

// NewPermissionGroupRepository constructs clinic-scoped permission-group persistence.
func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	groups := make([]model.PermissionGroup, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
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
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Rules", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
	}
	return &group, nil
}

// LockByIDForUpdate serializes authorization-policy writers before they capture
// the old audit snapshot. The caller must provide an ambient transaction.
func (r *permissionGroupRepository) LockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"permission group mutation lock requires an ambient transaction",
		)
	}
	var group model.PermissionGroup
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Rules", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&group).
		Error
	if err != nil {
		return nil, apperrors.FromGORM(
			err,
			"permission_group",
			fmt.Sprintf("%d", id),
		)
	}
	return &group, nil
}

// BE-refactor.md X-7: dbOrTx で ambient tx に参加する。clinic_service.CreateClinic は
// clinic 作成 + デフォルト権限グループ2件の作成を transactor.WithTx で包むが、Create が
// r.db.WithContext(ctx) のまま tx 非参加だと、2件目の作成が失敗しても1件目は既に
// オートコミット済みで WithTx のロールバックが効かず、デフォルト権限グループが片方だけの
// 孤児クリニックが生成しうるバグがあった。
func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from
	// INSERT and may write the DB default back into the struct.
	wantActive := group.IsActive
	if err := db.Create(group).Error; err != nil {
		return apperrors.FromGORM(err, "permission_group", "")
	}
	if !wantActive {
		if err := db.Model(group).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", group.ID))
		}
		group.IsActive = false
	}
	return nil
}

func (r *permissionGroupRepository) CreateWithRules(
	ctx context.Context,
	group *model.PermissionGroup,
	rules []model.PermissionGroupRule,
) (*model.PermissionGroup, error) {
	var result *model.PermissionGroup
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := r.Create(txCtx, group); err != nil {
			return err
		}
		if err := r.replaceRules(txCtx, group.ID, rules); err != nil {
			return err
		}
		var readErr error
		result, readErr = r.FindByID(txCtx, group.ClinicID, group.ID)
		return readErr
	}); err != nil {
		return nil, apperrors.Wrap(
			err,
			"failed to create permission group with rules",
		)
	}
	return result, nil
}

func (r *permissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error) {
	if err := persistence.UpdateScopedByID(
		ctx,
		persistence.DBOrTx(ctx, r.db),
		&model.PermissionGroup{},
		"permission_group",
		clinicID,
		id,
		fields,
	); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *permissionGroupRepository) UpdateWithRules(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
	rules []model.PermissionGroupRule,
) (*model.PermissionGroup, error) {
	var result *model.PermissionGroup
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		var group model.PermissionGroup
		if err := persistence.DBOrTx(txCtx, r.db).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"id = ? AND clinic_id = ? AND deleted_at IS NULL",
				id,
				clinicID,
			).
			First(&group).
			Error; err != nil {
			return apperrors.FromGORM(
				err,
				"permission_group",
				fmt.Sprintf("%d", id),
			)
		}
		if len(fields) > 0 {
			if err := persistence.UpdateScopedByID(
				txCtx,
				persistence.DBOrTx(txCtx, r.db),
				&model.PermissionGroup{},
				"permission_group",
				clinicID,
				id,
				fields,
			); err != nil {
				return err
			}
		}
		if err := r.replaceRules(txCtx, id, rules); err != nil {
			return err
		}
		var readErr error
		result, readErr = r.FindByID(txCtx, clinicID, id)
		return readErr
	}); err != nil {
		return nil, apperrors.Wrap(
			err,
			"failed to update permission group with rules",
		)
	}
	return result, nil
}

func (r *permissionGroupRepository) replaceRules(
	ctx context.Context,
	groupID uint64,
	rules []model.PermissionGroupRule,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Unscoped().
		Where("group_id = ?", groupID).
		Delete(&model.PermissionGroupRule{}).
		Error; err != nil {
		return apperrors.FromGORM(err, "permission_group_rule", "")
	}
	if len(rules) == 0 {
		return nil
	}
	persistedRules := make([]model.PermissionGroupRule, len(rules))
	copy(persistedRules, rules)
	for i := range persistedRules {
		// Force explicit bool columns on INSERT. GORM can omit zero-value
		// booleans when a column has a default tag, which historically made
		// "uncheck" (true→false) look like a successful no-op after reload.
		persistedRules[i].ID = 0
		persistedRules[i].GroupID = groupID
	}
	if err := db.
		Select(
			"GroupID",
			"Resource",
			"CanView",
			"CanCreate",
			"CanEdit",
			"CanDelete",
		).
		CreateInBatches(persistedRules, 100).Error; err != nil {
		return apperrors.FromGORM(err, "permission_group_rule", "")
	}
	return nil
}

func (r *permissionGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&model.PermissionGroup{}).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
		}
		result := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Where(`NOT EXISTS (
				SELECT 1 FROM staff_permission_groups
				JOIN staffs ON staffs.id = staff_permission_groups.staff_id
				  AND staffs.deleted_at IS NULL
				WHERE staff_permission_groups.group_id = permission_groups.id
			)`).
			Delete(&model.PermissionGroup{})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return r.normalizePermissionGroupDeleteMiss(persistence.WithTxValue(ctx, tx), clinicID, id)
		}
		return nil
	})
}

func (r *permissionGroupRepository) normalizePermissionGroupDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("この権限グループはスタッフに割り当てられているため削除できません")
}

// DeleteSoftDeletedByClinicID hard-deletes only rows already soft-deleted in the
// target clinic. Zero matching rows is a successful idempotent cleanup.
func (r *permissionGroupRepository) DeleteSoftDeletedByClinicID(ctx context.Context, clinicID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Unscoped().
		Where("clinic_id = ? AND deleted_at IS NOT NULL", clinicID).
		Delete(&model.PermissionGroup{})
	if result.Error != nil {
		return apperrors.FromGORM(
			result.Error,
			"permission_group",
			fmt.Sprintf("clinic_id=%d", clinicID),
		)
	}
	return nil
}

// UpdateRules はトランザクション内で権限グループの全ルールを置き換える（全削除→再挿入）。
// 親グループを clinic_id + active row で FOR UPDATE 取得してから子行を変更し、
// 越境IDと並行する親削除を同じ write transaction 内で拒否・直列化する。
// BE-refactor.md X-7: dbOrTx(ctx, r.db).Transaction(...) にすることで ambient tx があれば
// SAVEPOINT として参加する（Create と同じ tx 参加方針、R1-1 と同一パターン）。
func (r *permissionGroupRepository) UpdateRules(
	ctx context.Context,
	clinicID, groupID uint64,
	rules []model.PermissionGroupRule,
) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var group model.PermissionGroup
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", groupID, clinicID).
			First(&group).Error; err != nil {
			return apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", groupID))
		}

		txCtx := persistence.WithTxValue(ctx, tx)
		return r.replaceRules(txCtx, groupID, rules)
	}); err != nil {
		return apperrors.Wrap(err, "failed to set permission group rules")
	}
	return nil
}

// CountUsageByGroupID は指定グループを参照しているスタッフ数を返す（削除前の依存チェック用）
// permission_groups テーブルが clinic_id を持つため JOIN でテナント分離を行う
func (r *permissionGroupRepository) CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
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
	err := persistence.DBOrTx(ctx, r.db).
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
				AND pgr.deleted_at IS NULL
			WHERE spg.staff_id = ?
			GROUP BY pgr.resource
		`, clinicID, staffID).
		Scan(&rules).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "effective_permissions", fmt.Sprintf("staff:%d clinic:%d", staffID, clinicID))
	}
	return rules, nil
}

// FindAllGroupIDsByStaffID は認証済みクリニック内でスタッフが所属する権限グループIDを返す。
// staff_permission_groups は clinic_id を持たないため、permission_groups を経由して
// clinic_id と deleted_at の両方を制約し、他院または削除済みグループのID漏洩を防ぐ。
func (r *permissionGroupRepository) FindAllGroupIDsByStaffID(ctx context.Context, clinicID, staffID uint64) ([]uint64, error) {
	var rows []struct {
		GroupID uint64
	}
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.StaffPermissionGroup{}).
		Select("staff_permission_groups.group_id").
		Joins(`
			JOIN permission_groups
				ON permission_groups.id = staff_permission_groups.group_id
				AND permission_groups.clinic_id = ?
				AND permission_groups.deleted_at IS NULL
		`, clinicID).
		Where("staff_permission_groups.staff_id = ?", staffID).
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d clinic:%d", staffID, clinicID))
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.GroupID)
	}
	return ids, nil
}

// UpdateStaffGroups はスタッフの権限グループを全置換する（DELETE + INSERT）。
// staff_clinic_assignments の active 行を同じ transaction 内で FOR UPDATE 取得し、
// 所属解除との TOCTOU と並行する権限全置換を直列化する。
func (r *permissionGroupRepository) UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	db := persistence.DBOrTx(ctx, r.db)
	// BE-refactor.md X-7: dbOrTx(ctx, r.db).Transaction(...) で ambient tx があれば SAVEPOINT として参加する。
	return persistence.ReplaceJunctionInTransaction(db, func(tx *gorm.DB) error {
		var assignment model.StaffClinicAssignment
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"staff_id = ? AND clinic_id = ? AND deleted_at IS NULL",
				staffID,
				clinicID,
			).
			First(&assignment).Error; err != nil {
			return apperrors.FromGORM(
				err,
				"staff_clinic_assignment",
				fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
			)
		}

		// テナント越境 write 防止: 紐付け対象の権限グループIDも置換と同じ
		// transaction 内で検証し、失敗時は既存リンクを保持する。
		if err := persistence.ValidateClinicScopedMasterIDs(
			ctx,
			tx,
			clinicID,
			groupIDs,
			&model.PermissionGroup{},
			"permission_group",
			"group_ids contains invalid permission group",
		); err != nil {
			return err
		}

		// 既存の紐付けを全削除（BE-refactor.md H-1: staff_permission_groups は自前 clinic_id を
		// 持たないため、DELETE を staff_id のみでスコープすると、多施設所属スタッフ（Staff は
		// staff_clinic_assignments で複数クリニックに所属しうる）の場合、clinicID の属さない
		// 他クリニック分の紐付けまで無警告で消えてしまう。group 側の clinic_id サブクエリで
		// clinicID に属する紐付けのみを削除対象にスコープする）。
		if err := persistence.DeleteJunctionViaMasterClinicScope(tx, clinicID, staffID,
			&model.StaffPermissionGroup{}, &model.PermissionGroup{}, "group_id",
			"staff_permission_group", fmt.Sprintf("staff:%d", staffID)); err != nil {
			return err
		}
		if len(groupIDs) == 0 {
			return nil
		}
		rows := make([]model.StaffPermissionGroup, 0, len(groupIDs))
		for _, gid := range groupIDs {
			rows = append(rows, model.StaffPermissionGroup{StaffID: staffID, GroupID: gid})
		}
		return persistence.InsertJunctionRowsInBatches(tx, rows, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
	}, "failed to replace staff permission groups")
}

// Reorder は指定されたIDリストの順序でソート順を更新する。
// GORM の論理削除は Model 呼び出しで自動適用されないため、明示的に deleted_at IS NULL を指定する。
// BE-refactor.md X-7: dbOrTx(ctx, r.db).Transaction(...) で ambient tx があれば SAVEPOINT として参加する。
func (r *permissionGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.PermissionGroup{}).
				Scopes(persistence.ClinicScope(clinicID)).Where("id = ? AND deleted_at IS NULL", id).
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
