package repository

// examination_repository_tx_atomicity_test.go — BE-refactor.md R1-2 (D1): exam_results 置換の
// tx 内原子性（DB-backed）。
//
// ReplaceItemsByExamID が caller の ambient transaction に参加し、置換後に同一 tx 内の後続処理
// （= 監査書込）が失敗したら削除がロールバックして既存 items が DB に残存することを実 DB で実証する。
// checkup_field_result_tx_atomicity_test.go と同型（#211 パターンを exam_results に適用）。
//
//   - この原子性は mock では検証不可能（#211 security M2 の教訓）。repository が dbOrTx で
//     ambient tx に join することが正本である。
//   - temp-revert RED: examination_repository.go の ReplaceItemsByExamID / FindAllItemsByExamID の
//     dbOrTx(ctx, r.db) を r.db.WithContext(ctx) に戻すと、削除が独立 tx で即 commit され、
//     ambient tx の rollback では巻き戻らない → 既存 items が消えて RollsBack ケースが RED になる。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ambient tx 内で置換後に後続処理（監査書込）が失敗 → 削除がロールバックし既存 items が残存する。
// fail-closed（監査なしの検査結果削除を許さない）の DB-backed 正本テスト。
func TestExaminationRepository_ReplaceItemsByExamID_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査（原子性RB）")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
	})

	repo := NewExaminationRepository(db)
	_, _, err := repo.ReplaceItemsByExamID(ctx, clinicA, exam.ID, []model.ExamResult{
		{Name: "WBC", InspectionValue: "10.0"},
	})
	require.NoError(t, err, "既存 item の seed")

	// ambient tx 内で別の値へ置換した直後に、監査書込失敗を模倣して tx を中断する。
	tx := NewTransactor(db)
	sentinel := errors.New("simulated post-replace audit failure")
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, _, e := repo.ReplaceItemsByExamID(txCtx, clinicA, exam.ID, []model.ExamResult{
			{Name: "Glucose", InspectionValue: "90"},
		}); e != nil {
			return e
		}
		return sentinel // fail-closed: 置換後の監査書込失敗 → tx を中断
	})
	require.Error(t, txErr, "ambient tx 内の後続失敗で WithTx はエラーを返す")

	// 削除はロールバックされ、元の item が残存する。新しい item は挿入されていない。
	after, e := repo.FindAllItemsByExamID(ctx, clinicA, exam.ID)
	require.NoError(t, e)
	require.Len(t, after, 1, "監査失敗で削除はロールバックされ既存 item が残存しなければならない（fail-closed）")
	assert.Equal(t, "WBC", after[0].Name, "残存するのは置換前の元 item")
	assert.Equal(t, "10.0", after[0].InspectionValue)
}

// ambient tx 内で置換が成功し commit されると新 items が永続化される（原子コミット）。
// tx 内 read-your-writes（dbOrTx 化した FindAllItemsByExamID）で置換結果が返ることも確認する。
func TestExaminationRepository_ReplaceItemsByExamID_CommitsWithinAmbientTx(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査（原子性CM）")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC),
	})

	repo := NewExaminationRepository(db)

	// 既存 1 件を seed: tx 外の独立 tx で挿入する。
	// ambient tx 内の ReplaceItemsByExamID がこの 1 件を削除し deleted=1 を返す。
	_, _, err := repo.ReplaceItemsByExamID(ctx, clinicA, exam.ID, []model.ExamResult{
		{Name: "WBC（初期）", InspectionValue: "9.0"},
	})
	require.NoError(t, err, "既存 item の seed")

	tx := NewTransactor(db)
	var savedInTx []model.ExamResult
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		s, deleted, e := repo.ReplaceItemsByExamID(txCtx, clinicA, exam.ID, []model.ExamResult{
			{Name: "Glucose", InspectionValue: "90"},
		})
		if e != nil {
			return e
		}
		require.EqualValues(t, 1, deleted, "既存 1 件を置換＝実削除数 1（#211 監査ゲートの根拠と同方針）")
		savedInTx = s
		return nil
	})
	require.NoError(t, txErr)
	require.Len(t, savedInTx, 1, "tx 内 read-your-writes で置換結果が返る")
	assert.Equal(t, "Glucose", savedInTx[0].Name)

	after, e := repo.FindAllItemsByExamID(ctx, clinicA, exam.ID)
	require.NoError(t, e)
	require.Len(t, after, 1, "commit 後は新 item が永続化される")
	assert.Equal(t, "Glucose", after[0].Name)
	assert.Equal(t, "90", after[0].InspectionValue)
}

// BE-refactor.md H-8d: examinationRepository.Delete が dbOrTx(ctx, r.db) で ambient tx に参加する
// ことを実証する（examinationService.Delete が finalize ロック確認・FK チェック・Delete を
// s.transactor.WithTx で束ねるようになったための追加）。ambient tx 内で削除後に後続処理が失敗すると
// 削除がロールバックし、対象 exam が DB に残存することを確認する。
//
//   - temp-revert RED: examination_repository.go の Delete の dbOrTx(ctx, r.db) を r.db.WithContext(ctx)
//     に戻すと、削除が独立 tx で即 commit され、ambient tx の rollback では巻き戻らない → 削除後の
//     exam が消えたままになり RollsBack ケースが RED になる。
func TestExaminationRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査（Delete原子性RB）")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC),
	})

	repo := NewExaminationRepository(db)

	tx := NewTransactor(db)
	sentinel := errors.New("simulated post-delete failure")
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if e := repo.Delete(txCtx, clinicA, exam.ID); e != nil {
			return e
		}
		return sentinel // fail-closed: 削除後の後続失敗 → tx を中断
	})
	require.Error(t, txErr, "ambient tx 内の後続失敗で WithTx はエラーを返す")

	// 削除はロールバックされ、exam が残存する。
	found, e := repo.FindByID(ctx, clinicA, exam.ID)
	require.NoError(t, e, "後続失敗で削除はロールバックされ exam が残存しなければならない（fail-closed）")
	assert.Equal(t, exam.ID, found.ID)
}

// ambient tx 内で削除が成功し commit されると exam が永続化され消える（原子コミット）。
func TestExaminationRepository_Delete_CommitsWithinAmbientTx(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査（Delete原子性CM）")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC),
	})

	repo := NewExaminationRepository(db)

	tx := NewTransactor(db)
	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Delete(txCtx, clinicA, exam.ID)
	})
	require.NoError(t, txErr)

	_, e := repo.FindByID(ctx, clinicA, exam.ID)
	require.Error(t, e, "commit 後は exam が削除されている")
	assert.True(t, apperrors.IsNotFound(e))
}
