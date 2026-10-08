# C06 — SQL/クエリ parameterization inventory (インジェクション防御棚卸し)

Campaign: emr-332-security-defense / Unit: U03-INPUT-OUTPUT / Attempt: att-u03-1
Template: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)
Method note: 本ケースは **実行しないことが目的のケース** (injection を実行しない)。static 棚卸しで「ユーザー入力が SQL 文へ構文として到達しない」ことを確認する。

## 判定レコード

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | C06 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w03` — campaign base `de28e9765a29091339a872a85f1875a947e07e5b` (u01-safety-record)。staging 実測なし |
| 3 | 実施方式 (method) | `static` — `rg` による `db.Raw` / `.Exec` / `Sprintf`-into-query / LIKE 経路の全件棚卸し + 各 site の file:line 読取。クエリ実行・DB 接続なし (安全条件: DB アクセス禁止) |
| 4 | 正常条件 | 期待: ユーザー由来の値は常に bind parameter (`?`/`$N`) 経路のみを通り、SQL 文テキストは constant またはコード由来 fragment のみで構成される / 観測: 棚卸しした全 107 site (Raw 49 + Exec 58) で値引数は bind 化されており、SQL 文へのユーザー入力の構文混入は**0件**。動的に組み立てられる断片はすべて (a) package 内 const/静的リスト, (b) `pgx.Identifier{}.Sanitize()` / `quotePostgresIdentifier` (二重引用符 escape) による識別子引用, (c) `information_schema`/`pg_catalog` 由来の識別子 (test/fixture 経路のみ), (d) allowlist で制約された ORDER BY キー |
| 5 | 拒否条件 | 期待: marker 文字列 (例: `SECCHK20261009C06R01` クラスの sentinel) が SQL 構文として解釈されない — bind 値としてのみ渡される / 観測: marker が構文へ混入する経路は存在しない。代表経路の確認: `owner/repository.go:121-125`・`pet/repository.go:226-238`・`identitylink/repository.go:56-92`・`medicalrecord/vaccination_repository.go:101-119`・`owner/ltv_repository_query.go:136`・`lstep/lstep_tag_cache_repository.go:180-182` の ILIKE/LIKE 述語はすべて `textsearch.EscapeLike` (`internal/textsearch/textsearch.go:102-106` — `\`→`\\`, `%`→`\%`, `_`→`\_`) + `ESCAPE '\'` + bind 値の三点組。LIKE パターン自体にユーザー入力を連結する site は 0 件 |
| 6 | 防御判定 | **PASS** (static の範囲) — 全 raw/dynamic SQL site が bound または code-constant 由来。unsafe (ユーザー入力を SQL 文へ構文連結) site は 0 件、したがって FAIL 対象・remediation candidate なし |
| 7 | 副作用呼出回数 | **0** — クエリを一度も実行していない (static review のみ) |
| 8 | 検出判定 | **N/A** — 実行しない検証に対する検出経路は適用外。SQL ログ監視 (実 DB への marker 到達観測) は本安全条件下では行わない |
| 9 | 安全な証跡参照 | 本ファイル末尾の「全 site 棚卸し表」 (file:line 一覧)。代表読取箇所: `internal/lstep/checkup_sync_repository.go:90-114` (Sprintf は `checkupSyncPreviewSQL` const + helper 生成の定数 fragment), `internal/owner/ltv_repository_query.go:380-405` (ORDER BY allowlist), `internal/csvimport/cutover_import_copy.go:105,210-211` (`pgx.Identifier{}.Sanitize()`), `internal/testdb/testdb.go:650-652` (`quotePostgresIdentifier` = `"` を `""` へ escape) |
| 10 | 制約・未検証範囲 | 未検証: (a) GORM の `Where("col = ?", v)` 型 fluent API (数千 site) は bind 前提の API contract であり個別には列挙していない — 棚卸し対象は `Raw`/`Exec`/Sprintf-into-query/LIKE-連結 のリスククラスのみ。(b) `internal/testdb`/`internal/clinicale2e`/`internal/csvimport` は test/migration-support 経路であり production request path ではない (分類欄に明記)。(c) DB 実測による plan-level の確認 (marker が実クエリログへ現れないこと) は DB アクセス禁止のため未実施 — static での根拠に留まる |
| 11 | 修復案 | 該当なし — unsafe site 0 件。新規コードへの guardrail 候補 (スコープ外・候補として記録): `Raw(fmt.Sprintf(...))` を lint で禁止する rule、または `textsearch.EscapeLike` の使用強制 |
| 12 | 再検証結果 | 未実施 — 修復対象が存在しないため。本棚卸し自体が検証である |

