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

// D2 Update-with-Rules evidence uses a test-local staged/committed TX double.
// It proves Update(Rules) participates in WithTx and discards staged
// metadata/rules/audit on error. It is not real Postgres atomicity evidence.

type d2MutateTxKey struct{}

type d2MutateGroupState struct {
	name  string
	rules []model.PermissionGroupRule
}

type d2MutateStage struct {
	groups map[uint64]d2MutateGroupState
	audits []AuthAuditEntry
	events []string
}

type d2MutateWorld struct {
	mu             sync.Mutex
	committed      map[uint64]d2MutateGroupState
	committedAudit []AuthAuditEntry
	lastEvents     []string
}

func newD2MutateWorld(seedGroup uint64, seed d2MutateGroupState) *d2MutateWorld {
	committed := make(map[uint64]d2MutateGroupState, 1)
	committed[seedGroup] = cloneD2MutateState(seed)
	return &d2MutateWorld{committed: committed}
}

func cloneD2MutateState(state d2MutateGroupState) d2MutateGroupState {
	return d2MutateGroupState{
		name:  state.name,
		rules: append([]model.PermissionGroupRule(nil), state.rules...),
	}
}

func (w *d2MutateWorld) snapshot(groupID uint64) d2MutateGroupState {
	w.mu.Lock()
	defer w.mu.Unlock()
	return cloneD2MutateState(w.committed[groupID])
}

func (w *d2MutateWorld) auditCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.committedAudit)
}

func (w *d2MutateWorld) events() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]string(nil), w.lastEvents...)
}

type d2MutateTransactor struct {
	world *d2MutateWorld
}

func (t *d2MutateTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	stage := &d2MutateStage{groups: make(map[uint64]d2MutateGroupState)}
	txCtx := context.WithValue(ctx, d2MutateTxKey{}, stage)
	err := fn(txCtx)
	t.world.mu.Lock()
	t.world.lastEvents = append([]string(nil), stage.events...)
	t.world.mu.Unlock()
	if err != nil {
		return err
	}
	t.world.mu.Lock()
	defer t.world.mu.Unlock()
	for groupID, state := range stage.groups {
		t.world.committed[groupID] = cloneD2MutateState(state)
	}
	t.world.committedAudit = append(t.world.committedAudit, stage.audits...)
	return nil
}

func d2MutateStageFrom(ctx context.Context) *d2MutateStage {
	stage, _ := ctx.Value(d2MutateTxKey{}).(*d2MutateStage)
	return stage
}

type d2MutateAuditLogger struct {
	err error
}

func (a *d2MutateAuditLogger) LogEntryTx(ctx context.Context, entry AuthAuditEntry) error {
	stage := d2MutateStageFrom(ctx)
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

func d2MutateRepo(
	t *testing.T,
	world *d2MutateWorld,
	staffGroupIDs []uint64,
	effective []model.PermissionGroupRule,
	effectiveErr error,
) *atomicPermissionGroupRepositoryStub {
	t.Helper()
	base := &mockPermissionGroupRepository{
		findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
			require.NotNil(t, d2MutateStageFrom(ctx), "FindByID must run inside WithTx")
			state := world.snapshot(id)
			if stage := d2MutateStageFrom(ctx); stage != nil {
				if staged, ok := stage.groups[id]; ok {
					state = cloneD2MutateState(staged)
				}
			}
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     state.name,
				Rules:    append([]model.PermissionGroupRule(nil), state.rules...),
			}, nil
		},
		lockByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
			require.NotNil(t, d2MutateStageFrom(ctx), "LockByIDForUpdate must run inside WithTx")
			state := world.snapshot(id)
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     state.name,
				Rules:    append([]model.PermissionGroupRule(nil), state.rules...),
			}, nil
		},
		findAllGroupIDsByStaffIDFn: func(_ context.Context, clinicID, staffID uint64) ([]uint64, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(10), staffID)
			// Reproduce the production precheck input. After the fix this
			// lookup must no longer gate Update-with-Rules.
			return append([]uint64(nil), staffGroupIDs...), nil
		},
		getEffectivePermissionsByStaffID: func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
			assert.Equal(t, uint64(10), staffID)
			assert.Equal(t, uint64(1), clinicID)
			if effectiveErr != nil {
				return nil, effectiveErr
			}
			stage := d2MutateStageFrom(ctx)
			require.NotNil(t, stage, "effective lookup must run inside WithTx")
			stage.events = append(stage.events, "effective")
			return append([]model.PermissionGroupRule(nil), effective...), nil
		},
	}
	return &atomicPermissionGroupRepositoryStub{
		mockPermissionGroupRepository: base,
		updateWithRulesFn: func(
			ctx context.Context,
			clinicID, id uint64,
			cmd UpdatePermissionGroupInput,
			rules []model.PermissionGroupRule,
		) (*model.PermissionGroup, error) {
			stage := d2MutateStageFrom(ctx)
			require.NotNil(t, stage, "UpdateWithRules must stage writes inside WithTx")
			assert.Equal(t, uint64(1), clinicID)
			stage.events = append(stage.events, "write")
			state := world.snapshot(id)
			if cmd.Name != nil {
				state.name = *cmd.Name
			}
			state.rules = append([]model.PermissionGroupRule(nil), rules...)
			stage.groups[id] = cloneD2MutateState(state)
			return &model.PermissionGroup{
				ID:       id,
				ClinicID: clinicID,
				Name:     state.name,
				Rules:    append([]model.PermissionGroupRule(nil), state.rules...),
			}, nil
		},
	}
}

