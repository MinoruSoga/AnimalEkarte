package auth

import (
	"context"
	"log/slog"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// CreatePermissionGroupInput is the permission-group creation input.
type CreatePermissionGroupInput struct {
	Name        string
	Description string
	Color       string
	IsActive    bool
	SortOrder   int
	Rules       []SetPermissionGroupRulesInput
}

// UpdatePermissionGroupInput contains optional PATCH fields.
type UpdatePermissionGroupInput struct {
	Name        *string
	Description *string
	Color       *string
	SortOrder   *int
	IsActive    *bool
	Rules       []SetPermissionGroupRulesInput
}

// SetPermissionGroupRulesInput is one replacement permission rule.
type SetPermissionGroupRulesInput struct {
	Resource  string
	CanView   bool
	CanCreate bool
	CanEdit   bool
	CanDelete bool
}

// PermissionMutationAudit is the typed request metadata required for one
// authorization-policy mutation. The use case validates clinic/action/resource
// and fills ResourceID/OldValue/NewValue from transaction-local state.
type PermissionMutationAudit struct {
	ClinicID     uint64
	ActorStaffID uint64
	Action       string
	Resource     string
	ResourceID   *uint64
	OldValue     any
	NewValue     any
	IPAddress    string
	UserAgent    string
}

// PermissionAuditTxLogger persists an authorization mutation audit entry in
// the caller's ambient transaction.
type PermissionAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry AuthAuditEntry) error
}

// PermissionGroupMutationLocker is the transaction-local row lock required
// before an authorization-policy mutation captures its old audit snapshot.
type PermissionGroupMutationLocker interface {
	LockByIDForUpdate(
		ctx context.Context,
		clinicID, id uint64,
	) (*model.PermissionGroup, error)
}

const (
	colPermissionGroupName        = "name"
	colPermissionGroupDescription = "description"
	colPermissionGroupColor       = "color"
	colPermissionGroupSortOrder   = "sort_order"
	colPermissionGroupIsActive    = "is_active"

	errMsgResourceNameEmpty = "リソース名が空です"
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

// PermissionGroupService provides clinic-scoped permission-group CRUD.
type PermissionGroupService interface {
	List(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	Create(
		ctx context.Context,
		clinicID uint64,
		input *CreatePermissionGroupInput,
		audit PermissionMutationAudit,
	) (*model.PermissionGroup, error)
	Update(
		ctx context.Context,
		clinicID, id uint64,
		input *UpdatePermissionGroupInput,
		audit PermissionMutationAudit,
	) (*model.PermissionGroup, error)
	Delete(
		ctx context.Context,
		clinicID, id uint64,
		audit PermissionMutationAudit,
	) error
	UpdateRules(
		ctx context.Context,
		clinicID, groupID uint64,
		inputs []SetPermissionGroupRulesInput,
		actorStaffID uint64,
		audit PermissionMutationAudit,
	) (*model.PermissionGroup, error)
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

// EffectivePermissionService provides clinic-scoped effective permissions.
type EffectivePermissionService interface {
	GetEffectivePermissions(
		ctx context.Context,
		staffID, clinicID uint64,
	) ([]model.PermissionGroupRule, error)
}

// PermissionGroupApplication combines permission-group mutation and effective
// permission lookup for type-safe application composition.
type PermissionGroupApplication interface {
	PermissionGroupService
	EffectivePermissionService
}

type permissionGroupService struct {
	repo       PermissionGroupRepository
	transactor Transactor
	audit      PermissionAuditTxLogger
}

// NewPermissionGroupService constructs permission-group CRUD and effective
// permission lookup on one shared implementation.
func NewPermissionGroupService(
	repo PermissionGroupRepository,
	transactor Transactor,
	audit PermissionAuditTxLogger,
) PermissionGroupApplication {
	return &permissionGroupService{
		repo:       repo,
		transactor: transactor,
		audit:      audit,
	}
}

func (s *permissionGroupService) validateAuditedMutation(
	clinicID uint64,
	audit PermissionMutationAudit,
	expectedAction, expectedResource string,
) error {
	if s.transactor == nil || s.audit == nil {
		return apperrors.WrapInternalServerError(
			"permission mutation audit dependencies are not configured",
		)
	}
	if clinicID == 0 || audit.ClinicID != clinicID ||
		audit.ActorStaffID == 0 ||
		audit.Action != expectedAction ||
		audit.Resource != expectedResource {
		return apperrors.WrapInternalServerError(
			"permission mutation audit metadata is invalid",
		)
	}
	return nil
}

func permissionAuditEntry(
	audit PermissionMutationAudit,
	resourceID uint64,
	oldValue, newValue any,
) AuthAuditEntry {
	clinicID := audit.ClinicID
	actorID := audit.ActorStaffID
	audit.ResourceID = &resourceID
	audit.OldValue = oldValue
	audit.NewValue = newValue
	return AuthAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     audit.Action,
		Resource:   audit.Resource,
		ResourceID: audit.ResourceID,
		OldValue:   audit.OldValue,
		NewValue:   audit.NewValue,
		IPAddress:  audit.IPAddress,
		UserAgent:  audit.UserAgent,
	}
}

func permissionGroupAuditSnapshot(
	group *model.PermissionGroup,
) map[string]any {
	return map[string]any{
		"name":        group.Name,
		"description": group.Description,
		"color":       group.Color,
		"is_active":   group.IsActive,
		"sort_order":  group.SortOrder,
		"rules":       permissionRulesAuditSnapshot(group.Rules),
	}
}

func permissionRulesAuditSnapshot(
	rules []model.PermissionGroupRule,
) []map[string]any {
	snapshot := make([]map[string]any, 0, len(rules))
	for i := range rules {
		rule := &rules[i]
		snapshot = append(snapshot, map[string]any{
			"resource":   rule.Resource,
			"can_view":   rule.CanView,
			"can_create": rule.CanCreate,
			"can_edit":   rule.CanEdit,
			"can_delete": rule.CanDelete,
		})
	}
	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i]["resource"].(string) <
			snapshot[j]["resource"].(string)
	})
	return snapshot
}

