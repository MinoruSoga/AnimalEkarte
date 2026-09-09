package staff

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	domainaudit "github.com/animal-ekarte/backend/internal/audit"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type observingStaffPermissionGroupRepository struct {
	PermissionGroupRepository
	contenderPID chan int
	once         sync.Once
}

func (r *observingStaffPermissionGroupRepository) LockPermissionPolicy(ctx context.Context, clinicID uint64) error {
	if tx := persistence.TxFromContext(ctx); tx != nil {
		var pid int
		if err := tx.WithContext(ctx).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
			return err
		}
		r.once.Do(func() {
			select {
			case r.contenderPID <- pid:
			default:
			}
		})
	}
	return r.PermissionGroupRepository.LockPermissionPolicy(ctx, clinicID)
}

type staffPermissionAuditAdapter struct {
	inner domainaudit.TxLogger
}

func (a staffPermissionAuditAdapter) LogEntryTx(ctx context.Context, entry *PermissionAssignmentAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &domainaudit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}

func setupSelfAdminGrantGroups(t *testing.T) (*gorm.DB, uint64, []model.PermissionGroup) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.PermissionGroupRule{},
		&model.StaffPermissionGroup{},
		&model.AuditLog{},
	))
	testdb.Truncate(
		t,
		db,
		"audit_logs",
		"staff_permission_groups",
		"permission_group_rules",
		"permission_groups",
		"staff_clinic_assignments",
		"staffs",
		"clinics",
		"companies",
	)
	company := &model.Company{Name: "staff permission lock company"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: "staff permission lock clinic"}
	require.NoError(t, db.Create(clinic).Error)
	require.NoError(t, db.Create(&model.Staff{
		ID: 17, ClinicID: clinic.ID, Name: "self admin actor", IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: 17, ClinicID: clinic.ID, IsMain: true,
	}).Error)
	groups := []model.PermissionGroup{
		{ClinicID: clinic.ID, Name: "first admin grant", IsActive: true},
		{ClinicID: clinic.ID, Name: "second admin grant", IsActive: true},
	}
	for i := range groups {
		require.NoError(t, db.Create(&groups[i]).Error)
		require.NoError(t, db.Create(&model.PermissionGroupRule{
			GroupID: groups[i].ID, Resource: string(model.ResourceMasterPermission), CanView: true, CanEdit: true,
		}).Error)
		require.NoError(t, db.Create(&model.StaffPermissionGroup{StaffID: 17, GroupID: groups[i].ID}).Error)
	}
	return db, clinic.ID, groups
}

func waitForStaffPolicyLockContenderBlocked(
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
			require.Failf(t, "self-unassign finished before policy-lock wait was observed", "err=%v", err)
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
			require.Failf(t, "self-unassign did not enter advisory policy-lock wait behind holder",
				"holderPID=%d contenderPID=%d err=%v", holderPID, contenderPID, ctx.Err())
		}
	}
}

