# C06 — SQL/クエリ parameterization inventory (インジェクション防御棚卸し)

Campaign: emr-332-security-defense / Unit: U03-INPUT-OUTPUT / Attempt: att-u03-2 (att-u03-1 証跡訂正 — verdict 維持)
Template: `docs/work/emr-332-security/baseline/case-verdict-template.md` (12 fields)
Method note: 本ケースは **実行しないことが目的のケース** (injection を実行しない)。static 棚卸しで「ユーザー入力が SQL 文へ構文として到達しない」ことを確認する。

## 判定レコード

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | C06 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w03` — campaign base `de28e9765a29091339a872a85f1875a947e07e5b` (baseline/safety-record)。staging 実測なし |
| 3 | 実施方式 (method) | `static` — `rg` による `db.Raw` / `.Exec` / `Sprintf`-into-query / LIKE 経路の全件棚卸し + 各 site の file:line 読取。クエリ実行・DB 接続なし (安全条件: DB アクセス禁止) |
| 4 | 正常条件 | 期待: ユーザー由来の値は常に bind parameter (`?`/`$N`) 経路のみを通り、SQL 文テキストは constant またはコード由来 fragment のみで構成される / 観測: 棚卸しした全 **119 site (Raw 58 + Exec 61)** で値引数は bind 化されており、SQL 文へのユーザー入力の構文混入は**0件**。動的に組み立てられる断片はすべて (a) package 内 const/静的リスト, (b) `pgx.Identifier{}.Sanitize()` / `quotePostgresIdentifier` (二重引用符 escape) による識別子引用, (c) `information_schema`/`pg_catalog` 由来の識別子 (test/fixture 経路のみ), (d) allowlist で制約された ORDER BY キー |
| 5 | 拒否条件 | 期待: marker 文字列 (例: `SECCHK20261009C06R01` クラスの sentinel) が SQL 構文として解釈されない — bind 値としてのみ渡される / 観測: marker が構文へ混入する経路は存在しない。代表経路の確認: `owner/repository.go:121-125`・`pet/repository.go:226-238`・`identitylink/repository.go:56-92`・`medicalrecord/vaccination_repository.go:101-119`・`owner/ltv_repository_query.go:136`・`lstep/lstep_tag_cache_repository.go:181-182` の ILIKE/LIKE 述語はすべて `textsearch.EscapeLike` (`internal/textsearch/textsearch.go:102-106` — `\`→`\\`, `%`→`\%`, `_`→`\_`) + `ESCAPE '\'` + bind 値の三点組。LIKE パターン自体にユーザー入力を連結する site は 0 件 |
| 6 | 防御判定 | **PASS** (static の範囲) — 全 raw/dynamic SQL site が bound または code-constant 由来。unsafe (ユーザー入力を SQL 文へ構文連結) site は 0 件、したがって FAIL 対象・remediation candidate なし |
| 7 | 副作用呼出回数 | **0** — クエリを一度も実行していない (static review のみ) |
| 8 | 検出判定 | **N/A** — 実行しない検証に対する検出経路は適用外。SQL ログ監視 (実 DB への marker 到達観測) は本安全条件下では行わない |
| 9 | 安全な証跡参照 | 本ファイル末尾の「全 site 棚卸し表」 (file:line 一覧、全 site grep 行に紐付く)。代表読取箇所: `internal/lstep/checkup_sync_repository.go:90-114` (Sprintf は `checkupSyncPreviewSQL` const + helper 生成の定数 fragment), `internal/owner/ltv_repository_query.go:380-405` (ORDER BY allowlist), `internal/csvimport/cutover_import_copy.go:105,210-211` (`pgx.Identifier{}.Sanitize()`), `internal/testdb/testdb.go:655-657` (`quotePostgresIdentifier` = `"` を `""` へ escape) |
| 10 | 制約・未検証範囲 | 未検証: (a) GORM の `Where("col = ?", v)` 型 fluent API (数千 site) は bind 前提の API contract であり個別には列挙していない — 棚卸し対象は `Raw`/`Exec`/Sprintf-into-query/LIKE-連結 のリスククラスのみ。(b) `internal/testdb`/`internal/clinicale2e`/`internal/csvimport`/`internal/seedlogin` は test/migration-support/seed 経路であり production request path ではない (分類欄に明記)。(c) DB 実測による plan-level の確認 (marker が実クエリログへ現れないこと) は DB アクセス禁止のため未実施 — static での根拠に留まる。(d) `database/sql` の `ExecContext`/`QueryRowContext` は本棚卸しの grep 範囲外 — `internal/seedlogin` (seed/CLI 経路) と `internal/testdb/truncate.go` (test-support 経路) に存在し補助確認で個別判定済み (下表; truncate.go は `:89` QueryRowContext + `:94,:97,:100,:103,:114,:137` ExecContext — `:114` の `truncateSQL` は const `coreTruncateSQL` (:23) または `isSafeTruncateTableName` (:38-49) + `quotePostgresIdentifier` 済み識別子のみで構成、組立て :65-70)。(e) pgx の `.QueryRow(`/`.Query(` 系は宣言 grep scope (`\.Raw\(|\.Exec\(` + 行先頭 continuation) の外 — 同一 pgx 経路の `.Exec(` は棚卸し対象なのに Query 系は対象外という非対称な境界をここに開示する。27 non-test site は全て `internal/csvimport` に集中 (cutover_import_copy.go ×14, cutover_import_validate.go ×5, cutover_payment_target.go ×4, import.go:88, failure_rehearsal.go:384, cutover_window_zero_target.go:45, cutover_import_resume.go:32 — :32 は `cutoverQuerier` 実装の pgx passthrough)。個別確認で全て bound (const SQL + `$N` bind) / catalog-derived (`pg_get_serial_sequence` 結果) / `pgx.Identifier{}.Sanitize()` 引用のいずれか — unsafe 0 件 |
| 11 | 修復案 | 該当なし — unsafe site 0 件。新規コードへの guardrail 候補 (スコープ外・候補として記録): `Raw(fmt.Sprintf(...))` を lint で禁止する rule、または `textsearch.EscapeLike` の使用強制 |
| 12 | 再検証結果 | 実施済み (att-u03-2) — 棚卸し grep を行先頭 chain-continuation を捉える 2 コマンドに訂正し全 site を再列挙・verb 属性を file:line 読取で確認 (下記「証跡コマンド」)。unsafe site は引き続き 0 件 |

