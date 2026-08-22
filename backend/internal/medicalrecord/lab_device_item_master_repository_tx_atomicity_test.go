package medicalrecord

// lab_device_item_master_repository_tx_atomicity_test.go — F-5
//
// labDeviceItemMasterRepository の DBOrTx 参加を実証する。writer は ambient WithTx の
// sentinel rollback で永続化されない。reader は同一 tx の未コミット write を見て、
// rollback 後は 0 件になる。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errLabDeviceMasterTxSentinel = errors.New("simulated post-write failure in ambient tx")

func newLabDeviceCatalogRow(clinicID uint64, code string) model.LabDeviceItemMaster {
	return model.LabDeviceItemMaster{
		ClinicID:       clinicID,
		SourceType:     string(model.LabImportSourceTypeFujiNX600),
		DeviceItemCode: code,
		Unit:           "mg/dl",
		ValueShape:     model.LabDeviceValueShapeNumeric,
		SortOrder:      1,
		IsActive:       true,
	}
}

func newLabDeviceRow(clinicID uint64, name string) *model.LabDevice {
	return &model.LabDevice{
		ClinicID:   clinicID,
		SourceType: string(model.LabImportSourceTypeFujiNX600),
		Name:       name,
		IsActive:   true,
		SortOrder:  10,
	}
}

func TestLabDeviceItemMasterRepository_WritersRollBackWhenAmbientTxFails(t *testing.T) {
	const clinicID = uint64(1)

	tests := []struct {
		name          string
		seed          func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository)
		write         func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository)
		afterRollback func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository)
	}{
		{
			name: "EnsureCatalog",
			write: func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				inserted, err := repo.EnsureCatalog(txCtx, []model.LabDeviceItemMaster{
					newLabDeviceCatalogRow(clinicID, "TX-BUN-P"),
				})
				require.NoError(t, err)
				assert.EqualValues(t, 1, inserted)
			},
			afterRollback: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				rows, err := repo.List(ctx, clinicID, "")
				require.NoError(t, err)
				assert.Empty(t, rows)
			},
		},
		{
			name: "Update",
			seed: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				_, err := repo.EnsureCatalog(ctx, []model.LabDeviceItemMaster{
					newLabDeviceCatalogRow(clinicID, "TX-BUN-P"),
				})
				require.NoError(t, err)
			},
			write: func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				rows, err := repo.List(txCtx, clinicID, "")
				require.NoError(t, err)
				require.Len(t, rows, 1)
				updated, err := repo.Update(txCtx, clinicID, rows[0].ID, map[string]any{"unit": "g/dl"})
				require.NoError(t, err)
				assert.Equal(t, "g/dl", updated.Unit)
			},
			afterRollback: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				rows, err := repo.List(ctx, clinicID, "")
				require.NoError(t, err)
				require.Len(t, rows, 1)
				assert.Equal(t, "mg/dl", rows[0].Unit)
			},
		},
		{
			name: "CreateDevice",
			write: func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				require.NoError(t, repo.CreateDevice(txCtx, newLabDeviceRow(clinicID, "TX-NX600")))
			},
			afterRollback: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				devices, err := repo.ListDevices(ctx, clinicID)
				require.NoError(t, err)
				assert.Empty(t, devices)
			},
		},
		{
			name: "UpdateDevice",
			seed: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				require.NoError(t, repo.CreateDevice(ctx, newLabDeviceRow(clinicID, "TX-NX600")))
			},
			write: func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				devices, err := repo.ListDevices(txCtx, clinicID)
				require.NoError(t, err)
				require.Len(t, devices, 1)
				updated, err := repo.UpdateDevice(txCtx, clinicID, devices[0].ID, map[string]any{"name": "changed"})
				require.NoError(t, err)
				assert.Equal(t, "changed", updated.Name)
			},
			afterRollback: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				devices, err := repo.ListDevices(ctx, clinicID)
				require.NoError(t, err)
				require.Len(t, devices, 1)
				assert.Equal(t, "TX-NX600", devices[0].Name)
			},
		},
		{
			name: "EnsureDevices",
			write: func(t *testing.T, txCtx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				inserted, err := repo.EnsureDevices(txCtx, []model.LabDevice{*newLabDeviceRow(clinicID, "TX-NX600")})
				require.NoError(t, err)
				assert.EqualValues(t, 1, inserted)
			},
			afterRollback: func(t *testing.T, ctx context.Context, repo LabDeviceItemMasterRepository) {
				t.Helper()
				devices, err := repo.ListDevices(ctx, clinicID)
				require.NoError(t, err)
				assert.Empty(t, devices)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupLabDeviceItemMasterTestDB(t)
			repo := NewLabDeviceItemMasterRepository(db)
			ctx := context.Background()
			if tt.seed != nil {
				tt.seed(t, ctx, repo)
			}

			txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
				tt.write(t, txCtx, repo)
				return errLabDeviceMasterTxSentinel
			})
			require.ErrorIs(t, txErr, errLabDeviceMasterTxSentinel)
			tt.afterRollback(t, ctx, repo)
		})
	}
}

