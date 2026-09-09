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
	attempted chan struct{}
	once      sync.Once
}

func (r *observingStaffPermissionGroupRepository) LockPermissionPolicy(ctx context.Context, clinicID uint64) error {
	r.once.Do(func() { close(r.attempted) })
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

func TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation(t *testing.T) {
	db, clinicID, groups := setupSelfAdminGrantGroups(t)
	repo := authdomain.NewPermissionGroupRepository(db)
	auditKernel := domainaudit.NewService(domainaudit.NewRepository(db))
	guardPassed := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseFirst()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	groupSvc := authdomain.NewPermissionGroupService(repo, persistence.NewTransactor(db), pauseAfterPermissionGuardForStaff{
		inner: committedAuthAudit{inner: auditKernel}, guardPassed: guardPassed, release: release,
	})
	attempted := make(chan struct{})
	staffSvc := NewServiceWithAudits(
		NewRepository(db),
		nil,
		nil,
		nil,
		nil,
		&observingStaffPermissionGroupRepository{PermissionGroupRepository: repo, attempted: attempted},
		nil,
		nil,
		nil,
		persistence.NewTransactor(db),
		nil,
		staffPermissionAuditAdapter{inner: auditKernel},
	)
	deactivate := make(chan error, 1)
	unassign := make(chan error, 1)
	go func() {
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
	go func() {
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
	select {
	case <-attempted:
	case <-ctx.Done():
		t.Fatal("self-unassign did not start")
	}
	select {
	case err := <-unassign:
		t.Fatalf("self-unassign bypassed clinic policy lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	releaseFirst()
	require.NoError(t, <-deactivate)
	require.ErrorIs(t, <-unassign, apperrors.ErrForbidden)
	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("clinic_id = ? AND action = ?", clinicID, model.AuditActionStaffPermissionGroupsReplace).
		Count(&auditCount).Error)
	require.Equal(t, int64(0), auditCount, "rejected unassign must not persist an audit entry")
}

type pauseAfterPermissionGuardForStaff struct {
	inner       authdomain.PermissionAuditTxLogger
	guardPassed chan struct{}
	release     <-chan struct{}
}

func (a pauseAfterPermissionGuardForStaff) LogEntryTx(ctx context.Context, entry authdomain.AuthAuditEntry) error {
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
