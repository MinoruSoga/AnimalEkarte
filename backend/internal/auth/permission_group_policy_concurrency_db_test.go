package auth_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	domainaudit "github.com/animal-ekarte/backend/internal/audit"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"

	"gorm.io/gorm"
)

type pauseAfterPermissionGuard struct {
	inner       authdomain.PermissionAuditTxLogger
	guardPassed chan struct{}
	holderPID   chan int
	release     <-chan struct{}
}

func (a pauseAfterPermissionGuard) LogEntryTx(ctx context.Context, entry authdomain.AuthAuditEntry) error {
	if tx := persistence.TxFromContext(ctx); tx != nil && a.holderPID != nil {
		var pid int
		if err := tx.WithContext(ctx).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err == nil && pid != 0 {
			select {
			case a.holderPID <- pid:
			default:
			}
		}
	}
	close(a.guardPassed)
	select {
	case <-a.release:
		return a.inner.LogEntryTx(ctx, entry)
	case <-ctx.Done():
		return ctx.Err()
	}
}

type pidReportingLockByIDRepository struct {
	authdomain.PermissionGroupRepository
	locker       authdomain.PermissionGroupMutationLocker
	contenderPID chan int
	once         sync.Once
}

func (r *pidReportingLockByIDRepository) LockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	if tx := persistence.TxFromContext(ctx); tx != nil {
		var pid int
		if err := tx.WithContext(ctx).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
			return nil, err
		}
		r.once.Do(func() {
			select {
			case r.contenderPID <- pid:
			default:
			}
		})
	}
	return r.locker.LockByIDForUpdate(ctx, clinicID, id)
}

func setupTwoAdminGrantGroups(t *testing.T) (*gorm.DB, uint64, []model.PermissionGroup) {
	t.Helper()
	db, clinicID := setupPermissionAuditRollbackDB(t)
	groups := []model.PermissionGroup{
		{ClinicID: clinicID, Name: "first admin grant", IsActive: true},
		{ClinicID: clinicID, Name: "second admin grant", IsActive: true},
	}
	for i := range groups {
		require.NoError(t, db.Create(&groups[i]).Error)
		require.NoError(t, db.Create(&model.PermissionGroupRule{
			GroupID: groups[i].ID, Resource: string(model.ResourceMasterPermission), CanView: true, CanEdit: true,
		}).Error)
		require.NoError(t, db.Create(&model.StaffPermissionGroup{StaffID: 17, GroupID: groups[i].ID}).Error)
	}
	return db, clinicID, groups
}

func waitForPolicyLockContenderBlocked(
	t *testing.T,
	ctx context.Context,
	db *gorm.DB,
	holderPID, contenderPID int,
	contenderDone <-chan error,
) {
	t.Helper()
	require.NotZero(t, holderPID)
	require.NotZero(t, contenderPID)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-contenderDone:
			require.Failf(t, "contender finished before policy-lock wait was observed", "err=%v", err)
		case <-ticker.C:
			var blockedByHolder bool
			require.NoError(t, db.WithContext(ctx).Raw(
				`SELECT EXISTS (
					SELECT 1
					FROM pg_locks blocked
					JOIN pg_locks holder
					  ON holder.locktype = blocked.locktype
					 AND holder.database IS NOT DISTINCT FROM blocked.database
					 AND holder.classid IS NOT DISTINCT FROM blocked.classid
					 AND holder.objid IS NOT DISTINCT FROM blocked.objid
					 AND holder.objsubid IS NOT DISTINCT FROM blocked.objsubid
					 AND holder.pid IS DISTINCT FROM blocked.pid
					WHERE blocked.pid = ?
					  AND holder.pid = ?
					  AND NOT blocked.granted
					  AND holder.granted
					  AND blocked.locktype = 'advisory'
				)`,
				contenderPID,
				holderPID,
			).Scan(&blockedByHolder).Error)
			if blockedByHolder {
				return
			}
		case <-ctx.Done():
			require.Failf(t, "contender did not enter advisory policy-lock wait behind holder",
				"holderPID=%d contenderPID=%d err=%v", holderPID, contenderPID, ctx.Err())
		}
	}
}

func receivePermissionWriterResult(
	t *testing.T,
	ctx context.Context,
	result <-chan error,
	label string,
) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		require.Failf(t, "writer did not finish before deadline", "label=%s err=%v", label, ctx.Err())
		return ctx.Err()
	}
}