## 全 site 棚卸し表 (107 sites — 全数 file:line)

凡例 — **分類**:
- `bound` : SQL 文は constant; 全可変値は `?`/`$N` bind 引数
- `bound+constfrag` : SQL 文にコード内 const/静的 fragment を Sprintf/連結で埋込むが、可変値は全て bind。fragment にユーザー入力は混入しない
- `ident-quoted` : 識別子を `pgx.Identifier{}.Sanitize()` または `quotePostgresIdentifier` (二重引用符の `""` escape) で引用
- `catalog-derived` : 識別子を DB catalog (information_schema/pg_constraint) から取得 — test/fixture 経路のみ
- `static-list` : テーブル/列名がコード内の静的リテラル列挙からのみ来る
- `DDL-const` : スキーマ定数から組み立てた DDL (test/infra 経路のみ)

### Raw (49 sites)

| file:line | 分類 | 備考 |
|-----------|------|------|
| internal/billing/accounting_repository_reports_allocation.go:63,77,93,137,167 | bound | 月次配賦レポート集計 SQL (const) |
| internal/billing/accounting_repository_reports_close_queries.go:17,37,56,82,146,164 | bound | 締めレポート集計 (const) |
| internal/billing/accounting_repository_reports_monthly_scan.go:33,57,79 | bound | 月次スキャン (const) |
| internal/billing/accounting_repository_reports_shared.go:27 | bound | 共有レポート SQL (const) |
| internal/billing/accounting_repository_unpaid.go:259 | bound+constfrag | `Raw(fmt.Sprintf(..., unpaidAmountSQL, ownerHasPeriodBillingSQL))` — 両者 package const (unpaid_amount.go:18, accounting_repository_unpaid.go:98); 可変値は末尾 args で bind |
| internal/billing/billing_item_repository_unbilled.go:29,123,208 | bound | 未請求抽出 (const) |
| internal/billing/synthetic_closing_audit.go:51,61 | bound | synthetic fixture 監査確認 (test-support) |
| internal/billing/synthetic_closing_fixture.go:559 | bound | synthetic fixture (test-support) |
| internal/clinicale2e/fixture.go:549,568 | catalog-derived | information_schema/pg_constraint から識別子列挙 (E2E teardown) |
| internal/identitylink/link_write_auth.go:101,132,152,178 | bound | system admin / active clinic 判定 (const SQL + `?`) |
| internal/lstep/checkup_sync_repository.go:114 | bound+constfrag | `Sprintf(checkupSyncPreviewSQL, where, filterClause)` — where/filter は `buildCheckupSyncWhere`/`buildCheckupSyncPostJoinFilters` (:121-181) が生成する**定数 fragment + `?` プレースホルダ**; 全値 args で bind |
| internal/lstep/lstep_sync_error_counter_repository.go:30 | bound | INSERT..ON CONFLICT (const + `?`) |
| internal/lstep/lstep_tag_cache_repository.go:154,197,215 | bound+constfrag | baseSQL 連結; `LIKE ? ESCAPE '\'` + `EscapeLike` 済み値 |
| internal/medicalrecord/checkup_repository.go:153 | bound | 健診一覧集計 (const) |
| internal/medicalrecord/medical_record_image_deletion_store.go:37,58,97,119 | bound | 画像削除 store (const) |
| internal/medicalrecord/medical_record_image_upload_quota.go:242,255 | bound | quota 確認 (const) |
| internal/medicalrecord/vaccination_repository.go:180 | bound | ワクチン履歴 ROW_NUMBER (const + `?`) |
| internal/owner/ltv_repository.go:119 | bound+constfrag | `Sprintf(ownerLTVSelectSQL, ...fragments)` — fragment は全てコード生成 const; ORDER BY は allowlist (`buildOrderBy` ltv_repository_query.go:380-405) |
| internal/pet/chronic_condition_repository.go:107 | bound | 慢性疾患 (const) |
| internal/reservation/reservation_repository_slots.go:87,103,126,147,175 | bound | 予約枠ロック/競合 (const + `?`) |
| internal/reservation/reservation_type_occupation_repository.go:115,141 | bound | 予約種別×職種 (const) |
| internal/reservation/synthetic_closing_appointment.go:31 | bound | synthetic fixture (test-support) |
| internal/staff/staff_provisioning_repository.go:224 | bound | 権限判定 (const + `?`) |
| internal/staff/synthetic_closing_staff.go:52 | bound | `SELECT to_regclass('shift_entries')` — 完全 const |
| internal/testdb/testdb.go:424,435,639 | bound / DDL-const | test DB 基盤; :435 は schema 定数 CREATE TYPE; :424,:639 は catalog 照会 + `?` |

