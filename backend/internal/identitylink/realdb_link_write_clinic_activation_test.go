package identitylink

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// grantIdentityLinkEdit grants identity-links edit on each clinic via real
// permission-group rows (same join shape as loadIdentityLinkEditClinics).
func grantIdentityLinkEdit(t *testing.T, db *gorm.DB, staffID uint64, clinicIDs ...uint64) {
	t.Helper()
	for _, clinicID := range clinicIDs {
		group := &model.PermissionGroup{
			ClinicID: clinicID,
			Name:     fmt.Sprintf("link-edit-%d-%d", staffID, clinicID),
			IsActive: true,
		}
		require.NoError(t, db.Create(group).Error)
		require.NoError(t, db.Create(&model.PermissionGroupRule{
			GroupID:  group.ID,
			Resource: string(model.ResourceIdentityLinks),
			CanEdit:  true,
		}).Error)
		require.NoError(t, db.Create(&model.StaffPermissionGroup{
			StaffID: staffID,
			GroupID: group.ID,
		}).Error)
	}
}

// linkWriteTestAccount creates an account row and links it to staffID; the
// link-write authorizer's loadCurrentSystemAdmin INNER JOINs accounts, so a
// staff without an account row is classified as "not found". Email is derived
// from the (monotonic, never truncated) clinic serials so reruns cannot
// collide on idx_accounts_email.
func linkWriteTestAccount(t *testing.T, db *gorm.DB, staffID, clinicID uint64, systemAdmin bool) {
	t.Helper()
	account := &model.Account{
		Email:         fmt.Sprintf("link-auth-%d-%d@example.test", staffID, clinicID),
		PasswordHash:  "x",
		IsActive:      true,
		IsSystemAdmin: systemAdmin,
	}
	require.NoError(t, db.Create(account).Error)
	require.NoError(t, db.Model(&model.Staff{}).
		Where("id = ?", staffID).
		Update("account_id", account.ID).Error)
}

// A3: clinics.is_active must be re-read inside the link-write transaction so a
// clinic deactivated after the request-level scope resolution cannot authorize
// the write — including for system-admin staff.
func TestLinkWriteAuth_ClinicActivationRecheckedInsideTx(t *testing.T) {
	db := setupIdentityLinkTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.StaffClinicAssignment{},
		&model.StaffPermissionGroup{},
	))
	fx := testdb.SeedDualClinicGrantFixture(t, db, "link-auth-inactive")
	linkWriteTestAccount(t, db, fx.StaffID, fx.ClinicA, false)
	grantIdentityLinkEdit(t, db, fx.StaffID, fx.ClinicA, fx.ClinicB)

	authorizer := NewDBLinkWriteAuthorizer(db)
	tx := persistence.NewTransactor(db)
	decisionFor := func(t *testing.T, staffID uint64, clinics ...uint64) LinkWriteDecision {
		t.Helper()
		var decision LinkWriteDecision
		require.NoError(t, tx.WithTx(context.Background(), func(txCtx context.Context) error {
			var err error
			decision, err = authorizer.AuthorizeLinkWrite(txCtx, staffID, clinics)
			return err
		}))
		return decision
	}

	t.Run("active affected clinics pass the recheck", func(t *testing.T) {
		decision := decisionFor(t, fx.StaffID, fx.ClinicA, fx.ClinicB)
		assert.Empty(t, decision.InactiveClinicIDs)
		assert.Empty(t, decision.NonMemberClinicIDs)
		assert.Empty(t, decision.EditDeniedClinicIDs)
	})

	// Deactivate clinic B after the request-level scope was resolved.
	require.NoError(t, db.Model(&model.Clinic{}).
		Where("id = ?", fx.ClinicB).
		Update("is_active", false).Error)

	t.Run("deactivated clinic lands in InactiveClinicIDs", func(t *testing.T) {
		decision := decisionFor(t, fx.StaffID, fx.ClinicA, fx.ClinicB)
		assert.Equal(t, []uint64{fx.ClinicB}, decision.InactiveClinicIDs)
		assert.Empty(t, decision.NonMemberClinicIDs)
		assert.Empty(t, decision.EditDeniedClinicIDs)
	})

	t.Run("system admin cannot bypass the deactivated clinic", func(t *testing.T) {
		admin := &model.Staff{
			ClinicID:  fx.ClinicA,
			Name:      "link-auth-admin",
			StaffType: model.StaffTypeDoctor,
			IsActive:  true,
		}
		require.NoError(t, db.Create(admin).Error)
		linkWriteTestAccount(t, db, admin.ID, fx.ClinicB, true)

		decision := decisionFor(t, admin.ID, fx.ClinicA, fx.ClinicB)
		assert.True(t, decision.SystemAdmin)
		assert.Equal(t, []uint64{fx.ClinicB}, decision.InactiveClinicIDs)
	})

	t.Run("write request rejects the deactivated clinic", func(t *testing.T) {
		repo := NewRepository(db)
		svc := NewService(repo, tx, audit.NewService(audit.NewRepository(db)), authorizer)
		ownerA := seedClinicOwner(t, db, fx.ClinicA, "ActiveClinicOwner")
		ownerB := seedClinicOwner(t, db, fx.ClinicB, "InactiveClinicOwner")
		actor := ActorContext{
			StaffID:         fx.StaffID,
			HomeClinicID:    fx.ClinicA,
			VerifiedClinics: []uint64{fx.ClinicA, fx.ClinicB},
			IPAddress:       "127.0.0.1",
			UserAgent:       "test",
		}
		_, _, err := svc.CreateOwnerGroup(context.Background(), actor, []OwnerMemberRef{
			{ClinicID: fx.ClinicA, OwnerID: ownerA.ID},
			{ClinicID: fx.ClinicB, OwnerID: ownerB.ID},
		})
		require.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Contains(t, err.Error(), errLinkWriteClinicInactive)

		var groups int64
		require.NoError(t, db.Model(&model.OwnerIdentityGroup{}).Count(&groups).Error)
		assert.Equal(t, int64(0), groups, "no partial write on deactivated clinic")
	})

	// Restore for other suites sharing this DB handle.
	require.NoError(t, db.Model(&model.Clinic{}).
		Where("id = ?", fx.ClinicB).
		Update("is_active", true).Error)
}