## 全 site 棚卸し表 (119 sites — 全数 file:line)

凡例 — **分類**:
- `bound` : SQL 文は constant; 全可変値は `?`/`$N` bind 引数
- `bound+constfrag` : SQL 文にコード内 const/静的 fragment を Sprintf/連結で埋込むが、可変値は全て bind。fragment にユーザー入力は混入しない
- `ident-quoted` : 識別子を `pgx.Identifier{}.Sanitize()` または `quotePostgresIdentifier` (二重引用符の `""` escape) で引用
- `catalog-derived` : 識別子を DB catalog (information_schema/pg_constraint) から取得 — test/fixture 経路のみ
- `static-list` : テーブル/列名がコード内の静的リテラル列挙からのみ来る
- `DDL-const` : スキーマ定数から組み立てた DDL (test/infra 経路のみ)

### Raw (58 sites)

| file:line | 分類 | 備考 |
|-----------|------|------|
| internal/auth/permission_group_repository.go:376 | bound | 権限集約 (`bool_or`) クエリ — const SQL + `?` bind。**[att-u03-2 追加: chain-continuation site]** |
| internal/billing/accounting_repository_ltv.go:147 | bound | `ownerAnnualRevenueTopPercentSQL` const + args bind。**[att-u03-2 追加]** |
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
| internal/lstep/lstep_tag_cache_repository.go:150,161,187,197,215 | bound+constfrag | baseSQL 連結; `LIKE ? ESCAPE '\'` + `EscapeLike` 済み値 (:181-182)。**[att-u03-2 訂正: :150,:161,:187 追加; :154 は :150 クエリ内 ORDER BY テキストのため call site ではなく削除]** |
| internal/medicalrecord/checkup_repository.go:153 | bound | 健診一覧集計 (const) |
| internal/medicalrecord/medical_record_image_deletion_store.go:58,97 | bound | 画像削除 SELECT (const)。**[att-u03-2 訂正: :37,:119 は `.Exec` — Exec 表へ移動]** |
| internal/medicalrecord/medical_record_image_upload_quota.go:242,255 | bound | quota 確認 (const) |
| internal/medicalrecord/medical_record_repository.go:100 | bound | `SELECT pg_try_advisory_xact_lock(hashtextextended(?, 0))` + lockKey bind。**[att-u03-2 追加]** |
| internal/medicalrecord/vaccination_repository.go:180 | bound | ワクチン履歴 ROW_NUMBER (const + `?`) |
| internal/owner/ltv_repository.go:119 | bound+constfrag | `Sprintf(ownerLTVSelectSQL, ...fragments)` — fragment は全てコード生成 const; ORDER BY は allowlist (`buildOrderBy` ltv_repository_query.go:380-405) |
| internal/pet/chronic_condition_repository.go:107 | bound | 慢性疾患 (const) |
| internal/reservation/reservation_intent_repository.go:315 | bound | candidate-status UPDATE..RETURNING (const query + `?` args)。**[att-u03-2 追加]** |
| internal/reservation/reservation_repository_slots.go:103,126,147,175 | bound | 予約枠ロック/競合 (const + `?`)。**[att-u03-2 訂正: :87 は `.Exec` — Exec 表へ移動し重複解消]** |
| internal/reservation/reservation_type_occupation_repository.go:115,141 | bound | 予約種別×職種 (const) |
| internal/reservation/synthetic_closing_appointment.go:31 | bound | synthetic fixture (test-support) |
| internal/staff/staff_provisioning_repository.go:224 | bound | 権限判定 (const + `?`) |
| internal/staff/synthetic_closing_staff.go:52 | bound | `SELECT to_regclass('shift_entries')` — 完全 const |
| internal/testdb/testdb.go:424,639 | bound | test DB 基盤; :424 enum label 照会 + `?`, :639 `pg_database` EXISTS + `?`。**[att-u03-2 訂正: :435 は `.Exec` — Exec 表へ移動]** |

