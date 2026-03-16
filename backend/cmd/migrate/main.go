package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 環境変数から DB 接続情報を取得
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		logger.Error("Missing required environment variables",
			slog.String("DB_HOST", dbHost),
			slog.String("DB_USER", dbUser),
			slog.Bool("DB_PASSWORD_SET", dbPassword != ""),
			slog.String("DB_NAME", dbName))
		os.Exit(1)
	}

	if dbPort == "" {
		dbPort = "5432"
	}

	// SSL mode を取得
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "require"
	}

	// PostgreSQL 接続文字列を構築
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode,
	)

	// DB に接続
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// 接続テスト
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Connected to database", slog.String("host", dbHost), slog.String("dbname", dbName))

	// DB_RESET 環境変数をチェック（本番リセット時のみ true）
	resetDB := os.Getenv("DB_RESET") == "true"
	if resetDB {
		logger.Warn("⚠️ DB_RESET=true: Dropping and recreating schema")
		if err := resetSchema(db, logger); err != nil {
			logger.Error("Failed to reset schema", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	// マイグレーション実行
	if err := runMigrations(db, logger); err != nil {
		logger.Error("Migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("✓ All migrations completed successfully")
}

// runMigrations はマイグレーションを順序通りに実行
func runMigrations(db *sql.DB, logger *slog.Logger) error {
	migrationsDir := "/app/migrations"

	// ディレクトリが存在するか確認
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	// マイグレーションファイルを読み込み
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// SQL ファイルを集める（読み込み時に自動的にソートされる）
	migrationFiles := []string{}
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	if len(migrationFiles) == 0 {
		logger.Warn("No migration files found")
		return nil
	}

	// マイグレーションを実行
	for _, filename := range migrationFiles {
		filePath := filepath.Join(migrationsDir, filename)

		logger.Info("Executing migration", slog.String("file", filename))

		// SQL ファイルを読み込み
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// トランザクション内で実行
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		logger.Info("✓ Migration completed", slog.String("file", filename))
	}

	return nil
}

// resetSchema はスキーマを削除して再作成する
func resetSchema(db *sql.DB, logger *slog.Logger) error {
	logger.Info("Dropping public schema...")

	// トランザクション内で実行
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// スキーマを削除（CASCADE で関連オブジェクトも削除）
	if _, err := tx.Exec("DROP SCHEMA public CASCADE"); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	logger.Info("Creating new public schema...")

	// 新しいスキーマを作成
	if _, err := tx.Exec("CREATE SCHEMA public"); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema reset: %w", err)
	}

	logger.Info("✓ Schema reset completed")
	return nil
}
