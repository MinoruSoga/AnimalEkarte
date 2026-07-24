package repository

import (
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// db_setup_test.go — shared implementation lives in backend/internal/testdb（サブパッケージ
// から import 可能にするため）。ここは既存 165 テストファイルの呼び出し口を無変更に保つための
// 薄い alias のみを持つ。詳細な設計意図・キャッシュ戦略のコメントは testdb/testdb.go 側を参照。

// TestMain は internal/repository package の全テストで共有する DB 接続プールを管理する。
// 個々のテストは接続を閉じず（testdb.SetupTestDB 内の sync.Once で一度だけ確立・全テストで再利用）、
// プロセス終了時にここで一度だけ閉じる。
func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupTestDB(t)
}

func setupIsolatedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupIsolatedTestDB(t)
}

func ensureClinicSettingsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	testdb.EnsureClinicSettingsTable(t, db)
}

func makeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

func ensureAutoMigrated(db *gorm.DB, models ...any) error {
	return testdb.EnsureAutoMigrated(db, models...)
}

func markAutoMigrated(models ...any) {
	testdb.MarkAutoMigrated(models...)
}
