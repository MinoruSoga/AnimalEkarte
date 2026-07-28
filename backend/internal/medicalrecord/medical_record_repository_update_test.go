package medicalrecord

// medical_record_repository_update_test.go — BE-refactor.md R29: 800行上限超過のため
// medical_record_repository_test.go から Update-version 系と行ロック（LockByIDForUpdate）系の
// テスト関数を逐語移動したもの。ヘルパー（setupMedicalRecordListTestDB 等）は同一パッケージのため
// medical_record_repository_test.go 側のものをそのまま共有する。

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestMedicalRecordRepository_LockByIDForUpdate は X-11 finalize-child-write-race fix の
// LockByIDForUpdate の基本挙動（clinic_id 隔離・存在確認・status に関わらずロック取得できること）を
// 検証する（並行ロック挙動そのものの実証は medical_record_finalize_lock_concurrency_test.go）。
func TestMedicalRecordRepository_LockByIDForUpdate(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "LockByIDForUpdate飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "LockByIDForUpdateペット")
	draftRec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "LDB-001", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
		Status: model.MedicalRecordStatusDraft,
	})
	finalizedRec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
		ClinicID: clinicA, RecordNo: "LDB-002", Date: time.Now(), OwnerID: &ownerA.ID, PetID: &petA.ID,
		Status: model.MedicalRecordStatusFinalized,
	})

	t.Run("同一医院からは draft レコードを取得できる", func(t *testing.T) {
		got, err := repo.LockByIDForUpdate(ctx, clinicA, draftRec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, draftRec.ID, got.ID)
		assert.Equal(t, model.MedicalRecordStatusDraft, got.Status)
	})

	t.Run("status に関わらずロック取得できる（finalized も返す。判定は呼び出し元）", func(t *testing.T) {
		got, err := repo.LockByIDForUpdate(ctx, clinicA, finalizedRec.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, finalizedRec.ID, got.ID)
		assert.Equal(t, model.MedicalRecordStatusFinalized, got.Status)
	})

	t.Run("他院からは NotFound（clinic_id 越境防止）", func(t *testing.T) {
		_, err := repo.LockByIDForUpdate(ctx, clinicB, draftRec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.LockByIDForUpdate(ctx, clinicA, uint64(999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestMedicalRecordRepository_Update は draft のみ更新可能な業務ルールと
// clinic_id 越境更新の拒否を検証する。
func TestMedicalRecordRepository_Update(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "Update飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "Updateペット")

	t.Run("draft カルテは更新できる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-DRAFT", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		updated, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-DRAFT-EDITED"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "UP-DRAFT-EDITED", updated.RecordNo)
	})

	t.Run("finalized カルテは Conflict で更新できない", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "UP-FIN", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			Status: model.MedicalRecordStatusFinalized,
		})
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-FIN-EDITED"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテの更新は Conflict であるべき: %v", err)
	})

	t.Run("他院からの更新は対象0件で Conflict（越境更新の拒否）", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-XCLINIC", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		_, err := repo.Update(ctx, clinicB, rec.ID, map[string]any{"record_no": "HACKED"}, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))

		// 実データが書き換わっていないことも確認する。
		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "UP-XCLINIC", got.RecordNo, "他院からの更新でレコードが変更されてはならない")
	})

	// ─── BE-refactor.md X-10 (mr-version-check-not-atomic): version 述語の原子性 ───

	t.Run("expectedVersion 一致時は更新でき version がインクリメントされる", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-VER-OK", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		require.Equal(t, 1, rec.Version, "makeFullMedicalRecord 直後の既定 version は 1 のはず")
		expected := rec.Version
		updated, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "UP-VER-OK-EDITED", "version": rec.Version + 1}, &expected)
		require.NoError(t, err)
		assert.Equal(t, "UP-VER-OK-EDITED", updated.RecordNo)
		assert.Equal(t, rec.Version+1, updated.Version)
	})

	t.Run("expectedVersion 不一致時は「他のユーザーが変更した」Conflict になり書き込まれない", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicA, RecordNo: "UP-VER-STALE", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
		staleVersion := rec.Version - 1 // 既に他ユーザーが version を進めた状態を模す
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "SHOULD-NOT-APPLY", "version": rec.Version + 1}, &staleVersion)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "他のユーザーがこのカルテを変更しました", "version不一致は not-draft とは異なる文言で区別されるべき")

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "UP-VER-STALE", got.RecordNo, "version不一致時はフィールドが書き換わってはならない")
	})

	t.Run("finalized カルテへの expectedVersion 付き更新は version不一致メッセージではなく not-draft を返す", func(t *testing.T) {
		rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{
			ClinicID: clinicA, RecordNo: "UP-VER-FIN", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID,
			Status: model.MedicalRecordStatusFinalized,
		})
		expected := rec.Version
		_, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"record_no": "SHOULD-NOT-APPLY"}, &expected)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "not in draft status", "確定済みは version不一致ではなく従来の not-draft 文言を維持するべき")
	})
}

// TestMedicalRecordRepository_Update_VersionPredicate_ConcurrentUpdates_OnlyOneSucceeds は
// BE-refactor.md X-10 の受け入れ条件: 同一 version=N を起点にした同一カルテへの同時2件の
// Update で、正確に1件のみ成功し他方は「他のユーザーがこのカルテを変更しました」Conflict を
// 受け取り、lost update が発生しないことを実DBの2 goroutineで検証する。
func TestMedicalRecordRepository_Update_VersionPredicate_ConcurrentUpdates_OnlyOneSucceeds(t *testing.T) {
	db := setupMedicalRecordListTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "並行更新飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "並行更新ペット")
	rec := makeFullMedicalRecord(t, db, &model.MedicalRecord{ClinicID: clinicID, RecordNo: "UP-CONCURRENT", Date: time.Now(), OwnerID: &owner.ID, PetID: &pet.ID})
	startVersion := rec.Version

	const attempts = 2
	notes := []string{"スタッフAによる更新", "スタッフBによる更新"}
	results := make([]error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func(idx int) {
			defer wg.Done()
			expected := startVersion
			_, err := repo.Update(ctx, clinicID, rec.ID, map[string]any{
				"record_no": notes[idx],
				"version":   startVersion + 1,
			}, &expected)
			results[idx] = err
		}(i)
	}
	wg.Wait()

	successCount, conflictCount := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successCount++
		case apperrors.IsConflict(err):
			conflictCount++
		default:
			t.Fatalf("unexpected error from concurrent version-checked update: %v", err)
		}
	}
	assert.Equal(t, 1, successCount, "同一versionを起点にした同時更新はちょうど1件だけ成功するべき（lost update防止）")
	assert.Equal(t, 1, conflictCount, "もう1件は version 不一致 Conflict になるべき")

	got, err := repo.FindByID(ctx, clinicID, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, startVersion+1, got.Version, "version はちょうど1回だけインクリメントされるべき")
	assert.Contains(t, notes, got.RecordNo, "書き込まれた内容は勝者側のnotesのいずれかであるべき")
}
