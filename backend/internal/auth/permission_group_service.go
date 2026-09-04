package auth

import (
	"context"
	"log/slog"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
