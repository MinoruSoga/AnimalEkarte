package repository

// repositories_test.go — Repositories (DIコンテナ) の統合テスト。
//
// 保護する不変条件:
//   - NewRepositories はすべてのリポジトリフィールドを初期化する（nilが残らない）。
//   - DB() はコンストラクタに渡した *gorm.DB をそのまま返す。
//   - Transaction は TransactionFn が未設定の場合、実際のDBトランザクションを開始し、
//     コールバックが成功すればコミット、エラーを返せばロールバックして
//     "transaction failed" でラップしたエラーを返す。
//   - Transaction は TransactionFn が設定されていればそちらを使う（テスト用差し替え経路）。

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupRepositoriesTestDB は Repositories.Transaction 検証用に clinic_integrations テーブルを用意する。
func setupRepositoriesTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.ClinicIntegration{}))
	db.Exec("TRUNCATE TABLE clinic_integrations CASCADE")
	return db
}

// TestNewRepositories_InitializesAllRepositories は NewRepositories が生成する全リポジトリ
// フィールドが nil でないことを reflect で検証する。フィールド追加時も自動的に対象へ入る。
func TestNewRepositories_InitializesAllRepositories(t *testing.T) {
	db := setupRepositoriesTestDB(t)
	repos := NewRepositories(db)

	v := reflect.ValueOf(*repos)
	tp := v.Type()
	checked := 0
	for i := 0; i < v.NumField(); i++ {
		field := tp.Field(i)
		if !field.IsExported() {
			continue // db は非公開フィールド
		}
		if field.Name == "TransactionFn" {
			continue // 関数フィールドで、本番動作ではnilが正しい既定値
		}
		fv := v.Field(i)
		switch fv.Kind() { //nolint:exhaustive // interfaceフィールドのみ判定すれば十分
		case reflect.Interface, reflect.Ptr:
			assert.False(t, fv.IsNil(), "field %s should be initialized by NewRepositories", field.Name)
			checked++
		}
	}
	assert.Greater(t, checked, 50, "NewRepositoriesが初期化するリポジトリの大半を検査できていること")
}

func TestRepositories_DB(t *testing.T) {
	db := setupRepositoriesTestDB(t)
	repos := NewRepositories(db)
	assert.Same(t, db, repos.DB(), "DB()はコンストラクタに渡したインスタンスをそのまま返す")
}

func TestRepositories_Transaction_CommitsOnSuccess(t *testing.T) {
	db := setupRepositoriesTestDB(t)
	repos := NewRepositories(db)
	ctx := context.Background()

	err := repos.Transaction(ctx, func(txRepos *Repositories) error {
		return txRepos.LstepSettings.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: 501, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "tx-committed",
		})
	})
	require.NoError(t, err)

	var stored model.ClinicIntegration
	require.NoError(t, db.Where("clinic_id = ?", 501).First(&stored).Error)
	assert.Equal(t, "tx-committed", stored.KeyValue)
}

func TestRepositories_Transaction_RollsBackOnError(t *testing.T) {
	db := setupRepositoriesTestDB(t)
	repos := NewRepositories(db)
	ctx := context.Background()

	sentinel := errors.New("boom")
	err := repos.Transaction(ctx, func(txRepos *Repositories) error {
		require.NoError(t, txRepos.LstepSettings.Upsert(ctx, &model.ClinicIntegration{
			ClinicID: 502, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLstepAPIKey, KeyValue: "should-not-persist",
		}))
		return sentinel
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transaction failed")

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).Where("clinic_id = ?", 502).Count(&count).Error)
	assert.Equal(t, int64(0), count, "エラー時はロールバックされ行が残らないこと")
}

func TestRepositories_Transaction_UsesTransactionFnOverrideWhenSet(t *testing.T) {
	db := setupRepositoriesTestDB(t)
	repos := NewRepositories(db)

	called := false
	repos.TransactionFn = func(_ context.Context, fn func(*Repositories) error) error {
		called = true
		return fn(repos) // 実DBトランザクションを介さずインラインで実行
	}

	err := repos.Transaction(context.Background(), func(_ *Repositories) error { return nil })
	require.NoError(t, err)
	assert.True(t, called, "TransactionFn が設定されている場合はそちらが使われるべき")
}