### Exec (61 sites)

| file:line | 分類 | 備考 |
|-----------|------|------|
| internal/auth/permission_group_policy_lock.go:23 | bound | advisory lock (const + `?`) |
| internal/billing/cash_register_close_repository.go:71 | bound | `pg_advisory_xact_lock(hashtextextended(?,0))` |
| internal/billing/estimate_repository.go:282 | bound | 同上 (採番ロック) |
| internal/billing/synthetic_closing_audit.go:70,79,85 | bound | sentinel audit UPDATE (const + `?`) |
| internal/billing/synthetic_closing_fixture.go:483,540 | static-list | `Sprintf("ALTER TABLE %s ...", table)` — `table` は静的リスト `syntheticClosingAppendOnlyTables` (:287-297) のみ |
| internal/billing/synthetic_closing_fixture.go:492,505 | bound | synthetic teardown UPDATE/DELETE |
| internal/clinic/closing_special_period_repository.go:76,131 | bound | advisory lock |
| internal/clinicale2e/fixture.go:597,606,611 | bound | SAVEPOINT/RELEASE/ROLLBACK TO SAVEPOINT |
| internal/clinicale2e/fixture.go:604 | catalog-derived | `DELETE FROM "`+target.table+`" WHERE `+target.where — target は上記 catalog 由来識別子 + `"` 囲み; clinicID は `?` bind |
| internal/csvimport/cutover_import_copy.go:105,116 | ident-quoted | `LOCK TABLE `+`pgx.Identifier{}.Sanitize()` |
| internal/csvimport/cutover_import_copy.go:243,319,323,342,371,408,678 | ident-quoted / bound | staging テーブル生成・COPY 後処理 — 識別子全て `pgx.Identifier{}.Sanitize()`; 値は `$N` bind |
| internal/csvimport/cutover_import_resume.go:36,131,135,137,202,205,248,251 | bound | advisory lock/set_config (const + `$N`); :137,:205,:251 の set_config('app.bypass_rls',...) は固定リテラル |
| internal/csvimport/cutover_import_validate.go:61,64,67 | bound | 同上 validate 経路 |
| internal/csvimport/failure_rehearsal.go:255,258,275 | bound | rehearsal 制御 + 意図的 FK 違反行 INSERT (`injectionErr` = FK 違反の発生確認であって攻撃ペイロードではない; 値は `$N` bind) |
| internal/csvimport/import.go:108 | static-list | `DELETE FROM "`+table+`"` — `table` は :100-106 の静的リテラル列挙のみ |
| internal/lstep/lstep_delivery_trigger_log_repository.go:102 | bound | advisory lock |
| internal/medicalrecord/medical_record_image_deletion_store.go:37,119 | bound | :37 INSERT / :119 UPDATE (const + `?`)。**[att-u03-2 訂正: Raw 表から移動 — 実 verb は `.Exec`]** |
| internal/medicalrecord/medical_record_image_upload_quota.go:232 | bound | advisory lock |
| internal/reservation/reservation_intent_repository.go:239,703 | bound | `SELECT pg_advisory_xact_lock(hashtextextended(?, 0))` + lockKey bind (:239 lifecycle / :703 next-visit-intent)。**[att-u03-2 追加]** |
| internal/reservation/reservation_repository_slots.go:87 | bound | booking advisory lock (const + `?`)。**[att-u03-2 訂正: Raw 表との重複を解消 — 実 verb は `.Exec`]** |
| internal/reservation/synthetic_closing_appointment.go:39 | bound | synthetic DELETE |
| internal/staff/staff_provisioning_repository.go:48 | bound | `SELECT pg_advisory_xact_lock(hashtextextended(?, 0))` + lockKey bind (staff-provision batch)。**[att-u03-2 追加]** |
| internal/staff/synthetic_closing_staff.go:58 | bound | `DELETE FROM shift_entries WHERE clinic_id = ?` |
| internal/testdb/clinic_grant_fixture.go:188 | bound+static-list | `Sprintf` — `table` は :179-183 の switch で `companies`/`clinics`/`staffs` のみ許可 (それ以外は `t.Fatalf`) |
| internal/testdb/testdb.go:169,219,235 | DDL-const | CREATE TABLE IF NOT EXISTS 等の静的 DDL (test DB schema 構築) |
| internal/testdb/testdb.go:435,447,454,457 | DDL-const | enum create/reconcile — `name`/`v`/`create` は `SharedTestSchemaEnumTypes` (001_init.sql 由来のコード定数) に限定。**[att-u03-2 訂正: :435 を Raw 表から移動 — `db.Exec(create)`]** |
| internal/testdb/testdb.go:631,635 | bound | advisory lock/unlock (`hashtext(?)`) |
| internal/testdb/testdb.go:646 | ident-quoted | `CREATE DATABASE "+quotePostgresIdentifier(name)` — `"`→`""` escape (:655-657) |

