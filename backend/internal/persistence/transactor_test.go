package persistence

// transactor_test.go — Transactor (gormTransactor.WithTx / txFromContext) の単体テスト。
//
// 保護する不変条件:
//   - WithTx はコールバックが成功すればコミットする。
//   - WithTx はコールバックがエラーを返せばロールバックし、"transaction failed" でラップしたエラーを返す。
//   - WithTx 内では txFromContext がトランザクション付き *gorm.DB を返す。
//   - アンビエントtxが無いコンテキストでは txFromContext は nil を返す。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupTransactorTestDB は WithTx 検証用に clinic_integrations テーブルを用意する。
func setupTransactorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClinicIntegration{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE clinic_integrations CASCADE").Error)
	return db
}

func TestTxFromContext_NoAmbientTx_ReturnsNil(t *testing.T) {
	assert.Nil(t, TxFromContext(context.Background()))
}

func TestGormTransactor_WithTx_ExposesTxViaContext(t *testing.T) {
	db := setupTransactorTestDB(t)
	tx := NewTransactor(db)

	var sawTx *gorm.DB
	err := tx.WithTx(context.Background(), func(txCtx context.Context) error {
		sawTx = TxFromContext(txCtx)
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, sawTx, "WithTx 内部では txFromContext がトランザクションを返すべき")
}

func TestGormTransactor_WithTx_CommitsOnSuccess(t *testing.T) {
	db := setupTransactorTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		return TxFromContext(txCtx).Create(&model.ClinicIntegration{
			ClinicID: 601, Service: "lstep", KeyName: "k", KeyValue: "committed",
		}).Error
	})
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 601).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestGormTransactor_WithTx_RollsBackAndWrapsErrorOnFailure(t *testing.T) {
	db := setupTransactorTestDB(t)
	tx := NewTransactor(db)
	ctx := context.Background()

	sentinel := errors.New("boom")
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, TxFromContext(txCtx).Create(&model.ClinicIntegration{
			ClinicID: 602, Service: "lstep", KeyName: "k", KeyValue: "should-not-persist",
		}).Error)
		return sentinel
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transaction failed")
	assert.ErrorIs(t, err, sentinel, "元のエラーは Unwrap でたどれるはず")

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 602).Count(&count).Error)
	assert.Equal(t, int64(0), count, "エラー時はロールバックされる")
}
