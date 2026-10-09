package main

// Codex #1/#6/#7 (PR #606 remediation) の回帰テスト。
//
// 030_next_visit_recommended_date_to_tentative.sql は future-dated
// next_visit_recommended_date を tentative appointments へ backfill してから列を
// DROP する。committed migration は不変のため、データ損失ガードは runner 側の
// preflightNextVisitIntentBackfill が担う。本テストは捨て DB を作成し、001–029 を
// 適用した状態で runSQLMigrations に 030 以降を走らせることで、preflight による
// 中断 (#1)・general 区分の必須化 (#6)・backfill 結果・031 複合 FK (#7) を
// 実マイグレーション経路で検証する。
//
// testdb (ekarte_db_test) は AutoMigrate 由来の GORM double であり実
// マイグレーションの複合 FK / DROP COLUMN を持たないため、このテストは
// コンテナ内 Postgres 上に CREATE DATABASE する（compose の ekarte_user は
// superuser）。/app/migrations はコンテナマウント前提のため、ホスト実行や
// DB 環境変数なしでは skip する。

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/dbconn"
)

const migration030Filename = "030_next_visit_recommended_date_to_tentative.sql"

// requireMigrateScratchEnv は本テストの前提（コンテナ内 DB 接続・/app/migrations
// マウント）を満たさない場合に skip する。ホスト側 `go test ./cmd/migrate` で
// DB が無い環境でも既存の env-gated テスト群と同じく no-op にする。
func requireMigrateScratchEnv(t *testing.T) dbconn.ConnParams {
	t.Helper()
	if testing.Short() {
		t.Skip("scratch database tests are skipped in -short / offline verification")
	}
	conn, err := dbconn.FromEnv()
	if err != nil {
		t.Skipf("DB env vars not configured; skipping migration integration test: %v", err)
	}
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		t.Skipf("migrations directory %q not mounted (host run); skipping", migrationsDir)
	}
	return conn
}

// newScratchMigrateDatabase は一時データベースを作成して接続を返す。
// テスト終了時に DROP DATABASE ... WITH (FORCE) で破棄する。
func newScratchMigrateDatabase(t *testing.T, conn dbconn.ConnParams, label string) *sql.DB {
	t.Helper()
	// 識別子はテスト生成の安全文字のみ（[a-z0-9_]）。
	name := fmt.Sprintf("preflight_%s_%d", label, time.Now().UnixNano())

	admin, err := sql.Open("postgres", conn.DSN("postgres"))
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		_ = admin.Close()
		t.Fatalf("create scratch database %s: %v", name, err)
	}
	_ = admin.Close()

	db, err := sql.Open("postgres", conn.DSN(name))
	if err != nil {
		t.Fatalf("open scratch database %s: %v", name, err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		cleanup, cerr := sql.Open("postgres", conn.DSN("postgres"))
		if cerr != nil {
			t.Logf("open admin connection for cleanup: %v", cerr)
			return
		}
		defer cleanup.Close()
		if _, derr := cleanup.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)"); derr != nil {
			t.Logf("drop scratch database %s: %v", name, derr)
		}
	})
	if err := db.Ping(); err != nil {
		t.Fatalf("ping scratch database %s: %v", name, err)
	}
	return db
}

// applyMigrationsBefore は stopBefore に指定したファイル名より前の top-level DDL
// migration を昇順で適用し、schema_migrations に記録する。これにより
// runSQLMigrations が stopBefore 以降の未適用ファイルから再開する状態を再現する。
func applyMigrationsBefore(t *testing.T, db *sql.DB, stopBefore string) {
	t.Helper()
	if err := ensureMigrationsTable(db); err != nil {
		t.Fatalf("ensureMigrationsTable: %v", err)
	}
	dir := inventoryMigrationsDir(t)
	for _, name := range listInventorySQLFiles(t) {
		if name >= stopBefore {
			break
		}
		content, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // テスト用の固定ディレクトリ
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx for %s: %v", name, err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			t.Fatalf("apply migration %s: %v", name, err)
		}
		if err := recordMigration(tx, name, fileChecksum(content)); err != nil {
			_ = tx.Rollback()
			t.Fatalf("record migration %s: %v", name, err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit migration %s: %v", name, err)
		}
	}
}

