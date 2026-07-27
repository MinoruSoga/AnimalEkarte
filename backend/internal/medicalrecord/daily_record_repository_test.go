package medicalrecord

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - FindByHospitalizationID / FindByHospitalizationIDAndDate は clinic_id でテナント隔離される。
//   - FindByHospitalizationID は date DESC で返す。
//   - VitalRecords / CareLogs は clinic_id 述語付きで Preload される。
//   - FindOrCreateByDate は同一 (clinic_id, hospitalization_id, date) では既存行を再利用する（重複作成しない）。
//   - CreateVitalRecord / CreateCareLog / CreateStaffNote は対応するテーブルへ行を追加する。
//
// makeSpeciesAndPet/makeHospitalizationRec はフラット package の同名ヘルパーの複製
// （BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。
// makeTestOwner は testdb.MakeTestOwner に直接委譲する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeHospitalizationRec(t *testing.T, db *gorm.DB, clinicID, ownerID, petID uint64, cageID *uint64) *model.Hospitalization {
	t.Helper()
	now := time.Now()
	h := &model.Hospitalization{
		ClinicID:            clinicID,
		OwnerID:             ownerID,
		PetID:               petID,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now,
		CageID:              cageID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(h).Error)
	return h
}

// withTx mirrors repository.Transactor.WithTx (repohelpers-based ambient tx) without
// importing the flat repository package, which would create an import cycle
// (repository imports this subpackage via its facade).
// setupDailyRecordTestDB は daily_records と関連する vital_records / care_logs / staff_notes、
// および入院・ペットの前提テーブルを用意する。
func setupDailyRecordTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Hospitalization{},
		&model.DailyRecord{},
		&model.VitalRecord{},
		&model.CareLog{},
		&model.StaffNote{},
	))
	// 永続テストDBのスキーマドリフト自己修復（BUG-404 の再発防止）:
	// 旧モデル（type タグなしの time.Time）時代に AutoMigrate された永続テストDBでは
	// time 列が timestamptz のまま残る（AutoMigrate は既存列の型を変更しない）。
	// 本番 001_init.sql の正は TIME 型。このドリフトが「time.Time では TIME 列を
	// Scan できない」本番バグをテストから隠していた。冪等 ALTER で 001 に揃える。
	db.Exec(`ALTER TABLE care_logs ALTER COLUMN "time" TYPE time USING "time"::time`)
	db.Exec(`ALTER TABLE staff_notes ALTER COLUMN "time" TYPE time USING "time"::time`)
	db.Exec("TRUNCATE TABLE staff_notes CASCADE")
	db.Exec("TRUNCATE TABLE care_logs CASCADE")
	db.Exec("TRUNCATE TABLE vital_records CASCADE")
	db.Exec("TRUNCATE TABLE daily_records CASCADE")
	db.Exec("TRUNCATE TABLE hospitalizations CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeDailyRecord(t *testing.T, db *gorm.DB, clinicID, hospitalizationID uint64, date time.Time) *model.DailyRecord {
	t.Helper()
	dr := &model.DailyRecord{ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date}
	require.NoError(t, db.WithContext(context.Background()).Create(dr).Error)
	return dr
}

