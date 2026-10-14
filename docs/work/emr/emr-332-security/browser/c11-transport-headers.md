# U04 — Case Verdict Record: C11 (TLS / HSTS / CSP / cookie attributes / frame controls)

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](../test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U04-BROWSER-EDGE / Attempt: att-u04-1
Template: ../baseline/case-verdict-template.md (12 fields)

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | `C11` — トランスポート・セキュリティヘッダ (TLS/HSTS/CSP/cookie 属性/frame control) |
| 2 | 対象リビジョン | 静的・模擬対象: base `de28e9765a29091339a872a85f1875a947e07e5b` + w04 追加テスト。**staging 稼働リビジョン: undetermined** (U01 §2 照合試行の詳細は c08-csrf.md 補足 — repo 内に deploy 記録・Worker version stamp・既知差分証跡なし)。staging-observed は BLOCKED。 |
| 3 | 実施方式 (method) | `static` (middleware / Worker / `_headers` / wrangler / cookie 発行コードのレビュー) + `simulated` (実 `SecurityHeaders` middleware 経由 httptest)。staging-observed 行は BLOCKED。cookie 属性は login 禁止のため恒久的に static-only。 |
| 4 | 正常条件 | 期待: HTTPS 前提で API 応答に XCTO nosniff / XFO DENY / Referrer-Policy / CSP `default-src 'none'` / Permissions-Policy / production 時 HSTS が付く。観測 (simulated): `SecurityHeaders(true)` 経由の 200 応答に上記すべて + `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload` を確認; CSRF 403 拒否応答にも同一ヘッダ群が残る (global chain で先に set されるため)。 |
| 5 | 拒否条件 | 本ケースの拒否形状は「ヘッダ欠落時に応答が送出されること」— 観測: 欠落なし (non-production で HSTS のみ意図的省略: security_headers.go:19-22 が isProduction ゲート。STG/本番は `GIN_MODE=release` のため付与対象 — backend/wrangler.jsonc vars `GIN_MODE: "release"`)。 |
| 6 | 防御判定 | **PASS** (static + simulated 範囲)。staging-observed 判定は BLOCKED (稼働リビジョン未確定・送信 0 回)。 |
| 7 | 副作用呼出回数 | 0 — ヘッダ観測のみで保存・送信系呼出は発生しない (simulated 内でも downstream 呼出は制御用に計測済み: GET=1, CSRF 403=0)。 |
| 8 | 検出判定 | **N/A** — ヘッダ付与の観測・検出機構は存在しない設計 (ヘッダは能動防御でありイベント検出対象ではない)。marker-capture 証跡なしのため PASS には丸めない。 |
| 9 | 安全な証跡参照 | `backend/internal/middleware/security_headers.go:9-23`; `backend/cmd/api/main.go:203-208` (global Use); `backend/tests/security/browser/browser_edge_test.go` `TestEMR332C11C13_SecurityHeaderChain`; `docs/work/emr/emr-332-security/browser/evidence/middleware-tests.txt`; `frontend/public/_headers` (read-only); `backend/worker/index.ts:404-427,540-542` (read-only); `backend/internal/auth/http_types.go:87-93`, `http_session_cookies.go:13-23,97-115`; `backend/cmd/api/composition_runtime.go:455`; `backend/wrangler.jsonc` vars `GIN_MODE`; `frontend/wrangler.jsonc`。 |
| 10 | 制約・未検証範囲 | **TLS**: Cloudflare edge で終端 (workers.dev HTTPS; SECURITY.md:91 の 2026-10-03 疎通記録は commit revision の証明ではない)。TLS バージョン・cipher 等の握手実測は staging-observed BLOCKED のため未検証。**HSTS**: API は isProduction ゲート付き (release で付与)、frontend `_headers` /* に同一値を静的確認。**CSP**: API は `default-src 'none'` (JSON API で最小妥当); frontend は `default-src 'self'` + `frame-ancestors 'none'` + `object-src 'none'` + `base-uri 'self'`。**frame control**: XFO DENY (API+frontend) + frame-ancestors 'none' (frontend CSP)。**cookie 属性 (static-only)**: access+refresh とも `HttpOnly: true` 固定; production では `Secure: true, SameSite: None` (http_types.go:88-93, composition_runtime.go:455 `GinMode=="release"`); refresh cookie は `RefreshTokenCookiePath` でパス限定; ClearCookie (http_session_cookies.go:13-23) が同属性で `MaxAge:-1` 失効。**未検証**: cookie の live 観測 (login を伴うため不可)、staging 応答ヘッダの実測、production 環境全体 (未 live)。 |
| 11 | 修復案 | **適用なし (owned_paths 内 gap なし)**。候補 (read-only 層の観察所見・本ユニットでは非適用): edge 即応答 preflight (backend/worker/index.ts `buildCorsPreflightResponse` :404-427) は security headers を付けない — 204 no-body OPTIONS で renderable content を持たず実害リスク低だが、将来修正機会があれば `X-Content-Type-Options: nosniff` 程度の付与は候補になり得る (backend/worker/ は本ユニット read-only のため remediation candidate として記録のみ)。frontend Worker (frontend/worker/index.ts) もヘッダ非付与で `_headers` と API 透過に依存 — 現行設計どおり。 |
| 12 | 再検証結果 | 修復なしのため再検証不要。初回検証は scoped Docker で PASS (`TestEMR332C11C13_SecurityHeaderChain` 2 subtests; evidence-t01 verbatim)。 |
