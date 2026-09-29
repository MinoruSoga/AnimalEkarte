package medicalrecord

// checkup_type_field_tx_atomicity_test.go — EMR-225 フィールド定義 CRUD の ambient tx 参加（DB-backed）
//
// checkup_type_fields の write 系メソッド（CreateField / UpdateField / DeleteField /
// LockFieldByID / ReorderFields）が persistence.DBOrTx 経由で caller の ambient transaction に
// 参加し、tx 内の後続処理（= 監査書込相当）が失敗したら全変更がロールバックされることを
// 実 DB で実証する。マスタ定義の部分更新が残ると項目型変更・並び順・削除が中途半端に
// 永続化されるため fail-closed 必須（checkup_field_result_tx_atomicity_test.go と同型）。
//
//   - この原子性は mock では検証不可能。repository が DBOrTx で ambient tx に join することが正本。
//   - temp-revert RED: checkup_field_repository.go の checkupTypeFieldRepository 各メソッドの
//     persistence.DBOrTx(ctx, r.db) を r.db.WithContext(ctx) に戻すと、write が独立 tx で
//     即 commit され ambient tx の rollback では巻き戻らない → RollsBack ケースが RED になる。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// RollsBack: ambient tx 内で Lock→Update→Create→Reorder→Delete の全 write を実行した後に
// 後続失敗（監査書込を模倣）で tx を中断 → 全変更がロールバックされ初期状態が残存する。
func TestCheckupTypeFieldRepository_Writes_RollBackWhenAmbientTxFails(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	ct := makeCheckupTypeMaster(t, db, clinicA, "原子性健診")
	fieldA := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "体温",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 1,
	})
	fieldB := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "体重",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 2,
	})

	repo := NewCheckupTypeFieldRepository(db)
	sentinel := errors.New("simulated post-write audit failure")
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		if _, e := repo.LockFieldByID(txCtx, clinicA, ct.ID, fieldA.ID); e != nil {
			return e
		}
		if _, e := repo.UpdateField(txCtx, clinicA, ct.ID, fieldA.ID,
			map[string]any{"name": "体温（更新済）"}); e != nil {
			return e
		}
		fieldC := &model.CheckupTypeField{
			ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "触診",
			FieldType: model.CheckupFieldTypeBoolean, SortOrder: 3,
		}
		if e := repo.CreateField(txCtx, fieldC); e != nil {
			return e
		}
		if e := repo.ReorderFields(txCtx, clinicA, ct.ID, []uint64{fieldC.ID, fieldA.ID, fieldB.ID}); e != nil {
			return e
		}
		if e := repo.DeleteField(txCtx, clinicA, ct.ID, fieldB.ID); e != nil {
			return e
		}
		return sentinel // fail-closed: write 後の監査失敗で tx を中断
	})
	require.Error(t, txErr, "ambient tx 内の後続失敗で WithTx はエラーを返す")

	// 全 write がロールバックされ、初期の 2 件・元の name/sort_order が残存する。
	after, e := repo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, e)
	require.Len(t, after, 2, "create/delete はロールバックされ 2 件のまま残る")
	assert.Equal(t, "体温", after[0].Name, "UpdateField はロールバックされ元名が残る")
	assert.EqualValues(t, 1, after[0].SortOrder, "ReorderFields はロールバックされ元の並びが残る")
	assert.Equal(t, "体重", after[1].Name, "DeleteField はロールバックされ行が残る")
	assert.EqualValues(t, 2, after[1].SortOrder)
}

// CommitsWithinAmbientTx: ambient tx 内の write が成功し commit されると変更が永続化され、
// tx 内 read-your-writes（LockFieldByID / FindByCheckupTypeID 経由の同一 tx 参照）が成立する。
func TestCheckupTypeFieldRepository_Writes_CommitWithinAmbientTx(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	ct := makeCheckupTypeMaster(t, db, clinicA, "原子性健診2")
	fieldA := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "体温",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 1,
	})
	fieldB := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "体重",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 2,
	})

	repo := NewCheckupTypeFieldRepository(db)
	var fieldCID uint64
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		fieldC := &model.CheckupTypeField{
			ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "触診",
			FieldType: model.CheckupFieldTypeBoolean, SortOrder: 3,
		}
		if e := repo.CreateField(txCtx, fieldC); e != nil {
			return e
		}
		fieldCID = fieldC.ID
		if e := repo.ReorderFields(txCtx, clinicA, ct.ID, []uint64{fieldC.ID, fieldA.ID, fieldB.ID}); e != nil {
			return e
		}
		if e := repo.DeleteField(txCtx, clinicA, ct.ID, fieldB.ID); e != nil {
			return e
		}
		// tx 内 read-your-writes: 同一 tx からロック読みで作成行が参照できる。
		locked, e := repo.LockFieldByID(txCtx, clinicA, ct.ID, fieldC.ID)
		if e != nil {
			return e
		}
		assert.Equal(t, "触診", locked.Name)
		return nil
	})
	require.NoError(t, txErr)
	require.NotZero(t, fieldCID, "CreateField は tx 内で採番される")

	// commit 後は [C(sort 1), A(sort 2)] が生存し、B はソフトデリート済み。
	after, e := repo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, e)
	require.Len(t, after, 2, "delete 済み B は生存一覧に出ない")
	assert.Equal(t, "触診", after[0].Name)
	assert.EqualValues(t, 1, after[0].SortOrder)
	assert.Equal(t, "体温", after[1].Name)
	assert.EqualValues(t, 2, after[1].SortOrder)
}