func TestLabDeviceItemMasterRepository_ReadersSeeUncommittedWritesThenRollback(t *testing.T) {
	db := setupLabDeviceItemMasterTestDB(t)
	repo := NewLabDeviceItemMasterRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	catalog := newLabDeviceCatalogRow(clinicID, "TX-BUN-P")
	device := newLabDeviceRow(clinicID, "TX-NX600")

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		inserted, err := repo.EnsureCatalog(txCtx, []model.LabDeviceItemMaster{catalog})
		require.NoError(t, err)
		assert.EqualValues(t, 1, inserted)
		require.NoError(t, repo.CreateDevice(txCtx, device))

		listed, err := repo.List(txCtx, clinicID, "")
		require.NoError(t, err)
		require.Len(t, listed, 1)
		assert.Equal(t, "TX-BUN-P", listed[0].DeviceItemCode)

		found, err := repo.FindByID(txCtx, clinicID, listed[0].ID)
		require.NoError(t, err)
		assert.Equal(t, listed[0].ID, found.ID)

		byCodes, err := repo.FindByClinicSourceCodes(
			txCtx, clinicID, string(model.LabImportSourceTypeFujiNX600), []string{"TX-BUN-P"},
		)
		require.NoError(t, err)
		require.Len(t, byCodes, 1)
		assert.Equal(t, "TX-BUN-P", byCodes[0].DeviceItemCode)

		devices, err := repo.ListDevices(txCtx, clinicID)
		require.NoError(t, err)
		require.Len(t, devices, 1)
		assert.Equal(t, "TX-NX600", devices[0].Name)

		foundDevice, err := repo.FindDeviceByID(txCtx, clinicID, devices[0].ID)
		require.NoError(t, err)
		assert.Equal(t, devices[0].ID, foundDevice.ID)
		return errLabDeviceMasterTxSentinel
	})
	require.ErrorIs(t, txErr, errLabDeviceMasterTxSentinel)

	listed, err := repo.List(ctx, clinicID, "")
	require.NoError(t, err)
	assert.Empty(t, listed)
	devices, err := repo.ListDevices(ctx, clinicID)
	require.NoError(t, err)
	assert.Empty(t, devices)
}

func TestLabDeviceItemMasterRepository_ExamTypeReadersSeeUncommittedSeedThenRollback(t *testing.T) {
	db := setupLabDeviceItemMasterTestDB(t)
	repo := NewLabDeviceItemMasterRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	var examID, fieldID uint64
	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		exam := &model.ExaminationType{ClinicID: clinicID, Name: "tx-exam-type"}
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(exam).Error)
		field := &model.ExamTypeField{ClinicID: clinicID, ExamTypeID: exam.ID, Name: "BUN"}
		require.NoError(t, persistence.DBOrTx(txCtx, db).Create(field).Error)
		examID, fieldID = exam.ID, field.ID

		gotExam, err := repo.FindExamType(txCtx, clinicID, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, "tx-exam-type", gotExam.Name)

		gotField, err := repo.FindExamTypeField(txCtx, clinicID, field.ID)
		require.NoError(t, err)
		assert.Equal(t, "BUN", gotField.Name)

		fields, err := repo.FindExamTypeFields(txCtx, clinicID, []uint64{field.ID})
		require.NoError(t, err)
		require.Len(t, fields, 1)
		assert.Equal(t, "BUN", fields[field.ID].Name)
		return errLabDeviceMasterTxSentinel
	})
	require.ErrorIs(t, txErr, errLabDeviceMasterTxSentinel)

	_, err := repo.FindExamType(ctx, clinicID, examID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	_, err = repo.FindExamTypeField(ctx, clinicID, fieldID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	fields, err := repo.FindExamTypeFields(ctx, clinicID, []uint64{fieldID})
	require.NoError(t, err)
	assert.Empty(t, fields)
}
