package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/animal-ekarte/backend/internal/config"

	_ "github.com/lib/pq"
)

// Advisory lock ID（プロジェクト固有の定数。並行実行を防止する）
const migrationLockID = 7283946501

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// run はマイグレーション処理全体を実行し、エラーを返す。
// defer が os.Exit に打ち消されないよう main から分離している。
func run(logger *slog.Logger) error {
	if err := config.ConfigureTimeZone(); err != nil {
		return fmt.Errorf("timezone configuration failed: %w", err)
	}

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
		return fmt.Errorf("missing required environment variables")
	}

	if dbPort == "" {
		dbPort = "5432"
	}

	// SSL mode を取得
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	// PostgreSQL 接続文字列を構築
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode, config.JapanTimeZone,
	)

	// DB に接続
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}
	defer db.Close() //nolint:errcheck // 接続クローズ失敗は復旧不可のため無視

	logger.Info("Connected to database", slog.String("host", dbHost), slog.String("dbname", dbName))

	// -------------------------------------------------------
	// Advisory Lock を取得（並行実行防止）
	// pg_advisory_lock はセッション終了時に自動解放される
	// -------------------------------------------------------
	logger.Info("Acquiring migration lock...")
	if err := acquireAdvisoryLock(db); err != nil {
		return fmt.Errorf("failed to acquire migration lock (another migration may be running): %w", err)
	}
	logger.Info("Migration lock acquired")

	// DB_RESET 環境変数をチェック（デフォルト: false）
	resetDB := os.Getenv("DB_RESET") == "true"
	if resetDB {
		logger.Warn("⚠️ DB_RESET=true: Dropping and recreating schema")
		if err := resetSchema(db, logger); err != nil {
			return fmt.Errorf("failed to reset schema: %w", err)
		}
	}

	// schema_migrations テーブルを作成（存在しなければ）
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 既存DBへの初回適用: baseline 処理
	// schema_migrations が空だが既にテーブルが存在する場合、
	// 全マイグレーションを「適用済み」として記録しスキップする
	if err := baselineIfNeeded(db, logger); err != nil {
		return fmt.Errorf("baseline failed: %w", err)
	}

	// マイグレーション実行
	if err := runMigrations(db, logger); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Info("✓ All migrations completed successfully")
	return nil
}

// acquireAdvisoryLock は PostgreSQL Advisory Lock を取得する
// 別のマイグレーションプロセスがロックを保持している場合はブロックする
func acquireAdvisoryLock(db *sql.DB) error {
	_, err := db.Exec("SELECT pg_advisory_lock($1)", migrationLockID)
	return err
}

// ensureMigrationsTable は schema_migrations テーブルを作成する（冪等）
func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename    VARCHAR(255) PRIMARY KEY,
			checksum    VARCHAR(64)  NOT NULL,
			executed_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// baselineIfNeeded は既存DBへの初回適用時に全マイグレーションを「適用済み」として記録する
// 条件: schema_migrations が空 AND 既にアプリケーションテーブルが存在する
func baselineIfNeeded(db *sql.DB, logger *slog.Logger) error {
	// schema_migrations にレコードがあれば baseline 不要
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		return fmt.Errorf("failed to count schema_migrations: %w", err)
	}
	if count > 0 {
		return nil
	}

	// アプリケーションテーブルの存在チェック（clinics は最も基本的なテーブル）
	var exists bool
	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'clinics'
		)
	`).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check for existing tables: %w", err)
	}
	if !exists {
		// テーブルが存在しない → 新規DB。baseline 不要
		return nil
	}

	// 既存DB: 全マイグレーションファイルを適用済みとして記録
	logger.Warn("⚠️ Existing database detected with no migration history — running baseline")

	migrationsDir := "/app/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin baseline transaction: %w", err)
	}

	baselined := 0
	for _, entry := range files {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		checksum := fileChecksum(content)
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (filename, checksum, executed_at) VALUES ($1, $2, $3)",
			entry.Name(), checksum, time.Now(),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to baseline %s: %w", entry.Name(), err)
		}

		logger.Info("Baselined", slog.String("file", entry.Name()))
		baselined++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit baseline: %w", err)
	}

	logger.Info("✓ Baseline completed", slog.Int("files", baselined))
	return nil
}

// isAlreadyApplied は指定ファイルが既に適用済みかチェックする
// checksum が変更されている場合はエラーを返す（手動改変検出）
func isAlreadyApplied(db *sql.DB, filename, checksum string) (bool, error) {
	var storedChecksum string
	err := db.QueryRow(
		"SELECT checksum FROM schema_migrations WHERE filename = $1",
		filename,
	).Scan(&storedChecksum)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query schema_migrations: %w", err)
	}

	// 適用済みだが checksum が異なる → 改変されている
	if storedChecksum != checksum {
		return true, fmt.Errorf(
			"checksum mismatch for %s: applied=%s, current=%s — migration file was modified after execution",
			filename, storedChecksum, checksum,
		)
	}

	return true, nil
}

// recordMigration は実行済みマイグレーションを記録する
func recordMigration(tx *sql.Tx, filename, checksum string) error {
	_, err := tx.Exec(
		"INSERT INTO schema_migrations (filename, checksum, executed_at) VALUES ($1, $2, $3)",
		filename, checksum, time.Now(),
	)
	return err
}

// fileChecksum はファイル内容の SHA-256 ハッシュを返す
func fileChecksum(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// runMigrations はマイグレーションを順序通りに実行（実行済みはスキップ）
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

	applied := 0
	skipped := 0

	for _, filename := range migrationFiles {
		filePath := filepath.Join(migrationsDir, filename)

		// SQL ファイルを読み込み
		content, err := os.ReadFile(filePath) //nolint:gosec // マイグレーションファイルパスは管理者制御下の固定ディレクトリのみ
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		checksum := fileChecksum(content)

		// 適用済みチェック
		alreadyApplied, err := isAlreadyApplied(db, filename, checksum)
		if err != nil {
			return err
		}
		if alreadyApplied {
			logger.Info("⏭ Skipping (already applied)", slog.String("file", filename))
			skipped++
			continue
		}

		// トランザクション内で実行 + 記録
		logger.Info("Executing migration", slog.String("file", filename))

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if err := recordMigration(tx, filename, checksum); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		logger.Info("✓ Migration completed", slog.String("file", filename))
		applied++
	}

	logger.Info("Migration summary",
		slog.Int("applied", applied),
		slog.Int("skipped", skipped),
		slog.Int("total", len(migrationFiles)))

	return nil
}

// resetSchema はスキーマを削除して再作成する
func resetSchema(db *sql.DB, logger *slog.Logger) error {
	logger.Info("Dropping public schema...")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if _, err := tx.Exec("DROP SCHEMA public CASCADE"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	logger.Info("Creating new public schema...")

	if _, err := tx.Exec("CREATE SCHEMA public"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to create schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema reset: %w", err)
	}

	logger.Info("✓ Schema reset completed")
	return nil
}
