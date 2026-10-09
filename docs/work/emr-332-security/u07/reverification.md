# U07 — Re-verification Records (修復後再検証 — AC5 / 工程6)

Campaign: emr-332-security-defense / Unit: U07-REVERIFY-CONSOLIDATE / Attempt: att-u07-1
Recorded by: receiver-emr332-u07-01 / Recorded at: 2026-10-09
Ticket revision (consolidated): EMR-332 @ 2026-10-08T17:54:31Z
本ユニットは docs-only consolidation — 新たなテスト実行・コード変更は実施していない。以下の記録は U01–U06 attempt が残した scoped 証跡の引用 (path + test name のみ; 出力本文の転記なし)。

## 1. 適用修復の棚卸し (applied product-code fixes)

**適用された product-code 修復: 0 件 (ZERO)。**

U01–U06 を通じて本番経路への修復コード適用は一切ない。キャンペーンが生産した唯一のコード artifact は検証 scaffold:

| Artifact | Path | 状態 |
|----------|------|------|
| Strict JSON bind helper (unwired scaffold) | `backend/internal/middleware/emr332_u03/strict_bind.go` (`BindStrictJSON`) | **UNWIRED** — handler/本番経路への配線なし (helper doc comment :48-50 に「Wiring it into handlers is intentionally NOT done」と明記) |
| Scoped verification tests | `backend/internal/middleware/emr332_u03/bind_boundary_test.go` (6 tests), `strict_bind_test.go` (7 tests); `backend/internal/middleware/emr332_u04/browser_edge_test.go`; `backend/internal/imageupload/emr332_u05/sniff_contract_test.go`; `backend/internal/middleware/emr332_u06/{rate_limit,logging_requestid}_test.go`; `backend/internal/reservation/emr332_u06/order_idempotency_test.go`; `backend/internal/medicalrecord/emr332_u06/auto_create_idempotency_test.go`; `backend/internal/audit/emr332_u06/audit_validate_test.go` | test-only — production code 差分なし |

よって AC5「修復したケースは…変更範囲で再確認」に該当する修復済みケースは存在しない。**未修復 FAIL は未完了のまま残す: C05, C09 (scoped), C10 (scoped)。**

## 2. C05 — 同ケース ID の再検証証跡 (unwired scaffold の検証証拠)

C05 の防御判定は **FAIL のまま** (unknown/disallowed field silent-accept on 226 bind paths は解消されていない)。ただし U03 (att-u03-2) が `BindStrictJSON` helper とその scoped unit tests を同一ケース ID `C05` の下で実施しており、これらが「helper が期待する拒否契約を満たす」ことの同ケース ID 再検証証拠となる (helper の検証であり handler 配線の検証ではない — 配線は remediation candidate 段階)。

| 再検証項目 | 結果 | 証跡 (path + test name のみ — 出力本文は各証跡 doc を参照) |
|------------|------|----------------------------------------------------------|
| 正常条件 (valid body 受理) | PASS (記録済み) | `TestBindStrictJSON_AcceptsValidBody`; `TestC05ShouldBindJSON_TagEnforcement/normal:_valid_body`; `TestC05RealHandlerStruct/normal:_valid_addendum` — 記録 doc: `docs/work/emr-332-security/u03/c05-input-validation.md` 判定レコード field 4/12 + 「実測コマンド出力」 |
| 拒否条件 (unknown field / trailing value / oversize / tag 違反) | PASS (記録済み) | `TestBindStrictJSON_{RejectsUnknownField,RejectsTrailingJSONValue,RejectsOversizedBody,RejectsOversizedDeclaredLength,EnforcesBindingTags,RejectsMalformedAndEmpty}`; gap 実測 `TestC05ShouldBindJSON_UnknownFieldSilentlyAccepted`, `TestC05ShouldBindJSON_TrailingValueUnread`, `TestC05RealHandlerStruct/gap:_unknown_field_silently_accepted`, `TestC05DisallowUnknownFields_Contrast` — 同 doc |
| 例外条件 | N/A (本ケースに例外条件の定義なし) | — |
| 副作用呼出 (拒否時ゼロ) | PASS (記録済み) | `TestC05ZeroSideEffectInvocations` (save-path sentinel 0 を機械確認) — 同 doc field 7 |
| 検出 | N/A — bind 層に unknown-field 受理の検出経路は構造上存在しない (silent drop が性質) | 同 doc field 8 |
| 回帰 | PASS (記録済み) | `go test ./internal/middleware/` (sanitize_null_bytes 既存テスト含む) + `go vet` exit 0 — 同 doc field 12 |
| scoped コマンド | `docker run --rm --pull never --network none -v <w03>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine go test -count=1 -v ./internal/middleware/emr332_u03/` → 13 テスト関数全 PASS (att-u03-2 実測) | 同 doc 「実測コマンド出力」 (verbatim は当該 doc に保持) |

**残存する未検証**: helper の handler への配線 (wiring) — 配線後に上記と同じケース ID `C05` で「226 経路のうち配線済み経路で unknown field が 400 拒否される」実測が必要 (`remediation-plan.md` GAP-1 再検証条件)。

## 3. 他ユニットの初回 scoped 検証 (修復なしケースの初回検証記録 — 再検証ではない)

