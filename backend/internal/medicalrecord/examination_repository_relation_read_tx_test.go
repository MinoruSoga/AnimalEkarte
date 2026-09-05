package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestExaminationRepository_FindAllAndFindByJobIDJoinAmbientTx(t *testing.T) {
	fixture := setupClinicalRelationWriteFixture(t)
	repo := NewExaminationRepository(fixture.db)
	ctx := context.Background()
	jobID := uuid.New()
	sentinel := errors.New("rollback examination relation reads")

	err := fixture.writeTransactor.WithTx(ctx, func(txCtx context.Context) error {
		recordID := fixture.recordA.ID
		petID := fixture.petA.ID
		doctorID := fixture.assignedDoctor.ID
		examination := &model.Examination{
			ClinicID: fixture.clinicA, MedicalRecordID: &recordID, PetID: &petID,
			ExamTypeID: fixture.examinationType.ID, DoctorID: &doctorID,
			JobID: &jobID, Date: time.Now(), Status: model.ExaminationStatusPending,
		}
		require.NoError(t, repo.Create(txCtx, examination))

		listed, total, findErr := repo.FindAll(txCtx, fixture.clinicA, nil, nil, nil, nil, nil, nil, 1, 100, false)
		require.NoError(t, findErr)
		require.EqualValues(t, 1, total)
		require.Len(t, listed, 1)
		assert.Equal(t, examination.ID, listed[0].ID)

		byJob, findErr := repo.FindByJobID(txCtx, fixture.clinicA, jobID)
		require.NoError(t, findErr)
		require.Len(t, byJob, 1)
		assert.Equal(t, examination.ID, byJob[0].ID)
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	listed, total, err := repo.FindAll(ctx, fixture.clinicA, nil, nil, nil, nil, nil, nil, 1, 100, false)
	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, listed)
}
