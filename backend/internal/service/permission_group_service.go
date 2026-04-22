// Package service provides business logic implementations for PermissionGroup entity.
package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- PermissionGroupService ----

// CreatePermissionGroupInput は権限グループ作成のための入力データ
type CreatePermissionGroupInput struct {
	Name        string
	Description string
	Color       string
	IsActive    bool
	SortOrder   int
}

// UpdatePermissionGroupInput holds the fields that can be updated via PATCH.
// All fields are pointers: nil means "not provided / skip".
type UpdatePermissionGroupInput struct {
	Name        *string
	Description *string
	Color       *string
	SortOrder   *int
	IsActive    *bool
}

// SetPermissionGroupRulesInput は権限グループのルール設定のための入力データ
type SetPermissionGroupRulesInput struct {
	Resource  string
	CanView   bool
	CanCreate bool
	CanEdit   bool
	CanDelete bool
}

const (
	colPermissionGroupName        = "name"
	colPermissionGroupDescription = "description"
	colPermissionGroupColor       = "color"
	colPermissionGroupSortOrder   = "sort_order"
	colPermissionGroupIsActive    = "is_active"
)

func buildPermissionGroupUpdate(input *UpdatePermissionGroupInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colPermissionGroupName] = *input.Name
	}
	if input.Description != nil {
		fields[colPermissionGroupDescription] = *input.Description
	}
	if input.Color != nil {
		fields[colPermissionGroupColor] = *input.Color
	}
	if input.SortOrder != nil {
		fields[colPermissionGroupSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colPermissionGroupIsActive] = *input.IsActive
	}
	return fields
}

type PermissionGroupService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, clinicID uint64, input *CreatePermissionGroupInput) (*model.PermissionGroup, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// SetRules はグループのルールを全置換する。actorStaffID は自己参照チェックに使用される。
	SetRules(ctx context.Context, groupID uint64, inputs []SetPermissionGroupRulesInput, actorStaffID uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// EffectivePermissionService は有効権限取得の責務を持つ独立インターフェース。
// CRUD 操作の PermissionGroupService とは認可責務が異なるため分離する。
type EffectivePermissionService interface {
	GetEffectivePermissions(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error)
}

type permissionGroupService struct {
	repo repository.PermissionGroupRepository
}

func NewPermissionGroupService(repo repository.PermissionGroupRepository) PermissionGroupService {
	return &permissionGroupService{repo: repo}
}

// newPermissionGroupServiceImpl は PermissionGroupService と EffectivePermissionService の
// 両インターフェースを実装する具体型ポインタを返す。
// service.go の DI 配線のみで使用する。
func newPermissionGroupServiceImpl(repo repository.PermissionGroupRepository) *permissionGroupService {
	return &permissionGroupService{repo: repo}
}

func (s *permissionGroupService) List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list permission groups", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list permission groups")
	}
	return items, nil
}

func (s *permissionGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get permission group", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get permission group")
	}
	return result, nil
}

func (s *permissionGroupService) Create(ctx context.Context, clinicID uint64, input *CreatePermissionGroupInput) (*model.PermissionGroup, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	group := &model.PermissionGroup{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		IsActive:    input.IsActive,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		slog.ErrorContext(ctx, "failed to create permission group", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create permission group")
	}
	slog.InfoContext(ctx, "permission group created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", group.ID),
		slog.String("name", group.Name))
	return group, nil
}

func (s *permissionGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get permission group", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get permission group")
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := buildPermissionGroupUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}
	result, err := s.repo.Update(ctx, clinicID, id, fields)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update permission group", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update permission group")
	}
	slog.InfoContext(ctx, "permission group updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", id))
	return result, nil
}

func (s *permissionGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get permission group")
	}
	count, err := s.repo.CountUsageByGroupID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check permission group dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check permission group dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この権限グループはスタッフに割り当てられているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete permission group", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete permission group")
	}
	slog.InfoContext(ctx, "permission group deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", id))
	return nil
}

func (s *permissionGroupService) SetRules(ctx context.Context, groupID uint64, inputs []SetPermissionGroupRulesInput, actorStaffID uint64) error {
	// Input DTO を model.PermissionGroupRule に変換
	rules := make([]model.PermissionGroupRule, 0, len(inputs))
	for _, inp := range inputs {
		rules = append(rules, model.PermissionGroupRule{
			Resource:  inp.Resource,
			CanView:   inp.CanView,
			CanCreate: inp.CanCreate,
			CanEdit:   inp.CanEdit,
			CanDelete: inp.CanDelete,
		})
	}
	// BUG-146: 入力バリデーション — 空文字・存在しないリソース名・重複を拒否
	if err := validateNoDuplicateRules(rules); err != nil {
		return err
	}
	// BUG-140: 自分が所属するグループの master-permission edit を削除できないようにする
	// staffGroupIDs をサービス内で取得する（Handler が外部データを取得する責務を持たない）
	staffGroupIDs, err := s.repo.FindGroupIDsByStaffID(ctx, actorStaffID)
	if err != nil {
		// エラー時は空にして自己参照チェック不能なら許可方向（ベストエフォート、Handler層がエラーハンドリング）
		staffGroupIDs = []uint64{}
	}
	if err := validateNotSelfReference(groupID, rules, staffGroupIDs); err != nil {
		return err
	}
	if err := s.repo.SetRules(ctx, groupID, rules); err != nil {
		slog.ErrorContext(ctx, "failed to set permission group rules", "error", err, "id", groupID)
		return apperrors.Wrap(err, "failed to set permission group rules")
	}
	slog.InfoContext(ctx, "permission group rules set",
		slog.Uint64("permission_group_id", groupID),
		slog.Int("rule_count", len(rules)))
	return nil
}

// validateNoDuplicateRules は空文字・存在しないリソース名・重複を検証する（BUG-146）
func validateNoDuplicateRules(rules []model.PermissionGroupRule) error {
	seen := make(map[string]bool, len(rules))
	for _, r := range rules {
		if r.Resource == "" {
			return apperrors.WrapInvalidInput(ErrMsgResourceNameEmpty)
		}
		if !model.IsValidResource(r.Resource) {
			return apperrors.WrapInvalidInput("無効なリソース名: " + r.Resource)
		}
		if seen[r.Resource] {
			return apperrors.WrapInvalidInput("リソース名が重複しています: " + r.Resource)
		}
		seen[r.Resource] = true
	}
	return nil
}

// validateNotSelfReference は自分が所属するグループの master-permission edit を削除しないことを検証する（BUG-140）
func validateNotSelfReference(groupID uint64, rules []model.PermissionGroupRule, staffGroupIDs []uint64) error {
	isSelfGroup := false
	for _, gid := range staffGroupIDs {
		if gid == groupID {
			isSelfGroup = true
			break
		}
	}
	if !isSelfGroup {
		return nil
	}
	hasMasterPermEdit := false
	for _, r := range rules {
		if r.Resource == string(model.ResourceMasterPermission) && r.CanEdit {
			hasMasterPermEdit = true
			break
		}
	}
	if !hasMasterPermEdit {
		return apperrors.WrapInvalidInput("自分が所属するグループの権限管理権限（master-permission edit）を削除することはできません")
	}
	return nil
}

func (s *permissionGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder permission groups", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder permission groups")
	}
	slog.InfoContext(ctx, "permission groups reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

func (s *permissionGroupService) GetEffectivePermissions(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
	rules, err := s.repo.FindEffectivePermissionsByStaffID(ctx, staffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get effective permissions", "error", err, "id", staffID)
		return nil, apperrors.Wrap(err, "failed to get effective permissions")
	}
	return rules, nil
}
