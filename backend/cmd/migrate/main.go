package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/seedbundle"
)

// Advisory lock ID（プロジェクト固有の定数。並行実行を防止する）
const migrationLockID = 7283946501

// migrationsDir is the fixed on-disk location of both *.sql migration files
// and the seeds/ bundle directory (mounted read-only into the backend
// container). Shared by runSQLMigrations and runSeedBundles so both agree on
// the same root.
const migrationsDir = "/app/migrations"

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

	// 環境変数から DB 接続情報を取得（DB_NAME はこのツール固有の必須項目のため個別チェック）
	conn, err := dbconn.FromEnv()
	dbName := os.Getenv("DB_NAME")
	if err != nil || dbName == "" {
		logger.Error("Missing required environment variables",
			slog.String("DB_HOST", conn.Host),
			slog.String("DB_USER", conn.User),
			slog.Bool("DB_PASSWORD_SET", conn.Password != ""),
			slog.String("DB_NAME", dbName))
		return fmt.Errorf("missing required environment variables")
	}
	dbHost := conn.Host

	// PostgreSQL 接続文字列を構築
	connStr := conn.DSN(dbName)

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

	// migration履歴とアプリケーションschemaの存在が一致しないDBは、その完全性を
	// 検証できないためfail-closedにする。この読取り専用guardはlegacy key翻訳より先に
	// 実行し、矛盾状態へ001やseedの適用済み記録を書き込ませない。
	if err := guardEmptyMigrationHistory(db); err != nil {
		return fmt.Errorf("migration history safety check failed: %w", err)
	}

	// 旧形式（002-004 stub SQL 由来）の schema_migrations キー検出
	// 検出したキーは seeds/<bundle> の現行キーへ翻訳して適用済み記録する（P1-3, PR #186 review）。
	// seedキーの翻訳だけを理由とするresetは不要。ただし2026-07-27の001統合前checksumを持つDBは
	// 後続runSQLMigrationsで失敗し、USER承認済みのDB_RESET/再構築が必要になる。
	if err := detectLegacySeedKeys(db, logger); err != nil {
		return err
	}

	// フェーズ1: DDL migration（直下の *.sql）を昇順適用
	if err := runSQLMigrations(db, logger); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// フェーズ2: CSV シードバンドルを APP_ENV ゲート付き BundleOrder でロード。
	// production / empty / unknown は master のみ（demo の system admin を投入しない）。
	// connStr は pgx（CSVシードバンドルのCOPYロード用）でもそのまま使える
	// libpq形式のDSNのため、lib/pq用に構築したものを使い回す。
	if err := runSeedBundles(context.Background(), db, connStr, logger); err != nil {
		return fmt.Errorf("seed bundle load failed: %w", err)
	}

	// 両フェーズ完了後: ディスク上の期待キーと schema_migrations を突合する。
	// 欠落のみ fail-closed。余剰（統合・削除済みファイルの履歴）は情報ログのみ。
	if err := verifyMigrationKeyCoverage(db, logger); err != nil {
		return err
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

// detectLegacySeedKeys は 2026-07 の stub SQL 削除（002-004 の CSV-only 移行）より前の
// バイナリが記録した seed 由来の schema_migrations キー（例: "002_seed_master.sql"）が
// 残っていないか確認する。それらのキーは stub SQL ファイル自体が削除された現行バイナリでは
// 二度と生成されない。検出した場合は translateLegacySeedKeys で旧stubに対応する現行キー
// （"seeds/002_master" 等）を適用済み記録する。seedキーの翻訳自体はdb_resetを要求しない
// （P1-3, PR #186 review）が、DDLのchecksum検査を迂回するものではない。2026-07-27の
// 001統合前checksumを持つDBは後続runSQLMigrationsで失敗し、USER承認済みの
// DB_RESET/再構築が必要になる。
//
// DB アクセス（全 filename の読み出し）と判定ロジック（legacyKeysAmong）を分離しているのは、
// 判定ロジック単体を DB 接続なしでユニットテストできるようにするため。
func detectLegacySeedKeys(db *sql.DB, logger *slog.Logger) error {
	rows, err := db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to read schema_migrations for legacy seed key check: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query; close failure is not actionable here

	var applied []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return fmt.Errorf("failed to scan schema_migrations filename: %w", err)
		}
		applied = append(applied, filename)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate schema_migrations: %w", err)
	}

	found := legacyKeysAmong(applied)
	if len(found) == 0 {
		return nil
	}

	logger.Warn("⚠️ Detected legacy seed migration key(s) — baselining legacy-equivalent seeds/<bundle> keys",
		slog.String("legacy_keys", strings.Join(found, ", ")))
	if err := translateLegacySeedKeys(db, logger); err != nil {
		return fmt.Errorf("failed to translate legacy seed key(s) %s: %w", strings.Join(found, ", "), err)
	}
	return nil
}