func TestDailyRecordRepository_FindByHospitalizationID(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "入院飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "入院ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	hospOther := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "入院飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "入院ポチB")
	hospB := makeHospitalizationRec(t, db, clinicB, ownerB.ID, petB.ID, nil)

	recordEarlier := makeDailyRecord(t, db, clinicA, hospA.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	recordLater := makeDailyRecord(t, db, clinicA, hospA.ID, time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC))
	makeDailyRecord(t, db, clinicA, hospOther.ID, time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC))
	makeDailyRecord(t, db, clinicB, hospB.ID, time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC))
	makeDailyRecord(t, db, clinicA, hospB.ID, time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC))

	weight := 4.2
	vr := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, DailyRecordID: &recordLater.ID, Weight: &weight}
	require.NoError(t, repo.CreateVitalRecord(ctx, vr))

	cl := &model.CareLog{ClinicID: clinicA, DailyRecordID: recordLater.ID, Time: "12:00:00", Type: model.CareLogTypeFood}
	require.NoError(t, repo.CreateCareLog(ctx, cl))

	sn := &model.StaffNote{DailyRecordID: recordLater.ID, Time: "12:00:00", Content: "経過良好"}
	require.NoError(t, repo.CreateStaffNote(ctx, sn))

	t.Run("returns records ordered by date DESC with vital/care/staff notes preloaded", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, hospA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, recordLater.ID, got[0].ID)
		assert.Equal(t, recordEarlier.ID, got[1].ID)

		require.Len(t, got[0].VitalRecords, 1)
		assert.Equal(t, vr.ID, got[0].VitalRecords[0].ID)
		require.Len(t, got[0].CareLogs, 1)
		assert.Equal(t, cl.ID, got[0].CareLogs[0].ID)
		require.Len(t, got[0].StaffNotes, 1)
		assert.Equal(t, sn.ID, got[0].StaffNotes[0].ID)
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicB, hospA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("clinic isolation: polluted hospitalization_id parent is rejected", func(t *testing.T) {
		got, err := repo.FindByHospitalizationID(ctx, clinicA, hospB.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestDailyRecordRepository_FindByHospitalizationIDAndDate(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "ポチB")
	hospB := makeHospitalizationRec(t, db, clinicB, ownerB.ID, petB.ID, nil)
	date := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	record := makeDailyRecord(t, db, clinicA, hospA.ID, date)
	makeDailyRecord(t, db, clinicA, hospB.ID, date)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByHospitalizationIDAndDate(ctx, clinicA, hospA.ID, date)
		require.NoError(t, err)
		assert.Equal(t, record.ID, got.ID)
	})

	t.Run("not found for a date with no record", func(t *testing.T) {
		got, err := repo.FindByHospitalizationIDAndDate(ctx, clinicA, hospA.ID, time.Date(2026, 6, 6, 0, 0, 0, 0, time.UTC))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByHospitalizationIDAndDate(ctx, clinicB, hospA.ID, date)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: polluted hospitalization_id parent is rejected", func(t *testing.T) {
		got, err := repo.FindByHospitalizationIDAndDate(ctx, clinicA, hospB.ID, date)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDailyRecordRepository_FindOrCreateByDate(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	date := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)

	created, err := repo.FindOrCreateByDate(ctx, clinicA, hospA.ID, date)
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	t.Run("second call for the same date reuses the existing record", func(t *testing.T) {
		again, err := repo.FindOrCreateByDate(ctx, clinicA, hospA.ID, date)
		require.NoError(t, err)
		assert.Equal(t, created.ID, again.ID)

		var count int64
		require.NoError(t, db.Model(&model.DailyRecord{}).
			Where("hospitalization_id = ? AND date = ?", hospA.ID, date).
			Count(&count).Error)
		assert.Equal(t, int64(1), count, "重複作成されないこと")
	})
}

func TestDailyRecordRepository_CreateVitalRecord(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")

	weight := 3.5
	vr := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, Weight: &weight}
	require.NoError(t, repo.CreateVitalRecord(ctx, vr))
	assert.NotZero(t, vr.ID)
}

func TestDailyRecordRepository_CreateCareLog(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	record := makeDailyRecord(t, db, clinicA, hospA.ID, time.Now())

	cl := &model.CareLog{ClinicID: clinicA, DailyRecordID: record.ID, Time: "12:00:00", Type: model.CareLogTypeMedicine}
	require.NoError(t, repo.CreateCareLog(ctx, cl))
	assert.NotZero(t, cl.ID)
}

func TestDailyRecordRepository_CreateStaffNote(t *testing.T) {
	db := setupDailyRecordTestDB(t)
	repo := NewDailyRecordRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチA")
	hospA := makeHospitalizationRec(t, db, clinicA, ownerA.ID, petA.ID, nil)
	record := makeDailyRecord(t, db, clinicA, hospA.ID, time.Now())

	sn := &model.StaffNote{DailyRecordID: record.ID, Time: "12:00:00", Content: "テストノート"}
	require.NoError(t, repo.CreateStaffNote(ctx, sn))
	assert.NotZero(t, sn.ID)
}
