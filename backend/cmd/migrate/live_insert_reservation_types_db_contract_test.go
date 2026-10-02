package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EMR-220: live_insert_standard_reservation_types.sql の実行契約を使い捨て
// スキーマで固定する DB 依存 contract test。静的な文字列 pin
// （live_insert_reservation_types_contract_test.go）だけでは検出できない
// 「同名の inactive live 行が残ったまま成功扱いになる」旧不具合を、
// 実際にスクリプト全文を実行して防ぐ。
//
// 使い捨て lane: openMigrateTestPool と同じ DB（local db container のみ、
// dbconn.IsLocalHost が非 local を拒否）の中でテスト専用 schema を CREATE し、
// search_path をその schema に向けてからスクリプトを verbatim 実行する。
// public schema の実テーブルは一切触らず、終了時に schema を DROP CASCADE する。
// 本番/STG 共有 DB への適用は contract 外であり、このテストも行わない。

const liveInsertReservationTypesSQL = "../../migrations/seeds/live_insert_standard_reservation_types.sql"

// runStandardReservationTypesSeed は live SQL ファイルを読み、simple protocol
// （複数 statement + BEGIN/COMMIT を含む全文）で実行する。RAISE EXCEPTION で
// 全体 rollback された場合は error として返る。
func runStandardReservationTypesSeed(ctx context.Context, conn *pgxpool.Conn) error {
	contents, err := os.ReadFile(liveInsertReservationTypesSQL)
	if err != nil {
		return fmt.Errorf("read live insert SQL: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, string(contents)).ReadAll()
	if err != nil {
		// 失敗時は batch 内の COMMIT がスキップされ、conn に aborted tx が残る
		// （runbook の psql ON_ERROR_STOP=1 でセッションが切れて rollback される
		// のと同じ終端）。明示 ROLLBACK で閉じ、後続の検証クエリを可能にする。
		// 部分 commit は構造上起きないが、起きた場合は行数アサートが検出する。
		_, _ = conn.Exec(ctx, "ROLLBACK")
	}
	return err
}

// setupReservationTypesContractSchema はテスト専用 schema を作り、001_init.sql
// のうちスクリプトが依存する最小構成（companies/clinics/reservation_types と
// live 行限定の一意 index）を再現する。group_id/parent_id の FK 対象は
// スクリプト非依存のため省略し、列形状は実 DDL と揃える。
//
// 返却する conn はプールから acquire 済みのため、呼び出し側は必ず
// `defer teardownReservationTypesContractSchema(t, conn, schema)` を
// `defer pool.Close()` より後に登録すること（defer は LIFO なので
// teardown が先に走り、conn を release してから pool.Close が wait する。
// t.Cleanup で release すると pool.Close の defer が先に走って deadlock する）。
func setupReservationTypesContractSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (*pgxpool.Conn, string) {
	t.Helper()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}

	schema := fmt.Sprintf("emr220_contract_%d", time.Now().UnixNano())

	if _, err := conn.Exec(ctx, "CREATE SCHEMA "+quoteIdent(schema)); err != nil {
		conn.Release()
		t.Fatalf("create contract schema: %v", err)
	}
	if _, err := conn.Exec(ctx, "SET search_path TO "+quoteIdent(schema)); err != nil {
		conn.Release()
		t.Fatalf("set search_path: %v", err)
	}

	for _, ddl := range []string{
		`CREATE TABLE companies (
			id   BIGSERIAL PRIMARY KEY,
			name text NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE clinics (
			id         BIGSERIAL PRIMARY KEY,
			company_id bigint NOT NULL REFERENCES companies(id),
			name       text NOT NULL DEFAULT ''
		)`,
		`CREATE TYPE reservation_type_category AS ENUM ('general', 'trimming')`,
		`CREATE TABLE reservation_types (
			id                       BIGSERIAL   PRIMARY KEY,
			clinic_id                bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
			name                     text        NOT NULL,
			is_active                boolean     NOT NULL DEFAULT true,
			description              text        NOT NULL DEFAULT '',
			color                    text        NOT NULL DEFAULT '#3B82F6',
			sort_order               integer              DEFAULT 0,
			group_id                 bigint,
			reservation_display_name text        NOT NULL DEFAULT '',
			duration_minutes         int         NOT NULL DEFAULT 15,
			short_name               text        NOT NULL DEFAULT '',
			show_short_name          boolean     NOT NULL DEFAULT false,
			reservation_visible      boolean     NOT NULL DEFAULT true,
			reservation_comment      text        NOT NULL DEFAULT '',
			reservation_image_url    text        NOT NULL DEFAULT '',
			parent_id                bigint      REFERENCES reservation_types(id) ON DELETE RESTRICT,
			max_concurrent           integer              CHECK (max_concurrent > 0),
			reservation_day_option   text         NOT NULL DEFAULT 'none',
			is_internal              boolean      NOT NULL DEFAULT false,
			category                 reservation_type_category NOT NULL DEFAULT 'general',
			created_at               timestamptz  NOT NULL DEFAULT now(),
			updated_at               timestamptz  NOT NULL DEFAULT now(),
			deleted_at               timestamptz
		)`,
		`CREATE UNIQUE INDEX idx_reservation_types_clinic_name
			ON reservation_types(clinic_id, name) WHERE deleted_at IS NULL`,
	} {
		if _, err := conn.Exec(ctx, ddl); err != nil {
			_, _ = conn.Exec(ctx, "SET search_path TO public")
			_, _ = conn.Exec(ctx, "DROP SCHEMA IF EXISTS "+quoteIdent(schema)+" CASCADE")
			conn.Release()
			t.Fatalf("contract schema DDL failed: %v\n%s", err, ddl)
		}
	}
	return conn, schema
}

// teardownReservationTypesContractSchema は search_path を public へ戻してから
// 契約 schema を DROP CASCADE し、acquire した conn を release する。
// deferred pool.Close() が checkout 済み conn を待たないよう、必ず defer で
// pool.Close より先に呼ばれる順序にする。
func teardownReservationTypesContractSchema(t *testing.T, conn *pgxpool.Conn, schema string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// 失敗系ケースでは conn に aborted tx が残っている可能性があるため、
	// 先に ROLLBACK で閉じる（tx 外では警告のみで no-op）。
	_, _ = conn.Exec(ctx, "ROLLBACK")
	if _, err := conn.Exec(ctx, "SET search_path TO public"); err != nil {
		t.Logf("reset search_path: %v", err)
	}
	if _, err := conn.Exec(ctx, "DROP SCHEMA IF EXISTS "+quoteIdent(schema)+" CASCADE"); err != nil {
		t.Logf("drop contract schema %s: %v", schema, err)
	}
	conn.Release()
}

func insertContractClinic(t *testing.T, ctx context.Context, conn *pgxpool.Conn, name string) int64 {
	t.Helper()
	var companyID, clinicID int64
	if err := conn.QueryRow(ctx, `INSERT INTO companies (name) VALUES ($1) RETURNING id`, name+"-company").Scan(&companyID); err != nil {
		t.Fatalf("insert company: %v", err)
	}
	if err := conn.QueryRow(ctx, `INSERT INTO clinics (company_id, name) VALUES ($1, $2) RETURNING id`, companyID, name).Scan(&clinicID); err != nil {
		t.Fatalf("insert clinic: %v", err)
	}
	return clinicID
}

// insertContractReservationType は既存行の事前投入ヘルパー。isActive=false で
// 同名 inactive live 行（旧不具合の再現条件）を作れる。
func insertContractReservationType(t *testing.T, ctx context.Context, conn *pgxpool.Conn, clinicID int64, name string, isActive bool, description, color string, durationMinutes int) {
	t.Helper()
	_, err := conn.Exec(ctx, `
INSERT INTO reservation_types
  (clinic_id, name, is_active, description, color, duration_minutes, category)
VALUES ($1, $2, $3, $4, $5, $6, 'general')`,
		clinicID, name, isActive, description, color, durationMinutes)
	if err != nil {
		t.Fatalf("insert existing reservation_type: %v", err)
	}
}

func countContractReservationTypes(t *testing.T, ctx context.Context, conn *pgxpool.Conn) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM reservation_types WHERE deleted_at IS NULL`).Scan(&n); err != nil {
		t.Fatalf("count reservation_types: %v", err)
	}
	return n
}

// activeStandardTypeCount は (clinic × 標準4区分) のうち live かつ
// is_active=true の行を持つ組合せ数を返す。スクリプトの事後条件と同じ定義。
func activeStandardTypeCount(t *testing.T, ctx context.Context, conn *pgxpool.Conn) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(ctx, `
SELECT count(*)
FROM clinics c
CROSS JOIN (VALUES ('診察'), ('お手入れ'), ('ワクチン'), ('健診')) AS d(name)
WHERE EXISTS (
  SELECT 1 FROM reservation_types e
  WHERE e.clinic_id = c.id AND e.name = d.name
    AND e.deleted_at IS NULL AND e.is_active
)`).Scan(&n); err != nil {
		t.Fatalf("count active standard types: %v", err)
	}
	return n
}

// TestLiveInsertStandardReservationTypesAllActiveSucceeds は全医院への投入が
// 成功し、各医院に 4 区分の is_active=true live 行が揃うことを確認する。
func TestLiveInsertStandardReservationTypesAllActiveSucceeds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	clinicA := insertContractClinic(t, ctx, conn, "contract clinic A")
	clinicB := insertContractClinic(t, ctx, conn, "contract clinic B")

	if err := runStandardReservationTypesSeed(ctx, conn); err != nil {
		t.Fatalf("seed run failed: %v", err)
	}
	if got := activeStandardTypeCount(t, ctx, conn); got != 8 {
		t.Fatalf("active standard (clinic,type) pairs = %d, want 8", got)
	}
	for _, clinicID := range []int64{clinicA, clinicB} {
		var names []string
		rows, err := conn.Query(ctx, `
SELECT name FROM reservation_types
WHERE clinic_id = $1 AND deleted_at IS NULL AND is_active
ORDER BY sort_order`, clinicID)
		if err != nil {
			t.Fatalf("list types for clinic %d: %v", clinicID, err)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scan: %v", err)
			}
			names = append(names, name)
		}
		rows.Close()
		want := []string{"診察", "お手入れ", "ワクチン", "健診"}
		if !slices.Equal(names, want) {
			t.Fatalf("clinic %d active types = %v, want %v", clinicID, names, want)
		}
	}
	// 投入行の属性も契約どおりか確認する（category='general', reservation_visible=false）。
	var category string
	var visible, internal bool
	if err := conn.QueryRow(ctx, `
SELECT category, reservation_visible, is_internal
FROM reservation_types WHERE clinic_id = $1 AND name = '診察' AND deleted_at IS NULL`, clinicA).
		Scan(&category, &visible, &internal); err != nil {
		t.Fatalf("read inserted attrs: %v", err)
	}
	if category != "general" || visible || internal {
		t.Fatalf("inserted 診察 attrs = (category=%s, visible=%v, internal=%v), want (general,false,false)", category, visible, internal)
	}
}

// TestLiveInsertStandardReservationTypesInactiveSameNameFailsAndRollsBack は
// 同名の inactive live 行がある医院では run が失敗し、他の INSERT を含む
// transaction 全体が rollback されることを確認する（旧不具合: inactive 行を
// 残したまま成功扱いになっていた）。既存行が再活性化されないことも併せて見る。
func TestLiveInsertStandardReservationTypesInactiveSameNameFailsAndRollsBack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	clinicID := insertContractClinic(t, ctx, conn, "contract clinic inactive")
	insertContractReservationType(t, ctx, conn, clinicID, "診察", false, "既存の無効区分", "#999999", 30)

	err := runStandardReservationTypesSeed(ctx, conn)
	if err == nil {
		t.Fatal("seed run must fail when a same-name inactive live row exists, got nil")
	}
	if !strings.Contains(err.Error(), "postcondition mismatch") {
		t.Fatalf("seed error = %v, want postcondition mismatch", err)
	}

	// 全 INSERT が rollback され、事前の inactive 行だけが残る。
	if got := countContractReservationTypes(t, ctx, conn); got != 1 {
		t.Fatalf("reservation_types rows after rollback = %d, want 1 (pre-existing inactive row only)", got)
	}
	var isActive bool
	var description, color string
	var duration int
	if err := conn.QueryRow(ctx, `
SELECT is_active, description, color, duration_minutes
FROM reservation_types WHERE clinic_id = $1 AND name = '診察' AND deleted_at IS NULL`, clinicID).
		Scan(&isActive, &description, &color, &duration); err != nil {
		t.Fatalf("read surviving row: %v", err)
	}
	if isActive {
		t.Fatal("pre-existing inactive row was re-activated; contract requires it stays inactive")
	}
	if description != "既存の無効区分" || color != "#999999" || duration != 30 {
		t.Fatalf("pre-existing row attributes changed: (desc=%s, color=%s, duration=%d)", description, color, duration)
	}
}

// TestLiveInsertStandardReservationTypesPreservesExistingAttributes は既存の
// live かつ active な同名行の属性が一切上書きされず、残り 3 区分だけが追加
// されることを確認する。
func TestLiveInsertStandardReservationTypesPreservesExistingAttributes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	clinicID := insertContractClinic(t, ctx, conn, "contract clinic existing")
	insertContractReservationType(t, ctx, conn, clinicID, "診察", true, "既存の診察区分", "#ABCDEF", 45)

	if err := runStandardReservationTypesSeed(ctx, conn); err != nil {
		t.Fatalf("seed run failed: %v", err)
	}

	var isActive, visible bool
	var description, color, category string
	var duration int
	if err := conn.QueryRow(ctx, `
SELECT is_active, description, color, duration_minutes, reservation_visible, category
FROM reservation_types WHERE clinic_id = $1 AND name = '診察' AND deleted_at IS NULL`, clinicID).
		Scan(&isActive, &description, &color, &duration, &visible, &category); err != nil {
		t.Fatalf("read existing row: %v", err)
	}
	if !isActive || description != "既存の診察区分" || color != "#ABCDEF" || duration != 45 || visible != true || category != "general" {
		t.Fatalf("existing row overwritten: (active=%v desc=%s color=%s duration=%d visible=%v category=%s)",
			isActive, description, color, duration, visible, category)
	}
	// visible は既存行の既定 true のまま（seed 既定 false で上書きされていない）。
	if got := countContractReservationTypes(t, ctx, conn); got != 4 {
		t.Fatalf("reservation_types rows = %d, want 4 (1 existing + 3 inserted)", got)
	}
	if got := activeStandardTypeCount(t, ctx, conn); got != 4 {
		t.Fatalf("active standard pairs = %d, want 4", got)
	}
}

// TestLiveInsertStandardReservationTypesRerunIsNoOp は 2 回目の実行が行を
// 追加しない（既存 live 行 skip の冪等性）ことを確認する。
func TestLiveInsertStandardReservationTypesRerunIsNoOp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	clinicID := insertContractClinic(t, ctx, conn, "contract clinic rerun")

	if err := runStandardReservationTypesSeed(ctx, conn); err != nil {
		t.Fatalf("first seed run failed: %v", err)
	}
	if got := countContractReservationTypes(t, ctx, conn); got != 4 {
		t.Fatalf("rows after first run = %d, want 4", got)
	}
	var firstMaxID int64
	if err := conn.QueryRow(ctx, `SELECT max(id) FROM reservation_types WHERE clinic_id = $1`, clinicID).Scan(&firstMaxID); err != nil {
		t.Fatalf("max id: %v", err)
	}

	if err := runStandardReservationTypesSeed(ctx, conn); err != nil {
		t.Fatalf("second seed run failed: %v", err)
	}
	if got := countContractReservationTypes(t, ctx, conn); got != 4 {
		t.Fatalf("rows after second run = %d, want 4 (re-run must be a no-op)", got)
	}
	var secondMaxID int64
	if err := conn.QueryRow(ctx, `SELECT max(id) FROM reservation_types WHERE clinic_id = $1`, clinicID).Scan(&secondMaxID); err != nil {
		t.Fatalf("max id after rerun: %v", err)
	}
	if secondMaxID != firstMaxID {
		t.Fatalf("re-run inserted rows: max(id) %d -> %d", firstMaxID, secondMaxID)
	}
}

// TestLiveInsertStandardReservationTypesZeroClinicsRejected は医院 0 件の
// 環境への適用が拒否されることを確認する。
func TestLiveInsertStandardReservationTypesZeroClinicsRejected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	err := runStandardReservationTypesSeed(ctx, conn)
	if err == nil {
		t.Fatal("seed run must fail when no clinics exist, got nil")
	}
	if !strings.Contains(err.Error(), "no clinics found") {
		t.Fatalf("seed error = %v, want 'no clinics found'", err)
	}
	if got := countContractReservationTypes(t, ctx, conn); got != 0 {
		t.Fatalf("reservation_types rows = %d, want 0", got)
	}
}

// TestLiveInsertStandardReservationTypesSingleClinicDeficitDetected は複数医院
// のうち 1 医院だけが is_active=true の 4 区分を欠く場合でも検出され、
// 全医院分の INSERT を含む transaction 全体が rollback されることを確認する。
func TestLiveInsertStandardReservationTypesSingleClinicDeficitDetected(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool := openMigrateTestPool(t, ctx)
	defer pool.Close()
	conn, contractSchema := setupReservationTypesContractSchema(t, ctx, pool)
	defer teardownReservationTypesContractSchema(t, conn, contractSchema)

	clinicA := insertContractClinic(t, ctx, conn, "contract clinic healthy")
	clinicB := insertContractClinic(t, ctx, conn, "contract clinic deficit")
	insertContractReservationType(t, ctx, conn, clinicB, "ワクチン", false, "別医院の無効区分", "#111111", 20)

	err := runStandardReservationTypesSeed(ctx, conn)
	if err == nil {
		t.Fatal("seed run must fail when any single clinic lacks an active standard type, got nil")
	}
	if !strings.Contains(err.Error(), "postcondition mismatch") {
		t.Fatalf("seed error = %v, want postcondition mismatch", err)
	}

	// 健全側の医院への INSERT も含めて全 rollback。残るのは deficit 側の既存行のみ。
	if got := countContractReservationTypes(t, ctx, conn); got != 1 {
		t.Fatalf("reservation_types rows after rollback = %d, want 1", got)
	}
	var clinicARows int
	if err := conn.QueryRow(ctx, `SELECT count(*) FROM reservation_types WHERE clinic_id = $1`, clinicA).Scan(&clinicARows); err != nil {
		t.Fatalf("count clinic A rows: %v", err)
	}
	if clinicARows != 0 {
		t.Fatalf("clinic A received %d rows despite failed run; rollback did not cover other clinics' inserts", clinicARows)
	}
	var deficitActive bool
	if err := conn.QueryRow(ctx, `
SELECT is_active FROM reservation_types
WHERE clinic_id = $1 AND name = 'ワクチン' AND deleted_at IS NULL`, clinicB).Scan(&deficitActive); err != nil {
		t.Fatalf("read deficit row: %v", err)
	}
	if deficitActive {
		t.Fatal("deficit clinic's inactive row was re-activated; contract requires it stays inactive")
	}
}
