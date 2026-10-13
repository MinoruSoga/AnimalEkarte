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
	"sort"
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

	// フェーズ3: LoginForm と同じ合成デモログインを upsert する。CSV ではない。
	// 共通パスワードはコード定数。production と unknown APP_ENV はスキップ。
	if err := runLoginSeed(context.Background(), db, logger); err != nil {
		return fmt.Errorf("login seed failed: %w", err)
	}

	// 全フェーズ完了後: ディスク上の期待キーと schema_migrations を突合する。
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
// It returns only bundles that still exist in the current seed layout. Pure,
// no DB access — this is what the translation unit tests exercise directly.
//
// The pre-2026-07 binary recorded legacy stub keys together. Only 002_master
// remains a current bundle; retired 003_demo and 004_staging must not be
// checksummed or recorded because their source directories no longer exist.
func legacyTranslationTargets() []string {
	legacyBundleDirs := [...]string{"002_master"}
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
				"following docs/ops/deploy/seeds/LOCAL_DB_RESET.md",
		)
	}
	if migrationCount > 0 && !hasApplicationSchema {
		return errors.New(
			"migration history exists while application schema is missing; " +
				"schema completeness cannot be verified; rebuild with USER-approved DB_RESET " +
				"following docs/ops/deploy/seeds/LOCAL_DB_RESET.md",
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

	// Applied migration drift has no automatic repair path. A checksum alone
	// cannot prove that every schema delta was applied, so fail before any DDL
	// or schema_migrations update. Rebuilds remain an explicit operator action.
	if storedChecksum != checksum {
		fmt.Fprintf(os.Stderr, "checksum_mismatch_seen file=%s applied=%s current=%s\n", filename, storedChecksum, checksum)
		return true, fmt.Errorf(
			"checksum mismatch for %s: applied=%s, current=%s — automatic checksum repair is unsupported; rebuild with DB_RESET=true under an approved reset plan (local procedure: docs/ops/deploy/seeds/LOCAL_DB_RESET.md)",
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

	// Codex #1/#6 (PR #606): destructive DDL の前に fail-closed のデータ検査を行う
	// migration 固有 preflight。committed migration ファイルは不変のため、
	// データ損失防止のガードは runner 側に置く。ガードが失敗した場合は migration
	// 全体を中断し（return error）、030 の DROP COLUMN には到達しない。
	//
	// Codex 再指摘（PR #621 review #1）: ガードを migration とは別接続
	// （非 tx）で評価すると「guard 通過 → 直前の区分 deactivation / 新規
	// 未来日カルテ commit → 030 実行」の割り込みでサイレントなデータ損失が残る。
	// READ COMMITTED で同一 tx に移すだけでは guard 評価と backfill が別
	// スナップショットを読むため不十分 —— ガード対象 migration は専用 tx を
	// 先に開き、判定が依存するテーブルへの書込を凍結するテーブルロックを取得してから
	// guard・適用・記録を同一 tx で実行する。
	migrationGuardedApplies := map[string]migrationGuardSpec{
		"030_next_visit_recommended_date_to_tentative.sql": {
			// guard/backfill が読む全テーブルを列挙する —— どれかが guard 通過と
			// 030 実行の間に変わると判定が崩れる。
			//   reservation_types: 区分の deactivation / category 変更 / 削除
			//   medical_records:    新規・更新された未来日 next_visit_recommended_date
			//   pets / owners:      復活（deleted_at クリア）や status 変更で
			//                       join 対象セットが変わる
			//   appointments:       tentative の追加・削除で NOT EXISTS の評価が変わる
			lockTables: "reservation_types, medical_records, pets, owners, appointments",
			guard:      preflightNextVisitIntentBackfill,
		},
	}

	if len(migrationFiles) == 0 {
		logger.Warn("No migration files found")
		return nil
	}
	sort.Strings(migrationFiles)
	logger.Info("DDL migration files discovered", slog.Any("files", migrationFiles))

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

		// ガード対象 migration: テーブルロック取得 → guard → 適用 → 記録を
		// 同一 tx で不可分にする（lock は commit まで保持される）。
		if spec, ok := migrationGuardedApplies[filename]; ok {
			logger.Info("Executing migration", slog.String("file", filename))
			if err := applyGuardedSQLMigration(db, filename, content, checksum, spec); err != nil {
				return err
			}
			logger.Info("✓ Migration completed", slog.String("file", filename))
			applied++
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

// migrationGuardSpec はガード付き migration の適用規約。
// lockTables は guard 評価〜migration 適用〜記録の全期間で WRITE を凍結する
// テーブル集合（SHARE ROW EXCLUSIVE: SELECT は通し、INSERT/UPDATE/DELETE/TRUNCATE
// が持つ ROW EXCLUSIVE と衝突してブロックする）。guard は同一 tx 内で評価するため、
// 評価したスナップショットが migration の書込みまで保全される。
type migrationGuardSpec struct {
	lockTables string
	guard      func(tx *sql.Tx) error
}

// migrationGuardedPostLockHook はテスト用の観測点 —— テーブルロック取得直後
// （guard 評価前）に呼ばれる。本番では nil のまま。並行接続からの書込が
// ブロックされることを検証する barrier として使う。
var migrationGuardedPostLockHook func(tx *sql.Tx)

// applyGuardedSQLMigration はガード付き migration を不可分に適用する。
// 先に tx を開き SHARE ROW EXCLUSIVE で対象テーブルへの書込を凍結してから
// guard・migration 本体・schema_migrations 記録をこの tx で実行する。
// 凍結により「guard 通過後・適用前に別接続の commit が割り込む」race を排除する
// （Codex PR #621 review #1: guard と適用を同じロック境界で守る）。
// 中断時は tx ごとロールバックされるため、列もデータも変化しない。
func applyGuardedSQLMigration(db *sql.DB, filename string, content []byte, checksum string, spec migrationGuardSpec) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
	}
	// LOCK TABLE 名はコード固定の静的文字列（ユーザー入力ではない）。
	if _, err := tx.Exec("LOCK TABLE " + spec.lockTables + " IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to acquire guard locks for %s: %w", filename, err)
	}
	if migrationGuardedPostLockHook != nil {
		migrationGuardedPostLockHook(tx)
	}
	if err := spec.guard(tx); err != nil {
		_ = tx.Rollback()
		return err
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
	return nil
}

// preflightNextVisitIntentBackfill は 030_next_visit_recommended_date_to_tentative.sql
// 適用前の fail-closed ガード（Codex #1/#6, PR #606）。committed migration は不変の
// ため runner 側で検査する。呼び出し側（applyGuardedSQLMigration）が同一 tx 内で
// guard/backfill が読むテーブルへの書込を凍結済み —— ここで読む集合は 030 実行
// まで変化しない。
//
// 030 は medical_records.next_visit_recommended_date の未来日を tentative
// appointments へ backfill してから列を DROP するが、placeholder reservation_type
// を LATERAL で解決できない行（clinic に非削除の区分が0件）は WARNING のみで
// スキップし、DROP でその未来日は失われる（#1: データ損失）。
//
// さらに placeholder 解決は `(category='general') DESC, is_active DESC` 順のため、
// general が無い clinic では任意の非 general 区分・inactive 区分が割り当てられ得る
// （#6: placeholder 誤昇格）。
//
// 本ガードは「変換対象 intent を持つ clinic が active かつ非削除の general 区分を
// 持つこと」を全件要求する。active general が1件でも存在すれば 030 の ORDER BY が
// 必ずそれを選ぶため placeholder は必ず general になり、type 未解決によるスキップ
// も起きない。違反があれば migrate 全体を中断し、operator が区分を追加してから
// 再実行する。クエリの SELECT 部は 030 の latest/candidates 契約を踏襲する。
func preflightNextVisitIntentBackfill(tx *sql.Tx) error {
	const query = `
		WITH latest AS (
			SELECT DISTINCT ON (mr.clinic_id, mr.pet_id)
				mr.clinic_id,
				mr.pet_id,
				mr.next_visit_recommended_date AS intended_date
			FROM medical_records AS mr
			WHERE mr.deleted_at IS NULL
			  AND mr.pet_id IS NOT NULL
			ORDER BY mr.clinic_id, mr.pet_id, mr.id DESC
		)
		SELECT COUNT(*)
		FROM latest AS l
		JOIN pets AS p
		  ON p.id = l.pet_id
		 AND p.clinic_id = l.clinic_id
		 AND p.deleted_at IS NULL
		 AND p.status <> 'deceased'
		 AND p.deceased_at IS NULL
		JOIN owners AS o
		  ON o.id = p.owner_id
		 AND o.clinic_id = l.clinic_id
		 AND o.deleted_at IS NULL
		WHERE l.intended_date IS NOT NULL
		  AND l.intended_date >= (now() AT TIME ZONE 'Asia/Tokyo')::date
		  AND NOT EXISTS (
			SELECT 1
			FROM appointments AS a
			WHERE a.clinic_id = l.clinic_id
			  AND a.pet_id = l.pet_id
			  AND a.status = 'tentative'
			  AND a.deleted_at IS NULL
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM reservation_types AS rt
			WHERE rt.clinic_id = l.clinic_id
			  AND rt.deleted_at IS NULL
			  AND rt.category = 'general'
			  AND rt.is_active
		  )
	`
	var unconvertible int
	if err := tx.QueryRow(query).Scan(&unconvertible); err != nil {
		return fmt.Errorf("preflight for 030_next_visit_recommended_date_to_tentative.sql failed: %w", err)
	}
	if unconvertible > 0 {
		return fmt.Errorf(
			"migration 030_next_visit_recommended_date_to_tentative.sql aborted (Codex #1/#6): %d pet(s) keep a future next_visit_recommended_date whose clinic has no active general reservation type — the column drop would lose those dates. Add an active general reservation type to each affected clinic, then re-run migrate",
			unconvertible,
		)
	}
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
// 全環境で 002_master のみ（医院骨格 + 参照マスタ。003_demo / 004_staging は退役。
// SEC-CS2-F01: CSV に accounts を載せない。合成デモログインはフェーズ3）。
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
