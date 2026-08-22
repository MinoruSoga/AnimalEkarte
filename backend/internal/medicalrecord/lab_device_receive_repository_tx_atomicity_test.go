package medicalrecord

// lab_device_receive_repository_tx_atomicity_test.go — F-5
//
// labDeviceReceiveRepository.q は persistence.DBOrTx の private helper。公開 writer
// CreateJobWithItems 経由で ambient tx に参加し、sentinel rollback で永続化されないことを証明する。

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

var errLabDeviceReceiveTxSentinel = errors.New("simulated post-write failure in ambient tx")

func TestLabDeviceReceiveRepository_CreateJobWithItems_RollsBackWhenAmbientTxFails(t *testing.T) {
	db, _, _ := setupLabDeviceReceiveTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_job_items WHERE clinic_id = ?`, clinicID).Error)
	require.NoError(t, db.Exec(`DELETE FROM lab_import_jobs WHERE clinic_id = ?`, clinicID).Error)

	job := &model.LabImportJob{
		ID:                uuid.New(),
		ClinicID:          clinicID,
		SourceType:        model.LabImportSourceTypeFujiNX600,
		SourceFingerprint: "f5-q-create-job",
		Status:            model.LabImportJobStatusReceived,
		DeviceHint:        "NX600",
		RowCount:          1,
	}
	items := []model.LabImportJobItem{{
		DeviceItemCode: "BUN-P",
		ValueRaw:       "12",
		Unit:           "mg/dl",
		SortOrder:      1,
	}}

	txErr := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.CreateJobWithItems(txCtx, job, items); err != nil {
			return err
		}
		found, err := repo.FindJobByID(txCtx, clinicID, job.ID)
		require.NoError(t, err)
		assert.Equal(t, "f5-q-create-job", found.SourceFingerprint)
		return errLabDeviceReceiveTxSentinel
	})
	require.ErrorIs(t, txErr, errLabDeviceReceiveTxSentinel)

	_, err := repo.FindJobByID(ctx, clinicID, job.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var count int64
	require.NoError(t, db.Model(&model.LabImportJob{}).
		Where("clinic_id = ? AND source_fingerprint = ?", clinicID, "f5-q-create-job").
		Count(&count).Error)
	assert.Zero(t, count, "CreateJobWithItems must roll back with the ambient transaction")
}
