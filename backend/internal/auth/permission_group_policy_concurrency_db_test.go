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
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { firstResult <- firstMutate(ctx, first) }()
	select {
	case <-guardPassed:
	case <-ctx.Done():
		t.Fatal("first guard did not pass")
	}
	go func() { secondResult <- secondMutate(ctx, second) }()
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
	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).Where("action = ?", successAuditAction).Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount, "rejected transaction must not persist an audit entry")
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
}
