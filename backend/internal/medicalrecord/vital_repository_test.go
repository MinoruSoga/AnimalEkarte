package medicalrecord

// vital_repository_test.go — VitalRepository の統合テスト（内部カバレッジ向上）。
// 移動元 internal/repository（BE9-2D sub-batch④a）。setupTestDB/ensureAutoMigrated は
// testdb.SetupTestDB/EnsureAutoMigrated 直呼びへ、makeTestOwner はこの package の共有
// repotest ラッパーを再利用（vaccination_repository_test.go 先例）。
//
// 対象: FindByMedicalRecordID / FindByID / Create / Update / Delete
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ、recorded_at 昇順ソート。

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

// setupVitalTestDB は vital_records と、その FK 先である
// pets/animal_species/staffs/daily_records を整備する（medical_records は core AutoMigrate 済み）。
func setupVitalTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Staff{}, &model.DailyRecord{}, &model.VitalRecord{},
	))
	db.Exec("TRUNCATE TABLE vital_records CASCADE")
	db.Exec("TRUNCATE TABLE daily_records CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func vitalFloatPtr(v float64) *float64 { return &v }

func makeVitalSpecies(t *testing.T, db *gorm.DB, name string) *model.AnimalSpecies {
	t.Helper()
	s := &model.AnimalSpecies{Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

func makeVitalPet(t *testing.T, db *gorm.DB, clinicID, ownerID, speciesID uint64, name string) *model.Pet {
	t.Helper()
	p := &model.Pet{ClinicID: clinicID, OwnerID: ownerID, AnimalSpeciesID: speciesID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func makeVitalMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{ClinicID: clinicID, RecordNo: recordNo, Date: date, Status: model.MedicalRecordStatusDraft}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

// TestVitalRepository_Create_FindByID は作成と単件取得（clinic_id 隔離・NotFound）を検証する。
func TestVitalRepository_Create_FindByID(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	species := makeVitalSpecies(t, db, "犬")
	petA := makeVitalPet(t, db, clinicA, ownerA.ID, species.ID, "ペットA")

	vital := &model.VitalRecord{
		ClinicID:    clinicA,
		PetID:       petA.ID,
		RecordedAt:  time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Temperature: vitalFloatPtr(38.5),
		Notes:       "初回検温",
	}

	t.Run("Create で新規作成され ID が採番される", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, vital))
		assert.NotZero(t, vital.ID)
	})

	t.Run("FindByID は自クリニックのレコードを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, vital.ID, got.ID)
		assert.Equal(t, "初回検温", got.Notes)
		require.NotNil(t, got.Temperature)
		assert.InDelta(t, 38.5, *got.Temperature, 0.001)
	})

	t.Run("FindByID は別クリニックのレコードを返さない（NotFound）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, vital.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("FindByID は存在しないIDに対して NotFound を返す", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestVitalRepository_FindByMedicalRecordID は診察記録単位の一覧取得（昇順ソート・
// clinic_id 隔離・ソフトデリート除外）を検証する。
func TestVitalRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	species := makeVitalSpecies(t, db, "猫")
	petA := makeVitalPet(t, db, clinicA, ownerA.ID, species.ID, "ペットA")
	mrA := makeVitalMedicalRecord(t, db, clinicA, "MR-VITAL-001", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))

	later := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &mrA.ID, RecordedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), Notes: "12時"}
	earlier := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &mrA.ID, RecordedAt: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC), Notes: "9時"}
	require.NoError(t, repo.Create(ctx, later))
	require.NoError(t, repo.Create(ctx, earlier))

	deleted := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, MedicalRecordID: &mrA.ID, RecordedAt: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC), Notes: "削除済み"}
	require.NoError(t, repo.Create(ctx, deleted))
	require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))

	// 別クリニックの vital が同じ medical_record_id を偽装しても、親 medical_records.clinic と
	// 相関しない行は返らない（SEC-SWEEP-02-MR-B1）。
	crossClinicVital := &model.VitalRecord{ClinicID: clinicB, PetID: petA.ID, MedicalRecordID: &mrA.ID, RecordedAt: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), Notes: "越境"}
	require.NoError(t, repo.Create(ctx, crossClinicVital))

	t.Run("recorded_at 昇順で返す・ソフトデリート除外", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2, "ソフトデリート済み・越境行は含まれない")
		assert.Equal(t, "9時", got[0].Notes)
		assert.Equal(t, "12時", got[1].Notes)
	})

	t.Run("別クリニックで問い合わせると空を返す（clinic_id 隔離 + 親相関）", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicB, mrA.ID)
		require.NoError(t, err)
		// clinicB の vital は親 medical_records(clinicA) と clinic が不一致のため除外される。
		assert.Empty(t, got, "cross-clinic vital linked to foreign parent medical record must be excluded")
	})

	t.Run("該当する medical_record_id が無ければ空スライスを返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, 999999)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

// TestVitalRepository_Update は部分更新の正常系・clinic_id 隔離・NotFound を検証する。
func TestVitalRepository_Update(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	species := makeVitalSpecies(t, db, "犬")
	petA := makeVitalPet(t, db, clinicA, ownerA.ID, species.ID, "ペットA")

	vital := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, RecordedAt: time.Now(), Notes: "更新前"}
	require.NoError(t, repo.Create(ctx, vital))

	t.Run("正常系: notes が更新される", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, vital.ID, map[string]any{"notes": "更新後"}))

		got, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Notes)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, vital.ID, map[string]any{"notes": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 9999999, map[string]any{"notes": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestVitalRepository_Delete はソフトデリートの正常系・clinic_id 隔離・NotFound を検証する。
func TestVitalRepository_Delete(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	species := makeVitalSpecies(t, db, "犬")
	petA := makeVitalPet(t, db, clinicA, ownerA.ID, species.ID, "ペットA")

	vital := &model.VitalRecord{ClinicID: clinicA, PetID: petA.ID, RecordedAt: time.Now(), Notes: "削除対象"}
	require.NoError(t, repo.Create(ctx, vital))

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, vital.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正常系: ソフトデリートされ FindByID から除外されるが行自体は残る", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, vital.ID))

		_, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.VitalRecord
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", vital.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "物理削除ではなくソフトデリートであるべき")
	})

	t.Run("存在しないIDの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestVitalRepository_FindByMedicalRecordID_CorrelatesMedicalRecordClinic
// SEC-SWEEP-02-MR-B1: vital_records.medical_record_id 読みは medical_records.clinic_id と相関必須。
// 子 vital の clinic だけ一致して親カルテが他院でも返ってしまう旧 failure mode を固定する。
func TestVitalRepository_FindByMedicalRecordID_CorrelatesMedicalRecordClinic(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	species := makeVitalSpecies(t, db, "犬")
	petA := makeVitalPet(t, db, clinicA, ownerA.ID, species.ID, "ペットA")

	mrA := makeVitalMedicalRecord(t, db, clinicA, "MR-VITAL-PARENT-A", time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC))
	mrB := makeVitalMedicalRecord(t, db, clinicB, "MR-VITAL-PARENT-B", time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC))

	// 正常: 同一 clinic の親カルテに紐づく vital
	legit := &model.VitalRecord{
		ClinicID:        clinicA,
		PetID:           petA.ID,
		MedicalRecordID: &mrA.ID,
		RecordedAt:      time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC),
		Notes:           "正常",
	}
	require.NoError(t, repo.Create(ctx, legit))

	// 汚染: 子は clinicA だが親カルテは clinicB — 旧実装は ClinicScope + medical_record_id だけで返す。
	polluted := &model.VitalRecord{
		ClinicID:        clinicA,
		PetID:           petA.ID,
		MedicalRecordID: &mrB.ID,
		RecordedAt:      time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC),
		Notes:           "親他院",
	}
	require.NoError(t, repo.Create(ctx, polluted))

	// ソフトデリート済み（同一 clinic・正常親）— 既存 lifecycle 維持
	deleted := &model.VitalRecord{
		ClinicID:        clinicA,
		PetID:           petA.ID,
		MedicalRecordID: &mrA.ID,
		RecordedAt:      time.Date(2026, 7, 28, 8, 0, 0, 0, time.UTC),
		Notes:           "削除済",
	}
	require.NoError(t, repo.Create(ctx, deleted))
	require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))

	t.Run("excludes vital whose parent medical record is foreign-clinic", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrB.ID)
		require.NoError(t, err)
		assert.Empty(t, got, "vital with foreign-clinic medical_record parent must not be returned")
	})

	t.Run("returns same-clinic vitals ordered and excludes soft-deleted", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "正常", got[0].Notes)
		assert.Equal(t, legit.ID, got[0].ID)
	})
}