func TestPermissionGroupService_UpdateWithRules_D2_AllowsSelfGroupStripWhenOtherGroupKeepsViewAndEdit(
	t *testing.T,
) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	before := world.snapshot(1)
	beforeAudits := world.auditCount()

	repo := d2MutateRepo(t, world, []uint64{1, 2}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}, nil)
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)
	name := "renamed"

	result, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{
			Name: &name,
			Rules: []SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanView:  true,
			}},
		},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.NoError(t, err, "other active group OR grant must allow Update-with-Rules")
	require.NotNil(t, result)
	require.Equal(t, seed.name, before.name)
	require.Len(t, before.rules, 1)
	assert.Equal(t, 0, beforeAudits)
	got := world.snapshot(1)
	assert.Equal(t, "renamed", got.name)
	require.Len(t, got.rules, 1)
	assert.Equal(t, string(model.ResourceOwners), got.rules[0].Resource)
	assert.True(t, got.rules[0].CanView)
	assert.Equal(t, 1, world.auditCount())
	assert.Equal(t, []string{"write", "effective", "audit"}, world.events())
}

func d2MutateStripAdminRules() []SetPermissionGroupRulesInput {
	return []SetPermissionGroupRulesInput{{
		Resource: string(model.ResourceOwners),
		CanView:  true,
	}}
}

func TestPermissionGroupService_UpdateWithRules_D2_RejectsViewOnlyLoss(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  false,
		CanEdit:  true,
	}}, nil)
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)

	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Rules: d2MutateStripAdminRules()},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshot(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateWithRules_D2_RejectsEditOnlyLoss(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  false,
	}}, nil)
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)

	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Rules: d2MutateStripAdminRules()},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshot(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateWithRules_D2_RejectsBothLoss(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceOwners),
		CanView:  true,
	}}, nil)
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)

	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Rules: d2MutateStripAdminRules()},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshot(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateWithRules_D2_FailClosedOnEffectiveLookupError(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1, 2}, nil, errors.New("lookup failed"))
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)

	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Rules: d2MutateStripAdminRules()},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.False(t, errors.Is(err, apperrors.ErrForbidden))
	assert.Equal(t, seed, world.snapshot(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateWithRules_D2_FailClosedOnAuditError(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1, 2}, []model.PermissionGroupRule{{
		Resource: string(model.ResourceMasterPermission),
		CanView:  true,
		CanEdit:  true,
	}}, nil)
	audit := &d2MutateAuditLogger{err: errors.New("audit write failed")}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)

	_, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{Rules: d2MutateStripAdminRules()},
		testPermissionMutationAudit(
			1,
			10,
			model.AuditActionPermissionGroupUpdate,
			"permission_group",
		),
	)

	require.Error(t, err)
	assert.Equal(t, seed, world.snapshot(1))
	assert.Equal(t, 0, world.auditCount())
}

func TestPermissionGroupService_UpdateWithRules_D2_SystemAdminExemptFromSelfLockout(t *testing.T) {
	seed := d2MutateGroupState{
		name: "d2 mutate group",
		rules: []model.PermissionGroupRule{{
			Resource: string(model.ResourceMasterPermission),
			CanView:  true,
			CanEdit:  true,
		}},
	}
	world := newD2MutateWorld(1, seed)
	repo := d2MutateRepo(t, world, []uint64{1}, nil, nil)
	audit := &d2MutateAuditLogger{}
	svc := NewPermissionGroupService(repo, &d2MutateTransactor{world: world}, audit)
	mutation := testPermissionMutationAudit(
		1,
		10,
		model.AuditActionPermissionGroupUpdate,
		"permission_group",
	)
	mutation.ActorIsSystemAdmin = true

	result, err := svc.Update(
		context.Background(),
		1,
		1,
		&UpdatePermissionGroupInput{
			Rules: []SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanView:  true,
			}},
		},
		mutation,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	got := world.snapshot(1)
	require.Len(t, got.rules, 1)
	assert.Equal(t, string(model.ResourceOwners), got.rules[0].Resource)
	assert.Equal(t, 1, world.auditCount())
}