### 棚卸しから除外した一致 (call site ではない)

| file:line | 種別 | 理由 |
|-----------|------|------|
| internal/csvimport/cutover_import_resume.go:23 | interface 宣言 | `Exec(context.Context, string, ...any) (pgconn.CommandTag, error)` — pgx 互換 interface の method signature (呼出しではない) |
| internal/csvimport/cutover_import.go:93 | interface 宣言 | 同上 |
| internal/csvimport/import.go:22 | interface 宣言 | 同上 |
| internal/lstep/lstep_tag_cache_repository.go:154 | SQL テキスト内文字列 | :150 に始まる複数行 `Raw(...)` call のリテラル尾部行 (`ORDER BY owner_count DESC` + 閉じ側 `, clinicID).`) — 行自体に `Raw(`/`Exec(` トークンはなく、宣言した両 grep pattern (行内 `\.Raw\(|\.Exec\(` / 行先頭 `(Raw|Exec)\(`) のいずれにも一致しない。call site は :150 のみ (att-u03-1 は :154 を独立 site として誤計上) |

### 補助確認 (Raw/Exec 以外の構文連結経路)

| 経路 | file:line | 判定 |
|------|-----------|------|
| `.Where(fmt.Sprintf("%s = ?", parentFKColumn), ...)` | internal/persistence/junction.go:28 | bound+constfrag — doc コメント「compile-time literal」契約; 全 caller はリテラル (`shift_template_repository.go:148` `"shift_template_id"`, `shift_entry_repository.go:165` `"shift_entry_id"`) |
| `.Where("NOT "+fragment, ...)` / `.Where(fragment, ...)` | internal/pet/checkup_history.go:75,87,89 | bound+constfrag — `petCheckupHistoryCorrelation` は const (:56-65); :87,:89 の `windowed` fragment は const + `CAST(? AS date)` 片 (:83-85) で値は bind |
| `.Where(fragment, m, m)` | internal/billing/accounting_repository.go:29-68 | bound+constfrag — `hasPayment`/`hasSplit`/`methodMatch` はローカル const fragment; `method` 値は bind |
| `.Select(fmt.Sprintf(...))` | internal/billing/accounting_repository_unpaid.go:173,184,209,249,305 | bound+constfrag — 埋込みは `unpaidAmountSQL` 等の package const |
| `ORDER BY` 動的 | internal/owner/ltv_repository_query.go:380-405 | allowlist — sort key を switch で 8 値に限定、order は `asc`/`desc` 以外を `desc` へ矯正 |
| `ORDER BY` 動的 | internal/medicalrecord/medical_record_repository.go:224-234 | allowlist — `medicalRecordSortColumns` map 照合 + order `asc` 以外 `DESC` |
| LIKE パターン連結 | internal/trimming/trimming_repository.go:101 | bound — pattern は `"%"+trimmingImageStorageKeyPrefix+"%"` (const prefix)、ユーザー入力なし |
| `estimate_no LIKE ?` | internal/billing/estimate_repository.go:294 | bound — パターン `"EST-%"` リテラル |
| `tx`/`conn` の `ExecContext`/`QueryRowContext` (database/sql 直接) | internal/seedlogin/apply.go:103,109,122,135,166,179,192,217,229,236,251,259,267,275,290, internal/seedlogin/operator.go:94,115,120,134, internal/testdb/truncate.go:89,94,97,100,103,114,137 | bound / bound+constfrag — seed/CLI + test-support 経路 (production request path ではない)。apply.go:179 の `Sprintf` は生成済み `$N` placeholder 列のみを埋込み値は全て `args` bind、apply.go:290 は const statements 列挙 (:285-288)。truncate.go:114 の `truncateSQL` は const または `isSafeTruncateTableName`+`quotePostgresIdentifier` 済み識別子のみ (組立て :65-70) |

