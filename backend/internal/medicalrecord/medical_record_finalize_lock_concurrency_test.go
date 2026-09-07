package medicalrecord

// medical_record_finalize_lock_concurrency_test.go — BE-refactor.md Appendix A X-11
// (`finalize-child-write-race`) の実DB並行性証明。
//
// 背景: 確定(finalize)は medical_record_repository.Update の draft-only WHERE
// （X-10 で原子化済み）で行われる。子エンティティ（treatment 等）の追加は、旧実装では
// tx 外の素の FindByID で親カルテの status を確認していたため、確認直後に別リクエストが
// 確定を commit すると、確定済みカルテへの子追加がすり抜ける check-then-act レースがあった。
// LockByIDForUpdate は FOR UPDATE でカルテ行をロックし、finalize の UPDATE と直列化することで
// これを塞ぐ（同一行への UPDATE は postgres の行ロックにより FOR UPDATE 保持者の commit/rollback
// までブロックされる）。
//
// 子書込 tx の構築は treatment_service.go 実コードと同じ ctx-txKey 機構（Transactor.WithTx）を
// 使う（BE9-2D ④b で repo-swap 機構 Repositories.Transaction から移行）。旧実装では
// treatmentRepository.Create が `r.db.WithContext` 直参照で ambient tx 非参加だったため、
// WithTx と混在させると別コネクションの INSERT の FK チェック（FOR KEY SHARE）が
// LockByIDForUpdate の FOR UPDATE と自己デッドロックした（検証中に実際に踏んだ失敗モード）。
// ④b で treatmentRepository.Create/Update/Delete を dbOrTx 化し解消 — 本テストが WithTx 機構で
// green であること自体が「treatment 書込が同一 tx に参加している（自己デッドロックしない）」ことの
// 実 DB 証明を兼ねる。
//
// 検証方法（順方向）: 子書込トランザクション（LockByIDForUpdate → 意図的な待機 → Treatment.Create）
// が行ロックを保持している間、並行する finalize の UPDATE が完全にブロックされ、子書込
// トランザクションのコミット後にのみ finalize が完了することを実 DB のロック挙動で証明する
// （reservation_booking_lock_concurrency_test.go の X-9 パターンを踏襲）。
//
// 検証方法（逆方向）: finalize が先に commit した場合、子書込 tx が LockByIDForUpdate 取得時点で
// 既に finalized を観測し、Treatment が作成されないことを検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// TestMedicalRecordRepository_LockByIDForUpdate_SerializesFinalizeAgainstChildWrite は X-11 の
// 受け入れ条件（順方向）: 子書込 tx が LockByIDForUpdate の行ロックを保持している間、並行する
// finalize の UPDATE がブロックされ続け、子書込 tx の commit（ロック解放）後にのみ finalize が
// 完了することを検証する。
func TestMedicalRecordRepository_LockByIDForUpdate_SerializesFinalizeAgainstChildWrite(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
	))

	const clinicID = uint64(90101)
	owner := makeTestOwner(t, db, clinicID, "X-11 太郎")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "X-11 ポチ")

	medRecRepo := NewMedicalRecordRepository(db)
	treatmentRepo := NewTreatmentRepository(db)
	ctx := context.Background()

	petID := pet.ID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		PetID:    &petID,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, medRecRepo.Create(ctx, mr))

	lockAcquired := make(chan struct{})
	proceed := make(chan struct{})
	childDone := make(chan struct{})
	var childErr error

	go func() {
		defer close(childDone)
		childErr = withTx(context.Background(), db, func(txCtx context.Context) error {
			parent, err := medRecRepo.LockByIDForUpdate(txCtx, clinicID, mr.ID)
			if err != nil {
				return err
			}
			close(lockAcquired)
			<-proceed // 意図的に finalize の UPDATE がブロックされる時間を作る
			if parent.Status == model.MedicalRecordStatusFinalized {
				return apperrors.WrapConflict("確定済みカルテには治療を追加できません")
			}
			return treatmentRepo.Create(txCtx, &model.Treatment{MedicalRecordID: mr.ID})
		})
	}()

	<-lockAcquired

	finalizeStarted := make(chan struct{})
	finalizeDone := make(chan struct{})
	var finalizeErr error
	go func() {
		defer close(finalizeDone)
		close(finalizeStarted)
		_, finalizeErr = medRecRepo.Update(context.Background(), clinicID, mr.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
	}()
	<-finalizeStarted

	// finalize の UPDATE は子書込 tx が行ロックを保持している間ブロックされ続けるはずである
	// （postgres の行ロック挙動）。300ms 待っても完了しないことを確認する。
	select {
	case <-finalizeDone:
		t.Fatal("finalize が子書込 tx のロック保持中に完了してしまった（X-11 のレース防止が機能していない）")
	case <-time.After(300 * time.Millisecond):
	}

	close(proceed) // 子書込 tx を進行させ、commit させる（ロック解放）
	<-childDone
	<-finalizeDone

	require.NoError(t, childErr, "子書込 tx は finalize がまだ commit していない時点で draft を観測し成功するべき")
	require.NoError(t, finalizeErr, "子書込 tx の commit（ロック解放）後、finalize は成功するべき")

	var persistedTreatments int64
	require.NoError(t, db.Model(&model.Treatment{}).Where("medical_record_id = ?", mr.ID).Count(&persistedTreatments).Error)
	assert.Equal(t, int64(1), persistedTreatments, "治療は1件だけ永続化されているべき")

	final, err := medRecRepo.FindByID(ctx, clinicID, mr.ID)
	require.NoError(t, err)
	assert.Equal(t, model.MedicalRecordStatusFinalized, final.Status, "カルテは確定済みになっているべき")
}

