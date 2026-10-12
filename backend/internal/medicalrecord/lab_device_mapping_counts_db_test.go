package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const mappingCountsClinicB = uint64(9702)

func seedMappingCountsJob(
	t *testing.T,
	tx persistence.Transactor,
	repo LabDeviceReceiveRepository,
	clinicID uint64,
	petID *uint64,
	status model.LabImportJobStatus,
	initialCount int,
) uuid.UUID {
	t.Helper()
	jobID := uuid.New()
	job := &model.LabImportJob{
		ID:                jobID,
		ClinicID:          clinicID,
		SourceType:        model.LabImportSourceTypeFujiNX600,
		SourceFingerprint: "mappingcounts-" + jobID.String(),
		Status:            status,
		RowCount:          16,
		NeedsReviewCount:  initialCount,
		UnmappedItemCount: initialCount,
		PetID:             petID,
	}
	items, _ := mappingCountsItems(clinicID, jobID, 16)
	require.NoError(t, tx.WithTx(context.Background(), func(txCtx context.Context) error {
		return repo.CreateJobWithItems(txCtx, job, items)
	}))
	return jobID
}

func requireMappingCountsJobItemsUnchanged(t *testing.T, repo LabDeviceReceiveRepository, clinicID uint64, jobID uuid.UUID) {
	t.Helper()
	itemMap, err := repo.ListJobItems(context.Background(), clinicID, []uuid.UUID{jobID})
	require.NoError(t, err)
	items := itemMap[jobID]
	require.Len(t, items, 16)
	for _, item := range items {
		require.False(t, item.NeedsReview, "item %s review flag changed", item.DeviceItemCode)
		require.Nil(t, item.ExamTypeFieldID, "item %s field changed", item.DeviceItemCode)
	}
}

func TestLabDeviceJobMappingCountsDB_PartialAndZero(t *testing.T) {
	cases := []struct {
		name  string
		count int
	}{
		{name: "recalculated fifteen", count: 15},
		{name: "recalculated zero", count: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _, _ := setupLabDeviceReceiveTest(t)
			repo := NewLabDeviceReceiveRepository(db)
			tx := persistence.NewTransactor(db)
			ctx := context.Background()
			jobID := seedMappingCountsJob(t, tx, repo, 9701, nil, model.LabImportJobStatusNeedsReview, 16)

			require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
				return repo.UpdateJobMappingCounts(txCtx, 9701, jobID, tc.count)
			}))

			job, err := repo.FindJobByID(ctx, 9701, jobID)
			require.NoError(t, err)
			require.Equal(t, tc.count, job.UnmappedItemCount, "stored unmapped_item_count")
			require.Equal(t, tc.count, job.NeedsReviewCount, "stored needs_review_count")
			requireMappingCountsJobItemsUnchanged(t, repo, 9701, jobID)
		})
	}
}

func TestLabDeviceJobMappingCountsDB_RejectsInvalidWrites(t *testing.T) {
	db, _, _ := setupLabDeviceReceiveTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	tx := persistence.NewTransactor(db)
	ctx := context.Background()
	jobID := seedMappingCountsJob(t, tx, repo, 9701, nil, model.LabImportJobStatusNeedsReview, 16)

	t.Run("no ambient transaction", func(t *testing.T) {
		err := repo.UpdateJobMappingCounts(ctx, 9701, jobID, 3)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr), "got %T %v", err, err)
		require.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("negative count", func(t *testing.T) {
		err := tx.WithTx(ctx, func(txCtx context.Context) error {
			return repo.UpdateJobMappingCounts(txCtx, 9701, jobID, -1)
		})
		require.True(t, apperrors.IsInvalidInput(err), "got %v", err)
	})

	t.Run("foreign clinic does not leak existence", func(t *testing.T) {
		err := tx.WithTx(ctx, func(txCtx context.Context) error {
			return repo.UpdateJobMappingCounts(txCtx, mappingCountsClinicB, jobID, 3)
		})
		require.True(t, apperrors.IsNotFound(err), "got %v", err)
	})

	t.Run("missing job id", func(t *testing.T) {
		err := tx.WithTx(ctx, func(txCtx context.Context) error {
			return repo.UpdateJobMappingCounts(txCtx, 9701, uuid.New(), 3)
		})
		require.True(t, apperrors.IsNotFound(err), "got %v", err)
	})

	job, err := repo.FindJobByID(ctx, 9701, jobID)
	require.NoError(t, err)
	require.Equal(t, 16, job.UnmappedItemCount, "rejected writes must not change unmapped_item_count")
	require.Equal(t, 16, job.NeedsReviewCount, "rejected writes must not change needs_review_count")
	requireMappingCountsJobItemsUnchanged(t, repo, 9701, jobID)
}

type mappingCountsFailAfterWriteRepo struct {
	LabDeviceReceiveRepository
	fail error
}

func (r *mappingCountsFailAfterWriteRepo) UpdateJobMappingCounts(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	unmappedCount int,
) error {
	if err := r.LabDeviceReceiveRepository.UpdateJobMappingCounts(ctx, clinicID, jobID, unmappedCount); err != nil {
		return err
	}
	return r.fail
}

func TestLabDeviceJobMappingCountsDB_PostCountFailureRollsBack(t *testing.T) {
	db, _, petID := setupLabDevicePersistTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	tx := persistence.NewTransactor(db)
	ctx := context.Background()
	pet := petID
	jobID := seedMappingCountsJob(t, tx, repo, 9701, &pet, model.LabImportJobStatusNeedsReview, 7)

	_, codes := mappingCountsItems(9701, jobID, 16)
	masters := &mappingCountsMasters{resolution: &LabDeviceMasterResolution{UnmappedCodes: codes}}
	sentinel := errors.New("post-write counter failure")
	failRepo := &mappingCountsFailAfterWriteRepo{LabDeviceReceiveRepository: repo, fail: sentinel}
	jobSvc := NewLabImportJobService(NewLabImportJobRepository(db), NewLabImportEventRepository(db))
	persister := NewLabDeviceExamPersister(failRepo, masters, nil, jobSvc, NewLabImportEventRepository(db), nil, tx)

	err := persister.PersistLinkedJob(ctx, 9701, jobID, petID)
	require.Error(t, err)
	require.True(t, errors.Is(err, sentinel), "got %v", err)

	job, err := repo.FindJobByID(ctx, 9701, jobID)
	require.NoError(t, err)
	require.Equal(t, 7, job.UnmappedItemCount, "counter write must roll back")
	require.Equal(t, 7, job.NeedsReviewCount, "counter write must roll back")
	require.Equal(t, model.LabImportJobStatusNeedsReview, job.Status, "status must be unchanged")
	requireMappingCountsJobItemsUnchanged(t, repo, 9701, jobID)
}
