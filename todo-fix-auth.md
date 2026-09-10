# 認証・認可レビューの未完了TODO

> 作成日: 2026-09-08  
> 最終整理: 2026-09-10
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)
> 元レビュー: [todo-check-auth.md](todo-check-auth.md)

本書には未完了作業と、完了済みでも外部承認が残る境界だけを記載する。ローカル実装と disposable 実DB検証の詳細証跡は Git 差分と `/tmp/ae-auth-fix-evidence/` を参照する。

実装基準は `a4292a241`。認証・認可のローカル修正は `main` に統合済み。2026-09-10 に disposable Postgres（`ae-auth-fix-disposable-pg`）上で D2 / D3 / D1 合成 / D5 を実行した。

## 未完了サマリー

| 優先 | ID | 状態 | 残作業 | 完了条件 |
| --- | --- | --- | --- | --- |
| P1 | D1 | PARTIAL | 合成DBは PASS。本番付与・対象環境メールは未実施 | 承認済み本番手順と非機密 receipt、メール経路確認 |
| P2 | LINEAR | UNKNOWN | 対応チケットのライブ特定と投稿 | 対象 issue を確認し、承認後に未完了境界のみ投稿 |

## 完了済み（本整理で PASS）

| ID | 状態 | 証跡 |
| --- | --- | --- |
| D3 | PASS（disposable 実DB） | GET/HEAD 178 経路に `realdb-return-data:`。coverage gate remaining=0。package 実DB `go test -p 1 -run TestRealDB_` PASS（auth/inventory/staff/clinic/trimming/pet/owner/identitylink/billing/reservation/lstep/medicalrecord） |
| D2 | PASS（disposable 実DB） | 並行3テスト PASS。`-short` SKIP を PASS にしない |
| D5 | PASS（disposable 2プロセス） | `scripts/auth-d5-dualprocess.sh` / `backend/cmd/auth-d5-dualprocess`。旧 session 即時 401/403、未変更医院維持、p95 +2.5%。DB statements/request は UNKNOWN |
| D1 合成 | PASS | migrate 済み `auth_d1_db` で success / admin conflict / email conflict / audit rollback。証跡 `/tmp/ae-auth-fix-evidence/auth-fix-d1-evidence.md` |

## 共通の安全条件

- 共有 `ekarte_db`、`old-db-postgres`、STG、PRODをテストDBに使わない。
- agentは migration を共有環境へ apply しない。今回の disposable `auth_d1_db` への migrate はユーザー明示承認済み。
- 資格情報、接続文字列、メールアドレス、cookie、token、患者情報を本書・ログ・Gitへ記録しない。
- Linear投稿、本番付与、対象環境メール、STG/PROD変更は個別の明示承認を得る。

## 実行環境（2026-09-10）

- disposable: `ae-auth-fix-disposable-pg`（label `com.animalekarte.disposable=true`）、host publish `127.0.0.1:25432`、DB `auth_fix_db` / `auth_fix_db_test` / migrate 用 `auth_d1_db`
- 共有 compose `animalekarte-db-1` / `ekarte_db` は実DB検証に未使用
- 実DB package テストは `TEST_DATABASE_URL` → `auth_fix_db_test`、`-p 1`（共有プール TRUNCATE 競合回避）

## D3. GET/HEADの実DB返却データ分離 — PASS

正本は [get_head_permissions.json](backend/cmd/api/testdata/get_head_permissions.json)。D3 対象 178 経路すべてに `realdb-return-data:` を付与。gate: [get_head_permission_realdb_coverage_test.go](backend/cmd/api/get_head_permission_realdb_coverage_test.go)。

| class | 対象 | 現在 |
| --- | ---: | --- |
| clinic-fixed | 158 | 実装済み・disposable 実DB PASS |
| cross-clinic | 20 | 実装済み・disposable 実DB PASS（`GET /api/v1/clinics` 除外） |
| 除外 | 25 | 変更なし |

共有 harness: [clinic_grant_fixture.go](backend/internal/testdb/clinic_grant_fixture.go)。

ExtractClinicID のみだった clinic-fixed handler（inventory / clinic / trimming / lstep）には `RequireSelectedClinicGrant` を追加し、handler 直呼びでも grant A → 403 を証明可能にした。

## D2. 自己ロックアウト防止の実DB並行検証 — PASS

- `TestPermissionPolicyDB_ConcurrentGroupDeactivation`
- `TestPermissionPolicyDB_ConcurrentRuleReplacement`
- `TestPermissionPolicyDB_ConcurrentSelfUnassignWaitsForGroupDeactivation`

## D1. 初回管理者 — 合成 PASS / 本番 BLOCKED

入口:

- [first_system_admin.sql](backend/internal/auth/testdata/first_system_admin.sql)
- [FIRST_SYSTEM_ADMIN.md](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)
- `backend/internal/auth/first_system_admin_procedure_test.go`（静的契約 PASS）

合成（`auth_d1_db`）: success / 既存 admin 競合 / email 競合 / audit rollback PASS。

残作業（別承認）:

1. 対象環境・operator・承認記録・既存 staff/主所属の確定
2. 本番付与と通常 login 確認
3. 対象環境メール経路確認

## D5. 独立2プロセスの即時失効と性能測定 — PASS

- harness: `scripts/auth-d5-dualprocess.sh`、`backend/cmd/auth-d5-dualprocess`
- staff 無効化 / clinic 所属解除 / password-epoch: 次リクエストで旧 session 拒否
- 所属解除時の未変更医院アクセス維持
- baseline p95 対 dual p95 +2.5%（調査閾値 +100% 未満）
- DB statements/request: UNKNOWN（pg_stat_statements 未使用）
- 証跡: `/tmp/ae-auth-fix-evidence/auth-fix-d5-evidence.md`

## Linear反映

Ticket ID と現在状態は UNKNOWN。Linear MCP/CLI はこのセッションで利用不可。ローカル下書き: `/tmp/ae-auth-fix-evidence/linear-draft.md`。外部投稿は明示承認後。

反映すべき内容（承認後）:

- ローカル実装は `a4292a241` 以降の D2/D3/D5/D1合成まで disposable で PASS
- D1 本番付与・メール、Linear ライブ更新は未完了
- static / offline / `-short` / stub の PASS を実DB・本番完了へ変換しない

## 台帳全体ステータス

**INCOMPLETE（外部境界のみ）**。コード側の D2/D3/D5 と D1 合成は disposable で PASS。D1 本番・メールと Linear 投稿が残る。
