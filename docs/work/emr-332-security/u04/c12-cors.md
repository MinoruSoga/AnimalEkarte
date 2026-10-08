# U04 — Case Verdict Record: C12 (CORS origin handling)

Campaign: emr-332-security-defense / Unit: U04-BROWSER-EDGE / Attempt: att-u04-1
Template: ../u01-case-verdict-template.md (12 fields)

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | `C12` — CORS origin 判定 (許可 origin への ACAO+Vary、不許可 origin の非許可) |
| 2 | 対象リビジョン | 静的・模擬対象: base `de28e9765a29091339a872a85f1875a947e07e5b` + w04 追加テスト。**staging 稼働リビジョン: undetermined** (照合試行記録は c08-csrf.md 補足)。staging-observed は BLOCKED。 |
| 3 | 実施方式 (method) | `simulated` (実 `CORS` middleware 経由 httptest、staging 設定値の allowlist shape で実行) + `static` (edge preflight 複製・設定値レビュー)。staging-observed 行は BLOCKED。 |
| 4 | 正常条件 | 期待: 設定 origin の preflight → 204 + `Access-Control-Allow-Origin: <origin>` + `Vary: Origin` + `Timing-Allow-Origin: <origin>` + `Access-Control-Allow-Credentials: true`。観測: staging 設定の 2 origin (`https://stg.noah-karte.com`, `https://api.stg.noah-karte.com`) それぞれで全項目一致 + `Access-Control-Allow-Headers` に `x-requested-with` を含み wildcard `*` なし + downstream 呼出 0。 |
| 5 | 拒否条件 | 期待: 不許可 origin の preflight → ACAO/Timing-Allow-Origin/Vary を付与しない (grant を拒否)。観測: `https://disallowed.example.com` で 3 ヘッダとも未設定・204・downstream 0。GET 実リクエストでは応答自体は返るが ACAO 系は未設定 (CORS 拒否はブラウザ読取りゲートでありサーバ拒否ではない — middleware 仕様どおり)。 |
| 6 | 防御判定 | **PASS** (simulated 範囲)。staging-observed 判定は BLOCKED — かつ live OPTIONS は **edge worker が応答する経路** (backend/worker/index.ts:540-542 → `buildCorsPreflightResponse`) であり Gin の cors.go には到達しないため、実測しても cors.go の provenance 証明にはならない点を明示する (edge/Gin の provenance 分離を保持)。 |
| 7 | 副作用呼出回数 | 0 — preflight は Gin middleware で `AbortWithStatus(204)` (cors.go:40-42) / edge では Container 非起動 (worker/index.ts:537-542)。spy counter 実測: preflight 全件で downstream 0。 |
| 8 | 検出判定 | **UNKNOWN** — CORS 拒否のログ/マーカー捕捉証跡なし (logging.go は U06 所有・read-only; marker-capture 証拠なしのため UNKNOWN)。 |
| 9 | 安全な証跡参照 | `backend/internal/middleware/cors.go:12-46` (exact-match allowlist, Vary: Origin, unconditional ACAC/Methods/Headers/Max-Age, OPTIONS abort); `backend/internal/middleware/emr332_u04/browser_edge_test.go` `TestEMR332C12_CORSOriginJudgment`; `docs/work/emr-332-security/u04/evidence-t01-middleware-tests.txt`; edge 複製 `backend/worker/index.ts:376-427` (`isCorsOriginAllowed` exact-match :388-395, allowed のみ ACAO/Timing-Allow-Origin/Vary :409-421, 無条件 ACAC/Headers/Methods/Max-Age :422-425, 常時 204 :426) + dispatch :537-542; 設定値 `backend/wrangler.jsonc` vars `CORS_ALLOWED_ORIGIN: "https://stg.noah-karte.com,https://api.stg.noah-karte.com"` (非 secret); frontend は `frontend/wrangler.jsonc` — CORS var なし (assets 配信 + /api/* service binding 中継のみ)。 |
| 10 | 制約・未検証範囲 | **edge-vs-Gin 分離 (本ケースの核心)**: staging での OPTIONS preflight は edge worker (`buildCorsPreflightResponse`) が応答し Container/Gin は起こさない。両者はヘッダ契約を共有する複製実装 (worker/index.ts:376-379 のコメントが同期義務を明記) で、Go 側と値完全一致を静的に確認 (`CORS_PREFLIGHT_ALLOW_METHODS`/`_ALLOW_HEADERS`/`_MAX_AGE`/fallback origins = cors.go の既定)。Gin 側 cors.go が外部から観測されるのは非 OPTIONS 実リクエストのみ。**挙動注記**: ACAC/Allow-Headers/Allow-Methods/Max-Age は不許可 origin にも無条件で返る設計 (grant を伴わない metadata のみ — cors.go:30-38, worker :422-425)。**未検証**: staging 実測 (BLOCKED)、preflight 応答に security headers が付かない件の live 確認、edge/Gin 複製の将来 drift (機械ゲートなし — 手動同期契約)。 |
| 11 | 修復案 | **適用なし (owned_paths 内 gap なし)**。候補として記録のみ: (a) edge/Gin preflight 複製の値 drift を防ぐ機械的 contract テスト (両者とも read-only/別ユニット領域に跨るため本ユニットでは非適用); (b) edge preflight への最小 security header 付与 (c11-transport-headers.md 修復案欄と同一観察所見 — backend/worker/ read-only)。 |
| 12 | 再検証結果 | 修復なしのため再検証不要。初回検証 PASS: `TestEMR332C12_CORSOriginJudgment` 4 subtests 全 PASS (evidence-t01 verbatim)。既存 `cors_test.go` (denied origin ACAO 空、default/custom origin、OPTIONS 204、`*` 非含有) も scoped run で全 PASS。 |
