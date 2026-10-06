package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestLabMappingDraftRepository_WritesRequireAmbientTx(t *testing.T) {
	repo := NewLabMappingDraftRepository(nil)
	ctx := context.Background()
	require.Error(t, repo.SetGenerationStopped(ctx, 1, true))
	require.Error(t, repo.InsertDraft(ctx, &model.LabDeviceMappingDraft{ClinicID: 1}, nil))
	_, _, err := repo.LockDraft(ctx, 1, 1)
	require.Error(t, err)
	require.Error(t, repo.MarkConfirmed(ctx, 1, 1, 1, time.Now()))
	require.Error(t, repo.UpdateConfirmedItems(ctx, 1, 1, nil))
	require.Error(t, repo.UpsertDeviceMaster(ctx, model.LabDeviceItemMaster{}))
	require.Error(t, repo.UpsertSpeciesSpec(ctx, model.LabDeviceMappingSpec{}))
}

func TestFindMappingSpecs_EmptySpeciesDoesNotTouchDB(t *testing.T) {
	repo := NewLabDeviceItemMasterRepository(nil)
	rows, err := repo.FindMappingSpecs(context.Background(), 1, "fuji_nx600", "", []string{"BUN-P"})
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestLabMappingDraftRepository_ConfirmAndSpeciesOverlay(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.LabDeviceItemMaster{},
		&model.LabDeviceMappingDraft{},
		&model.LabDeviceMappingDraftItem{},
		&model.LabDeviceMappingSpec{},
		&model.LabDeviceMappingGeneration{},
	))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	examA := &model.ExaminationType{ClinicID: clinicA, Name: "mapping-a"}
	require.NoError(t, db.Create(examA).Error)
	fieldA := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examA.ID, Name: "BUN", Unit: "mg/dl"}
	require.NoError(t, db.Create(fieldA).Error)
	examB := &model.ExaminationType{ClinicID: clinicB, Name: "mapping-b"}
	require.NoError(t, db.Create(examB).Error)
	fieldB := &model.ExamTypeField{ClinicID: clinicB, ExamTypeID: examB.ID, Name: "BUN", Unit: "mg/dl"}
	require.NoError(t, db.Create(fieldB).Error)
	t.Cleanup(func() {
		db.Exec("DELETE FROM lab_device_mapping_draft_items WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_specs WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_drafts WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_generation WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_item_masters WHERE clinic_id IN (1,2) AND device_item_code IN ('NEW','KEEP')")
		db.Exec("DELETE FROM exam_type_fields WHERE id IN (?,?)", fieldA.ID, fieldB.ID)
		db.Exec("DELETE FROM exam_types WHERE id IN (?,?)", examA.ID, examB.ID)
	})

	store := NewLabMappingDraftRepository(db)
	masters := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	tx := persistence.NewTransactor(db)
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return store.InsertDraft(txCtx, &model.LabDeviceMappingDraft{
			ClinicID: clinicA, SourceType: "fuji_nx600", Status: model.LabMappingDraftStatusUnconfirmed,
			RawPayload: []byte{0x02, 0x03}, RawFingerprint: "draft-new",
		}, []model.LabDeviceMappingDraftItem{{
			DeviceItemCode: "NEW", ItemSource: model.LabMappingItemSourceBlank, ValueRaw: mappingSentinelValue,
		}})
	}))

	res, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, res.Mapped)

	var draft model.LabDeviceMappingDraft
	require.NoError(t, db.Where("clinic_id = ? AND raw_fingerprint = ?", clinicA, "draft-new").First(&draft).Error)
	require.NoError(t, db.Create(&model.LabDeviceMappingSpec{
		ClinicID: clinicA, SourceType: "fuji_nx600", Species: "canine", DeviceItemCode: "NEW",
		Unit: "g/dl", ValueShape: model.LabDeviceValueShapeNumeric, ExamTypeFieldID: fieldA.ID, ConfirmedDraftID: draft.ID,
	}).Error)
	canine, err := masters.ResolveItemsForSpecies(ctx, clinicA, "fuji_nx600", "canine", []string{"NEW"})
	require.NoError(t, err)
	require.Len(t, canine.Mapped, 1)
	assert.Equal(t, fieldA.ID, canine.Mapped[0].ExamTypeFieldID)
	plain, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, plain.Mapped)
	feline, err := masters.ResolveItemsForSpecies(ctx, clinicA, "fuji_nx600", "feline", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, feline.Mapped)

	require.NoError(t, db.Create(&model.LabDeviceItemMaster{
		ClinicID: clinicA, SourceType: "fuji_nx600", DeviceItemCode: "KEEP",
		Unit: "mg/dl", ValueShape: model.LabDeviceValueShapeNumeric, ExamTypeFieldID: &fieldA.ID, IsActive: true,
	}).Error)
	audit := &scriptAudit{}
	svc := NewLabMappingDraftService(store, masters, nil, audit, tx)
	require.NoError(t, svc.SetGenerationStopped(ctx, clinicA, true))
	kept, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"KEEP"})
	require.NoError(t, err)
	require.Len(t, kept.Mapped, 1)

	_, err = svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldB.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.True(t, apperrors.IsInvalidInput(err))

	audit.err = assert.AnError
	_, err = svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldA.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.Error(t, err)
	var masterCount int64
	require.NoError(t, db.Model(&model.LabDeviceItemMaster{}).
		Where("clinic_id = ? AND device_item_code = ?", clinicA, "NEW").Count(&masterCount).Error)
	assert.Zero(t, masterCount)
	var still model.LabDeviceMappingDraft
	require.NoError(t, db.First(&still, draft.ID).Error)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, still.Status)

	audit.err = nil
	confirmed, err := svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldA.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusConfirmed, confirmed.Draft.Status)
	require.NoError(t, db.Model(&model.LabDeviceItemMaster{}).
		Where("clinic_id = ? AND device_item_code = ?", clinicA, "NEW").Count(&masterCount).Error)
	assert.Equal(t, int64(1), masterCount)
}
