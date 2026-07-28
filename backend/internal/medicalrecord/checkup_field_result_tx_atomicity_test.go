package medicalrecord

// checkup_field_result_tx_atomicity_test.go — #211 監査の tx 内原子性（DB-backed）
//
// checkup_field_results の置換（既存全削除→一括挿入）が caller の ambient transaction に
// 参加し、置換後に同一 tx 内の後続処理（= 監査書込）が失敗したら削除がロールバックして
// 既存結果が DB に残存することを実 DB で実証する。
//
//   - この原子性は mock では検証不可能（前段 security M2 の教訓）。repository が dbOrTx で
//     ambient tx に join することが正本である。
//   - temp-revert RED: checkup_field_repository.go の ReplaceForCheckup / FindByCheckupID の
//     dbOrTx(ctx, r.db) を r.db.WithContext(ctx) に戻すと、削除が独立 tx で即 commit され、
//     ambient tx の rollback では巻き戻らない → 既存結果が消えて RollsBack ケースが RED になる。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// SC3: ambient tx 内で置換後に後続処理（監査書込）が失敗 → 削除がロールバックし既存結果が残存する。
// fail-closed（監査なしの患者結果削除を許さない）の DB-backed 正本テスト。
func TestCheckupFieldResultRepository_ReplaceForCheckup_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "原子性飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "原子性ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-ATOMIC-RB", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石除去必要の有無",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	resultRepo := NewCheckupFieldResultRepository(db)
	boolTrue := true
	_, _, err := resultRepo.ReplaceForCheckup(ctx, clinicA, checkupID, []model.CheckupFieldResult{
		{ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
			FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, ValueBool: &boolTrue, SortOrder: 1},
	})
	require.NoError(t, err, "既存結果の seed")

	// ambient tx 内で別の値（number）へ置換した直後に、監査書込失敗を模倣して tx を中断する。
	// withTx は prescription_repository_test.go の repohelpers 直結 ambient-tx ヘルパー（移動元の
	// repository.Transactor.WithTx を import cycle なしで再現する。医療記録パッケージ共有）。
	score := 2.0
	newResults := []model.CheckupFieldResult{
		{ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
			FieldName: "歯石付着度", FieldType: model.CheckupFieldTypeNumber, ValueNumber: &score, SortOrder: 1},
	}
	sentinel := errors.New("simulated post-replace audit failure")
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		if _, _, e := resultRepo.ReplaceForCheckup(txCtx, clinicA, checkupID, newResults); e != nil {
			return e
		}
		return sentinel // fail-closed: 置換後の監査書込失敗 → tx を中断
	})
	require.Error(t, txErr, "ambient tx 内の後続失敗で WithTx はエラーを返す")

	// 削除はロールバックされ、元の boolean 結果が残存する。新しい number 結果は挿入されていない。
	after, e := resultRepo.FindByCheckupID(ctx, clinicA, checkupID)
	require.NoError(t, e)
	require.Len(t, after, 1, "監査失敗で削除はロールバックされ既存結果が残存しなければならない（fail-closed）")
	assert.Equal(t, model.CheckupFieldTypeBoolean, after[0].FieldType, "残存するのは置換前の元結果（boolean）")
	require.NotNil(t, after[0].ValueBool)
	assert.True(t, *after[0].ValueBool, "元の value_bool が保持される")
}

// SC2: ambient tx 内で置換が成功し commit されると新結果が永続化される（原子コミット）。
// tx 内 read-your-writes（dbOrTx 化した FindByCheckupID）で置換結果が返ることも確認する。
func TestCheckupFieldResultRepository_ReplaceForCheckup_CommitsWithinAmbientTx(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "原子性飼主2")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "原子性ポチ2")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-ATOMIC-CM", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石付着度",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	resultRepo := NewCheckupFieldResultRepository(db)

	// 既存 1 件を seed: tx 外の独立 tx で boolean 結果を挿入する。
	// ambient tx 内の ReplaceForCheckup がこの 1 件を削除し deleted=1 を返す。
	boolSeed := true
	_, _, err := resultRepo.ReplaceForCheckup(ctx, clinicA, checkupID, []model.CheckupFieldResult{
		{ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
			FieldName: "歯石付着度（初期）", FieldType: model.CheckupFieldTypeBoolean, ValueBool: &boolSeed, SortOrder: 1},
	})
	require.NoError(t, err, "既存結果の seed")

	score := 3.0
	newResults := []model.CheckupFieldResult{
		{ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
			FieldName: "歯石付着度", FieldType: model.CheckupFieldTypeNumber, ValueNumber: &score, SortOrder: 1},
	}
	var savedInTx []model.CheckupFieldResult
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		s, deleted, e := resultRepo.ReplaceForCheckup(txCtx, clinicA, checkupID, newResults)
		if e != nil {
			return e
		}
		require.EqualValues(t, 1, deleted, "既存 1 件を置換＝実削除数 1（#211 監査ゲートの根拠）")
		savedInTx = s
		return nil
	})
	require.NoError(t, txErr)
	require.Len(t, savedInTx, 1, "tx 内 read-your-writes で置換結果が返る")

	after, e := resultRepo.FindByCheckupID(ctx, clinicA, checkupID)
	require.NoError(t, e)
	require.Len(t, after, 1, "commit 後は新結果が永続化される")
	assert.Equal(t, model.CheckupFieldTypeNumber, after[0].FieldType)
	require.NotNil(t, after[0].ValueNumber)
	assert.InDelta(t, 3.0, *after[0].ValueNumber, 1e-9)
}
