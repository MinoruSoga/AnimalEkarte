package repository

// db_test.go — NewDB / isUniqueConstraintErr / isFKConstraintErr の単体テスト。
//
// 保護する不変条件:
//   - NewDB は有効な DSN であればプール設定（MaxOpenConns=50 等）を適用した *gorm.DB を返す。
//   - NewDB は DSN 解析に失敗した場合 nil と "failed to open database connection" でラップしたエラーを返す
//     （ネットワーク接続を必要としない sslmode 不正で決定的に再現する）。
//   - isUniqueConstraintErr / isFKConstraintErr は pgconn.PgError の SQLSTATE (23505 / 23503) を
//     errors.As のラップ解除込みで判定し、非PGエラー・nilはfalseを返す。

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
)

// testDBConfig は getTestDatabaseConnection (ltv_repository_test.go) と同じ既定値で
// テスト用データベース (ekarte_db_test 相当) を指す *config.Config を組み立てる。
func testDBConfig(t *testing.T) *config.Config {
	t.Helper()
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "ekarte_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ekarte_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ekarte_db"
	}
	return &config.Config{
		DBHost:    dbHost,
		DBPort:    dbPort,
		DBUser:    dbUser,
		DBPass:    dbPassword,
		DBName:    dbName + "_test",
		DBSSLMode: "disable",
	}
}

func TestNewDB(t *testing.T) {
	t.Run("有効なDSNでプール設定込みの接続を返す", func(t *testing.T) {
		// setupTestDB を経由してテストDBの存在を保証する（CREATE DATABASE の副作用を再利用）。
		_ = setupTestDB(t)

		db, err := NewDB(testDBConfig(t))
		require.NoError(t, err)
		require.NotNil(t, db)

		sqlDB, err := db.DB()
		require.NoError(t, err)
		defer func() { _ = sqlDB.Close() }()

		require.NoError(t, sqlDB.Ping())
		stats := sqlDB.Stats()
		assert.Equal(t, 50, stats.MaxOpenConnections, "SetMaxOpenConns(50) が適用されていること")
	})

	t.Run("不正なsslmodeはDSN解析時点でラップ済みエラーを返す（DB接続不要で決定的）", func(t *testing.T) {
		cfg := testDBConfig(t)
		cfg.DBSSLMode = "not-a-real-sslmode"

		db, err := NewDB(cfg)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to open database connection")
	})
}

func TestIsUniqueConstraintErr(t *testing.T) {
	t.Run("23505は一意制約違反として検出される", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.True(t, isUniqueConstraintErr(err))
	})

	t.Run("fmt.Errorfでラップされていてもerrors.Asで解除して検出される", func(t *testing.T) {
		wrapped := fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"})
		assert.True(t, isUniqueConstraintErr(wrapped))
	})

	t.Run("別のPGエラーコードはfalse", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.False(t, isUniqueConstraintErr(err))
	})

	t.Run("PG以外のエラーはfalse", func(t *testing.T) {
		assert.False(t, isUniqueConstraintErr(errors.New("plain error")))
	})

	t.Run("nilはfalse", func(t *testing.T) {
		assert.False(t, isUniqueConstraintErr(nil))
	})
}

func TestIsFKConstraintErr(t *testing.T) {
	t.Run("23503はFK制約違反として検出される", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.True(t, isFKConstraintErr(err))
	})

	t.Run("fmt.Errorfでラップされていてもerrors.Asで解除して検出される", func(t *testing.T) {
		wrapped := fmt.Errorf("delete failed: %w", &pgconn.PgError{Code: "23503"})
		assert.True(t, isFKConstraintErr(wrapped))
	})

	t.Run("別のPGエラーコードはfalse", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.False(t, isFKConstraintErr(err))
	})

	t.Run("PG以外のエラーはfalse", func(t *testing.T) {
		assert.False(t, isFKConstraintErr(errors.New("plain error")))
	})

	t.Run("nilはfalse", func(t *testing.T) {
		assert.False(t, isFKConstraintErr(nil))
	})
}