### Exec (58 sites)

| file:line | 分類 | 備考 |
|-----------|------|------|
| internal/auth/permission_group_policy_lock.go:23 | bound | advisory lock (const + `?`) |
| internal/billing/cash_register_close_repository.go:71 | bound | `pg_advisory_xact_lock(hashtextextended(?,0))` |
| internal/billing/estimate_repository.go:282 | bound | 同上 (採番ロック) |
| internal/billing/synthetic_closing_audit.go:70,79,85 | bound | sentinel audit UPDATE (const + `?`) |
| internal/billing/synthetic_closing_fixture.go:483,540 | static-list | `Sprintf("ALTER TABLE %s ...", table)` — `table` は静的リスト `syntheticClosingAppendOnlyTables` (:287-297) のみ |
| internal/billing/synthetic_closing_fixture.go:492,505 | bound | synthetic teardown UPDATE/DELETE |
| internal/clinic/closing_special_period_repository.go:76,131 | bound | advisory lock |
| internal/clinicale2e/fixture.go:597,606,611 | bound | SAVEPOINT/RELEASE |
| internal/clinicale2e/fixture.go:604 | catalog-derived | `DELETE FROM "`+target.table+`" WHERE `+target.where — target は上記 catalog 由来識別子 + `"` 囲み; clinicID は `?` bind |
| internal/csvimport/cutover_import_copy.go:105,116 | ident-quoted | `LOCK TABLE `+`pgx.Identifier{}.Sanitize()` |
| internal/csvimport/cutover_import_copy.go:243,319,323,342,371,408,678 | ident-quoted / bound | staging テーブル生成・COPY 後処理 — 識別子全て `pgx.Identifier{}.Sanitize()`; 値は `$N` bind |
| internal/csvimport/cutover_import_resume.go:36,131,135,137,202,205,248,251 | bound | advisory lock/set_config (const + `$N`); :137,:205,:251 の set_config('app.bypass_rls',...) は固定リテラル |
| internal/csvimport/cutover_import_validate.go:61,64,67 | bound | 同上 validate 経路 |
| internal/csvimport/failure_rehearsal.go:255,258,275 | bound | rehearsal 制御 + 意図的 FK 違反行 INSERT (`injectionErr` = FK 違反の発生確認であって攻撃ペイロードではない; 値は `$N` bind) |
| internal/csvimport/import.go:108 | static-list | `DELETE FROM "`+table+`"` — `table` は :100-106 の静的リテラル列挙のみ |
| internal/lstep/lstep_delivery_trigger_log_repository.go:102 | bound | advisory lock |
| internal/medicalrecord/medical_record_image_deletion_store.go:119 | bound | DELETE (const + `?`) |
| internal/medicalrecord/medical_record_image_upload_quota.go:232 | bound | advisory lock |
| internal/reservation/reservation_repository_slots.go:87 | bound | 行ロック |
| internal/reservation/synthetic_closing_appointment.go:39 | bound | synthetic DELETE |
| internal/staff/synthetic_closing_staff.go:58 | bound | `DELETE FROM shift_entries WHERE clinic_id = ?` |
| internal/testdb/clinic_grant_fixture.go:188 | bound+static-list | `Sprintf` — `table` は :180-185 の switch で `companies`/`clinics`/`staffs` のみ許可 (それ以外は `t.Fatalf`) |
| internal/testdb/testdb.go:169,219,235 | DDL-const | CREATE TABLE IF NOT EXISTS 等の静的 DDL (test DB schema 構築) |
| internal/testdb/testdb.go:447,454,457 | DDL-const | enum reconcile — `name`/`v`/`create` は `SharedTestSchemaEnumTypes` (001_init.sql 由来のコード定数) のみ |
| internal/testdb/testdb.go:631,635 | bound | advisory lock/unlock (`hashtext(?)`) |
| internal/testdb/testdb.go:646 | ident-quoted | `CREATE DATABASE "+quotePostgresIdentifier(name)` — `"`→`""` escape (:650-652) |

