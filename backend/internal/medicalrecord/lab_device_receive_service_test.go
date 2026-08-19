package medicalrecord

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type stubLabDevicePetFinder struct {
	pets map[uint64]*model.Pet
}

func (s stubLabDevicePetFinder) FindByID(_ context.Context, clinicID, id uint64) (*model.Pet, error) {
	pet, ok := s.pets[id]
	if !ok || pet.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("pet", "")
	}
	return pet, nil
}

func setupLabDeviceReceiveTest(t *testing.T) (*gorm.DB, LabDeviceReceiveService, stubLabDevicePetFinder) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.LabImportJob{},
		&model.LabImportJobItem{},
		&model.LabDeviceWait{},
		&model.LabDeviceStationSettings{},
		&model.LabDeviceItemMaster{},
	))
	for _, value := range []string{"fuji_nx600", "fuji_au10v", "arkray_pu4010"} {
		require.NoError(t, db.Exec(`ALTER TYPE lab_import_source_type ADD VALUE IF NOT EXISTS '`+value+`'`).Error)
	}
	require.NoError(t, db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS testdb_uq_lab_import_jobs_clinic_source_fingerprint
  ON lab_import_jobs (clinic_id, source_type, source_fingerprint)
  WHERE source_fingerprint <> ''
`).Error)
	require.NoError(t, db.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS testdb_uq_lab_device_waits_clinic_active
  ON lab_device_waits (clinic_id)
  WHERE cleared_at IS NULL
`).Error)

	const clinicA, clinicB = uint64(9701), uint64(9702)
	require.NoError(t, db.Exec(
		`DELETE FROM lab_import_job_items WHERE clinic_id IN ?`,
		[]uint64{clinicA, clinicB},
	).Error)
	require.NoError(t, db.Exec(
		`DELETE FROM lab_import_jobs WHERE clinic_id IN ?`,
		[]uint64{clinicA, clinicB},
	).Error)
	require.NoError(t, db.Exec(
		`DELETE FROM lab_device_waits WHERE clinic_id IN ?`,
		[]uint64{clinicA, clinicB},
	).Error)
	require.NoError(t, db.Exec(
		`DELETE FROM lab_device_station_settings WHERE clinic_id IN ?`,
		[]uint64{clinicA, clinicB},
	).Error)
	require.NoError(t, db.Exec(
		`DELETE FROM lab_device_item_masters WHERE clinic_id IN ?`,
		[]uint64{clinicA, clinicB},
	).Error)

	alive := &model.Pet{ID: 11, ClinicID: clinicA, Name: "タロウ", Status: model.PetStatusAlive}
	dead := &model.Pet{ID: 12, ClinicID: clinicA, Name: "亡", Status: model.PetStatusDeceased}
	other := &model.Pet{ID: 21, ClinicID: clinicB, Name: "他院", Status: model.PetStatusAlive}
	finder := stubLabDevicePetFinder{pets: map[uint64]*model.Pet{
		11: alive, 12: dead, 21: other, 13: {ID: 13, ClinicID: clinicA, Name: "ジロウ", Status: model.PetStatusAlive},
	}}
	svc := NewLabDeviceReceiveService(
		NewLabDeviceReceiveRepository(db),
		NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db)),
		finder,
		persistence.NewTransactor(db),
		nil,
	)
	return db, svc, finder
}

func TestLabDeviceReceiveService_WaitLinksThenDuplicate(t *testing.T) {
	db, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)

	wait, err := svc.PutWait(ctx, clinicA, 7, 11)
	require.NoError(t, err)
	assert.Equal(t, "タロウ", wait.PetName)
	assert.True(t, wait.ExpiresAt.After(time.Now().Add(20*time.Minute)))

	payload := synthFujiAU10V()
	got, err := svc.ReceiveFrames(ctx, clinicA, payload, "AU10V")
	require.NoError(t, err)
	require.Len(t, got.Results, 1)
	assert.False(t, got.Results[0].Duplicate)
	require.NotNil(t, got.Results[0].Job.PetID)
	assert.Equal(t, uint64(11), *got.Results[0].Job.PetID)
	assert.Equal(t, "タロウ", got.Results[0].Job.PetName)
	assert.Equal(t, model.LabImportSourceTypeFujiAU10V, got.Results[0].Job.SourceType)
	assert.Equal(t, 1, got.Results[0].Job.ItemCount)
	assert.Equal(t, "<3.75", got.Results[0].Job.Items[0].ValueRaw)

	again, err := svc.ReceiveFrames(ctx, clinicA, payload, "AU10V")
	require.NoError(t, err)
	require.Len(t, again.Results, 1)
	assert.True(t, again.Results[0].Duplicate)
	assert.Equal(t, got.Results[0].Job.JobID, again.Results[0].Job.JobID)

	board, err := svc.Board(ctx, clinicA)
	require.NoError(t, err)
	require.NotNil(t, board.Wait)
	assert.Equal(t, "タロウ", board.Wait.PetName)
	require.Len(t, board.Saved, 1)
	assert.Empty(t, board.Unlinked)
	assert.Equal(t, labDeviceDefaultWaitTTLSeconds, board.Station.WaitTTLSeconds)
	assert.Contains(t, board.Station.SlotsJSON, "fuji_nx600")

	var jobCount int64
	require.NoError(t, db.Model(&model.LabImportJob{}).Where("clinic_id = ?", clinicA).Count(&jobCount).Error)
	assert.Equal(t, int64(1), jobCount)
}

