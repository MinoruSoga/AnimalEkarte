# U07 — Residual Known Issues (残件・既知問題記録)

Campaign: emr-332-security-defense / Unit: U07-REVERIFY-CONSOLIDATE / Attempt: att-u07-1
Recorded by: receiver-emr332-u07-01 / Recorded at: 2026-10-09
Ticket revision (consolidated): EMR-332 @ 2026-10-08T17:54:31Z
本ファイルは U01–U06 の verdict docs に記録された残件・既知問題の統合一覧 (record-only — ここでの記載は問題の解消・修復の実施を意味しない)。発見元の一次記録は各 case doc。

## 1. 未修復 FAIL (未完了として残す — AC5)

| ケース | ギャップ | 一次記録 |
|--------|----------|----------|
| C05 | unknown/disallowed field が全 226 bind 経路で黙殺受理; trailing 第 2 JSON value 未読黙殺。`strict_bind.go` は unwired scaffold で配線なし | `u03/c05-input-validation.md` |
| C09 (scoped) | env 駆動 egress プロバイダ (Discord/Plane/LLM/OPENAI_BASE_URL/SMTP) に host allowlist・dial 層 private-IP guard・redirect 抑止なし | `u05/c09-ssrf-egress.md` |
| C10 (scoped) | shared-file upload family が content sniffing を欠く (ext allowlist のみ) | `u05/c10-upload-validation.md` |

## 2. BLOCKED / UNKNOWN / NOT ASSESSED の残件

| 残件 | 状態 | 依存する人間操作 |
|------|------|------------------|
| staging-observed legs 全件 | BLOCKED — live 送信 0/20 requests、staging 稼働リビジョン undetermined | staging 稼働リビジョン照合手段 + 専用テストアカウント + 許可テナント識別子 |
| real-DB legs (`TestRealDB_*`, CAS/advisory-lock 実 SQL, audit repository 層, `TestAutoCreateFromReservation_Concurrent…`) | 未実行 (Compose `db` Postgres が本安全条件で使用不可) | test-DB 環境の供給 |
| C07 render leg (React render→DOM 実測) | BLOCKED (node_modules 不在・共有 compose 禁止) | isolated frontend container の供給 |
| C16 multi-instance 集約制限 | UNKNOWN (`RateLimitStore` は process-local map — `rate_limit.go:29-33`) | deployment topology の確定 |
| C18 通知経路到達性 | UNKNOWN (ReservationNotifier の配送確認証跡なし) | 通知配送証跡の提示 |
| C14 脆弱性状態 | NOT ASSESSED (`go vuln`/`pnpm audit`/Dependabot/OSV 未実行) | scanner 実行の許可・実行環境 |

## 3. 記録上の既知問題 (doc-level)

| 項目 | 内容 | 扱い |
|------|------|------|
| `u01/source-inventory.md:46` stale cite | `repository_db_test.go:33-36` と記載 — `u01-test-classification.md` §1 の `:31-34` と行番号が不整合 | **record-only — 当該ファイルは U01 owned のため U07 では編集しない** (後続の doc 保守で扱う) |
| U03 訂正履歴 | att-u03-1 の記録証跡不備 (合成サマリ行・site 数の混同・誤 cite) は att-u03-2 で訂正済み — verdict 不変 | `u03/c05-input-validation.md`, `u03/c06-sql-inventory.md` 末尾の訂正履歴表に記録済み |
| `attempts/att-u02-1.report.md` miscount | evidence-t06 に「568 checks」と記録 — 実際の再計測は **567** test-case entries (`=== RUN`/`--- PASS:` 各 567 行; PASS/0-FAIL verdict は正確) | **record-only — receiver report は履歴記録のため編集しない** (U07 docs 側は訂正済み — verdict-matrix.md・u07/reverification.md 訂正履歴) |
| `u06/c15-order-idempotency.md` field 9 miscount | 「TestC15_* 6 tests / 40 subcases」「8 tests / 20 subcases」と記録 — 実際の `func TestC15_*` 宣言数は **7** (`order_idempotency_test.go`) と **9** (`auto_create_idempotency_test.go`); subcase counts は独立再検証なし | **record-only — u06 owned file のため U07 では編集しない** (U07 docs 側は訂正済み — verdict-matrix.md・u07/reverification.md 訂正履歴) |

## 4. Remediation candidates (全て未適用・候補のみ)

`remediation-plan.md` Section 2 が正本。一覧: C05 wiring options (GAP-1 内), C09 R1–R4, C10 RC-1, C14 RC-1..4, U04 edge-preflight/headers observations (3 件), C06 lint guardrail, C07 client-side scheme allowlist, C18 `auth_staff_validation_fail_open` mislabel rename, C16 集約 store 検討。

## 5. 観察のみ記録 (verdict に影響しない設計注記)

- `auth_staff_validation_fail_open` (`composition_auth_routes.go:82-83`): 実挙動は fail-closed (503 拒否) で名前が逆 — 誤ラベル観察 (u06 c18 field 10)。
- CORS: ACAC/Allow-Headers/Allow-Methods/Max-Age は不許可 origin にも無条件で返る設計 (grant を伴わない metadata のみ — `u04/c12-cors.md` field 10)。
- C08: `forgot-password`/`reset-password` は CSRF 未適用だが匿名経路で cookie 認証非依存のため整合的 (token-mode 再設計は範囲外 — `u04/c08-csrf.md` field 11)。
- C14: tag-pin first-party actions・permissions 未宣言 workflow・deploy→migrate ≤150s window は残余リスク/known-issue として記録済み (`u05/c14-supply-chain.md` Section B/D、Section A 残余行)。
- C16: `FailureRateLimit` の over-admit (空 bucket 同時到着で最大 burst 分) は doc-comment どおりの許容設計 (`u06/c16-rate-limit.md` field 10)。