### 補助確認 (Raw/Exec 以外の構文連結経路)

| 経路 | file:line | 判定 |
|------|-----------|------|
| `.Where(fmt.Sprintf("%s = ?", parentFKColumn), ...)` | internal/persistence/junction.go:28 | bound+constfrag — doc コメント「compile-time literal」契約; 全 caller はリテラル (`shift_template_repository.go:148` `"shift_template_id"`, `shift_entry_repository.go:165` `"shift_entry_id"`) |
| `.Where("NOT "+fragment, ...)` | internal/pet/checkup_history.go:64-70 | bound+constfrag — `petCheckupHistoryCorrelation` は const (:56) |
| `.Where(fragment, m, m)` | internal/billing/accounting_repository.go:29-68 | bound+constfrag — `hasPayment`/`hasSplit`/`methodMatch` はローカル const fragment; `method` 値は bind |
| `.Select(fmt.Sprintf(...))` | internal/billing/accounting_repository_unpaid.go:173,184,209,249,305 | bound+constfrag — 埋込みは `unpaidAmountSQL` 等の package const |
| `ORDER BY` 動的 | internal/owner/ltv_repository_query.go:380-405 | allowlist — sort key を switch で 8 値に限定、order は `asc`/`desc` 以外を `desc` へ矯正 |
| `ORDER BY` 動的 | internal/medicalrecord/medical_record_repository.go:224-234 | allowlist — `medicalRecordSortColumns` map 照合 + order `asc` 以外 `DESC` |
| LIKE パターン連結 | internal/trimming/trimming_repository.go:101 | bound — pattern は `"%"+trimmingImageStorageKeyPrefix+"%"` (const prefix)、ユーザー入力なし |
| `estimate_no LIKE ?` | internal/billing/estimate_repository.go:288 | bound — パターン `"EST-%"` リテラル |

## 証跡コマンド (再現用)

```sh
# backend/internal 配下の全 raw query API 呼出しを列挙 (本棚卸しの入力)
grep -rn '\.Raw(\|\.Exec(' internal/ --include='*.go' | grep -v '_test.go'
# Sprintf が raw query に入る site のみ抽出
grep -rn '\.Raw(' internal/ --include='*.go' | grep -v '_test.go' | grep 'Sprintf'
# → internal/billing/accounting_repository_unpaid.go:259 の 1 件のみ
# LIKE 経路
grep -rn 'LIKE' internal/ --include='*.go' | grep -v '_test.go'
# → ユーザー入力を含む全 LIKE site で EscapeLike + ESCAPE '\' を確認 (上表)
```

*Marker 文字列は `SECCHK20261009C06Rnn` 形の英数字 sentinel のみ。実行可能な攻撃構文・実ペイロード・資格情報・PII は本ファイルに含まない (A5)。*
