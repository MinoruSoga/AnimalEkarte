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
)

type pauseAfterPermissionGuard struct {
	inner       authdomain.PermissionAuditTxLogger
	guardPassed chan struct{}
	release     <-chan struct{}
}

func (a pauseAfterPermissionGuard) LogEntryTx(ctx context.Context, entry authdomain.AuthAuditEntry) error {
	close(a.guardPassed)
	select {
	case <-a.release:
		return a.inner.LogEntryTx(ctx, entry)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Two different group rows cannot serialize this write skew: each transaction
// would observe the other grant until commit and both guards would pass.
func TestPermissionPolicyDB_ConcurrentGroupDeactivation(t *testing.T) {
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
	repo := authdomain.NewPermissionGroupRepository(db)
	audit := committedPermissionAudit{inner: domainaudit.NewService(domainaudit.NewRepository(db))}
	guardPassed := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseFirst()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first := authdomain.NewPermissionGroupService(repo, persistence.NewTransactor(db), pauseAfterPermissionGuard{
		inner: audit, guardPassed: guardPassed, release: release,
	})
	attempted := make(chan struct{})
	second := authdomain.NewPermissionGroupService(&observingPermissionGroupRepository{
		PermissionGroupRepository: repo, locker: repo, attempted: attempted,
	}, persistence.NewTransactor(db), audit)
	deactivate := func(app authdomain.PermissionGroupApplication, groupID uint64, result chan<- error) {
		inactive := false
		input := permissionAuditRollbackInput(clinicID, model.AuditActionPermissionGroupUpdate, "permission_group")
		input.ActorIsSystemAdmin = false
		_, err := app.Update(ctx, clinicID, groupID, &authdomain.UpdatePermissionGroupInput{IsActive: &inactive}, input)
		result <- err
	}
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go deactivate(first, groups[0].ID, firstResult)
	select {
	case <-guardPassed:
	case <-ctx.Done():
		t.Fatal("first guard did not pass")
	}
	go deactivate(second, groups[1].ID, secondResult)
	select {
	case <-attempted:
	case <-ctx.Done():
		t.Fatal("second writer did not start")
	}
	select {
	case err := <-secondResult:
		t.Fatalf("second writer bypassed clinic policy lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	releaseFirst()
	require.NoError(t, <-firstResult)
	require.ErrorIs(t, <-secondResult, apperrors.ErrForbidden)
	var activeCount, auditCount int64
	require.NoError(t, db.Model(&model.PermissionGroup{}).Where("clinic_id = ? AND is_active = true", clinicID).Count(&activeCount).Error)
	require.Equal(t, int64(1), activeCount)
	require.NoError(t, db.Model(&model.AuditLog{}).Where("clinic_id = ? AND action = ?", clinicID, model.AuditActionPermissionGroupUpdate).Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount, "rejected transaction must not persist an audit entry")
}
