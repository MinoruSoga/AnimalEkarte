package persistence

// rls_role_privilege_test.go — BE-refactor Phase A (#219): ローカルDB接続ロールの
// rolsuper / rolbypassrls 実測プリフライト。
//
// ─── 目的 ────────────────────────────────────────────────────────────────────────────
//
// rls_effectiveness_test.go の (e) FORCE subtest は「非 superuser の owner ロールなら
// FORCE ROW LEVEL SECURITY で RLS が効く」ことを probe テーブルで実証した。しかし
// superuser 接続は FORCE RLS すら bypass する（PostgreSQL 仕様）ため、実際にアプリが
// 接続する DB ロール自体が superuser/BYPASSRLS を持つ場合、将来 FORCE を全 clinic
// テーブルへ適用しても実効化されない。旧 be-refactor-followup-status.md（削除済み・git 履歴参照）は「アプリは
// POSTGRES_USER=テーブル owner（かつ superuser）で接続」と記載しているが、これは
// 001_init.sql のコメント根拠の質的記述であり、実測（pg_roles への実クエリ）ではなかった。
// 本テストはその実測を固定する（RLS full 実効化 architect レビューの preflight）。
//
// ⚠️ 本テストは特定のロール属性値を pass/fail 条件にしない。目的は「実測して t.Logf で
// 可視化する」ことであり、rolsuper=true は現状の意図的 baseline（001_init.sql:2895-2905）
// であって bug ではない。将来ロール構成を変更した際に本テストの出力が変わることで
// 変更に気づけるようにする（harness improvement P0: DB role privilege preflight）。
//
// 対象: ローカル/エフェメラルテスト DB のみ（setupTestDB）。STG/prod のロール構成は
// 未確認・未接続（別スコープ）。

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/testdb"
)

// dbRolePrivilegeRow は pg_roles から取得する接続ロールの権限属性。
type dbRolePrivilegeRow struct {
	RolName      string `gorm:"column:rol_name"`
	RolSuper     bool   `gorm:"column:rol_super"`
	RolBypassRLS bool   `gorm:"column:rol_bypass_rls"`
}

func TestDBConnectionRolePrivileges_LocalMeasurement(t *testing.T) {
	db := testdb.SetupTestDB(t)

	var row dbRolePrivilegeRow
	require.NoError(t, db.Raw(
		`SELECT rolname AS rol_name, rolsuper AS rol_super, rolbypassrls AS rol_bypass_rls `+
			`FROM pg_roles WHERE rolname = current_user`,
	).Scan(&row).Error, "pg_roles から接続ロールの権限属性を取得できること")

	require.NotEmpty(t, row.RolName, "current_user が pg_roles に存在すること")

	t.Logf("[#219 role privilege measurement] current_user=%s rolsuper=%v rolbypassrls=%v",
		row.RolName, row.RolSuper, row.RolBypassRLS)

	if row.RolSuper || row.RolBypassRLS {
		t.Logf("[#219] 接続ロールが superuser または BYPASSRLS のため、FORCE ROW LEVEL SECURITY を" +
			"適用しても RLS は常に bypass される（現状 dormant の実測確認）。" +
			"full 実効化には非 superuser かつ非 BYPASSRLS のアプリロールへの移行が前提条件" +
			"（旧 be-refactor-followup-status.md（git 履歴）記載の質的結論を実測で裏付け）。")
	} else {
		t.Logf("[#219] 接続ロールは非 superuser かつ非 BYPASSRLS。FORCE ROW LEVEL SECURITY 配線の" +
			"前提が満たされている可能性がある（他の前提条件は別途確認要）。")
	}
}