// TestMedicalRecordRepository_LockByIDForUpdate_ChildWriteRejectedAfterFinalizeCommits は X-11 の
// 受け入れ条件（逆方向）: finalize が先に commit した場合、子書込 tx が LockByIDForUpdate 取得時点で
// 既に status=finalized を観測し、Treatment が作成されず Conflict を返すことを検証する
// （旧・check-then-act のみの failure mode が再現しないことの確認）。
func TestMedicalRecordRepository_LockByIDForUpdate_ChildWriteRejectedAfterFinalizeCommits(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
	))

	const clinicID = uint64(90102)
	owner := makeTestOwner(t, db, clinicID, "X-11 花子")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "X-11 タマ")

	medRecRepo := NewMedicalRecordRepository(db)
	treatmentRepo := NewTreatmentRepository(db)
	ctx := context.Background()

	petID := pet.ID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		PetID:    &petID,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, medRecRepo.Create(ctx, mr))

	_, err := medRecRepo.Update(ctx, clinicID, mr.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
	require.NoError(t, err)

	childErr := withTx(ctx, db, func(txCtx context.Context) error {
		parent, err := medRecRepo.LockByIDForUpdate(txCtx, clinicID, mr.ID)
		if err != nil {
			return err
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテには治療を追加できません")
		}
		return treatmentRepo.Create(txCtx, &model.Treatment{MedicalRecordID: mr.ID})
	})

	require.Error(t, childErr)
	assert.True(t, apperrors.IsConflict(childErr))

	var persistedTreatments int64
	require.NoError(t, db.Model(&model.Treatment{}).Where("medical_record_id = ?", mr.ID).Count(&persistedTreatments).Error)
	assert.Equal(t, int64(0), persistedTreatments, "確定済みカルテには治療が混入しないべき")
}

