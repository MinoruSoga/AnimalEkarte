package auth

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *permissionGroupService) UpdateRules(
	ctx context.Context,
	clinicID, groupID uint64,
	inputs []SetPermissionGroupRulesInput,
	actorStaffID uint64,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	if actorStaffID == 0 || actorStaffID != audit.ActorStaffID {
		return nil, apperrors.WrapInternalServerError(
			"permission mutation actor is invalid",
		)
	}
	if err := s.validateAuditedMutation(
		clinicID,
		audit,
		model.AuditActionPermissionRulesUpdate,
		"permission_group_rules",
	); err != nil {
		return nil, err
	}
	var result *model.PermissionGroup
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		oldGroup, oldErr := s.lockByIDForUpdate(
			txCtx,
			clinicID,
			groupID,
		)
		if oldErr != nil {
			return oldErr
		}
		if updateErr := s.updateRules(
			txCtx,
			clinicID,
			groupID,
			inputs,
			actorStaffID,
		); updateErr != nil {
			return updateErr
		}
		var readErr error
		result, readErr = s.GetByID(txCtx, clinicID, groupID)
		if readErr != nil {
			return readErr
		}
		entry := permissionAuditEntry(
			audit,
			groupID,
			map[string]any{
				"rules": permissionRulesAuditSnapshot(oldGroup.Rules),
			},
			map[string]any{
				"rules": permissionRulesAuditSnapshot(result.Rules),
			},
		)
		if auditErr := s.audit.LogEntryTx(txCtx, entry); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to audit permission group rules update",
			)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionGroupService) updateRules(
	ctx context.Context,
	clinicID, groupID uint64,
	inputs []SetPermissionGroupRulesInput,
	actorStaffID uint64,
) error {
	rules := permissionRuleModels(inputs)
	if err := validateNoDuplicateRules(rules); err != nil {
		return err
	}
	staffGroupIDs, err := s.repo.FindAllGroupIDsByStaffID(ctx, clinicID, actorStaffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find staff group IDs for self-reference check",
			"error", err, "clinic_id", clinicID, "actor_staff_id", actorStaffID)
		return apperrors.Wrap(err, "failed to find staff group IDs")
	}
	if err := validateNotSelfReference(groupID, rules, staffGroupIDs); err != nil {
		return err
	}
	if err := s.repo.UpdateRules(ctx, clinicID, groupID, rules); err != nil {
		slog.ErrorContext(ctx, "failed to set permission group rules", "error", err, "id", groupID)
		return apperrors.Wrap(err, "failed to set permission group rules")
	}
	slog.InfoContext(ctx, "permission group rules set",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", groupID),
		slog.Int("rule_count", len(rules)))
	return nil
}

func permissionRuleModels(
	inputs []SetPermissionGroupRulesInput,
) []model.PermissionGroupRule {
	rules := make([]model.PermissionGroupRule, 0, len(inputs))
	for _, input := range inputs {
		rules = append(rules, model.PermissionGroupRule{
			Resource:  input.Resource,
			CanView:   input.CanView,
			CanCreate: input.CanCreate,
			CanEdit:   input.CanEdit,
			CanDelete: input.CanDelete,
		})
	}
	return rules
}

func validateNoDuplicateRules(rules []model.PermissionGroupRule) error {
	seen := make(map[string]bool, len(rules))
	for i := range rules {
		rule := &rules[i]
		if rule.Resource == "" {
			return apperrors.WrapInvalidInput(errMsgResourceNameEmpty)
		}
		if !model.IsValidResource(rule.Resource) {
			return apperrors.WrapInvalidInput("無効なリソース名: " + rule.Resource)
		}
		if seen[rule.Resource] {
			return apperrors.WrapInvalidInput("リソース名が重複しています: " + rule.Resource)
		}
		seen[rule.Resource] = true
	}
	return nil
}

func validateNotSelfReference(
	groupID uint64,
	rules []model.PermissionGroupRule,
	staffGroupIDs []uint64,
) error {
	isSelfGroup := false
	for _, candidateID := range staffGroupIDs {
		if candidateID == groupID {
			isSelfGroup = true
			break
		}
	}
	if !isSelfGroup {
		return nil
	}
	hasMasterPermissionEdit := false
	for i := range rules {
		rule := &rules[i]
		if rule.Resource == string(model.ResourceMasterPermission) && rule.CanEdit {
			hasMasterPermissionEdit = true
			break
		}
	}
	if !hasMasterPermissionEdit {
		return apperrors.WrapInvalidInput("自分が所属するグループの権限管理権限（master-permission edit）を削除することはできません")
	}
	return nil
}

func (s *permissionGroupService) Reorder(
	ctx context.Context,
	clinicID uint64,
	ids []uint64,
) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
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

func (s *permissionGroupService) GetEffectivePermissions(
	ctx context.Context,
	staffID, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	rules, err := s.repo.FindAllEffectivePermissionsByStaffID(ctx, staffID, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get effective permissions", "error", err, "staff_id", staffID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get effective permissions")
	}
	return rules, nil
}
