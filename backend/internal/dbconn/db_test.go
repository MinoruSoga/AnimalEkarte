package dbconn

// db_test.go — dbconn.OpenGORM の runtime integration test。
//
// 保護する不変条件:
//   - dbconn.OpenGORM は有効な DSN であればプール設定（MaxOpenConns=50 等）を適用した *gorm.DB を返す。
//   - dbconn.OpenGORM は DSN 解析に失敗した場合 nil と "failed to open database connection" でラップしたエラーを返す
//     （ネットワーク接続を必要としない sslmode 不正で決定的に再現する）。

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/testdb"
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
		// config.Load() を経由しない手組み Config のため、DB接続プール上限
		// (fe6a27303 で追加)も config.go の既定値と揃えて明示する。
		// 未設定のままだとゼロ値になり TestNewDB の SetMaxOpenConns(50) 検証が失敗する。
		DBMaxOpenConns: 50,
		DBMaxIdleConns: 25,
	}
}

func TestOpenGORMRuntime(t *testing.T) {
	t.Run("有効なDSNでプール設定込みの接続を返す", func(t *testing.T) {
		// setupTestDB を経由してテストDBの存在を保証する（CREATE DATABASE の副作用を再利用）。
		_ = testdb.SetupTestDB(t)

		db, err := OpenGORM(testDBConfig(t))
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

		db, err := OpenGORM(cfg)
		require.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to open database connection")
	})
}
