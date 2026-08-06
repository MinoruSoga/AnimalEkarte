package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type blockingMedicalRecordDeleteRepository struct {
	MedicalRecordRepository
	locked  chan struct{}
	proceed chan struct{}
}

func (r *blockingMedicalRecordDeleteRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	record, err := r.MedicalRecordRepository.LockByIDForUpdate(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	close(r.locked)
	<-r.proceed
	return record, nil
}

func makeDraftRecordForDeleteEstimateRace(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	return makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	})
}

func createEstimateForLockedMedicalRecord(ctx context.Context, db *gorm.DB, repo MedicalRecordRepository, clinicID, medicalRecordID uint64, estimateNo string) error {
	return testTransactor{db: db}.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.LockByIDForUpdate(txCtx, clinicID, medicalRecordID); err != nil {
			return err
		}
		return persistence.DBOrTx(txCtx, db).Create(&model.Estimate{
			ClinicID:        clinicID,
			EstimateNo:      estimateNo,
			MedicalRecordID: &medicalRecordID,
			Status:          model.EstimateStatusDraft,
		}).Error
	})
}

func TestMedicalRecordService_DeleteSerializesWithEstimateCreation(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("estimate commit wins and delete conflicts", func(t *testing.T) {
		db := setupMedicalRecordListTestDB(t)
		record := makeDraftRecordForDeleteEstimateRace(t, db, clinicID, "DEL-ESTIMATE-FIRST")
		repo := NewMedicalRecordRepository(db)
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testTransactor{db: db})
		estimateReady := make(chan struct{})
		releaseEstimate := make(chan struct{})
		estimateDone := make(chan error, 1)
		go func() {
			estimateDone <- testTransactor{db: db}.WithTx(context.Background(), func(txCtx context.Context) error {
				if _, err := repo.LockByIDForUpdate(txCtx, clinicID, record.ID); err != nil {
					return err
				}
				if err := persistence.DBOrTx(txCtx, db).Create(&model.Estimate{
					ClinicID: clinicID, EstimateNo: "EST-FIRST", MedicalRecordID: &record.ID,
					Status: model.EstimateStatusDraft,
				}).Error; err != nil {
					return err
				}
				close(estimateReady)
				<-releaseEstimate
				return nil
			})
		}()
		<-estimateReady

		deleteDone := make(chan error, 1)
		go func() { deleteDone <- svc.Delete(context.Background(), clinicID, record.ID) }()
		select {
		case err := <-deleteDone:
			require.Failf(t, "delete did not wait for estimate transaction", "err=%v", err)
		case <-time.After(100 * time.Millisecond):
		}
		close(releaseEstimate)
		require.NoError(t, <-estimateDone)

		err := <-deleteDone
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		var persisted model.MedicalRecord
		require.NoError(t, db.First(&persisted, record.ID).Error)
	})

	t.Run("delete commit wins and later estimate is rejected", func(t *testing.T) {
		db := setupMedicalRecordListTestDB(t)
		record := makeDraftRecordForDeleteEstimateRace(t, db, clinicID, "DEL-DELETE-FIRST")
		baseRepo := NewMedicalRecordRepository(db)
		blockingRepo := &blockingMedicalRecordDeleteRepository{
			MedicalRecordRepository: baseRepo,
			locked:                  make(chan struct{}),
			proceed:                 make(chan struct{}),
		}
		svc := NewMedicalRecordServiceWithTxAudit(blockingRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, testTransactor{db: db})
		deleteDone := make(chan error, 1)
		go func() { deleteDone <- svc.Delete(context.Background(), clinicID, record.ID) }()
		select {
		case <-blockingRepo.locked:
		case err := <-deleteDone:
			require.Failf(t, "delete skipped the medical-record row lock", "err=%v", err)
			return
		case <-time.After(time.Second):
			require.Fail(t, "delete did not acquire the medical-record row lock")
			return
		}

		estimateDone := make(chan error, 1)
		go func() {
			estimateDone <- createEstimateForLockedMedicalRecord(
				context.Background(), db, baseRepo, clinicID, record.ID, "EST-AFTER-DELETE",
			)
		}()
		select {
		case err := <-estimateDone:
			require.Failf(t, "estimate did not wait for delete transaction", "err=%v", err)
		case <-time.After(100 * time.Millisecond):
		}
		close(blockingRepo.proceed)
		require.NoError(t, <-deleteDone)
		require.Error(t, <-estimateDone)

		var count int64
		require.NoError(t, db.Model(&model.Estimate{}).
			Where("medical_record_id = ? AND deleted_at IS NULL", record.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	})
}