func receiveStaffPermissionWriterResult(
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

func TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation(t *testing.T) {
	db, clinicID, groups := setupSelfAdminGrantGroups(t)
	repo := authdomain.NewPermissionGroupRepository(db)
	auditKernel := domainaudit.NewService(domainaudit.NewRepository(db))
	guardPassed := make(chan struct{})
	holderPIDCh := make(chan int, 1)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseFirst)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	groupSvc := authdomain.NewPermissionGroupService(repo, persistence.NewTransactor(db), pauseAfterPermissionGuardForStaff{
		inner: committedAuthAudit{inner: auditKernel}, guardPassed: guardPassed, holderPID: holderPIDCh, release: release,
	})
	contenderPIDCh := make(chan int, 1)
	staffSvc := NewServiceWithAudits(
		NewRepository(db),
		nil,
		nil,
		nil,
		nil,
		&observingStaffPermissionGroupRepository{PermissionGroupRepository: repo, contenderPID: contenderPIDCh},
		nil,
		nil,
		nil,
		persistence.NewTransactor(db),
		nil,
		staffPermissionAuditAdapter{inner: auditKernel},
	)

	deactivate := make(chan error, 1)
	unassign := make(chan error, 1)
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
				t.Errorf("staff permission-policy writer %s did not finish within cleanup deadline", w.name)
			}
		}
	})

	deactivateDone := make(chan struct{})
	started = append(started, startedWorker{name: "deactivate", done: deactivateDone})
	go func() {
		defer close(deactivateDone)
		inactive := false
		input := authdomain.PermissionMutationAudit{
			ClinicID: clinicID, ActorStaffID: 17, Action: model.AuditActionPermissionGroupUpdate,
			Resource: "permission_group", IPAddress: "127.0.0.1", UserAgent: "staff-permission-lock-db-test",
		}
		_, err := groupSvc.Update(ctx, clinicID, groups[0].ID, &authdomain.UpdatePermissionGroupInput{IsActive: &inactive}, input)
		deactivate <- err
	}()

	select {
	case <-guardPassed:
	case <-ctx.Done():
		t.Fatal("group deactivation guard did not pass")
	}
	var holderPID int
	select {
	case holderPID = <-holderPIDCh:
	case <-ctx.Done():
		t.Fatal("holder backend pid was not reported")
	}

	unassignDone := make(chan struct{})
	started = append(started, startedWorker{name: "unassign", done: unassignDone})
	go func() {
		defer close(unassignDone)
		unassign <- staffSvc.SetPermissionGroupIDs(
			withPermissionAssignmentAudit(ctx, PermissionAssignmentAudit{
				ClinicID: clinicID, ActorStaffID: 17, TargetStaffID: 17,
				IPAddress: "127.0.0.1", UserAgent: "staff-permission-lock-db-test",
			}),
			clinicID,
			17,
			[]uint64{groups[0].ID},
		)
	}()

	var contenderPID int
	select {
	case contenderPID = <-contenderPIDCh:
	case err := <-unassign:
		t.Fatalf("self-unassign finished before reporting pid: %v", err)
	case <-ctx.Done():
		t.Fatal("contender backend pid was not reported")
	}

	waitForStaffPolicyLockContenderBlocked(t, ctx, db, holderPID, contenderPID, unassign)
	releaseFirst()
	require.NoError(t, receiveStaffPermissionWriterResult(t, ctx, deactivate, "deactivate"))
	require.ErrorIs(t, receiveStaffPermissionWriterResult(t, ctx, unassign, "unassign"), apperrors.ErrForbidden)

	var deactivateAudit int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("clinic_id = ? AND action = ?", clinicID, model.AuditActionPermissionGroupUpdate).
		Count(&deactivateAudit).Error)
	require.Equal(t, int64(1), deactivateAudit, "successful deactivation must leave exactly one audit")
	var rejectAudit int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("clinic_id = ? AND action = ?", clinicID, model.AuditActionStaffPermissionGroupsReplace).
		Count(&rejectAudit).Error)
	require.Equal(t, int64(0), rejectAudit, "rejected unassign must not persist an audit entry")
	var assigned []uint64
	require.NoError(t, db.Model(&model.StaffPermissionGroup{}).
		Where("staff_id = ?", 17).
		Pluck("group_id", &assigned).Error)
	require.ElementsMatch(t, []uint64{groups[0].ID, groups[1].ID}, assigned, "rejected unassign must restore prior group IDs")
	rules, err := repo.FindAllEffectivePermissionsByStaffID(context.Background(), 17, clinicID)
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
	require.True(t, hasView && hasEdit, "actor must retain master-permission view+edit after rejected unassign")
}

type pauseAfterPermissionGuardForStaff struct {
	inner       authdomain.PermissionAuditTxLogger
	guardPassed chan struct{}
	holderPID   chan int
	release     <-chan struct{}
}

func (a pauseAfterPermissionGuardForStaff) LogEntryTx(ctx context.Context, entry authdomain.AuthAuditEntry) error {
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

type committedAuthAudit struct {
	inner domainaudit.TxLogger
}

func (a committedAuthAudit) LogEntryTx(ctx context.Context, entry authdomain.AuthAuditEntry) error {
	return a.inner.LogEntryTx(ctx, &domainaudit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	})
}
