package persistence

// rls_effectiveness_test.go — BE-refactor follow-up (RLS実効化のローカル実効性確認)
//
// ─── 目的 ────────────────────────────────────────────────────────────────────────────
//
// 現行の RLS は 001_init.sql で ENABLE のみ（FORCE なし）で構築され、アプリはテーブル owner ロールで
// 接続し `app.current_clinic_ids` GUC を SET していないため **runtime では dormant**（owner は非 FORCE
// RLS を bypass）。これは migration のコメント（001_init.sql:2895-2905）に明記された意図的 baseline。
//
// 本テストは「RLS ポリシー式 `app_private.has_clinic_access(clinic_id)` 自体は正しく、FORCE + GUC を
// 適切に設定すれば DB レベルで越境 read/write を遮断できる」ことを probe テーブルで**実証**する
// （= 将来 RLS を実効化する際の readiness と、has_clinic_access ポリシーの correctness を固定）。
// 本テストは適用済みスキーマやアプリを一切変更しない（probe テーブル/ロールはテスト内で作成/破棄）。
//
// 実証する経路（すべて実 SQL を実行して検証。名前だけの主張はしない）:
//   (a) USING       — clinic 1/2 セッションが自クリニックの行のみ read できる（差分で証明）
//   (b) WITH CHECK  — 同一 clinic INSERT は成功し（positive control）、越境 clinic INSERT は拒否される
//   (c) bypass_rls  — batch セッションは全 clinic を横断 read できる
//   (d) fail-closed — GUC 未設定セッションは 0 行（非 owner ロールで clinic context 必須を示す）
//   (e) FORCE       — 非 FORCE では table owner が RLS を bypass（＝現行 dormant の根本原因）し、
//                     `FORCE ROW LEVEL SECURITY` 適用後は owner も RLS に従う。本番実効化では
//                     app の DB ロールが clinic テーブルの owner のため、この FORCE 経路が load-bearing。
//
// ⚠️ full 実効化（全 clinic テーブルへの FORCE + 非 owner/owner ロール整理 + 全 tx での SET LOCAL
// app.current_clinic_ids 配線 + batch bypass 経路）は接続層/middleware を跨ぐ高リスクな all-or-nothing
// 改修でフォローアップのスコープ外（BLOCKED）。FORCE を配線なしで適用するとアプリ全クエリが遮断され機能停止する。
// 本テストは「ポリシーと FORCE が期待通り動く」ことだけを probe で固定し、アプリ配線は行わない。

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupAppPrivateRLSFunctions は 001_init.sql:2907-2939 の app_private スキーマ + RLS 関数を
// テスト DB に再現する（テスト DB は AutoMigrate ベースで migration の RLS ブロックを持たないため）。
// 3 関数（current_clinic_ids / bypass_rls / has_clinic_access）の本体は migration と文字一致。
// 権限付与は「同等（PostgreSQL のデフォルト権限に依存）」であり byte-for-byte 一致ではない:
// migration は明示 `GRANT EXECUTE ... TO PUBLIC` + `REVOKE ALL ON FUNCTION apply_rls_policy FROM PUBLIC`
// を行うが、本セットアップは `CREATE FUNCTION` の既定（EXECUTE を PUBLIC へ自動付与）に依存する。
// fresh な postgres:18-alpine テスト DB では既定が有効なため結果は等価。
// これらの定義が migration に存在することは model.TestRLSMigrationCoversTenantTables が固定する。
func setupAppPrivateRLSFunctions(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE SCHEMA IF NOT EXISTS app_private`,
		`CREATE OR REPLACE FUNCTION app_private.current_clinic_ids() RETURNS bigint[] LANGUAGE sql STABLE AS $$
			SELECT COALESCE(
				string_to_array(NULLIF(current_setting('app.current_clinic_ids', true), ''), ',')::bigint[],
				ARRAY[]::bigint[]
			);
		$$`,
		`CREATE OR REPLACE FUNCTION app_private.bypass_rls() RETURNS boolean LANGUAGE sql STABLE AS $$
			SELECT COALESCE(NULLIF(current_setting('app.bypass_rls', true), '')::boolean, false);
		$$`,
		`CREATE OR REPLACE FUNCTION app_private.has_clinic_access(row_clinic_id bigint) RETURNS boolean LANGUAGE sql STABLE AS $$
			SELECT app_private.bypass_rls() OR row_clinic_id = ANY(app_private.current_clinic_ids());
		$$`,
		`GRANT USAGE ON SCHEMA app_private TO PUBLIC`,
	}
	for _, s := range stmts {
		require.NoError(t, db.Exec(s).Error, "setup app_private RLS function")
	}
}

// TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC は has_clinic_access ポリシーが
// SET LOCAL app.current_clinic_ids の下で (a) 越境 SELECT を隠し(USING)、(b) 同一 clinic INSERT を許し
// 越境 INSERT を拒否し(WITH CHECK)、(c) bypass_rls で batch 全件 read を許し、(d) GUC 未設定で fail-closed に
// なり、(e) FORCE ROW LEVEL SECURITY が table owner にも RLS を適用することを、実 SQL で実証する。
//
// SET LOCAL を tx 内で使うのは production で必要な「接続毎に clinic context を流す」パターンと同一
// （gorm.Transaction は単一コネクションを使うため SET LOCAL が効く）。
// テスト DB のログインユーザーは superuser で、superuser は FORCE RLS すら bypass するため、RLS を実際に
// 効かせるには **非 superuser ロール**が要る。非 owner ロール（USING/WITH CHECK/bypass/fail-closed 検証用）と
// 非 superuser owner ロール（FORCE 検証用）を作り、各 tx で SET LOCAL ROLE する。これは production 実効化で
// 必要な「非 superuser ロール + 接続毎の clinic context」パターンそのもの。
//
// M-1（並行実行安全性）: role/table 名はテスト毎に PID+nanotime で unique 化する。PostgreSQL の
// ロールは cluster スコープ（DB より広い blast radius）なので、固定名だと同一 dev DB への並行 go test で
// 「片方の tx が SET LOCAL ROLE 中に、もう片方が DROP ROLE/再作成」の衝突が起きうる。既存の
// lstep_csv_import_service_integration_test.go（DB 名を nanotime で unique 化）と同じ回避策。
func TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()

	// テスト DB は setupTestDB が AutoMigrate で作るため 001_init.sql の RLS ブロック（app_private
	// スキーマ + 関数）が入っていない。実ポリシー式を検証するため migration と同一の関数群を再現する。
	setupAppPrivateRLSFunctions(t, db)

	exec := func(sql string) {
		t.Helper()
		require.NoError(t, db.WithContext(ctx).Exec(sql).Error, sql)
	}

	// M-1: cluster スコープのロール/テーブル名をテスト毎に unique 化（token は数値のみで injection 安全）。
	token := fmt.Sprintf("%d_%d", os.Getpid(), time.Now().UnixNano())
	probeTable := fmt.Sprintf("rls_probe_%s", token)
	probeSeq := probeTable + "_id_seq"
	probeUser := fmt.Sprintf("rls_probe_user_%s", token)   // 非 owner・非 superuser（USING/WITH CHECK 等）
	probeOwner := fmt.Sprintf("rls_probe_owner_%s", token) // 非 superuser の table owner（FORCE 検証）

	// probe テーブル（clinic_id 直持ち）に 001_init.sql と同じポリシー式を適用する。
	exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, probeTable))
	exec(fmt.Sprintf(`CREATE TABLE %s (id bigserial PRIMARY KEY, clinic_id bigint NOT NULL, note text NOT NULL DEFAULT '')`, probeTable))
	exec(fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL SECURITY`, probeTable))
	exec(fmt.Sprintf(`CREATE POLICY tenant_isolation ON %s FOR ALL `+
		`USING (app_private.has_clinic_access(clinic_id)) `+
		`WITH CHECK (app_private.has_clinic_access(clinic_id))`, probeTable))

	// seed は superuser（RLS bypass）で clinic 1 / clinic 2 の行を投入する。
	exec(fmt.Sprintf(`INSERT INTO %s (clinic_id, note) VALUES (1, 'A'), (2, 'B')`, probeTable))

	// 非 owner ロール（RLS が常に効く subject）。
	exec(fmt.Sprintf(`DROP ROLE IF EXISTS %s`, probeUser))
	exec(fmt.Sprintf(`CREATE ROLE %s NOLOGIN`, probeUser))
	exec(fmt.Sprintf(`GRANT SELECT, INSERT ON %s TO %s`, probeTable, probeUser))
	exec(fmt.Sprintf(`GRANT USAGE, SELECT ON SEQUENCE %s TO %s`, probeSeq, probeUser))

	// 非 superuser owner ロール（FORCE 検証用）。ALTER TABLE ... OWNER TO は PG15+ で新 owner が
	// schema への CREATE 権限を要するため付与する。所有権移管後も既存 GRANT は object に紐づき残る。
	exec(fmt.Sprintf(`DROP ROLE IF EXISTS %s`, probeOwner))
	exec(fmt.Sprintf(`CREATE ROLE %s NOLOGIN`, probeOwner))
	exec(fmt.Sprintf(`GRANT CREATE ON SCHEMA public TO %s`, probeOwner))
	exec(fmt.Sprintf(`ALTER TABLE %s OWNER TO %s`, probeTable, probeOwner))

	t.Cleanup(func() {
		// best-effort（エラーは無視）。DROP TABLE で object 依存の GRANT が消え、DROP OWNED BY で
		// schema-level GRANT（CREATE ON SCHEMA public）が消えてから DROP ROLE できる。
		db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, probeTable))
		db.Exec(fmt.Sprintf(`DROP OWNED BY %s`, probeUser))
		db.Exec(fmt.Sprintf(`DROP ROLE IF EXISTS %s`, probeUser))
		db.Exec(fmt.Sprintf(`DROP OWNED BY %s`, probeOwner))
		db.Exec(fmt.Sprintf(`DROP ROLE IF EXISTS %s`, probeOwner))
	})

	// asRole は 指定ロール + 指定 clinic context の tx で fn を実行する
	// （SET LOCAL ROLE + SET LOCAL app.* は tx スコープ・同一コネクションで効く）。
	asRole := func(t *testing.T, role, setup string, fn func(tx *gorm.DB) error) error {
		t.Helper()
		return db.Transaction(func(tx *gorm.DB) error {
			if setup != "" {
				if err := tx.Exec(setup).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec(fmt.Sprintf(`SET LOCAL ROLE %s`, role)).Error; err != nil {
				return err
			}
			return fn(tx)
		})
	}
	selectDistinctClinics := fmt.Sprintf(`SELECT DISTINCT clinic_id FROM %s ORDER BY clinic_id`, probeTable)

	t.Run("USING: clinic 1 セッションは clinic 1 の行のみ見える", func(t *testing.T) {
		var got []int64
		require.NoError(t, asRole(t, probeUser, `SET LOCAL app.current_clinic_ids = '1'`, func(tx *gorm.DB) error {
			return tx.Raw(selectDistinctClinics).Scan(&got).Error
		}))
		assert.Equal(t, []int64{1}, got, "clinic 1 セッションで clinic 2 の行が見えてはならない")
	})

	t.Run("USING: clinic 2 セッションは clinic 2 の行のみ見える", func(t *testing.T) {
		var got []int64
		require.NoError(t, asRole(t, probeUser, `SET LOCAL app.current_clinic_ids = '2'`, func(tx *gorm.DB) error {
			return tx.Raw(selectDistinctClinics).Scan(&got).Error
		}))
		assert.Equal(t, []int64{2}, got)
	})

	t.Run("WITH CHECK: 同一clinic INSERT は成功・越境clinic INSERT は拒否される", func(t *testing.T) {
		// positive control (H-1): 同一 clinic の INSERT は WITH CHECK を通過して成功する。
		// これが無いと「越境を正しく拒否」と「全 INSERT を拒否（grant 欠落等）」を区別できない。
		require.NoError(t, asRole(t, probeUser, `SET LOCAL app.current_clinic_ids = '1'`, func(tx *gorm.DB) error {
			return tx.Exec(fmt.Sprintf(`INSERT INTO %s (clinic_id, note) VALUES (1, 'same-clinic')`, probeTable)).Error
		}), "同一 clinic の INSERT は WITH CHECK を通過して成功しなければならない")

		// negative: clinic 1 セッションからの clinic 2 行 INSERT は WITH CHECK で拒否される。
		err := asRole(t, probeUser, `SET LOCAL app.current_clinic_ids = '1'`, func(tx *gorm.DB) error {
			return tx.Exec(fmt.Sprintf(`INSERT INTO %s (clinic_id, note) VALUES (2, 'cross-clinic')`, probeTable)).Error
		})
		require.Error(t, err, "clinic 1 セッションからの clinic 2 行 INSERT は WITH CHECK で拒否されるべき")
	})

	t.Run("bypass_rls: batch セッションは全 clinic の行を見える", func(t *testing.T) {
		var got []int64
		require.NoError(t, asRole(t, probeUser, `SET LOCAL app.bypass_rls = 'on'`, func(tx *gorm.DB) error {
			return tx.Raw(selectDistinctClinics).Scan(&got).Error
		}))
		assert.Equal(t, []int64{1, 2}, got, "bypass_rls=on は全 clinic を横断アクセスできる（batch/system 経路）")
	})

	t.Run("GUC 未設定セッションは何も見えない（fail-closed の DB 版）", func(t *testing.T) {
		var count int64
		require.NoError(t, asRole(t, probeUser, "", func(tx *gorm.DB) error {
			// current_clinic_ids 未設定 + bypass なし → has_clinic_access は常に false。
			return tx.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM %s`, probeTable)).Scan(&count).Error
		}))
		assert.Equal(t, int64(0), count, "clinic context 未設定では非 owner ロールで全行が不可視（app 配線が必須の理由）")
	})

	t.Run("FORCE: 非FORCEでは owner が RLS を bypass、FORCE 適用後は owner も従う", func(t *testing.T) {
		// before FORCE: table owner は clinic context を無視して全行が見える（ENABLE のみで owner は
		// RLS を bypass）。これが「現在アプリが owner 接続で RLS dormant」の根本原因そのもの。
		var beforeClinics []int64
		require.NoError(t, asRole(t, probeOwner, `SET LOCAL app.current_clinic_ids = '1'`, func(tx *gorm.DB) error {
			return tx.Raw(selectDistinctClinics).Scan(&beforeClinics).Error
		}))
		assert.Equal(t, []int64{1, 2}, beforeClinics,
			"非 FORCE では table owner は clinic context を無視して全行が見える（＝現在 RLS が dormant な理由）")

		// FORCE を有効化する。
		exec(fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL SECURITY`, probeTable))

		// after FORCE: table owner も RLS に従い clinic 1 のみになる。本番実効化で app の DB ロールが
		// clinic テーブルの owner のため、この FORCE 経路が実際に必要（= FORCE が load-bearing）。
		var afterClinics []int64
		require.NoError(t, asRole(t, probeOwner, `SET LOCAL app.current_clinic_ids = '1'`, func(tx *gorm.DB) error {
			return tx.Raw(selectDistinctClinics).Scan(&afterClinics).Error
		}))
		assert.Equal(t, []int64{1}, afterClinics,
			"FORCE ROW LEVEL SECURITY 適用後は table owner も RLS に従う（本番実効化で app ロールが owner のため必須）")
	})
}
