package testdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unicode"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupIsolated001TestDB は最終 backend/migrations/001_init.sql だけを適用した
// 使い捨て PostgreSQL データベースを確保し、その *gorm.DB を返す。
//
// 既存の SetupTestDB / SetupIsolatedTestDB は GORM AutoMigrate の double であり、
// 001 が持つ複合 FK・RLS policy・deferred constraint trigger を再現しない。
// 001 の実 DDL を source にした検証が必要な呼び出し元だけが本入口を使う。
//
// ライフサイクルは SetupIsolatedTestDB に合わせる:
//
//	skipShortDatabase(t) → 使い捨てDB確保 → 001 適用 → cleanup で close + DROP DATABASE
//
// 接続先は環境変数 DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME
// （既定値は connectTestDatabase と同じ）から組み立てた、このテスト専用の
// 新規データベースのみ。共有テスト DB・開発 DB・STG・本番には触れない。
func SetupIsolated001TestDB(t *testing.T) *gorm.DB {
	t.Helper()
	skipShortDatabase(t)

	host, port, user, password, baseName := testDBConnEnv()

	mainDB, err := gorm.Open(postgres.Open(postgresDSN(user, password, host, port, baseName)), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to main db for 001 fixture: %v", err)
	}
	if sqlMain, derr := mainDB.DB(); derr == nil {
		t.Cleanup(func() { _ = sqlMain.Close() })
	}

	dbName := disposable001DatabaseName(baseName)
	if err := mainDB.Exec("CREATE DATABASE " + quotePostgresIdentifier(dbName)).Error; err != nil {
		t.Fatalf("create disposable 001 database %q: %v", dbName, err)
	}

	db, err := openDisposable001DB(user, password, host, port, dbName)
	if err != nil {
		if dropErr := mainDB.Exec("DROP DATABASE IF EXISTS " + quotePostgresIdentifier(dbName) + " WITH (FORCE)").Error; dropErr != nil {
			t.Errorf("cleanup after failed connect also failed; database %q may remain: %v", dbName, dropErr)
		}
		t.Fatalf("connect to disposable 001 database %q: %v", dbName, err)
	}
	// DB を落とす前に接続を閉じる必要があるため、close → DROP DATABASE を 1 つの
	// cleanup で順に実行する（mainDB 側の cleanup はその後に走る）。
	t.Cleanup(func() {
		if sqlDB, derr := db.DB(); derr == nil {
			_ = sqlDB.Close()
		}
		if err := mainDB.Exec("DROP DATABASE IF EXISTS " + quotePostgresIdentifier(dbName) + " WITH (FORCE)").Error; err != nil {
			t.Errorf("drop disposable 001 database %q: %v", dbName, err)
		}
	})

	schemaPath, err := migrations001SQLPath()
	if err != nil {
		t.Fatalf("resolve 001_init.sql path: %v", err)
	}
	ddl, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}

	// psql / cmd/migrate と同じく、1 ファイルをそのまま 1 回で適用する。
	// PreferSimpleProtocol により複数ステートメントが simple query として送られる。
	if err := db.Exec(string(ddl)).Error; err != nil {
		t.Fatalf("apply %s to disposable database %q: %v", filepath.Base(schemaPath), dbName, err)
	}

	return db
}

// testDBConnEnv は connectTestDatabase と同じ既定値で接続パラメータを返す。
func testDBConnEnv() (host, port, user, password, baseName string) {
	host = envOrDefault("DB_HOST", "db")
	port = envOrDefault("DB_PORT", "5432")
	user = envOrDefault("DB_USER", "ekarte_user")
	password = envOrDefault("DB_PASSWORD", "ekarte_password")
	baseName = envOrDefault("DB_NAME", "ekarte_db")
	return host, port, user, password, baseName
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func postgresDSN(user, password, host, port, dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
}

func openDisposable001DB(user, password, host, port, dbName string) (*gorm.DB, error) {
	return gorm.Open(postgres.New(postgres.Config{
		DSN:                  postgresDSN(user, password, host, port, dbName),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
}

// disposable001DatabaseName は 001 fixture 専用の使い捨て DB 名を作る。
// PostgreSQL の識別子制限（63 バイト・小文字英数字と _）を守る。
func disposable001DatabaseName(base string) string {
	safe := make([]rune, 0, len(base))
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			safe = append(safe, r)
		case unicode.IsUpper(r):
			safe = append(safe, unicode.ToLower(r))
		default:
			safe = append(safe, '_')
		}
	}
	name := string(safe) + "_001"
	// PostgreSQL 識別子は63バイト上限。base 側を先に切詰め、一意化 suffix（タイムスタンプ）
	// が欠落しないようにする（長い DB_NAME での並列・連続実行衝突を防ぐ）。
	suffix := fmt.Sprintf("_%d", time.Now().UnixNano())
	maxBase := 63 - len(suffix)
	if len(name) > maxBase {
		name = name[:maxBase]
	}
	return name + suffix
}

// migrations001SQLPath は backend/migrations/001_init.sql の絶対パスを runtime.Caller で
// 解決する（testdb_test.go と同じ方式。相対パス決め打ちをしない）。
func migrations001SQLPath() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed resolving testdb source path")
	}
	// backend/internal/testdb -> backend/migrations/001_init.sql
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "001_init.sql"), nil
}
