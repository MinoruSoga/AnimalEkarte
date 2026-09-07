package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	domainaudit "github.com/animal-ekarte/backend/internal/audit"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errPermissionAuditRollback = errors.New(
	"force permission mutation rollback after audit insert",
)

type rollbackAfterPermissionAudit struct {
	inner domainaudit.TxLogger
}

func writePermissionAuditTx(
	ctx context.Context,
	inner domainaudit.TxLogger,
	entry authdomain.AuthAuditEntry,
) error {
	return inner.LogEntryTx(ctx, &domainaudit.Entry{
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

func (a rollbackAfterPermissionAudit) LogEntryTx(
	ctx context.Context,
	entry authdomain.AuthAuditEntry,
) error {
	if err := writePermissionAuditTx(ctx, a.inner, entry); err != nil {
		return err
	}
	return errPermissionAuditRollback
}

type committedPermissionAudit struct {
	inner domainaudit.TxLogger
}

func (a committedPermissionAudit) LogEntryTx(
	ctx context.Context,
	entry authdomain.AuthAuditEntry,
) error {
	return writePermissionAuditTx(ctx, a.inner, entry)
}

type pausingPermissionGroupRepository struct {
	authdomain.PermissionGroupRepository
	locker  authdomain.PermissionGroupMutationLocker
	locked  chan struct{}
	release <-chan struct{}
	once    sync.Once
}

func (r *pausingPermissionGroupRepository) LockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	group, err := r.locker.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	r.once.Do(func() { close(r.locked) })
	select {
	case <-r.release:
		return group, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type observingPermissionGroupRepository struct {
	authdomain.PermissionGroupRepository
	locker    authdomain.PermissionGroupMutationLocker
	attempted chan struct{}
	once      sync.Once
}

func (r *observingPermissionGroupRepository) LockByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.PermissionGroup, error) {
	r.once.Do(func() { close(r.attempted) })
	return r.locker.LockByIDForUpdate(ctx, clinicID, id)
}

func setupPermissionAuditRollbackDB(t *testing.T) (*gorm.DB, uint64) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
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
		"staffs",
		"clinics",
		"companies",
	)
	company := &model.Company{Name: "permission audit rollback company"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "permission audit rollback clinic",
	}
	require.NoError(t, db.Create(clinic).Error)
	require.NoError(t, db.Create(&model.Staff{
		ID:       17,
		ClinicID: clinic.ID,
		Name:     "permission audit actor",
		IsActive: true,
	}).Error)
	return db, clinic.ID
}

func permissionAuditRollbackService(
	db *gorm.DB,
) authdomain.PermissionGroupApplication {
	auditKernel := domainaudit.NewService(domainaudit.NewRepository(db))
	return authdomain.NewPermissionGroupService(
		authdomain.NewPermissionGroupRepository(db),
		persistence.NewTransactor(db),
		rollbackAfterPermissionAudit{inner: auditKernel},
	)
}

func permissionAuditRollbackInput(
	clinicID uint64,
	action, resource string,
) authdomain.PermissionMutationAudit {
	return authdomain.PermissionMutationAudit{
		ClinicID:     clinicID,
		ActorStaffID: 17,
		Action:       action,
		Resource:     resource,
		IPAddress:    "127.0.0.1",
		UserAgent:    "permission-audit-db-test",
	}
}

func assertPermissionAuditRolledBack(
	t *testing.T,
	db *gorm.DB,
	action string,
) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&model.AuditLog{}).
		Where("action = ?", action).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestPermissionGroupAuditDB_RollsBackDomainAndAuditRows(
	t *testing.T,
) {
	tests := []struct {
		name     string
		action   string
		resource string
		invoke   func(
			context.Context,
			authdomain.PermissionGroupApplication,
			*gorm.DB,
			uint64,
		) error
		assertDomain func(*testing.T, *gorm.DB, uint64)
	}{
		{
			name:     "create",
			action:   model.AuditActionPermissionGroupCreate,
			resource: "permission_group",
			invoke: func(
				ctx context.Context,
				app authdomain.PermissionGroupApplication,
				_ *gorm.DB,
				clinicID uint64,
			) error {
				_, err := app.Create(
					ctx,
					clinicID,
					&authdomain.CreatePermissionGroupInput{
						Name:  "rolled back create",
						Color: "#123456",
					},
					permissionAuditRollbackInput(
						clinicID,
						model.AuditActionPermissionGroupCreate,
						"permission_group",
					),
				)
				return err
			},
			assertDomain: func(t *testing.T, db *gorm.DB, clinicID uint64) {
				var count int64
				require.NoError(t, db.Model(&model.PermissionGroup{}).
					Where("clinic_id = ? AND name = ?", clinicID, "rolled back create").
					Count(&count).Error)
				assert.Zero(t, count)
			},
		},
		{
			name:     "update",
			action:   model.AuditActionPermissionGroupUpdate,
			resource: "permission_group",
			invoke: func(
				ctx context.Context,
				app authdomain.PermissionGroupApplication,
				db *gorm.DB,
				clinicID uint64,
			) error {
				group := &model.PermissionGroup{
					ClinicID: clinicID,
					Name:     "before update",
					Color:    "#123456",
				}
				require.NoError(t, db.Create(group).Error)
				updatedName := "rolled back update"
				_, err := app.Update(
					ctx,
					clinicID,
					group.ID,
					&authdomain.UpdatePermissionGroupInput{Name: &updatedName},
					permissionAuditRollbackInput(
						clinicID,
						model.AuditActionPermissionGroupUpdate,
						"permission_group",
					),
				)
				return err
			},
			assertDomain: func(t *testing.T, db *gorm.DB, clinicID uint64) {
				var group model.PermissionGroup
				require.NoError(t, db.Where(
					"clinic_id = ? AND name = ?",
					clinicID,
					"before update",
				).First(&group).Error)
			},
		},
		{
			name:     "delete",
			action:   model.AuditActionPermissionGroupDelete,
			resource: "permission_group",
			invoke: func(
				ctx context.Context,
				app authdomain.PermissionGroupApplication,
				db *gorm.DB,
				clinicID uint64,
			) error {
				group := &model.PermissionGroup{
					ClinicID: clinicID,
					Name:     "rolled back delete",
					Color:    "#123456",
				}
				require.NoError(t, db.Create(group).Error)
				return app.Delete(
					ctx,
					clinicID,
					group.ID,
					permissionAuditRollbackInput(
						clinicID,
						model.AuditActionPermissionGroupDelete,
						"permission_group",
					),
				)
			},
			assertDomain: func(t *testing.T, db *gorm.DB, clinicID uint64) {
				var count int64
				require.NoError(t, db.Model(&model.PermissionGroup{}).
					Where("clinic_id = ? AND name = ?", clinicID, "rolled back delete").
					Count(&count).Error)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:     "rules",
			action:   model.AuditActionPermissionRulesUpdate,
			resource: "permission_group_rules",
			invoke: func(
				ctx context.Context,
				app authdomain.PermissionGroupApplication,
				db *gorm.DB,
				clinicID uint64,
			) error {
				group := &model.PermissionGroup{
					ClinicID: clinicID,
					Name:     "rolled back rules",
					Color:    "#123456",
				}
				require.NoError(t, db.Create(group).Error)
				require.NoError(t, db.Create(&model.PermissionGroupRule{
					GroupID:  group.ID,
					Resource: "owners",
					CanView:  true,
				}).Error)
				_, err := app.UpdateRules(
					ctx,
					clinicID,
					group.ID,
					[]authdomain.SetPermissionGroupRulesInput{{
						Resource: "reservations",
						CanView:  true,
					}},
					17,
					permissionAuditRollbackInput(
						clinicID,
						model.AuditActionPermissionRulesUpdate,
						"permission_group_rules",
					),
				)
				return err
			},
			assertDomain: func(t *testing.T, db *gorm.DB, clinicID uint64) {
				var rules []model.PermissionGroupRule
				require.NoError(t, db.Joins(
					"JOIN permission_groups ON permission_groups.id = permission_group_rules.group_id",
				).Where(
					"permission_groups.clinic_id = ?",
					clinicID,
				).Find(&rules).Error)
				require.Len(t, rules, 1)
				assert.Equal(t, "owners", rules[0].Resource)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, clinicID := setupPermissionAuditRollbackDB(t)
			app := permissionAuditRollbackService(db)

			err := test.invoke(context.Background(), app, db, clinicID)

			require.ErrorIs(t, err, errPermissionAuditRollback)
			test.assertDomain(t, db, clinicID)
			assertPermissionAuditRolledBack(t, db, test.action)
		})
	}
}

func TestPermissionGroupAuditDB_SerializesAuditOldSnapshot(
	t *testing.T,
) {
	db, clinicID := setupPermissionAuditRollbackDB(t)
	group := &model.PermissionGroup{
		ClinicID: clinicID,
		Name:     "A",
		Color:    "#123456",
	}
	require.NoError(t, db.Create(group).Error)

	baseRepo := authdomain.NewPermissionGroupRepository(db)
	auditKernel := domainaudit.NewService(domainaudit.NewRepository(db))
	audit := committedPermissionAudit{inner: auditKernel}
	transactor := persistence.NewTransactor(db)
	releaseFirst := make(chan struct{})
	firstLocked := make(chan struct{})
	secondAttempted := make(chan struct{})
	first := authdomain.NewPermissionGroupService(
		&pausingPermissionGroupRepository{
			PermissionGroupRepository: baseRepo,
			locker:                    baseRepo,
			locked:                    firstLocked,
			release:                   releaseFirst,
		},
		transactor,
		audit,
	)
	second := authdomain.NewPermissionGroupService(
		&observingPermissionGroupRepository{
			PermissionGroupRepository: baseRepo,
			locker:                    baseRepo,
			attempted:                 secondAttempted,
		},
		transactor,
		audit,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	firstResult := make(chan error, 1)
	nameB := "B"
	go func() {
		_, err := first.Update(
			ctx,
			clinicID,
			group.ID,
			&authdomain.UpdatePermissionGroupInput{Name: &nameB},
			permissionAuditRollbackInput(
				clinicID,
				model.AuditActionPermissionGroupUpdate,
				"permission_group",
			),
		)
		firstResult <- err
	}()

	select {
	case <-firstLocked:
	case <-ctx.Done():
		t.Fatal("first mutation did not acquire the parent-row lock")
	}

	secondResult := make(chan error, 1)
	nameC := "C"
	go func() {
		_, err := second.Update(
			ctx,
			clinicID,
			group.ID,
			&authdomain.UpdatePermissionGroupInput{Name: &nameC},
			permissionAuditRollbackInput(
				clinicID,
				model.AuditActionPermissionGroupUpdate,
				"permission_group",
			),
		)
		secondResult <- err
	}()

	select {
	case <-secondAttempted:
	case <-ctx.Done():
		t.Fatal("second mutation did not attempt the parent-row lock")
	}
	select {
	case err := <-secondResult:
		t.Fatalf("second mutation bypassed the parent-row lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	close(releaseFirst)
	require.NoError(t, <-firstResult)
	require.NoError(t, <-secondResult)

	var persisted model.PermissionGroup
	require.NoError(t, db.First(&persisted, group.ID).Error)
	assert.Equal(t, "C", persisted.Name)

	var logs []model.AuditLog
	require.NoError(t, db.
		Where(
			"action = ? AND resource_id = ?",
			model.AuditActionPermissionGroupUpdate,
			group.ID,
		).
		Order("id ASC").
		Find(&logs).
		Error)
	require.Len(t, logs, 2)
	transitions := make(map[string]string, len(logs))
	for i := range logs {
		var oldSnapshot struct {
			Name string `json:"name"`
		}
		var newSnapshot struct {
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(logs[i].OldValue, &oldSnapshot))
		require.NoError(t, json.Unmarshal(logs[i].NewValue, &newSnapshot))
		transitions[newSnapshot.Name] = oldSnapshot.Name
	}
	assert.Equal(t, "A", transitions["B"])
	assert.Equal(t, "B", transitions["C"])
}
