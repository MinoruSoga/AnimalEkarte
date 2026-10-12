# U01 — Existing Security-Relevant Test Classification (既存テスト分類)

Campaign: emr-332-security-defense / Unit: U01-SAFETY-BASELINE / Attempt: att-u01-1
Basis: `feat/emr-332-security-defense` base `de28e9765a29091339a872a85f1875a947e07e5b` の `backend/` ツリーを import・helper 呼出しまで確認した機械的分類。
Classification: `memory-only` (DB/外部通信/永続化なし) / `DB-dependent` (testdb = Compose `db` ホストの Postgres が必要) / `N/A` (テスト非存在)。

分類規則: `internal/testdb` (`testdb.go:564-615`) は `DB_HOST` (既定 `db` = docker-compose の db サービス名) へ Postgres 接続を張る。`internal/testdb` を import する、または `setupAuditTestDB`/`SetupIsolatedTestDB`/`SetupTestDB` 系 helper を呼ぶテストは DB-dependent。httptest + インタフェース stub/fake (含む `driver.ErrBadConn` 等のエラー値 fake) のみのものは memory-only。

## 1. A3 指定対象 (named scope)

| クラス | 対象ファイル | 分類 | 根拠 |
|--------|--------------|------|------|
| `backend/internal/middleware/*_test.go` | `auth_current_access_test.go`, `auth_test.go`, `cors_test.go`, `csrf_test.go`, `liff_auth_test.go`, `logging_test.go`, `rate_limit_test.go`, `sanitize_null_bytes_test.go`, `security_headers_test.go` (9 ファイル) | **memory-only** | 全ファイル `net/http/httptest` + stub 実装。`auth_test.go`/`auth_current_access_test.go` は `database/sql/driver` の `driver.ErrBadConn` を依存失敗の fake 値として使うのみで実 DB 接続なし。`testdb` import なし |
| `backend/internal/authjwt/` | `claims.go` のみ、`*_test.go` 0 件 | **N/A — no existing tests** | `ls internal/authjwt/` = `claims.go` のみ。テストを捏造しない |
| `backend/internal/audit/` (service 層) | `service_test.go`, `service_operations_test.go`, `service_doubles_test.go`, `repository_test.go` | **memory-only** | `testdb`/`gorm` DB 接続なし。`repository_test.go` の `TestRepositoryCreateTxRequiresAmbientTransaction` は `NewRepository(nil)` の fail-closed contract 確認で DB 不要 |
| `backend/internal/audit/` (repository 層) | `repository_db_test.go`, `repository_real_ddl_test.go`, `repository_test_helpers_test.go` | **DB-dependent → not-runnable-without-db** | `repository_db_test.go:31-34` の `setupAuditTestDB` → `setupAuditRealDDLTestDB` → `testdb.SetupIsolatedTestDB` (real_ddl_test.go:49)。`repository_test_helpers_test.go` は TestMain + `001_init.sql` DDL 抽出 helper (テストファイル自体は DB helper の土台) |
| `backend/internal/imageupload/` | `imageupload.go` のみ、`*_test.go` 0 件 | **N/A — no existing tests** | `ls internal/imageupload/` = `imageupload.go` のみ |

## 2. 実行可能クラスごとの隔離 Docker コマンド

ホストで `go` を直接実行しない規約 (backend/CLAUDE.md) に従い、イメージ `golang:1.26-alpine` (go.mod `go 1.26.0`・docker-compose.yml:131・Dockerfile.production:4 と一致) で worktree を read-only マウントして実行する。`<worktree>` = `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-emr-332-security-defense-w01`。

### memory-only: middleware

```sh
docker run --rm --pull never \
  -v <worktree>/backend:/app:ro \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build \
  -w /app golang:1.26-alpine \
  go test ./internal/middleware/...
```

### memory-only: audit service 層 (同 package に DB テストを含むため -run で選別)

```sh
docker run --rm --pull never \
  -v <worktree>/backend:/app:ro \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build \
  -w /app golang:1.26-alpine \
  go test ./internal/audit/ -run 'TestService|TestAuditService|TestValidateAuditLog|TestAuditActionConstants|TestMarshalJSON|TestRepositoryCreateTxRequiresAmbientTransaction'
```

注: `internal/audit` は package 単位でコンパイルされるが、TestMain (`repository_test_helpers_test.go`) が呼ぶ `testdb.CloseSharedTestDB()` は共有 DB 未オープン時 no-op (testdb.go:127-133) のため、service 系 `-run` 選別は DB 接続なしで完結する。

### DB-dependent (本条件では実行不可 — Compose `db` ホストの Postgres が必要)

`internal/audit` の repository 系 (上表) および `testdb` 利用の全テストは、isolated-Docker では **not-runnable-without-db**。`docker compose` の db サービスを立てない限り実行しない (ticket: 実 DB 接続・TRUNCATE・seed を伴う既存テストは本条件では実行しない)。

### N/A — テスト非存在のためコマンドなし

