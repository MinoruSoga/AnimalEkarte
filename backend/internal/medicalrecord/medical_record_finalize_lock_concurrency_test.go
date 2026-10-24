package medicalrecord

// medical_record_finalize_lock_concurrency_test.go — BE-refactor.md Appendix A X-11
// (`finalize-child-write-race`) の実DB並行性証明。
//
// 背景: 子エンティティ（treatment 等）の追加は、旧実装では tx 外の素の FindByID で親カルテを確認していたため、
// 確認直後に別リクエストが親を更新すると check-then-act のレースがあった。LockByIDForUpdate は
// FOR UPDATE でカルテ行をロックし、親 header の UPDATE と直列化することでこれを塞ぐ（同一行への UPDATE は
// postgres の行ロックにより FOR UPDATE 保持者の commit/rollback までブロックされる）。
//
// EMR-495 P1: medical_records.status 列と finalize（status=finalized への UPDATE とその後の子書込拒否）は
// 廃止した。本ファイルは行ロックによる直列化（順方向）のみを残し、「finalize が先に commit したら子書込が
// 拒否される」逆方向のテスト 3 件（ChildWriteRejected / TreatmentDeleteRejected / TreatmentBulkSortOrderRejected
// AfterFinalizeCommits）は挙動ごと廃止したため削除した。親 writer は status 非依存の header 書込
// （medicalRecordParentWrite）に置き換えた。
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
// 検証方法: 子書込トランザクション（LockByIDForUpdate → 意図的な待機 → Treatment.Create）
// が行ロックを保持している間、並行する親 header の UPDATE が完全にブロックされ、子書込
// トランザクションのコミット後にのみ親 UPDATE が完了することを実 DB のロック挙動で証明する
// （reservation_booking_lock_concurrency_test.go の X-9 パターンを踏襲）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// TestMedicalRecordRepository_LockByIDForUpdate_SerializesParentWriteAgainstChildWrite は X-11 の
// 受け入れ条件: 子書込 tx が LockByIDForUpdate の行ロックを保持している間、並行する
// 親 header の UPDATE がブロックされ続け、子書込 tx の commit（ロック解放）後にのみ完了することを検証する。
func TestMedicalRecordRepository_LockByIDForUpdate_SerializesParentWriteAgainstChildWrite(t *testing.T) {
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
			_ = parent
			<-proceed // 意図的に親 UPDATE がブロックされる時間を作る
			return treatmentRepo.Create(txCtx, &model.Treatment{MedicalRecordID: mr.ID})
		})
	}()

	<-lockAcquired

	parentWriteStarted := make(chan struct{})
	parentWriteDone := make(chan struct{})
	var parentWriteErr error
	go func() {
		defer close(parentWriteDone)
		close(parentWriteStarted)
		_, parentWriteErr = medRecRepo.Update(context.Background(), clinicID, mr.ID, medicalRecordParentWrite(), nil)
	}()
	<-parentWriteStarted

	// 親 UPDATE は子書込 tx が行ロックを保持している間ブロックされ続けるはずである
	// （postgres の行ロック挙動）。300ms 待っても完了しないことを確認する。
	select {
	case <-parentWriteDone:
		t.Fatal("親 UPDATE が子書込 tx のロック保持中に完了してしまった（X-11 のレース防止が機能していない）")
	case <-time.After(300 * time.Millisecond):
	}

	close(proceed) // 子書込 tx を進行させ、commit させる（ロック解放）
	<-childDone
	<-parentWriteDone

	require.NoError(t, childErr, "子書込 tx は成功するべき")
	require.NoError(t, parentWriteErr, "子書込 tx の commit（ロック解放）後、親 UPDATE は成功するべき")

	var persistedTreatments int64
	require.NoError(t, db.Model(&model.Treatment{}).Where("medical_record_id = ?", mr.ID).Count(&persistedTreatments).Error)
	assert.Equal(t, int64(1), persistedTreatments, "治療は1件だけ永続化されているべき")

	final, err := medRecRepo.FindByID(ctx, clinicID, mr.ID)
	require.NoError(t, err)
	require.NotNil(t, final.VisitType)
	assert.Equal(t, model.VisitTypeRevisit, *final.VisitType, "親 header の書込が反映されているべき")
}