以下は修復適用ゼロのため「初回検証」の記録であり、同ケース ID の修復後再検証ではない。全て attempt 時点で全 PASS が記録されている (verbatim 出力は各 case doc に保持):

| ケース | 初回検証証跡 (path + test name) |
|--------|--------------------------------|
| C01–C04, C17 (U02) | `docs/work/emr-332-security/u02/evidence-t01-middleware-auth.txt` (TestAuth*/TestLiffAuth*), `evidence-t02-token-service.txt` (TestTokenService_*), `evidence-t03-http-permission.txt` (TestHTTPHandler_*), `evidence-t04-current-access.txt` (TestCurrentAccessResolver_*), `evidence-t05-domain-membership.txt` (MembershipAB 系 6 packages), `evidence-t06-route-gates.txt` (TestGETHEAD*/TestRouteCompositionSmoke — **567 checks PASS / 0 FAIL**), `evidence-t07-realdb-unreachable.txt` (TestRealDB_* — DB 未到達の FAIL は実行制約の記録) |
| C06, C07 (U03) | static 棚卸しのみ — `u03/c06-sql-inventory.md` (119 sites 表), `u03/c07-xss-render.md` (sink 表); render leg BLOCKED |
| C08, C11–C13 (U04) | `u04/evidence-t01-middleware-tests.txt` — `TestEMR332C08_CSRFBlocksStateChangingRequestsWithoutHeader`, `TestEMR332C11C13_SecurityHeaderChain`, `TestEMR332C12_CORSOriginJudgment` (`backend/internal/middleware/emr332_u04/browser_edge_test.go`) |
| C09, C14 (U05) | static のみ — `u05/c09-ssrf-egress.md`, `u05/c14-supply-chain.md` |
| C10 (U05) | `backend/internal/imageupload/emr332_u05/sniff_contract_test.go` — 5 tests (PNG 受諾 / HTML-as-PNG 拒否 / rewind / oversize+ext+MIME 拒否 / `NewStoredName`) — `u05/c10-upload-validation.md` Section C |
| C15, C16, C18 (U06) | `backend/internal/reservation/emr332_u06/order_idempotency_test.go` (TestC15_* 7 tests), `backend/internal/medicalrecord/emr332_u06/auto_create_idempotency_test.go` (TestC15_* 9 tests), `backend/internal/middleware/emr332_u06/rate_limit_test.go` (TestC16_* 8 tests), `backend/internal/middleware/emr332_u06/logging_requestid_test.go` (TestC18_* 4 tests), `backend/internal/audit/emr332_u06/audit_validate_test.go` (TestC18_* 8 tests) — verbatim `ok` 行は `u06/` 各 case doc field 9 |

## 4. 本ユニット自身の検証 (docs-only gate)

本ユニットの検証は documentary self-check のみ (テスト・Docker・staging/DB 実行なし):

- verdict-matrix.md に C01–C18 全 18 行が存在し、defense/detection verdict が enum (`PASS|FAIL|SKIP|UNKNOWN|BLOCKED|N/A`) のみで構成されること — `grep -o 'C[0-9][0-9]'` / verdict 列の目視照合。
- remediation-plan.md の全エントリに AC4 フィールド (case id / code·config basis / minimal change or candidate / 担当 / priority / re-verification condition / evidence ref) が存在すること。
- index.md に 4 状態の assurance boundary・残件・未検証範囲・人間操作の各節が存在すること。
- 本 worktree での書込みが owned paths (`docs/work/emr-332-security/{index.md,verdict-matrix.md,remediation-plan.md,u07/**}`) に限定されること — `git status` で確認。

## 訂正履歴 (correction lineage — att-u07-1, 2026-10-09)

att-u07-1 の引用カウントに対する G2 IMPL_DEFECT 指摘への訂正。**検証証跡の verdict 記録自体は変更なし** — PASS / 0 FAIL の結論は正確であり、訂正はカウント記録値のみ。

| 箇所 | 誤り (訂正前) | 訂正後 |
|------|----------------|--------|
| §3 C01–C04, C17 行 — evidence-t06 count | `568 checks PASS / 0 FAIL` — `attempts/att-u02-1.report.md` の記録値を verbatim 転記 | **567 checks PASS / 0 FAIL** — `u02/evidence-t06-route-gates.txt` の独立再計測で `=== RUN` 567 行・`--- PASS:` 567 行・FAIL 0 を確認 (568 は upstream の miscount) |
| §3 C15, C16, C18 行 — TestC15_* 件数 | `order_idempotency_test.go` (TestC15_* 6 tests/40 subcases), `auto_create_idempotency_test.go` (TestC15_* 8 tests/20 subcases) — `u06/c15-order-idempotency.md` field 9 の記録値を verbatim 転記 | `order_idempotency_test.go` (TestC15_* **7 tests**), `auto_create_idempotency_test.go` (TestC15_* **9 tests**) — 両ファイルの `func TestC15_*` 宣言を独立再計測。`40 subcases`/`20 subcases` の qualifier は独立再検証されていないため除去 |

verdicts unchanged — PASS/0-FAIL verdict は正確であり全検証記録の結論は不変。upstream 記録そのもの (att-u02-1.report.md・u06 docs) の誤記録は `u07/residual-known-issues.md` §3 に record-only で記録 (receiver report・u06 owned file のため編集しない)。