## 証跡コマンド (再現用)

```sh
# backend/internal 配下の全 raw query API 呼出しを列挙 (本棚卸しの入力 — 2 コマンド併用が必須)
# (1) receiver.method 形式 (`.Raw(` / `.Exec(`) — 行内どこでも検出
rg -n '\.Raw\(|\.Exec\(' internal/ -g '*.go' -g '!*_test.go'
# (2) method-chain の行先頭 continuation (`DBOrTx(ctx, db).` の次行の `Raw(`/`Exec(`) — (1) では捕捉不能
rg -n '^\s+(Raw|Exec)\(' internal/ -g '*.go' -g '!*_test.go'
# → (1) 107 件 + (2) 15 件 = 122 file:line 一致。うち interface 宣言 3 件 (上表「除外した一致」) を除き
#   Raw 58 + Exec 61 = 119 call sites。att-u03-1 が使った単一 grep は (2) の検出を欠き
#   Raw 9 site / Exec 3 site を漏らした上、verb 誤属性と非-call-site 誤計上を含んでいた。
# Sprintf が raw query に入る site のみ抽出
rg -n '\.Raw\(fmt\.Sprintf|\.Exec\(fmt\.Sprintf' internal/ -g '*.go' -g '!*_test.go'
# → Raw+Sprintf は internal/billing/accounting_repository_unpaid.go:259 の 1 件 (package const のみ)
#   (上記 grep の Exec+Sprintf literal 一致は synthetic_closing_fixture.go:483,540 /
#    testdb.go:447,454 — いずれも静的リスト/コード定数、上表の分類参照。
#    clinic_grant_fixture.go:188 は `db.Exec(sql)` で sql を :184-187 の `Sprintf` で生成する
#    間接形のため literal pattern には一致しないが同一クラスとして棚卸し済み — `table` は
#    :179-183 の allowlist で `companies`/`clinics`/`staffs` のみ)
# LIKE 経路
rg -n 'LIKE' internal/ -g '*.go' -g '!*_test.go'
# → ユーザー入力を含む全 LIKE site で EscapeLike + ESCAPE '\' を確認 (上表)
```

## 訂正履歴 (correction lineage — att-u03-2)

att-u03-1 の記録証跡に対する G2 IMPL_DEFECT 指摘への訂正。**判定 (verdict) 自体は独立検証で正しいことが確認済みであり変更なし** (C06 PASS / static scope 維持 — 追加列挙した site も全て bound)。

| 箇所 | 誤り (att-u03-1) | 訂正後 (att-u03-2) |
|------|------------------|--------------------|
| site 合計 | 「107 sites (Raw 49 + Exec 58)」 — 表自体もラベルと不整合 (Raw 56 行) | **119 sites (Raw 58 + Exec 61)** — 2 コマンド grep の 122 一致から非-call-site 3 を減算 |
| 再現 grep | `grep -rn '\.Raw(\|\.Exec(' internal/ --include='*.go' \| grep -v '_test.go'` 単独 — 行先頭 chain-continuation (`DBOrTx(...).` 次行の `Raw(`/`Exec(`) を捕捉不能 | 2 コマンド併用に訂正 (上記「証跡コマンド」) — (2) が line-leading `Raw(`/`Exec(` continuation を捕捉 |
| Raw 表 | `medical_record_image_deletion_store.go:37,58,97,119` と記載 | :37,:119 は `.Exec` → Exec 表へ。Raw は :58,:97 |
| Raw 表 | `reservation_repository_slots.go:87,103,126,147,175` — :87 を Raw に誤計上 | :87 は `.Exec` → Exec 表へ確定。Raw は :103,:126,:147,:175 |
| Raw 表 | `testdb.go:424,435,639` | :435 は `db.Exec(create)` → Exec 表へ。Raw は :424,:639 |
| Raw 表 | `lstep_tag_cache_repository.go:154,197,215` | :154 は SQL テキスト内行 (非-call-site)。実 site は :150,:161,:187,:197,:215 — :150,:161,:187 を追加 |
| Raw 表 | 欠落: line-leading chain-continuation site | 追加: `permission_group_repository.go:376`, `accounting_repository_ltv.go:147`, `medical_record_repository.go:100`, `reservation_intent_repository.go:315` |
| Exec 表 | 欠落: line-leading chain-continuation site | 追加: `reservation_intent_repository.go:239,:703`, `staff_provisioning_repository.go:48` |
| Exec 表 | `medical_record_image_deletion_store.go:119` のみ (:37 欠落) | :37,:119 の 2 site に訂正 |
| field 10/補助表 | `database/sql` `ExecContext`/`QueryRowContext` 経路が未言及 | seedlogin (seed/CLI 経路) として補助確認へ追記 + field 10 (d) にスコープ注記 |
| hygiene pass (att-u03-2 G2 residual F-1..F-7) | field 10(d) が database/sql Context 系を seedlogin のみと記載 (testdb/truncate.go 7 site 未記載)・operator.go は :94 のみ列挙・pgx QueryRow/Query の scope 非対称が未開示; `checkup_history.go:64-70` / `estimate_repository.go:288` の誤 cite; 証跡コマンド comment が `clinic_grant_fixture.go:188` を `Exec+Sprintf` literal 一致として帰属 (実際は `db.Exec(sql)` の間接形); lstep_tag_cache :154 除外理由の誤検出メカ説明が不正確 | field 10(d) に truncate.go :89,94,97,100,103,114,137 を追記・新設 (e) で pgx QueryRow/Query (csvimport 27 site) の scope 非対称を開示; operator.go :115,:120,:134 と truncate.go conn.* を補助表へ追加; `checkup_history.go` → :75,:87,:89、`estimate_repository.go` → :294; 証跡コマンド comment を literal 一致 (synthetic_closing_fixture.go:483,540 / testdb.go:447,454) と間接形 (clinic_grant_fixture.go:184-188) に分離; :154 除外理由を「両 pattern に一致しない」へ精密化。verdict (PASS)・119 site 数は不変 |
| hygiene pass (att-u03-2 G2 residual — residual cite fix) | Exec 表 `clinic_grant_fixture.go:188` 備考が allowlist switch を `:180-185` と誤 cite (実際は switch :179 / case :180 / default+`t.Fatalf` :181-182 / close :183; :184 以降は `Sprintf` 呼出) | `:179-183 の switch` へ訂正 — 証跡コマンド comment (:142) の `:179-183` と整合。verdict (PASS)・119 site 数は不変 |

*Marker 文字列は `SECCHK20261009C06Rnn` 形の英数字 sentinel のみ。実行可能な攻撃構文・実ペイロード・資格情報・PII は本ファイルに含まない (A5)。*
