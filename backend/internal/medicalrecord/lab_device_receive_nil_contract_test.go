package medicalrecord

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLabDeviceReceiveRepository_FindJobByFingerprint_MissIsNotFound(t *testing.T) {
	db, _, _ := setupLabDeviceReceiveTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	ctx := context.Background()
	const clinicA = uint64(9701)
	source := string(model.LabImportSourceTypeFujiAU10V)

	tests := []struct {
		name        string
		fingerprint string
	}{
		{name: "empty fingerprint", fingerprint: ""},
		{name: "missing job", fingerprint: "no-such-fingerprint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := repo.FindJobByFingerprint(ctx, clinicA, source, tt.fingerprint)
			assert.Nil(t, job)
			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
		})
	}
}

func TestLabDeviceReceiveService_Receive_MissingFingerprintJobIsNotDuplicate(t *testing.T) {
	_, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)

	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiAU10V(), "AU10V")
	require.NoError(t, err)
	require.Len(t, got.Results, 1)
	assert.False(t, got.Results[0].Duplicate)
	assert.NotEqual(t, uuid.Nil, got.Results[0].Job.JobID)
}

func TestLabDeviceReceiveRepository_GetStation_MissIsNotFoundAndEnsureUpserts(t *testing.T) {
	db, svc, _ := setupLabDeviceReceiveTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	ctx := context.Background()
	const clinicA = uint64(9701)

	row, err := repo.GetStation(ctx, clinicA)
	assert.Nil(t, row)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	station, err := svc.GetStation(ctx, clinicA)
	require.NoError(t, err)
	require.NotNil(t, station)
	assert.Equal(t, labDeviceDefaultWaitTTLSeconds, station.WaitTTLSeconds)
	assert.Equal(t, labDeviceDefaultSlotsJSON, station.SlotsJSON)

	persisted, err := repo.GetStation(ctx, clinicA)
	require.NoError(t, err)
	require.NotNil(t, persisted)
	assert.Equal(t, labDeviceDefaultWaitTTLSeconds, persisted.WaitTTLSeconds)
}

func TestLabDeviceReceiveRepository_FindActiveWait_MissDoesNotFailBoardOrPutWait(t *testing.T) {
	db, svc, _ := setupLabDeviceReceiveTest(t)
	repo := NewLabDeviceReceiveRepository(db)
	ctx := context.Background()
	const clinicA = uint64(9701)

	wait, err := repo.FindActiveWait(ctx, clinicA)
	assert.Nil(t, wait)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	board, err := svc.Board(ctx, clinicA)
	require.NoError(t, err)
	require.NotNil(t, board)
	assert.Nil(t, board.Wait)

	require.NoError(t, svc.ClearWait(ctx, clinicA))

	view, err := svc.PutWait(ctx, clinicA, 7, 11)
	require.NoError(t, err)
	require.NotNil(t, view)
	assert.Equal(t, uint64(11), view.PetID)
}