func (s *permissionGroupService) lockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	locker, ok := s.repo.(PermissionGroupMutationLocker)
	if !ok {
		return nil, apperrors.WrapInternalServerError(
			"permission group mutation lock is not configured",
		)
	}
	group, err := locker.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			"failed to lock permission group for mutation",
		)
	}
	if group == nil || group.ID != id || group.ClinicID != clinicID {
		return nil, apperrors.WrapInternalServerError(
			"permission group mutation lock returned an invalid identity",
		)
	}
	return group, nil
}

func (s *permissionGroupService) List(
	ctx context.Context,
	clinicID uint64,
) ([]model.PermissionGroup, error) {
	items, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list permission groups", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list permission groups")
	}
	return items, nil
}

func (s *permissionGroupService) GetByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get permission group", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get permission group")
	}
	if result == nil {
		return nil, apperrors.WrapInternalServerError(
			"permission group repository returned an empty result",
		)
	}
	return result, nil
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
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintPermissionGroupName,
				apperrors.CodePermissionGroupNameConflict,
			); conflict != nil {
				return nil, conflict
			}
			slog.ErrorContext(ctx, "failed to create permission group", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to create permission group")
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
			if conflict := apperrors.AsNameUniqueConflict(
				err,
				input.Name,
				apperrors.ConstraintPermissionGroupName,
				apperrors.CodePermissionGroupNameConflict,
			); conflict != nil {
				return nil, conflict
			}
			slog.ErrorContext(ctx, "failed to create permission group with rules", "error", err, "clinic_id", clinicID)
			return nil, apperrors.Wrap(
				err,
				"failed to create permission group with rules",
			)
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
		slog.ErrorContext(ctx, "failed to get permission group", "error", err, "id", id, "clinic_id", clinicID)
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
		if conflict := apperrors.AsNameUniqueConflict(
			err,
			nameForConflict,
			apperrors.ConstraintPermissionGroupName,
			apperrors.CodePermissionGroupNameConflict,
		); conflict != nil {
			return nil, conflict
		}
		slog.ErrorContext(ctx, "failed to update permission group", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update permission group")
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