// seedPreflightOrg は company/clinic/animal_species の最小骨格を投入する。
// 明示 id は各 scratch DB でユニークな連番として扱う。
func seedPreflightOrg(t *testing.T, db *sql.DB) {
	t.Helper()
	stmts := []string{
		`INSERT INTO companies (id, name) VALUES (900, 'preflight co')`,
		`INSERT INTO clinics (id, company_id, name) VALUES
			(900, 900, 'preflight clinic A'),
			(901, 900, 'preflight clinic B'),
			(902, 900, 'preflight clinic C'),
			(903, 900, 'preflight clinic D')`,
		`INSERT INTO animal_species (id, name) VALUES (900, 'preflight species')`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed org: %v\nstmt: %s", err, stmt)
		}
	}
}

// seedFutureRecommendation は owner/pet/medical_record の最小行を投入し、
// その pet の最新カルテに未来日の next_visit_recommended_date を付ける。
func seedFutureRecommendation(t *testing.T, db *sql.DB, clinicID, seq int64, deceased bool) {
	t.Helper()
	status := "alive"
	if deceased {
		status = "deceased"
	}
	stmts := []string{
		fmt.Sprintf(`INSERT INTO owners (id, clinic_id, name) VALUES (%d, %d, 'owner %d')`, 9000+seq, clinicID, seq),
		fmt.Sprintf(`INSERT INTO pets (id, clinic_id, owner_id, name, animal_species_id, status)
			VALUES (%d, %d, %d, 'pet %d', 900, '%s')`, 9100+seq, clinicID, 9000+seq, seq, status),
		fmt.Sprintf(`INSERT INTO medical_records (id, clinic_id, record_no, date, pet_id, next_visit_recommended_date)
			VALUES (%d, %d, 'PFL-%d', CURRENT_DATE, %d, CURRENT_DATE + 30)`, 9200+seq, clinicID, seq, 9100+seq),
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed recommendation seq=%d: %v\nstmt: %s", err, seq, stmt)
		}
	}
}

func seedReservationType(t *testing.T, db *sql.DB, id, clinicID int64, name, category string, isActive bool) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO reservation_types (id, clinic_id, name, category, is_active) VALUES ($1, $2, $3, $4, $5)`,
		id, clinicID, name, category, isActive,
	); err != nil {
		t.Fatalf("seed reservation_type %d: %v", id, err)
	}
}

// TestMigration030PreflightAbortsWithoutActiveGeneralType は Codex #1/#6:
// active な general 区分を持たない医院に変換対象の未来日が残っている場合、
// runSQLMigrations が 030 適用前に中断し、列とデータが保全されることを固定する。
// 3 医院パターン（general なし / general inactive / 区分なし）を同一 DB で検証する。
func TestMigration030PreflightAbortsWithoutActiveGeneralType(t *testing.T) {
	conn := requireMigrateScratchEnv(t)
	db := newScratchMigrateDatabase(t, conn, "abort")
	applyMigrationsBefore(t, db, migration030Filename)

	seedPreflightOrg(t, db)
	// clinic 900: active だが non-general (trimming) 区分のみ → #6: 非 general への
	// placeholder 昇格を起こさないためガード対象。
	seedReservationType(t, db, 9300, 900, "trimming only", "trimming", true)
	seedFutureRecommendation(t, db, 900, 1, false)
	// clinic 901: general 区分はあるが inactive → #6/#15: inactive は昇格不可。
	seedReservationType(t, db, 9301, 901, "inactive general", "general", false)
	seedFutureRecommendation(t, db, 901, 2, false)
	// clinic 902: 区分がゼロ → #1: 030 の LATERAL が型を解決できず行がスキップされ
	// DROP で未来日が失われる経路。
	seedFutureRecommendation(t, db, 902, 3, false)
	// clinic 903: active tentative 既存 → NOT EXISTS で対象外（カウント 3 固定の根拠）。
	// appointments.reservation_type_id は NOT NULL のため参照用の区分を 1 件用意する。
	seedReservationType(t, db, 9313, 903, "trimming D", "trimming", true)
	seedFutureRecommendation(t, db, 903, 4, false)
	if _, err := db.Exec(`
		INSERT INTO appointments (clinic_id, start_time, end_time, reservation_type_id, pet_id, owner_id, status, intended_date, reservation_route)
		VALUES (903, now(), now() + interval '15 minutes', 9313, 9104, 9004, 'tentative', CURRENT_DATE + 30, 'record_intent')`); err != nil {
		t.Fatalf("seed existing tentative: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := runSQLMigrations(db, logger)
	if err == nil {
		t.Fatal("runSQLMigrations must abort when unconvertible future dates remain")
	}
	if !strings.Contains(err.Error(), "030_next_visit_recommended_date_to_tentative.sql aborted") {
		t.Fatalf("error must identify the aborted migration, got: %v", err)
	}
	if !strings.Contains(err.Error(), "3 pet(s)") {
		t.Fatalf("guard must count exactly the 3 unconvertible pets (existing tentative excluded), got: %v", err)
	}

	// #1: 中断時に列もデータも残っていること。
	var colExists bool
	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'medical_records' AND column_name = 'next_visit_recommended_date'
		)`).Scan(&colExists); err != nil {
		t.Fatal(err)
	}
	if !colExists {
		t.Fatal("preflight abort must leave medical_records.next_visit_recommended_date intact")
	}
	var tentativeCount int
	if err := db.QueryRow(`SELECT count(*) FROM appointments WHERE status = 'tentative'`).Scan(&tentativeCount); err != nil {
		t.Fatal(err)
	}
	if tentativeCount != 1 {
		t.Fatalf("abort must not write new tentative rows, got %d (want 1: only the pre-existing)", tentativeCount)
	}
	var appliedAfterGuard int
	if err := db.QueryRow(`
		SELECT count(*) FROM schema_migrations
		WHERE filename >= '030_'`).Scan(&appliedAfterGuard); err != nil {
		t.Fatal(err)
	}
	if appliedAfterGuard != 0 {
		t.Fatalf("no migration at or after 030 may be recorded on abort, got %d", appliedAfterGuard)
	}
}

