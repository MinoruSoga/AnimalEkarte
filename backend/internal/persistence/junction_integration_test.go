package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type persistenceTestMaster struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ClinicID  uint64 `gorm:"not null;index"`
	Name      string
	DeletedAt gorm.DeletedAt
}

func (persistenceTestMaster) TableName() string {
	return "persistence_test_masters"
}

type persistenceTestJunction struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ClinicID  uint64 `gorm:"not null;index"`
	StaffID   uint64 `gorm:"not null;index"`
	MasterID  uint64 `gorm:"not null;index"`
	Value     string
	DeletedAt gorm.DeletedAt
}

func (persistenceTestJunction) TableName() string {
	return "persistence_test_junctions"
}

type persistenceTestChild struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	ParentID  uint64 `gorm:"not null;uniqueIndex:ux_persistence_child_parent_value"`
	Value     string `gorm:"not null;uniqueIndex:ux_persistence_child_parent_value"`
	DeletedAt gorm.DeletedAt
}

func (persistenceTestChild) TableName() string {
	return "persistence_test_children"
}

func setupJunctionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&persistenceTestMaster{},
		&persistenceTestJunction{},
		&persistenceTestChild{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE persistence_test_junctions, persistence_test_children, persistence_test_masters RESTART IDENTITY CASCADE",
	).Error)
	return db
}

func TestReplaceChildRowsByParentIDReplacesOnlyTheSelectedParent(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, db.Create(&[]persistenceTestChild{
		{ParentID: 10, Value: "old"},
		{ParentID: 20, Value: "foreign"},
	}).Error)

	rows := []persistenceTestChild{
		{ParentID: 999, Value: "first"},
		{ParentID: 999, Value: "second"},
	}
	err := ReplaceChildRowsByParentID(
		db,
		10,
		rows,
		&persistenceTestChild{},
		"parent_id",
		"persistence_test_child",
		func(row *persistenceTestChild, parentID uint64) {
			row.ParentID = parentID
		},
		"replace children",
	)
	require.NoError(t, err)

	var own []persistenceTestChild
	require.NoError(t, db.Where("parent_id = ?", 10).Order("value").Find(&own).Error)
	require.Len(t, own, 2)
	assert.Equal(t, []string{"first", "second"}, []string{own[0].Value, own[1].Value})
	assert.Equal(t, uint64(10), own[0].ParentID)
	assert.Equal(t, uint64(10), own[1].ParentID)

	var foreignCount int64
	require.NoError(t, db.Model(&persistenceTestChild{}).
		Where("parent_id = ? AND value = ?", 20, "foreign").
		Count(&foreignCount).Error)
	assert.Equal(t, int64(1), foreignCount)
}

func TestReplaceChildRowsByParentIDRollsBackDeleteWhenInsertFails(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, db.Create(&persistenceTestChild{
		ParentID: 10,
		Value:    "old",
	}).Error)

	rows := []persistenceTestChild{
		{Value: "duplicate"},
		{Value: "duplicate"},
	}
	err := ReplaceChildRowsByParentID(
		db,
		10,
		rows,
		&persistenceTestChild{},
		"parent_id",
		"persistence_test_child",
		func(row *persistenceTestChild, parentID uint64) {
			row.ParentID = parentID
		},
		"replace children",
	)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&persistenceTestChild{}).
		Where("parent_id = ? AND value = ?", 10, "old").
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "the delete must roll back with the failed batch insert")
}

func TestReplaceChildRowsByParentIDAcceptsAnEmptyReplacement(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, db.Create(&persistenceTestChild{
		ParentID: 10,
		Value:    "old",
	}).Error)

	require.NoError(t, ReplaceChildRowsByParentID(
		db,
		10,
		[]persistenceTestChild{},
		&persistenceTestChild{},
		"parent_id",
		"persistence_test_child",
		func(row *persistenceTestChild, parentID uint64) {
			row.ParentID = parentID
		},
		"replace children",
	))

	var count int64
	require.NoError(t, db.Model(&persistenceTestChild{}).
		Where("parent_id = ?", 10).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestValidateClinicScopedMasterIDsRejectsForeignAndDeletedMasters(t *testing.T) {
	db := setupJunctionTestDB(t)
	ctx := context.Background()
	own := &persistenceTestMaster{ClinicID: 1, Name: "own"}
	foreign := &persistenceTestMaster{ClinicID: 2, Name: "foreign"}
	deleted := &persistenceTestMaster{ClinicID: 1, Name: "deleted"}
	require.NoError(t, db.Create(own).Error)
	require.NoError(t, db.Create(foreign).Error)
	require.NoError(t, db.Create(deleted).Error)
	require.NoError(t, db.Delete(deleted).Error)

	require.NoError(t, ValidateClinicScopedMasterIDs(
		ctx,
		db,
		1,
		nil,
		&persistenceTestMaster{},
		"persistence_test_master",
		"invalid master",
	))
	require.NoError(t, ValidateClinicScopedMasterIDs(
		ctx,
		db,
		1,
		[]uint64{own.ID},
		&persistenceTestMaster{},
		"persistence_test_master",
		"invalid master",
	))

	for _, id := range []uint64{foreign.ID, deleted.ID} {
		err := ValidateClinicScopedMasterIDs(
			ctx,
			db,
			1,
			[]uint64{id},
			&persistenceTestMaster{},
			"persistence_test_master",
			"invalid master",
		)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	}
}

func TestDeleteJunctionViaMasterClinicScopeDoesNotCrossClinics(t *testing.T) {
	db := setupJunctionTestDB(t)
	own := &persistenceTestMaster{ClinicID: 1, Name: "own"}
	foreign := &persistenceTestMaster{ClinicID: 2, Name: "foreign"}
	require.NoError(t, db.Create(own).Error)
	require.NoError(t, db.Create(foreign).Error)
	require.NoError(t, db.Create(&[]persistenceTestJunction{
		{ClinicID: 1, StaffID: 7, MasterID: own.ID, Value: "own"},
		{ClinicID: 2, StaffID: 7, MasterID: foreign.ID, Value: "foreign"},
	}).Error)

	require.NoError(t, DeleteJunctionViaMasterClinicScope(
		db,
		1,
		7,
		&persistenceTestJunction{},
		&persistenceTestMaster{},
		"master_id",
		"persistence_test_junction",
		"staff 7",
	))

	var ownCount, foreignCount int64
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("master_id = ?", own.ID).
		Count(&ownCount).Error)
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("master_id = ?", foreign.ID).
		Count(&foreignCount).Error)
	assert.Zero(t, ownCount)
	assert.Equal(t, int64(1), foreignCount)
}

