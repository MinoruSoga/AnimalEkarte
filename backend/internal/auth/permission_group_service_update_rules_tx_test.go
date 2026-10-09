package auth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// D2 unit evidence uses a test-local staged/committed TX double. This proves
// UpdateRules participates in WithTx and discards staged rules/audit on error.
// It is not real Postgres atomicity evidence.

type d2RulesTxKey struct{}

type d2RulesStage struct {
	rulesByGroup map[uint64][]model.PermissionGroupRule
	audits       []AuthAuditEntry
	events       []string
}

type d2RulesWorld struct {
	mu              sync.Mutex
	committedRules  map[uint64][]model.PermissionGroupRule
	committedAudits []AuthAuditEntry
	lastEvents      []string
}

func newD2RulesWorld(seedGroup uint64, seed []model.PermissionGroupRule) *d2RulesWorld {
	rules := make(map[uint64][]model.PermissionGroupRule, 1)
	if seed != nil {
		cp := append([]model.PermissionGroupRule(nil), seed...)
		rules[seedGroup] = cp
	}
	return &d2RulesWorld{committedRules: rules}
}

func (w *d2RulesWorld) snapshotRules(groupID uint64) []model.PermissionGroupRule {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]model.PermissionGroupRule(nil), w.committedRules[groupID]...)
}

func (w *d2RulesWorld) auditCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.committedAudits)
}

func (w *d2RulesWorld) events() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string(nil), w.lastEvents...)
}

type d2RulesTransactor struct {
	world *d2RulesWorld
}

func (t *d2RulesTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	stage := &d2RulesStage{rulesByGroup: make(map[uint64][]model.PermissionGroupRule)}
	txCtx := context.WithValue(ctx, d2RulesTxKey{}, stage)
	err := fn(txCtx)
	t.world.mu.Lock()
	t.world.lastEvents = append([]string(nil), stage.events...)
	t.world.mu.Unlock()
	if err != nil {
		return err
	}
	t.world.mu.Lock()
	defer t.world.mu.Unlock()
	for groupID, rules := range stage.rulesByGroup {
		t.world.committedRules[groupID] = append([]model.PermissionGroupRule(nil), rules...)
	}
	t.world.committedAudits = append(t.world.committedAudits, stage.audits...)
	return nil
}

func d2RulesStageFrom(ctx context.Context) *d2RulesStage {
	stage, _ := ctx.Value(d2RulesTxKey{}).(*d2RulesStage)
	return stage
}

type d2RulesAuditLogger struct {
	err error
}

func (a *d2RulesAuditLogger) LogEntryTx(ctx context.Context, entry AuthAuditEntry) error {
	stage := d2RulesStageFrom(ctx)
	if stage == nil {
		return errors.New("audit write outside WithTx")
	}
	if a.err != nil {
		return a.err
	}
	stage.events = append(stage.events, "audit")
	stage.audits = append(stage.audits, entry)
	return nil
}

func d2UpdateRulesRepo(
	t *testing.T,
	world *d2RulesWorld,
	staffGroupIDs []uint64,
	effective []model.PermissionGroupRule,
	effectiveErr error,
) *mockPermissionGroupRepository {
	t.Helper()
	return &mockPermissionGroupRepository{
		lockByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
			require.NotNil(t, d2RulesStageFrom(ctx), "LockByIDForUpdate must run inside WithTx")
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     "d2 group",
				Rules:    world.snapshotRules(id),
			}, nil
		},
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
			require.NotNil(t, d2RulesStageFrom(ctx), "GetByID must run inside WithTx")
			stage := d2RulesStageFrom(ctx)
			rules := world.snapshotRules(id)
			if staged, ok := stage.rulesByGroup[id]; ok {
				rules = append([]model.PermissionGroupRule(nil), staged...)
			}
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     "d2 group",
				Rules:    rules,
			}, nil
		},
		setRulesFn: func(ctx context.Context, clinicID, groupID uint64, rules []model.PermissionGroupRule) error {
			assert.Equal(t, uint64(1), clinicID)
			stage := d2RulesStageFrom(ctx)
			require.NotNil(t, stage, "UpdateRules must stage writes inside WithTx")
			stage.events = append(stage.events, "write")
			stage.rulesByGroup[groupID] = append([]model.PermissionGroupRule(nil), rules...)
			return nil
		},
		getGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), staffID)
			return append([]uint64(nil), staffGroupIDs...), nil
		},
		getEffectivePermissionsByStaffID: func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
			assert.Equal(t, uint64(10), staffID)
			assert.Equal(t, uint64(1), clinicID)
			if effectiveErr != nil {
				return nil, effectiveErr
			}
			stage := d2RulesStageFrom(ctx)
			require.NotNil(t, stage, "effective lookup must run inside WithTx")
			stage.events = append(stage.events, "effective")
			return append([]model.PermissionGroupRule(nil), effective...), nil
		},
	}
}