// legacyTranslationTargets returns the schema_migrations keys that
// translateLegacySeedKeys marks applied whenever ANY legacy key is found.
// It always returns all three bundles that have a legacy stub equivalent,
// never only the bundles whose specific legacy filename was found in
// schema_migrations. Bundles introduced after the stub era must remain
// eligible for normal application. Pure, no DB access — this is what the
// translation unit tests exercise directly.
//
// Why "all", not "only the ones found" (PR #186 security review, HIGH): the
// pre-2026-07 binary applied every *.sql file unconditionally in one pass, so
// a routinely-migrated DB carries all three legacy keys together. But nothing
// guarantees that invariant for every real DB (e.g. one hand-curated to skip
// demo/staging on purpose) — marking only the found subset would leave the
// other seeds/<bundle> keys "unapplied", and the runSeedBundles call right
// after this would then auto-COPY those CSV bundles onto what may be a real
// database. guardEmptyMigrationHistoryは履歴が空の既存schemaを拒否するが、legacy keyが
// 存在するDBはその対象ではないため、translateLegacySeedKeysがlegacy相当3件を一括して
// 適用済み記録する保守的な姿勢を維持し、found key単位では判断しない。
func legacyTranslationTargets() []string {
	legacyBundleDirs := [...]string{"002_master", "003_demo", "004_staging"}
	keys := make([]string, 0, len(legacyBundleDirs))
	for _, bundleDir := range legacyBundleDirs {
		keys = append(keys, seedbundle.BundleMigrationKey(bundleDir))
	}
	return keys
}

// translateLegacySeedKeys records every legacy-equivalent seeds/<bundle> key
// (legacyTranslationTargets) not already recorded — idempotent (an EXISTS
// check guards each insert, so a rerun or an already-translated/normally-
// applied bundle is a no-op). It never deletes the legacy row(s) that
// triggered it: leaving them in place is harmless once the current keys
// exist, and avoids treating an audit trail row as disposable. Normal seed
// 記録と同じrecordMigration/bundleChecksum helperを使うが、これはlegacy stub相当を
// "mark applied without loading"する翻訳であり、live CSV importではないため
// application dataには触れない。
func translateLegacySeedKeys(db *sql.DB, logger *slog.Logger) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin legacy seed key translation transaction: %w", err)
	}

	translated := 0
	for _, newKey := range legacyTranslationTargets() {
		var exists bool
		if err := tx.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)", newKey,
		).Scan(&exists); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to check existing seed bundle key %s: %w", newKey, err)
		}
		if exists {
			// Already translated or normally applied under the current key.
			continue
		}

		bundleDir := strings.TrimPrefix(newKey, "seeds/")
		checksum, err := bundleChecksum(migrationsDir, bundleDir)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to compute checksum for seed bundle %s: %w", bundleDir, err)
		}

		if err := recordMigration(tx, newKey, checksum); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record translated seed bundle key %s: %w", newKey, err)
		}

		logger.Info("Translated legacy seed key set to current bundle key", slog.String("current", newKey))
		translated++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit legacy seed key translation: %w", err)
	}

	if translated > 0 {
		logger.Info("✓ Legacy seed key translation completed", slog.Int("translated", translated))
	}
	return nil
}

// legacyKeysAmong returns which of seedbundle.LegacyStubFilenames appear in
// appliedFilenames (the full set of schema_migrations.filename values
// currently recorded), preserving LegacyStubFilenames order. Pure function,
// no DB access — this is what detectLegacySeedKeys's unit tests exercise
// directly.
func legacyKeysAmong(appliedFilenames []string) []string {
	applied := make(map[string]bool, len(appliedFilenames))
	for _, f := range appliedFilenames {
		applied[f] = true
	}

	var found []string
	for _, legacy := range seedbundle.LegacyStubFilenames {
		if applied[legacy] {
			found = append(found, legacy)
		}
	}
	return found
}

