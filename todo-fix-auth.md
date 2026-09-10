# 認証・認可レビューの未完了TODO

> 作成日: 2026-09-08  
> 最終整理: 2026-09-11  
> 統合コミット: `a0e569a9a`（`main`）  
> 正規設計書: [docs/architecture/auth.md](docs/architecture/auth.md)  
> 元レビュー: [todo-check-auth.md](todo-check-auth.md)

本書は **いま残っている外部境界だけ** を未完了として扱う。ローカル実装・disposable 実DB検証は完了済み。証跡の詳細は Git 履歴を参照する。

実装基準の起点は `a4292a241`。2026-09-10 に disposable Postgres（`ae-auth-fix-disposable-pg`）で D2 / D3 / D1 合成 / D5 を実行し、結果を `a0e569a9a` に統合した。

## 対応状況（要約）

| 区分 | 状態 |
| --- | --- |
| ローカル実装 + disposable 実DB（D2 / D3 / D5 / D1 合成） | **完了**（`a0e569a9a`） |
| D1 本番付与・対象環境メール | **未完了**（対象環境未定・別承認） |
| Linear ライブ特定・投稿 | **BLOCKED**（MCP作成試行が free issue limit exceeded で拒否） |

## 未完了サマリー

| 優先 | ID | 状態 | 残作業 | 完了条件 |
| --- | --- | --- | --- | --- |
| P1 | D1 | PARTIAL | 合成DBは PASS。対象環境は未定で、本番付与・対象環境メールは未実施 | 対象環境・承認済み本番手順と非機密 receipt、メール経路確認 |
| P2 | LINEAR | BLOCKED | 承認済み新規起票を Linear MCP で試行したが、workspace の free issue limit exceeded により HTTP 400 で拒否 | workspace 管理者が新規 issue 作成枠を用意し、同じ下書きで再試行 |

## 完了済み（`a0e569a9a`）

| ID | 状態 | 証跡 |
| --- | --- | --- |
| D3 | PASS（disposable 実DB） | GET/HEAD 178 経路に `realdb-return-data:`。coverage gate remaining=0。package 実DB `go test -p 1 -run TestRealDB_` PASS（auth/inventory/staff/clinic/trimming/pet/owner/identitylink/billing/reservation/lstep/medicalrecord） |
| D2 | PASS（disposable 実DB） | 並行3テスト PASS。`-short` SKIP を PASS にしない |
| D5 | PASS（disposable 2プロセス） | `scripts/auth-d5-dualprocess.sh` / `backend/cmd/auth-d5-dualprocess`。旧 session 即時 401/403、未変更医院維持、p95 +2.5%。DB statements/request は UNKNOWN |
| D1 合成 | PASS | migrate 済み `auth_d1_db` で success / admin conflict / email conflict / audit rollback |

## 共通の安全条件

- 共有 `ekarte_db`、`old-db-postgres`、STG、PRODをテストDBに使わない。
- agentは migration を共有環境へ apply しない。今回の disposable `auth_d1_db` への migrate はユーザー明示承認済みだった。
- 資格情報、接続文字列、メールアドレス、cookie、token、患者情報を本書・ログ・Gitへ記録しない。
- Linear投稿、本番付与、対象環境メール、STG/PROD変更は個別の明示承認を得る。

## 実行環境（2026-09-10）

- disposable: `ae-auth-fix-disposable-pg`（label `com.animalekarte.disposable=true`）、host publish `127.0.0.1:25432`、DB `auth_fix_db` / `auth_fix_db_test` / migrate 用 `auth_d1_db`
- 共有 compose `animalekarte-db-1` / `ekarte_db` は実DB検証に未使用
- 実DB package テストは `TEST_DATABASE_URL` → `auth_fix_db_test`、`-p 1`

## D3. GET/HEADの実DB返却データ分離 — PASS

正本は [get_head_permissions.json](backend/cmd/api/testdata/get_head_permissions.json)。D3 対象 178 経路すべてに `realdb-return-data:` を付与。gate: [get_head_permission_realdb_coverage_test.go](backend/cmd/api/get_head_permission_realdb_coverage_test.go)。

| class | 対象 | 現在 |
| --- | ---: | --- |
| clinic-fixed | 158 | 実装済み・disposable 実DB PASS |
| cross-clinic | 20 | 実装済み・disposable 実DB PASS（`GET /api/v1/clinics` 除外） |
| 除外 | 25 | 変更なし |

共有 harness: [clinic_grant_fixture.go](backend/internal/testdb/clinic_grant_fixture.go)。

ExtractClinicID のみだった clinic-fixed handler（inventory / clinic / trimming / lstep）には `RequireSelectedClinicGrant` を追加した。

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

## Linear反映

2026-09-11 に Linear MCP を使用してライブ確認と、明示承認済みの新規起票を 1 件試行した。作成は workspace の free issue limit exceeded により HTTP 400 で拒否され、issue は作成されていない。直接対応する Ticket ID と現在状態は UNKNOWN のままであり、既存 issue の変更・削除、課金または契約変更での解消は承認範囲外である。

承認後に反映する内容:

- `a0e569a9a` まで、D2 / D3 / D5 / D1 合成は disposable で PASS
- D1 は対象環境未定のまま、本番付与・メールは未完了。Linear は free issue limit の解消後に新規起票を再試行する
- static / offline / `-short` / stub の PASS を実DB・本番完了へ変換しない

## 台帳全体ステータス

**LOCAL COMPLETE / EXTERNAL INCOMPLETE**。

- ローカル実装と disposable 実DB検証（D2 / D3 / D5 / D1 合成）は完了し `main` の `a0e569a9a` に入っている。
- 残るのは D1 本番付与・対象環境メールと Linear 投稿のみ。