// TestMigration030BackfillAndCompositeFK は正常系:
// active general 区分を持つ医院では 030 が適用され、未来日が general 型の
// tentative へ変換され、031 の複合 FK が医院越境 provenance を拒否することを固定する。
func TestMigration030BackfillAndCompositeFK(t *testing.T) {
	conn := requireMigrateScratchEnv(t)
	db := newScratchMigrateDatabase(t, conn, "apply")
	applyMigrationsBefore(t, db, migration030Filename)

	seedPreflightOrg(t, db)
	// clinic 900: general + trimming の両方が active → #6: general が選ばれること。
	seedReservationType(t, db, 9310, 900, "general type", "general", true)
	seedReservationType(t, db, 9311, 900, "trimming type", "trimming", true)
	// clinic 901: general のみ。
	seedReservationType(t, db, 9312, 901, "general type", "general", true)

	seedFutureRecommendation(t, db, 900, 10, false) // → backfill 対象
	seedFutureRecommendation(t, db, 901, 14, false) // → backfill 対象

	// pet 911: 未来日 + 既存 tentative → NOT EXISTS で二重化しない。
	seedFutureRecommendation(t, db, 900, 11, false)
	if _, err := db.Exec(`
		INSERT INTO appointments (clinic_id, start_time, end_time, reservation_type_id, pet_id, owner_id, status, intended_date, reservation_route)
		VALUES (900, now(), now() + interval '15 minutes', 9310, 9111, 9011, 'tentative', CURRENT_DATE + 30, 'record_intent')`); err != nil {
		t.Fatalf("seed existing tentative: %v", err)
	}
	// pet 912: 過去日は変換しない。
	{
		if _, err := db.Exec(`
			INSERT INTO owners (id, clinic_id, name) VALUES (9012, 900, 'owner 12')`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
			INSERT INTO pets (id, clinic_id, owner_id, name, animal_species_id) VALUES (9112, 900, 9012, 'pet 12', 900)`); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`
			INSERT INTO medical_records (id, clinic_id, record_no, date, pet_id, next_visit_recommended_date)
			VALUES (9212, 900, 'PFL-12', CURRENT_DATE, 9112, CURRENT_DATE - 10)`); err != nil {
			t.Fatal(err)
		}
	}
	// pet 913: deceased は対象外。
	seedFutureRecommendation(t, db, 900, 13, true)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runSQLMigrations(db, logger); err != nil {
		t.Fatalf("runSQLMigrations on convertible fixture: %v", err)
	}

	// 030/031 が記録され、列が DROP されている。
	var colExists bool
	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'medical_records' AND column_name = 'next_visit_recommended_date'
		)`).Scan(&colExists); err != nil {
		t.Fatal(err)
	}
	if colExists {
		t.Fatal("migration 030 must drop next_visit_recommended_date after backfill")
	}
	for _, want := range []string{migration030Filename, "031_appointments_source_medical_record_clinic_fk.sql"} {
		var ok bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)`, want).Scan(&ok); err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("schema_migrations must record %s", want)
		}
	}

	// backfill 内容: pet 910/914 のみ新規 tentative。general 型・route・provenance を固定。
	type backfilledRow struct {
		petID, typeID, mrID int64
		route               string
		intendedDate        string
	}
	rows, err := db.Query(`
		SELECT pet_id, reservation_type_id, source_medical_record_id, reservation_route, intended_date::text
		FROM appointments
		WHERE status = 'tentative' AND source_medical_record_id IS NOT NULL
		ORDER BY pet_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []backfilledRow
	for rows.Next() {
		var r backfilledRow
		if err := rows.Scan(&r.petID, &r.typeID, &r.mrID, &r.route, &r.intendedDate); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	// 期待日は Go 側ではなく DB セッションの CURRENT_DATE を使う
	// （JST 深夜跨ぎでの flake を防ぐ）。
	var wantDate string
	if err := db.QueryRow(`SELECT (CURRENT_DATE + 30)::text`).Scan(&wantDate); err != nil {
		t.Fatal(err)
	}
	want := []backfilledRow{
		{petID: 9110, typeID: 9310, mrID: 9210, route: "record_intent", intendedDate: wantDate},
		{petID: 9114, typeID: 9312, mrID: 9214, route: "record_intent", intendedDate: wantDate},
	}
	if len(got) != len(want) {
		t.Fatalf("backfilled tentative rows = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("backfilled row %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// 既存 tentative (pet 911)・過去日 (pet 912)・deceased (pet 913) に新規行を作らない。
	var total int
	if err := db.QueryRow(`SELECT count(*) FROM appointments WHERE status = 'tentative'`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Fatalf("tentative count = %d, want 3 (pet 910 backfill + pet 911 pre-existing + pet 914 backfill)", total)
	}

	// #7: 031 複合 FK の存在と旧単一列 FK の削除。
	var composite bool
	if err := db.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_appointments_source_medical_record_clinic')`).Scan(&composite); err != nil {
		t.Fatal(err)
	}
	if !composite {
		t.Fatal("fk_appointments_source_medical_record_clinic must exist after 031")
	}
	var singleCol bool
	if err := db.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'appointments_source_medical_record_id_fkey')`).Scan(&singleCol); err != nil {
		t.Fatal(err)
	}
	if singleCol {
		t.Fatal("single-column appointments_source_medical_record_id_fkey must be dropped by 031")
	}

	// #7 機能検証: 他医院の medical_record を provenance に持つ行は拒否される。
	_, err = db.Exec(`
		INSERT INTO appointments (clinic_id, start_time, end_time, reservation_type_id, pet_id, owner_id, source_medical_record_id)
		VALUES (901, now(), now() + interval '15 minutes', 9312, 9114, 9014, 9210)`)
	if err == nil {
		t.Fatal("cross-clinic source_medical_record_id must violate fk_appointments_source_medical_record_clinic")
	}

	// #7 機能検証: medical_record 削除で source_medical_record_id のみ NULL になる。
	if _, err := db.Exec(`DELETE FROM medical_records WHERE id = 9210`); err != nil {
		t.Fatalf("delete provenance record: %v", err)
	}
	var cleared bool
	if err := db.QueryRow(`
		SELECT source_medical_record_id IS NULL FROM appointments
		WHERE clinic_id = 900 AND pet_id = 9110 AND status = 'tentative'`).Scan(&cleared); err != nil {
		t.Fatal(err)
	}
	if !cleared {
		t.Fatal("ON DELETE SET NULL must clear only source_medical_record_id, keeping the tentative row")
	}
}
