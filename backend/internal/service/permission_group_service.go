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

type PermissionGroupService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	Create(ctx context.Context, clinicID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	GetEffectivePermissions(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error)
}

type permissionGroupService struct {
	repo repository.PermissionGroupRepository
}

func NewPermissionGroupService(repo repository.PermissionGroupRepository) PermissionGroupService {
	return &permissionGroupService{repo: repo}
}

func (s *permissionGroupService) List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list permission groups")
	}
	return items, nil
}

func (s *permissionGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get permission group")
	}
	return result, nil
}

func (s *permissionGroupService) Create(ctx context.Context, clinicID uint64, input CreatePermissionGroupInput) (*model.PermissionGroup, error) {
	group := &model.PermissionGroup{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		IsActive:    input.IsActive,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, group); err != nil {
		return nil, apperrors.Wrap(err, "failed to create permission group")
	}
	slog.InfoContext(ctx, "permission group created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("group_id", group.ID),
		slog.String("name", group.Name))
	return group, nil
}

func (s *permissionGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
	fields := buildPermissionGroupUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update permission group")
	}
	slog.InfoContext(ctx, "permission group updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("group_id", id))
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get permission group after update")
	}
	return result, nil
}

func (s *permissionGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountStaffsByGroupID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check permission group dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この権限グループはスタッフに割り当てられているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete permission group")
	}
	slog.InfoContext(ctx, "permission group deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("group_id", id))
	return nil
}

func (s *permissionGroupService) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if err := s.repo.SetRules(ctx, groupID, rules); err != nil {
		return apperrors.Wrap(err, "failed to set permission group rules")
	}
	slog.InfoContext(ctx, "permission group rules set",
		slog.Uint64("group_id", groupID),
		slog.Int("rule_count", len(rules)))
	return nil
}

func (s *permissionGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder permission groups")
	}
	slog.InfoContext(ctx, "permission groups reordered",
		slog.Uint64("clinic_id", clinicID),
		slog.Int("count", len(ids)))
	return nil
}

func (s *permissionGroupService) GetEffectivePermissions(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
	rules, err := s.repo.GetEffectivePermissionsByStaffID(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get effective permissions")
	}
	return rules, nil
}

func buildPermissionGroupUpdateFields(input *UpdatePermissionGroupInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Color != nil {
		fields["color"] = *input.Color
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}
