package medicalrecord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestAssertSingleExamType(t *testing.T) {
	t.Run("empty and single exam type are allowed", func(t *testing.T) {
		require.NoError(t, AssertSingleExamType(nil))
		require.NoError(t, AssertSingleExamType([]LabDeviceResolvedItem{
			{DeviceItemCode: "Na-P", ExamTypeID: 10, ExamTypeFieldID: 1},
			{DeviceItemCode: "K-P", ExamTypeID: 10, ExamTypeFieldID: 2},
		}))
	})
	t.Run("two exam types are rejected", func(t *testing.T) {
		err := AssertSingleExamType([]LabDeviceResolvedItem{
			{DeviceItemCode: "Na-P", ExamTypeID: 10, ExamTypeFieldID: 1},
			{DeviceItemCode: "vf-SAA", ExamTypeID: 20, ExamTypeFieldID: 3},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), LabDeviceMultipleExamTypesMessage)
		assert.Equal(t, []uint64{10, 20}, UniqueMappedExamTypeIDs([]LabDeviceResolvedItem{
			{ExamTypeID: 10}, {ExamTypeID: 20}, {ExamTypeID: 10},
		}))
	})
}

func setupLabDeviceItemMasterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.LabDeviceItemMaster{},
	))
	require.NoError(t, db.Exec(`ALTER TABLE lab_device_item_masters DROP COLUMN IF EXISTS display_name`).Error)
	db.Exec("TRUNCATE TABLE lab_device_item_masters, exam_type_fields, exam_types CASCADE")
	return db
}

func TestLabDeviceItemMasterService_EnsureDefaultsAndIsolation(t *testing.T) {
	db := setupLabDeviceItemMasterTestDB(t)
	svc := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	inserted, items, err := svc.EnsureDefaults(ctx, clinicA)
	require.NoError(t, err)
	assert.Equal(t, int64(LabDeviceItemCatalogCount), inserted)
	require.Len(t, items, LabDeviceItemCatalogCount)

	var bun *model.LabDeviceItemMaster
	for i := range items {
		if items[i].DeviceItemCode == "BUN-P" {
			bun = &items[i]
			break
		}
	}
	require.NotNil(t, bun)
	assert.Nil(t, bun.ExamTypeFieldID)

	examA := &model.ExaminationType{ClinicID: clinicA, Name: "血液A"}
	require.NoError(t, db.Create(examA).Error)
	fieldA := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examA.ID, Name: "BUN"}
	require.NoError(t, db.Create(fieldA).Error)
	_, err = svc.Update(ctx, clinicA, bun.ID, UpdateLabDeviceItemMasterInput{
		Unit:            bun.Unit,
		ExamTypeFieldID: &fieldA.ID,
		IsActive:        true,
	})
	require.NoError(t, err)

	inserted2, items2, err := svc.EnsureDefaults(ctx, clinicA)
	require.NoError(t, err)
	assert.Equal(t, int64(0), inserted2)
	require.Len(t, items2, LabDeviceItemCatalogCount)
	for _, item := range items2 {
		if item.DeviceItemCode == "BUN-P" {
			require.NotNil(t, item.ExamTypeFieldID)
			assert.Equal(t, fieldA.ID, *item.ExamTypeFieldID, "ensure must not overwrite hospital exam_type_field_id")
		}
	}

	other, err := svc.List(ctx, clinicB, "")
	require.NoError(t, err)
	assert.Empty(t, other)

	_, err = svc.Update(ctx, clinicB, bun.ID, UpdateLabDeviceItemMasterInput{
		IsActive: true,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestLabDeviceItemMasterService_UpdateAndResolve(t *testing.T) {
	db := setupLabDeviceItemMasterTestDB(t)
	svc := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	_, items, err := svc.EnsureDefaults(ctx, clinicA)
	require.NoError(t, err)
	var bun *model.LabDeviceItemMaster
	for i := range items {
		if items[i].DeviceItemCode == "BUN-P" {
			bun = &items[i]
			break
		}
	}
	require.NotNil(t, bun)

	examA := &model.ExaminationType{ClinicID: clinicA, Name: "血液A"}
	require.NoError(t, db.Create(examA).Error)
	fieldA := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examA.ID, Name: "BUN"}
	require.NoError(t, db.Create(fieldA).Error)
	examB := &model.ExaminationType{ClinicID: clinicB, Name: "血液B"}
	require.NoError(t, db.Create(examB).Error)
	fieldB := &model.ExamTypeField{ClinicID: clinicB, ExamTypeID: examB.ID, Name: "BUN"}
	require.NoError(t, db.Create(fieldB).Error)

	_, err = svc.Update(ctx, clinicA, bun.ID, UpdateLabDeviceItemMasterInput{
		Unit:            bun.Unit,
		ExamTypeFieldID: &fieldB.ID,
		IsActive:        true,
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	updated, err := svc.Update(ctx, clinicA, bun.ID, UpdateLabDeviceItemMasterInput{
		Unit:            bun.Unit,
		ExamTypeFieldID: &fieldA.ID,
		IsActive:        true,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.ExamTypeFieldID)
	assert.Equal(t, fieldA.ID, *updated.ExamTypeFieldID)

	res, err := svc.ResolveItems(ctx, clinicA, string(model.LabImportSourceTypeFujiNX600), []string{"BUN-P", "GOT-P"})
	require.NoError(t, err)
	require.Len(t, res.Mapped, 1)
	assert.Equal(t, "BUN-P", res.Mapped[0].DeviceItemCode)
	assert.Equal(t, examA.ID, res.Mapped[0].ExamTypeID)
	assert.Equal(t, []string{"GOT-P"}, res.UnmappedCodes)
	require.NoError(t, AssertSingleExamType(res.Mapped))

	cleared, err := svc.Update(ctx, clinicA, bun.ID, UpdateLabDeviceItemMasterInput{
		Unit:            bun.Unit,
		ExamTypeFieldID: nil,
		IsActive:        true,
	})
	require.NoError(t, err)
	assert.Nil(t, cleared.ExamTypeFieldID)

	res2, err := svc.ResolveItems(ctx, clinicA, string(model.LabImportSourceTypeFujiNX600), []string{"BUN-P"})
	require.NoError(t, err)
	assert.Empty(t, res2.Mapped)
	assert.Equal(t, []string{"BUN-P"}, res2.UnmappedCodes)
}