`internal/authjwt`・`internal/imageupload` は実行対象テストが存在しない。下流ユニットが検証を必要とする場合は新規テスト作成または静的レビューで扱う (本ユニットは作成しない)。

## 3. 観測された補足 (A3 指定外・下流ユニットの再導出防止)

| パッケージ | 規模 | 分類 | 備考 |
|-----------|------|------|------|
| `backend/internal/auth/` | `*_test.go` 55 件 | 混在 (**DB-dependent 14 / memory-only 41**) | `http_*_security_test.go` (http_password_security_test.go, http_session_security_test.go), `refresh_token_family_security_test.go` 等を含む。**DB-dependent 14** = `internal/testdb` を直接 import する 11 ファイル (`account_password_change_concurrency_test.go`, `account_repository_test.go`, `current_access_staff_reader_db_test.go`, `password_reset_concurrency_test.go`, `password_reset_token_repository_test.go`, `permission_group_audit_db_test.go`, `permission_group_effective_permissions_test.go`, `permission_group_repository_test.go`, `permission_group_staff_clinic_isolation_test.go`, `realdb_selected_clinic_b_grant_a_isolation_test.go`, `token_blacklist_repository_test.go`) + testdb を直接 import しないが testdb 由来 helper を呼ぶ推移的 DB-dependent 3 ファイル: `persistence_ambient_tx_test.go` (`setupAccountTestDB` :19 → `account_repository_test.go:30` → `testdb.SetupTestDB`; ほか :61 `setupPasswordResetTokenTestDB`→`password_reset_token_repository_test.go:35`, :122 `setupTokenBlacklistTestDB`→`token_blacklist_repository_test.go:30`, :182,:236,:262,:297 `setupPermissionGroupRepositoryTestDB`→`permission_group_repository_test.go:29`, :211 `setupPermissionGroupStaffIsolationTestDB`→`permission_group_staff_clinic_isolation_test.go:29`), `permission_group_repository_delete_atomic_test.go` (`setupPermissionGroupRepositoryTestDB` :15,:34 → `permission_group_repository_test.go:29-31`), `permission_group_policy_concurrency_db_test.go` (`setupPermissionAuditRollbackDB` :74 → `permission_group_audit_db_test.go:110-112`)。残り **41 ファイルは memory-only** — `permission_group_test_helpers_test.go` は DB テスト用 factory helper (`makePermissionGroupTestClinic` :13 等) で opener を持たず規則上 memory-only 側に計上、`account_password_reset_logging_test.go`・`permission_group_policy_lock_test.go` は `gorm.Open` が `DisableAutomaticPing` (+`DryRun`) 指定で実接続しない |
| `backend/cmd/api/` | `*_test.go` 23 件 | memory-only (静的/httptest) | `testdb` import 0 件。`route_composition_smoke_test.go` (route 構成 smoke), `staging_access_cache_config_test.go` (staging access cache 設定), `get_head_permission_classification_test.go` / `get_head_permission_execution_test.go` / `get_head_permission_realdb_coverage_test.go` (GET/HEAD 210 経路の permission 分類ゲート — 名前に realdb とあるが実体は testdata/get_head_permissions.json の静的カバレッジ検査で DB 接続なし)。EMR-332 の C01/認可系で参照価値が高い |
| `backend/internal/persistence/rls_*_test.go` | `rls_effectiveness_test.go`, `rls_role_privilege_test.go` | DB-dependent | testdb 経由の RLS 実測 (SECURITY.md: RLS dormant の実測根拠)。本条件では実行しない |
| `backend/internal/lstep/` セキュリティ関連 | `line_link_signature_routing_test.go`, `routes_webhook_rate_limit_test.go` 等 | 要個別確認 (webhook 署名・rate limit 系は mock/stub 型が多いがファイル単位の確認は U05/U06 側の責務) | C09/webhook 評価時の参照先として記録 |
| `backend/internal/staff/` セキュリティ関連 | `staff_shift_security_integration_test.go`, `staff_update_security_test.go` | DB-dependent (integration 命名) | testdb 系。本条件では実行しない |
| `backend/internal/billing/` | `accounting_security_gates_test.go` | 要個別確認 | U02/U05 側スコープ |
| `backend/internal/model/` | `rls_migration_test.go`, `identity_link_rls_migration_test.go` | 静的/migration ファイル検査系 | 個別確認は各ケースユニット側 |

## 4. 実行ポリシーまとめ

- 本ユニットではテストを一切実行していない (documentary unit — 分類の記録のみ)。
- 下流ユニットは memory-only クラスを §2 の isolated Docker コマンドで実行できる。`-p 1` は共有 test DB 不使用の memory-only には不要だが、DB-dependent を誤って束ねる場合に備え各クラスを別コマンドで実行すること。
- DB-dependent クラスは Compose db が必要であり本キャンペーンの安全条件 (永続化禁止・DB アクセス禁止) では実行不可 → 対象テストの検証は静的レビューまたは明示的に許可された DB 経路のある別ユニットへ留保する。