func validateBaselineSafety(migrationCount int, hasApplicationSchema bool) error {
	if migrationCount == 0 && hasApplicationSchema {
		return errors.New(
			"existing application schema detected while schema_migrations is empty; " +
				"schema completeness cannot be verified; rebuild with USER-approved DB_RESET " +
				"following docs/ops/deploy/LOCAL_DB_RESET.md",
		)
	}
	if migrationCount > 0 && !hasApplicationSchema {
		return errors.New(
			"migration history exists while application schema is missing; " +
				"schema completeness cannot be verified; rebuild with USER-approved DB_RESET " +
				"following docs/ops/deploy/LOCAL_DB_RESET.md",
		)
	}
	return nil
}

// guardEmptyMigrationHistory rejects either direction of inconsistency between
// migration history and the application schema. It is read-only: current DDL
// and seed checksums must only be recorded after their normal application paths
// complete successfully.
func guardEmptyMigrationHistory(db *sql.DB) error {
	var migrationCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount); err != nil {
		return fmt.Errorf("failed to count schema_migrations: %w", err)
	}

	var hasApplicationSchema bool
	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'clinics'
		)
	`).Scan(&hasApplicationSchema); err != nil {
		return fmt.Errorf("failed to check for existing tables: %w", err)
	}
	return validateBaselineSafety(migrationCount, hasApplicationSchema)
}

// isAlreadyApplied は指定ファイルが既に適用済みかチェックする
// checksum が変更されている場合はエラーを返す（手動改変検出）
func isAlreadyApplied(db *sql.DB, filename, checksum string) (bool, error) {
	var storedChecksum string
	err := db.QueryRow(
		"SELECT checksum FROM schema_migrations WHERE filename = $1",
		filename,
	).Scan(&storedChecksum)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query schema_migrations: %w", err)
	}

	// 適用済みだが checksum が異なる → 改変されている
	if storedChecksum != checksum {
		repaired, repairErr := tryRepairKnownChecksumDrift(db, filename, storedChecksum, checksum)
		if repairErr != nil {
			return true, repairErr
		}
		if repaired {
			return true, nil
		}
		return true, fmt.Errorf(
			"checksum mismatch for %s: applied=%s, current=%s — migration file was modified after execution",
			filename, storedChecksum, checksum,
		)
	}

	return true, nil
}

// knownChecksumRepairs maps filename → appliedChecksum → currentChecksum for
// STG-safe drifts already reviewed (additive DDL only; never destructive).
// Companion side-effects for a pair live in tryRepairKnownChecksumDrift.
var knownChecksumRepairs = map[string]map[string]string{
	// 2026-08: lab_import_job_status gained 'reverted' (CREATE TYPE + ADD VALUE IF NOT EXISTS).
	"001_init.sql": {
		"28e954b32fd606a122e0cb29815ea277f8a96cb0966208f39e6fe69dd8cb9c4e": "287bfce66c810503c43c8a5c1d4cf414f561af2555314eb4119be74253ce77ce",
	},
}

func tryRepairKnownChecksumDrift(db *sql.DB, filename, applied, current string) (bool, error) {
	wantByApplied, ok := knownChecksumRepairs[filename]
	if !ok {
		return false, nil
	}
	want, ok := wantByApplied[applied]
	if !ok || want != current {
		return false, nil
	}

	// Ensure additive enum value exists before accepting the new checksum.
	// ADD VALUE IF NOT EXISTS cannot always run inside an explicit transaction.
	if filename == "001_init.sql" {
		if _, err := db.Exec(`ALTER TYPE lab_import_job_status ADD VALUE IF NOT EXISTS 'reverted'`); err != nil {
			return false, fmt.Errorf("checksum repair companion SQL failed for %s: %w", filename, err)
		}
	}

	res, err := db.Exec(
		`UPDATE schema_migrations SET checksum = $1 WHERE filename = $2 AND checksum = $3`,
		current, filename, applied,
	)
	if err != nil {
		return false, fmt.Errorf("checksum repair update failed for %s: %w", filename, err)
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return false, fmt.Errorf("checksum repair updated %d rows for %s (want 1)", n, filename)
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

// runSQLMigrations はフェーズ1: 直下の *.sql migration ファイル（2026-07-17 に
// 旧 incremental 002–011 を畳み込んだ統合スキーマ 001_init.sql と、それ以降に追加した
// append-only incremental）を昇順実行する（実行済みはスキップ）。CSV シードバンドルは
// フェーズ2の runSeedBundles が別途扱う — このファイル群には seed データは一切含まれない。
func runSQLMigrations(db *sql.DB, logger *slog.Logger) error {
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

// seedBundlesForEnv is the migrate-side selection of seed bundles for an
// APP_ENV value. Thin wrapper over seedbundle.BundleOrderForEnv so tests and
// runSeedBundles / verifyMigrationKeyCoverage share one plan.
func seedBundlesForEnv(env string) []string {
	return seedbundle.BundleOrderForEnv(env)
}

// seedBundlesForCurrentEnv reads APP_ENV (same name as .env.example /
// local-db-reset-contract). Unset or unknown values fail closed to master only.
func seedBundlesForCurrentEnv() []string {
	return seedBundlesForEnv(os.Getenv("APP_ENV"))
}

// runSeedBundles はフェーズ2: CSV シードバンドルを APP_ENV ゲート付き順でロードする。
// local development/test allowlist（development/local/dev/test）では
// 002_master → 003_demo → 004_staging。production / staging / empty / unknown は
// 002_master のみ（SEC-CS2-F01: staging へ privileged demo 資格情報を投入しない）。
// フェーズ1が全て commit した後にのみ実行される。各バンドルは applyCSVBundle が
// 単一トランザクションで完走した後にのみ schema_migrations へ
// seedbundle.BundleMigrationKey で記録されるため、CSVロードが失敗した場合は何も
// 記録されず次回実行でバンドル全体がリトライされる。
// 既知の限界: applyCSVBundle が commit した直後・このレコード用トランザクションの
// commit 前にプロセスが落ちた場合のみ、CSVは投入済みだが schema_migrations は
// 未記録という不整合が起こり得る（次回実行がCOPYを再試行しUNIQUE制約違反になる）。
// 単一オペレータの起動時マイグレーションツールでこの極小ウィンドウに2相コミットを
// 導入するのは過剰設計と判断し、既知の限界として明記するに留める（001_init.sql の
// フェーズと同じトレードオフ）。
func runSeedBundles(ctx context.Context, db *sql.DB, connStr string, logger *slog.Logger) error {
	applied := 0
	skipped := 0
	appEnv := os.Getenv("APP_ENV")
	bundles := seedBundlesForEnv(appEnv)

	logger.Info("Seed bundle plan",
		slog.String("APP_ENV", appEnv),
		slog.Any("bundles", bundles),
		slog.Int("total", len(bundles)))

	for _, bundleDir := range bundles {
		key := seedbundle.BundleMigrationKey(bundleDir)

		checksum, err := bundleChecksum(migrationsDir, bundleDir)
		if err != nil {
			return fmt.Errorf("failed to compute checksum for seed bundle %s: %w", bundleDir, err)
		}

		alreadyApplied, err := isAlreadyApplied(db, key, checksum)
		if err != nil {
			return err
		}
		if alreadyApplied {
			logger.Info("⏭ Skipping seed bundle (already applied)", slog.String("bundle", bundleDir))
			skipped++
			continue
		}

		logger.Info("Loading seed bundle", slog.String("bundle", bundleDir))
		if err := applyCSVBundle(ctx, connStr, migrationsDir, bundleDir, logger); err != nil {
			return fmt.Errorf("failed to load CSV seed bundle %s: %w", bundleDir, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction to record seed bundle %s: %w", bundleDir, err)
		}
		if err := recordMigration(tx, key, checksum); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record seed bundle %s: %w", bundleDir, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit seed bundle record %s: %w", bundleDir, err)
		}

		logger.Info("✓ Seed bundle loaded", slog.String("bundle", bundleDir))
		applied++
	}

	logger.Info("Seed bundle summary",
		slog.Int("applied", applied),
		slog.Int("skipped", skipped),
		slog.Int("total", len(bundles)))

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