// TestMedicalRecordRepository_LockByIDForUpdate_TreatmentDeleteRejectedAfterFinalizeCommits は
// H-8f（H-8b の並行性証明）: ChildWriteRejectedAfterFinalizeCommits と同型で、finalize が先に
// commit した場合、treatment.Delete 経路の tx が LockByIDForUpdate 取得時点で既に
// status=finalized を観測し、Conflict を返し治療が削除されないことを検証する
// （treatment_service.go Delete の Transactor.WithTx tx を模する）。
func TestMedicalRecordRepository_LockByIDForUpdate_TreatmentDeleteRejectedAfterFinalizeCommits(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
	))

	const clinicID = uint64(90103)
	owner := makeTestOwner(t, db, clinicID, "H-8f 太郎")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "H-8f ポチ")

	medRecRepo := NewMedicalRecordRepository(db)
	treatmentRepo := NewTreatmentRepository(db)
	ctx := context.Background()

	petID := pet.ID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		PetID:    &petID,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, medRecRepo.Create(ctx, mr))

	treatment := &model.Treatment{MedicalRecordID: mr.ID}
	require.NoError(t, treatmentRepo.Create(ctx, treatment))

	_, err := medRecRepo.Update(ctx, clinicID, mr.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
	require.NoError(t, err)

	childErr := withTx(ctx, db, func(txCtx context.Context) error {
		parent, err := medRecRepo.LockByIDForUpdate(txCtx, clinicID, mr.ID)
		if err != nil {
			return err
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの治療は削除できません")
		}
		return treatmentRepo.Delete(txCtx, clinicID, treatment.ID)
	})

	require.Error(t, childErr)
	assert.True(t, apperrors.IsConflict(childErr))

	var persistedTreatments int64
	require.NoError(t, db.Model(&model.Treatment{}).Where("medical_record_id = ? AND deleted_at IS NULL", mr.ID).Count(&persistedTreatments).Error)
	assert.Equal(t, int64(1), persistedTreatments, "確定済みカルテの治療は削除されず残存するべき")
}

// TestMedicalRecordRepository_LockByIDForUpdate_TreatmentBulkSortOrderRejectedAfterFinalizeCommits は
// H-8f（H-8c の並行性証明）: ChildWriteRejectedAfterFinalizeCommits と同型で、finalize が先に
// commit した場合、treatment.BulkUpdateSortOrder 経路の tx が LockByIDForUpdate 取得時点で既に
// status=finalized を観測し、Conflict を返し並び順が変更されないことを検証する
// （treatment_service.go BulkUpdateSortOrder の Transactor.WithTx tx を模する）。
func TestMedicalRecordRepository_LockByIDForUpdate_TreatmentBulkSortOrderRejectedAfterFinalizeCommits(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Consultation{},
		&model.Procedure{},
		&model.Medicine{},
		&model.InventoryItem{},
	))

	const clinicID = uint64(90104)
	owner := makeTestOwner(t, db, clinicID, "H-8f 花子")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "H-8f タマ")

	medRecRepo := NewMedicalRecordRepository(db)
	treatmentRepo := NewTreatmentRepository(db)
	ctx := context.Background()

	petID := pet.ID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		PetID:    &petID,
		Date:     time.Now(),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, medRecRepo.Create(ctx, mr))

	treatment := &model.Treatment{MedicalRecordID: mr.ID, SortOrder: 0}
	require.NoError(t, treatmentRepo.Create(ctx, treatment))

	_, err := medRecRepo.Update(ctx, clinicID, mr.ID, medicalRecordUpdateStatus(model.MedicalRecordStatusFinalized), nil)
	require.NoError(t, err)

	updates := []TreatmentSortUpdate{{ID: treatment.ID, ClinicID: clinicID, SortOrder: 99}}
	childErr := withTx(ctx, db, func(txCtx context.Context) error {
		parent, err := medRecRepo.LockByIDForUpdate(txCtx, clinicID, mr.ID)
		if err != nil {
			return err
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテの治療は編集できません")
		}
		return treatmentRepo.BulkUpdateSortOrder(txCtx, updates)
	})

	require.Error(t, childErr)
	assert.True(t, apperrors.IsConflict(childErr))

	var persisted model.Treatment
	require.NoError(t, db.First(&persisted, treatment.ID).Error)
	assert.Equal(t, 0, persisted.SortOrder, "確定済みカルテの治療は並び順が変更されず残存するべき")
}