func TestPermissionGroupService_UpdateRules_D2_AllowsSelfGroupStripWhenOtherGroupKeepsViewAndEdit(t *testing.T) {
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	beforeRules := world.snapshotRules(1)
	beforeAudits := world.auditCount()

	repo := d2UpdateRulesRepo(t, world, []uint64{1, 2}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}, nil)
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	result, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, beforeRules, 1)
	assert.Equal(t, 0, beforeAudits)
	got := world.snapshotRules(1)
	require.Len(t, got, 1)
	assert.Equal(t, string(model.ResourceOwners), got[0].Resource)
	assert.True(t, got[0].CanView)
	assert.Equal(t, 1, world.auditCount())
	assert.Equal(t, []string{"write", "effective", "audit"}, world.events())
}

func TestPermissionGroupService_UpdateRules_D2_RejectsViewOnlyLoss(t *testing.T) {
	// "viewだけ喪失": view lost, edit remains.
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  false,
		CanEdit:  true,
	}}, nil)
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshotRules(1), "failed TX must leave committed rules unchanged")
	assert.Equal(t, 0, world.auditCount(), "failed TX must not commit success audit")
}

func TestPermissionGroupService_UpdateRules_D2_RejectsEditOnlyLoss(t *testing.T) {
	// "editだけ喪失": edit lost, view remains.
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  false,
	}}, nil)
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshotRules(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateRules_D2_RejectsBothLoss(t *testing.T) {
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceOwners),
		CanView:  true,
	}}, nil)
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshotRules(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateRules_D2_FailClosedOnEffectiveLookupError(t *testing.T) {
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1}, nil, errors.New("lookup failed"))
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.Error(t, err)
	assert.False(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshotRules(1), "lookup failure must roll staged rules back")
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateRules_D2_FailClosedOnAuditError(t *testing.T) {
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1, 2}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}, nil)
	audit := &d2RulesAuditLogger{err: errors.New("audit write failed")}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)

	_, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionRulesUpdate,
			"permission_group_rules",
		),
	)

	require.Error(t, err)
	assert.Equal(t, seed, world.snapshotRules(1), "audit failure must roll staged rules back")
	assert.Equal(t, 0, world.auditCount(), "failed audit must not commit")
}

func TestPermissionGroupService_UpdateRules_D2_SystemAdminExemptFromSelfLockout(t *testing.T) {
	seed := []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}
	world := newD2RulesWorld(1, seed)
	repo := d2UpdateRulesRepo(t, world, []uint64{1}, nil, nil)
	audit := &d2RulesAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2RulesTransactor{world: world}, audit)
	mutation := testPermissionMutationAudit(
		1,
		10,
		model.AuditActionPermissionRulesUpdate,
		"permission_group_rules",
	)
	mutation.ActorIsSystemAdmin = true

	result, err := svc.UpdateRules(
		context.Background(),
		1,
		1,
		[]SetPermissionGroupRulesInput{{
			Resource: string(model.ResourceOwners),
			CanView:  true,
		}},
		10,
		mutation,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	got := world.snapshotRules(1)
	require.Len(t, got, 1)
	assert.Equal(t, string(model.ResourceOwners), got[0].Resource)
	assert.Equal(t, 1, world.auditCount())
}