func serializedPermissionWriters(
	t *testing.T,
	db *gorm.DB,
	firstMutate func(context.Context, authdomain.PermissionGroupApplication) error,
	secondMutate func(context.Context, authdomain.PermissionGroupApplication) error,
	successAuditAction string,
) {
	t.Helper()
	repo := authdomain.NewPermissionGroupRepository(db)
	audit := committedPermissionAudit{inner: domainaudit.NewService(domainaudit.NewRepository(db))}
	guardPassed := make(chan struct{})
	holderPIDCh := make(chan int, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseFirst)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	first := authdomain.NewPermissionGroupService(repo, persistence.NewTransactor(db), pauseAfterPermissionGuard{
		inner: audit, guardPassed: guardPassed, holderPID: holderPIDCh, release: release,
	})
	contenderPIDCh := make(chan int, 1)
	second := authdomain.NewPermissionGroupService(&pidReportingLockByIDRepository{
		PermissionGroupRepository: repo,
		locker:                    repo,
		contenderPID:              contenderPIDCh,
	}, persistence.NewTransactor(db), audit)

	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	type startedWorker struct {
		name string
		done <-chan struct{}
	}
	var started []startedWorker
	t.Cleanup(func() {
		releaseFirst()
		cancel()
		// Observe only workers that actually started. Do not spawn a
		// WaitGroup waiter goroutine that can outlive this cleanup.
		deadline := time.Now().Add(2 * time.Second)
		for _, w := range started {
			remaining := time.Until(deadline)
			if remaining < 0 {
				remaining = 0
			}
			timer := time.NewTimer(remaining)
			select {
			case <-w.done:
				if !timer.Stop() {
					<-timer.C
				}
			case <-timer.C:
				t.Errorf("permission-policy writer %s did not finish within cleanup deadline", w.name)
			}
		}
	})

	firstDone := make(chan struct{})
	started = append(started, startedWorker{name: "first", done: firstDone})
	go func() {
		defer close(firstDone)
		firstResult <- firstMutate(ctx, first)
	}()

	select {
	case <-guardPassed:
	case <-ctx.Done():
		t.Fatal("first guard did not pass")
	}
	var holderPID int
	select {
	case holderPID = <-holderPIDCh:
	case <-ctx.Done():
		t.Fatal("holder backend pid was not reported")
	}

	secondDone := make(chan struct{})
	started = append(started, startedWorker{name: "second", done: secondDone})
	go func() {
		defer close(secondDone)
		secondResult <- secondMutate(ctx, second)
	}()

	var contenderPID int
	select {
	case contenderPID = <-contenderPIDCh:
	case err := <-secondResult:
		t.Fatalf("second writer finished before reporting pid: %v", err)
	case <-ctx.Done():
		t.Fatal("contender backend pid was not reported")
	}

	waitForPolicyLockContenderBlocked(t, ctx, db, holderPID, contenderPID, secondResult)
	releaseFirst()
	require.NoError(t, receivePermissionWriterResult(t, ctx, firstResult, "first"))
	require.ErrorIs(t, receivePermissionWriterResult(t, ctx, secondResult, "second"), apperrors.ErrForbidden)

	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).Where("action = ?", successAuditAction).Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount, "exactly one success audit; rejected TX must not audit")
}

// Two different group rows cannot serialize this write skew: each transaction
// would observe the other grant until commit and both guards would pass.
func TestPermissionPolicyDB_ConcurrentGroupDeactivation(t *testing.T) {
	db, clinicID, groups := setupTwoAdminGrantGroups(t)
	deactivate := func(groupID uint64) func(context.Context, authdomain.PermissionGroupApplication) error {
		return func(ctx context.Context, app authdomain.PermissionGroupApplication) error {
			inactive := false
			input := permissionAuditRollbackInput(clinicID, model.AuditActionPermissionGroupUpdate, "permission_group")
			input.ActorIsSystemAdmin = false
			_, err := app.Update(ctx, clinicID, groupID, &authdomain.UpdatePermissionGroupInput{IsActive: &inactive}, input)
			return err
		}
	}
	serializedPermissionWriters(t, db, deactivate(groups[0].ID), deactivate(groups[1].ID), model.AuditActionPermissionGroupUpdate)
	var activeCount int64
	require.NoError(t, db.Model(&model.PermissionGroup{}).Where("clinic_id = ? AND is_active = true", clinicID).Count(&activeCount).Error)
	require.Equal(t, int64(1), activeCount)
	assertActorKeepsMasterPermissionViewAndEdit(t, db, clinicID, 17)
}

func TestPermissionPolicyDB_ConcurrentRuleReplacement(t *testing.T) {
	db, clinicID, groups := setupTwoAdminGrantGroups(t)
	stripAdmin := func(groupID uint64) func(context.Context, authdomain.PermissionGroupApplication) error {
		return func(ctx context.Context, app authdomain.PermissionGroupApplication) error {
			input := permissionAuditRollbackInput(clinicID, model.AuditActionPermissionRulesUpdate, "permission_group_rules")
			input.ActorIsSystemAdmin = false
			_, err := app.UpdateRules(ctx, clinicID, groupID, []authdomain.SetPermissionGroupRulesInput{{
				Resource: string(model.ResourceOwners),
				CanView:  true,
			}}, 17, input)
			return err
		}
	}
	serializedPermissionWriters(t, db, stripAdmin(groups[0].ID), stripAdmin(groups[1].ID), model.AuditActionPermissionRulesUpdate)
	var adminGrantCount int64
	require.NoError(t, db.Model(&model.PermissionGroupRule{}).
		Where("group_id IN ? AND resource = ? AND can_view = true AND can_edit = true AND deleted_at IS NULL",
			[]uint64{groups[0].ID, groups[1].ID}, string(model.ResourceMasterPermission)).
		Count(&adminGrantCount).Error)
	require.Equal(t, int64(1), adminGrantCount)
	assertActorKeepsMasterPermissionViewAndEdit(t, db, clinicID, 17)
}

func assertActorKeepsMasterPermissionViewAndEdit(t *testing.T, db *gorm.DB, clinicID, staffID uint64) {
	t.Helper()
	repo := authdomain.NewPermissionGroupRepository(db)
	rules, err := repo.FindAllEffectivePermissionsByStaffID(context.Background(), staffID, clinicID)
	require.NoError(t, err)
	hasView, hasEdit := false, false
	for i := range rules {
		rule := &rules[i]
		if rule.Resource != string(model.ResourceMasterPermission) {
			continue
		}
		hasView = hasView || rule.CanView
		hasEdit = hasEdit || rule.CanEdit
	}
	require.True(t, hasView, "actor must retain master-permission view after concurrent mutation")
	require.True(t, hasEdit, "actor must retain master-permission edit after concurrent mutation")
}