func TestLabDeviceReceiveService_NoWaitOrExpiredGoesUnlinked(t *testing.T) {
	_, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)

	got, err := svc.ReceiveFrames(ctx, clinicA, synthFujiAU10V(), "AU10V")
	require.NoError(t, err)
	assert.Nil(t, got.Results[0].Job.PetID)
	unlinked, err := svc.Unlinked(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, unlinked, 1)

	impl := svc.(*labDeviceReceiveService)
	impl.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	_, err = svc.PutWait(ctx, clinicA, 7, 11)
	require.NoError(t, err)
	impl.now = time.Now
	expired, err := svc.ReceiveFrames(ctx, clinicA, synthFujiNX600(), "NX600")
	require.NoError(t, err)
	assert.Nil(t, expired.Results[0].Job.PetID)
	unlinked, err = svc.Unlinked(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, unlinked, 2)
}

func TestLabDeviceReceiveService_AttachDetachAndIsolation(t *testing.T) {
	_, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(9701), uint64(9702)

	got, err := svc.ReceiveFrames(ctx, clinicA, synthUrinePU4010(), "PU-4010")
	require.NoError(t, err)
	jobID := got.Results[0].Job.JobID
	assert.Nil(t, got.Results[0].Job.PetID)

	_, err = svc.Attach(ctx, clinicB, jobID, 21)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	_, err = svc.Attach(ctx, clinicA, jobID, 21)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	_, err = svc.Attach(ctx, clinicA, jobID, 12)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not alive")

	linked, err := svc.Attach(ctx, clinicA, jobID, 11)
	require.NoError(t, err)
	require.NotNil(t, linked.PetID)
	assert.Equal(t, uint64(11), *linked.PetID)

	detached, err := svc.Detach(ctx, clinicA, jobID)
	require.NoError(t, err)
	assert.Nil(t, detached.PetID)

	_, err = svc.PutWait(ctx, clinicA, 7, 11)
	require.NoError(t, err)
	replaced, err := svc.PutWait(ctx, clinicA, 7, 13)
	require.NoError(t, err)
	assert.Equal(t, uint64(13), replaced.PetID)
	board, err := svc.Board(ctx, clinicA)
	require.NoError(t, err)
	require.NotNil(t, board.Wait)
	assert.Equal(t, "ジロウ", board.Wait.PetName)

	require.NoError(t, svc.ClearWait(ctx, clinicA))
	board, err = svc.Board(ctx, clinicA)
	require.NoError(t, err)
	assert.Nil(t, board.Wait)
}

func TestLabDeviceReceiveService_RejectsNonDeviceAndPersistedDetach(t *testing.T) {
	db, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)

	fixture := &model.LabImportJob{
		ClinicID:   clinicA,
		SourceType: model.LabImportSourceTypeFixture,
		Status:     model.LabImportJobStatusReceived,
	}
	require.NoError(t, db.Create(fixture).Error)
	_, err := svc.Attach(ctx, clinicA, fixture.ID, 11)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	persisted := &model.LabImportJob{
		ClinicID:   clinicA,
		SourceType: model.LabImportSourceTypeFujiAU10V,
		Status:     model.LabImportJobStatusPersisted,
	}
	require.NoError(t, db.Create(persisted).Error)
	_, err = svc.Detach(ctx, clinicA, persisted.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))

	_, err = svc.ReceiveFrames(ctx, clinicA, []byte("not-a-frame"), "NX600")
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Contains(t, err.Error(), "invalid_payload")

	_ = base64.StdEncoding.EncodeToString(synthFujiAU10V())
}

func TestLabDeviceReceiveService_ReplaceWaitKeepsOneActive(t *testing.T) {
	db, svc, _ := setupLabDeviceReceiveTest(t)
	ctx := context.Background()
	const clinicA = uint64(9701)
	_, err := svc.PutWait(ctx, clinicA, 7, 11)
	require.NoError(t, err)
	_, err = svc.PutWait(ctx, clinicA, 8, 13)
	require.NoError(t, err)
	var active int64
	require.NoError(t, db.Model(&model.LabDeviceWait{}).
		Where("clinic_id = ? AND cleared_at IS NULL", clinicA).
		Count(&active).Error)
	assert.Equal(t, int64(1), active)
}
