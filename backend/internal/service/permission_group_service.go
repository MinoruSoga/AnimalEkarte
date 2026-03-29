package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreatePermissionGroupInput はグループ作成の入力値
type CreatePermissionGroupInput struct {
	Name        string
	Description string
	Color       string
}

// UpdatePermissionGroupInput はグループ更新の入力値（ポインタ型でゼロ値問題を回避）
type UpdatePermissionGroupInput struct {
	Name        *string
	Description *string
	Color       *string
}

// SetPermissionGroupRulesInput はルール一括更新の入力値
type SetPermissionGroupRulesInput struct {
	Rules []RuleInput
}

// RuleInput は個別ルールの入力値
type RuleInput struct {
	Resource  model.Resource
	CanView   bool
	CanCreate bool
	CanEdit   bool
	CanDelete bool
}

// PermissionGroupService は権限グループのビジネスロジックインターフェース
type PermissionGroupService interface {
	List(ctx context.Context, companyID uint64) ([]model.PermissionGroup, error)
	GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, companyID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error)
	Update(ctx context.Context, id uint64, input UpdatePermissionGroupInput) error
	Delete(ctx context.Context, id uint64) error
	SetRules(ctx context.Context, groupID uint64, input SetPermissionGroupRulesInput) error
}

type permissionGroupService struct {
	repo repository.PermissionGroupRepository
}

// NewPermissionGroupService はPermissionGroupServiceを初期化して返す
func NewPermissionGroupService(repo repository.PermissionGroupRepository) PermissionGroupService {
	return &permissionGroupService{repo: repo}
}

func (s *permissionGroupService) List(ctx context.Context, companyID uint64) ([]model.PermissionGroup, error) {
	return s.repo.FindByCompanyID(ctx, companyID)
}

func (s *permissionGroupService) GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission group not found: %w", apperrors.ErrNotFound)
	}
	return g, nil
}

func (s *permissionGroupService) Create(ctx context.Context, companyID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error) {
	color := input.Color
	if color == "" {
		color = "#6B7280"
	}
	group := &model.PermissionGroup{
		CompanyID:   companyID,
		Name:        input.Name,
		Description: input.Description,
		Color:       color,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		if apperrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("permission group name already exists: %w", apperrors.ErrAlreadyExists)
		}
		return nil, fmt.Errorf("failed to create permission group: %w", err)
	}
	slog.InfoContext(ctx, "permission group created", "id", group.ID, "company_id", companyID, "name", group.Name)
	return group, nil
}

func (s *permissionGroupService) Update(ctx context.Context, id uint64, input UpdatePermissionGroupInput) error {
	fields := buildPermissionGroupUpdateFields(input)
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *permissionGroupService) Delete(ctx context.Context, id uint64) error {
	slog.InfoContext(ctx, "deleting permission group", "id", id)
	return s.repo.Delete(ctx, id)
}

func (s *permissionGroupService) SetRules(ctx context.Context, groupID uint64, input SetPermissionGroupRulesInput) error {
	rules := make([]model.PermissionGroupRule, len(input.Rules))
	for i, r := range input.Rules {
		rules[i] = model.PermissionGroupRule{
			GroupID:   groupID,
			Resource:  r.Resource,
			CanView:   r.CanView,
			CanCreate: r.CanCreate,
			CanEdit:   r.CanEdit,
			CanDelete: r.CanDelete,
		}
	}
	slog.InfoContext(ctx, "setting permission group rules", "group_id", groupID, "rule_count", len(rules))
	return s.repo.SetRules(ctx, groupID, rules)
}

func buildPermissionGroupUpdateFields(input UpdatePermissionGroupInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	return fields
}