func TestDeleteJunctionByClinicAndStaffDoesNotCrossClinics(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, db.Create(&[]persistenceTestJunction{
		{ClinicID: 1, StaffID: 7, MasterID: 10, Value: "own"},
		{ClinicID: 2, StaffID: 7, MasterID: 20, Value: "foreign"},
	}).Error)

	require.NoError(t, DeleteJunctionByClinicAndStaff(
		db,
		1,
		7,
		&persistenceTestJunction{},
		"persistence_test_junction",
		"staff 7",
	))

	var ownCount, foreignCount int64
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("clinic_id = ? AND staff_id = ?", 1, 7).
		Count(&ownCount).Error)
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("clinic_id = ? AND staff_id = ?", 2, 7).
		Count(&foreignCount).Error)
	assert.Zero(t, ownCount)
	assert.Equal(t, int64(1), foreignCount)
}

func TestInsertJunctionRowsInBatchesHandlesEmptyAndNonEmptyInput(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, InsertJunctionRowsInBatches(
		db,
		[]persistenceTestJunction{},
		"persistence_test_junction",
		"empty",
	))
	require.NoError(t, InsertJunctionRowsInBatches(
		db,
		[]persistenceTestJunction{
			{ClinicID: 1, StaffID: 7, MasterID: 10, Value: "first"},
			{ClinicID: 1, StaffID: 7, MasterID: 20, Value: "second"},
		},
		"persistence_test_junction",
		"staff 7",
	))

	var count int64
	require.NoError(t, db.Model(&persistenceTestJunction{}).Count(&count).Error)
	assert.Equal(t, int64(2), count)
}

func TestReplaceJunctionInTransactionCommitsOrRollsBackAsOneUnit(t *testing.T) {
	db := setupJunctionTestDB(t)
	sentinel := errors.New("force rollback")

	err := ReplaceJunctionInTransaction(db, func(tx *gorm.DB) error {
		require.NoError(t, tx.Create(&persistenceTestJunction{
			ClinicID: 1,
			StaffID:  7,
			MasterID: 10,
			Value:    "rollback",
		}).Error)
		return sentinel
	}, "replace junction")
	require.ErrorIs(t, err, sentinel)

	var count int64
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("value = ?", "rollback").
		Count(&count).Error)
	assert.Zero(t, count)

	require.NoError(t, ReplaceJunctionInTransaction(db, func(tx *gorm.DB) error {
		return tx.Create(&persistenceTestJunction{
			ClinicID: 1,
			StaffID:  7,
			MasterID: 10,
			Value:    "commit",
		}).Error
	}, "replace junction"))
	require.NoError(t, db.Model(&persistenceTestJunction{}).
		Where("value = ?", "commit").
		Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestPaginateAppliesStableOffsetAndLimit(t *testing.T) {
	db := setupJunctionTestDB(t)
	require.NoError(t, db.Create(&[]persistenceTestMaster{
		{ClinicID: 1, Name: "one"},
		{ClinicID: 1, Name: "two"},
		{ClinicID: 1, Name: "three"},
		{ClinicID: 1, Name: "four"},
		{ClinicID: 1, Name: "five"},
	}).Error)

	var page []persistenceTestMaster
	require.NoError(t, db.Order("id").
		Scopes(Paginate(2, 2)).
		Find(&page).Error)
	require.Len(t, page, 2)
	assert.Equal(t, []string{"three", "four"}, []string{page[0].Name, page[1].Name})
}
