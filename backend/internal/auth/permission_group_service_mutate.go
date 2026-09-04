package auth

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

func (s *permissionGroupService) Create(
	ctx context.Context,
	clinicID uint64,
	input *CreatePermissionGroupInput,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	if err := s.validateAuditedMutation(
		clinicID,
		audit,
		model.AuditActionPermissionGroupCreate,
		"permission_group",
	); err != nil {
		return nil, err
	}
	var group *model.PermissionGroup
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		var createErr error
		group, createErr = s.create(txCtx, clinicID, input)
		if createErr != nil {
			return createErr
		}
		entry := permissionAuditEntry(
			audit,
			group.ID,
			nil,
			permissionGroupAuditSnapshot(group),
		)
		if auditErr := s.audit.LogEntryTx(txCtx, entry); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to audit permission group create",
			)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *permissionGroupService) create(
	ctx context.Context,
	clinicID uint64,
	input *CreatePermissionGroupInput,
) (*model.PermissionGroup, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(
			sharedkernel.ErrMsgInputNotNil,
		)
	}
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	group := &model.PermissionGroup{
		ClinicID:    clinicID,
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		IsActive:    input.IsActive,
		SortOrder:   input.SortOrder,
	}
	if input.Rules == nil {
		if err := s.repo.Create(ctx, group); err != nil {
			return nil, mapPermissionGroupNameConflict(err, input.Name, "failed to create permission group")
		}
	} else {
		rules := permissionRuleModels(input.Rules)
		if err := validateNoDuplicateRules(rules); err != nil {
			return nil, err
		}
		writer, ok := s.repo.(PermissionGroupRulesAtomicWriter)
		if !ok {
			return nil, apperrors.WrapInternalServerError(
				"atomic permission group rule writer is not configured",
			)
		}
		var err error
		group, err = writer.CreateWithRules(ctx, group, rules)
		if err != nil {
			return nil, mapPermissionGroupNameConflict(err, input.Name, "failed to create permission group with rules")
		}
	}
	slog.InfoContext(ctx, "permission group created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", group.ID),
		slog.String("name", group.Name))
	return group, nil
}

func (s *permissionGroupService) Update(
	ctx context.Context,
	clinicID, id uint64,
	input *UpdatePermissionGroupInput,
	audit PermissionMutationAudit,
) (*model.PermissionGroup, error) {
	if err := s.validateAuditedMutation(
		clinicID,
		audit,
		model.AuditActionPermissionGroupUpdate,
		"permission_group",
	); err != nil {
		return nil, err
	}
	var result *model.PermissionGroup
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		oldGroup, oldErr := s.lockByIDForUpdate(txCtx, clinicID, id)
		if oldErr != nil {
			return oldErr
		}
		var updateErr error
		result, updateErr = s.update(
			txCtx,
			clinicID,
			id,
			input,
			audit.ActorStaffID,
		)
		if updateErr != nil {
			return updateErr
		}
		entry := permissionAuditEntry(
			audit,
			id,
			permissionGroupAuditSnapshot(oldGroup),
			permissionGroupAuditSnapshot(result),
		)
		if auditErr := s.audit.LogEntryTx(txCtx, entry); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to audit permission group update",
			)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *permissionGroupService) update(
	ctx context.Context,
	clinicID, id uint64,
	input *UpdatePermissionGroupInput,
	actorStaffID uint64,
) (*model.PermissionGroup, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get permission group")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	fields := buildPermissionGroupUpdate(input)
	if len(fields) == 0 && input.Rules == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	var result *model.PermissionGroup
	var err error
	if input.Rules == nil {
		result, err = s.repo.Update(ctx, clinicID, id, fields)
	} else {
		rules := permissionRuleModels(input.Rules)
		if validationErr := validateNoDuplicateRules(rules); validationErr != nil {
			return nil, validationErr
		}
		staffGroupIDs, groupIDsErr := s.repo.FindAllGroupIDsByStaffID(
			ctx,
			clinicID,
			actorStaffID,
		)
		if groupIDsErr != nil {
			return nil, apperrors.Wrap(
				groupIDsErr,
				"failed to find staff group IDs",
			)
		}
		if validationErr := validateNotSelfReference(
			id,
			rules,
			staffGroupIDs,
		); validationErr != nil {
			return nil, validationErr
		}
		writer, ok := s.repo.(PermissionGroupRulesAtomicWriter)
		if !ok {
			return nil, apperrors.WrapInternalServerError(
				"atomic permission group rule writer is not configured",
			)
		}
		result, err = writer.UpdateWithRules(
			ctx,
			clinicID,
			id,
			fields,
			rules,
		)
	}
	if err != nil {
		nameForConflict := ""
		if input.Name != nil {
			nameForConflict = *input.Name
		}
		return nil, mapPermissionGroupNameConflict(err, nameForConflict, "failed to update permission group")
	}
	if result == nil {
		return nil, apperrors.WrapInternalServerError(
			"permission group repository returned an empty update result",
		)
	}
	slog.InfoContext(ctx, "permission group updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("permission_group_id", id))
	return result, nil
}

func (s *permissionGroupService) Delete(
	ctx context.Context,
	clinicID, id uint64,
	audit PermissionMutationAudit,
) error {
	if err := s.validateAuditedMutation(
		clinicID,
		audit,
		model.AuditActionPermissionGroupDelete,
		"permission_group",
	); err != nil {
		return err
	}
	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		oldGroup, oldErr := s.lockByIDForUpdate(txCtx, clinicID, id)
		if oldErr != nil {
			return oldErr
		}
		if deleteErr := s.delete(txCtx, clinicID, id); deleteErr != nil {
			return deleteErr
		}
		entry := permissionAuditEntry(
			audit,
			id,
			permissionGroupAuditSnapshot(oldGroup),
			nil,
		)
		if auditErr := s.audit.LogEntryTx(txCtx, entry); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to audit permission group delete",
			)
		}
		return nil
	})
}

func (s *permissionGroupService) delete(
	ctx context.Context,
	clinicID, id uint64,
) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to get permission group")
	}
	count, err := s.repo.CountUsageByGroupID(ctx, clinicID, id)
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
		slog.Uint64("permission_group_id", id))
	return nil
}

func mapPermissionGroupNameConflict(err error, name, wrapMsg string) error {
	if conflict := apperrors.AsNameUniqueConflict(
		err,
		name,
		apperrors.ConstraintPermissionGroupName,
		apperrors.CodePermissionGroupNameConflict,
	); conflict != nil {
		return conflict
	}
	return apperrors.Wrap(err, wrapMsg)
}
