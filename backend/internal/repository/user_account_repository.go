package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UserAccountRepository はユーザーアカウントのデータアクセス層
type UserAccountRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.UserAccount, error)
	FindActiveByID(ctx context.Context, id uint64) (*model.UserAccount, error)
	FindByIDWithMemberships(ctx context.Context, id uint64) (*UserAccountWithMemberships, error)
	FindByClinicID(ctx context.Context, clinicID uint64) ([]model.UserAccount, error)
	Create(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	SetPermissionGroups(ctx context.Context, userID uint64, groupIDs []uint64) error
	FindPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error)
}

// EffectivePermissionRow は実効権限計算用のクエリ結果行（company単位）
type EffectivePermissionRow struct {
	Resource  string
	CanView   bool
	CanCreate bool
	CanEdit   bool
	CanDelete bool
}

// UserAccountWithMemberships はユーザー・所属クリニック・実効権限をまとめたクエリ結果
type UserAccountWithMemberships struct {
	model.UserAccount
	Memberships       []model.UserClinicMembership
	EffectivePermRows []EffectivePermissionRow
}

type userAccountRepository struct {
	db *gorm.DB
}

// NewUserAccountRepository はUserAccountRepositoryを初期化して返す
func NewUserAccountRepository(db *gorm.DB) UserAccountRepository {
	return &userAccountRepository{db: db}
}

// FindByEmail はメールアドレスでユーザーアカウントを取得する。
// アカウントが存在しない場合は ErrNotFound を返す。
func (r *userAccountRepository) FindByEmail(ctx context.Context, email string) (*model.UserAccount, error) {
	var account model.UserAccount
	err := r.db.WithContext(ctx).First(&account, "email = ? AND deleted_at IS NULL", email).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "user_account", email)
	}
	return &account, nil
}

// FindActiveByID はアクティブなユーザーをIDで取得する（アカウント停止・論理削除を除外）
func (r *userAccountRepository) FindActiveByID(ctx context.Context, id uint64) (*model.UserAccount, error) {
	var account model.UserAccount
	err := r.db.WithContext(ctx).
		First(&account, "id = ? AND status = 'active' AND deleted_at IS NULL", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "user_account", fmt.Sprintf("%d", id))
	}
	return &account, nil
}

// FindByClinicID はクリニックIDに紐づくユーザーアカウント一覧を返す。
func (r *userAccountRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.UserAccount, error) {
	var accounts []model.UserAccount
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_clinic_memberships ON user_clinic_memberships.user_id = user_accounts.id").
		Where("user_clinic_memberships.clinic_id = ? AND user_accounts.deleted_at IS NULL", clinicID).
		Find(&accounts).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user accounts by clinic id")
	}
	return accounts, nil
}

// Create はユーザーアカウントとクリニック所属を作成する。
func (r *userAccountRepository) Create(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error {
	account.StaffID = staffID
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create user account")
		}
		membership := &model.UserClinicMembership{
			UserID:   account.ID,
			ClinicID: clinicID,
			IsMain:   isMain,
		}
		if err := tx.Create(membership).Error; err != nil {
			return apperrors.Wrap(err, "create user clinic membership")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("create user account: %w", err)
	}
	return nil
}

// Update はユーザーアカウントを部分更新する。
func (r *userAccountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if err := r.db.WithContext(ctx).Model(&model.UserAccount{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error; err != nil {
		return apperrors.Wrap(err, "update user account")
	}
	return nil
}

// Delete はユーザーアカウントを論理削除する。
func (r *userAccountRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&model.UserAccount{}).Error; err != nil {
		return apperrors.Wrap(err, "delete user account")
	}
	return nil
}

// FindByIDWithMemberships はIDでユーザーを取得し、所属クリニックも合わせて返す。
func (r *userAccountRepository) FindByIDWithMemberships(ctx context.Context, id uint64) (*UserAccountWithMemberships, error) {
	var account model.UserAccount
	err := r.db.WithContext(ctx).
		Preload("Staff").
		Preload("JobTitle").
		First(&account, "id = ? AND deleted_at IS NULL", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "user_account", fmt.Sprintf("%d", id))
	}

	memberships := make([]model.UserClinicMembership, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Find(&memberships).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user clinic memberships")
	}

	effectivePerms, err := r.findEffectivePermissions(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "find effective permissions")
	}

	return &UserAccountWithMemberships{
		UserAccount:       account,
		Memberships:       memberships,
		EffectivePermRows: effectivePerms,
	}, nil
}

// SetPermissionGroups はユーザーのグループ割当を全置換する（トランザクション）
func (r *userAccountRepository) SetPermissionGroups(ctx context.Context, userID uint64, groupIDs []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserPermissionGroup{}).Error; err != nil {
			return err
		}
		if len(groupIDs) == 0 {
			return nil
		}
		rows := make([]model.UserPermissionGroup, len(groupIDs))
		for i, gid := range groupIDs {
			rows[i] = model.UserPermissionGroup{UserID: userID, GroupID: gid}
		}
		return tx.Create(&rows).Error
	}); err != nil {
		return fmt.Errorf("set user permission groups: %w", err)
	}
	return nil
}

// FindPermissionGroupIDs はユーザーに割り当てられているグループIDリストを返す
func (r *userAccountRepository) FindPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var rows []model.UserPermissionGroup
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("find user permission group ids: %w", err)
	}
	ids := make([]uint64, len(rows))
	for i, row := range rows {
		ids[i] = row.GroupID
	}
	return ids, nil
}

// findEffectivePermissions はユーザーの実効権限を取得する（company単位グループUNION計算）
func (r *userAccountRepository) findEffectivePermissions(ctx context.Context, userID uint64) ([]EffectivePermissionRow, error) {
	var rows []EffectivePermissionRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			pgr.resource,
			bool_or(pgr.can_view)   AS can_view,
			bool_or(pgr.can_create) AS can_create,
			bool_or(pgr.can_edit)   AS can_edit,
			bool_or(pgr.can_delete) AS can_delete
		FROM user_permission_groups upg
		JOIN permission_groups pg ON pg.id = upg.group_id AND pg.deleted_at IS NULL
		JOIN permission_group_rules pgr ON pgr.group_id = pg.id
		WHERE upg.user_id = ?
		GROUP BY pgr.resource
	`, userID).Scan(&rows).Error
	return rows, err
}
